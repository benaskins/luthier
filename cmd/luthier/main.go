package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/axon-talk/anthropic"
	"github.com/benaskins/axon-talk/openai"
	"github.com/benaskins/luthier/internal/analysis"
	"github.com/benaskins/luthier/internal/catalogue"
	"github.com/benaskins/luthier/internal/gaps"
	"github.com/benaskins/luthier/internal/snippets"
	"github.com/benaskins/luthier/internal/writer"
)

const defaultModel = "claude-sonnet-4-6"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "luthier:", err)
		os.Exit(1)
	}
}

func run() error {
	cataloguePath := flag.String("catalogue", "", "path to catalogue YAML (default: built-in axon catalogue)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		return fmt.Errorf("usage: luthier [-catalogue FILE] <prd.md>")
	}
	prdPath := args[0]

	prd, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	provider := os.Getenv("LUTHIER_PROVIDER")
	if provider == "" {
		provider = "anthropic"
	}
	client := newClient(provider)

	model := defaultModel
	if m := os.Getenv("LUTHIER_MODEL"); m != "" {
		model = m
	}

	// Determine system prompt: from catalogue file or built-in default
	var systemPrompt string
	if *cataloguePath != "" {
		cat, err := catalogue.Load(*cataloguePath)
		if err != nil {
			return fmt.Errorf("load catalogue: %w", err)
		}
		systemPrompt, err = cat.SystemPrompt()
		if err != nil {
			return fmt.Errorf("render catalogue: %w", err)
		}
		fmt.Fprintf(os.Stderr, "luthier: using %s catalogue (%s %s)\n", cat.Name, cat.Language, cat.Version)
	}

	fmt.Fprintln(os.Stderr, "luthier: analysing PRD...")
	var spec *analysis.ScaffoldSpec
	if systemPrompt != "" {
		spec, err = analysis.AnalyseWithPrompt(context.Background(), string(prd), client, model, systemPrompt)
	} else {
		spec, err = analysis.Analyse(context.Background(), string(prd), client, model)
	}
	if err != nil {
		return fmt.Errorf("analyse: %w", err)
	}

	if len(spec.Gaps) > 0 {
		resolver := gaps.New(model).WithIO(os.Stdin, os.Stderr)
		spec, err = resolver.Resolve(context.Background(), spec, client)
		if err != nil {
			return fmt.Errorf("resolve gaps: %w", err)
		}
	}

	// Compose glue code from snippets (only for catalogues with snippets)
	composed := composeFromSpec(spec)

	outDir := filepath.Join(".", spec.Name)
	fmt.Fprintf(os.Stderr, "luthier: writing scaffold to %s/\n", outDir)
	if err := writer.Write(spec, outDir, composed); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Fprintln(os.Stderr, "luthier: verifying scaffold builds...")
	if err := verifyBuild(outDir); err != nil {
		return fmt.Errorf("build verification: %w", err)
	}

	fmt.Println(outDir)
	return nil
}

func composeFromSpec(spec *analysis.ScaffoldSpec) *writer.ComposedOutput {
	reg := snippets.NewRegistry()
	for _, s := range snippets.GeneratedSnippets() {
		reg.Register(s)
	}

	// Collect selected modules that have wiring code
	var moduleNames []string
	for _, m := range spec.Modules {
		if s, ok := reg.Get(m.Name); ok {
			if s.Setup != "" || s.Helpers != "" {
				moduleNames = append(moduleNames, m.Name)
			}
		}
	}

	// Resolve transitive deps
	included := map[string]bool{}
	for _, n := range moduleNames {
		included[n] = true
	}
	queue := append([]string{}, moduleNames...)
	for i := 0; i < len(queue); i++ {
		s, ok := reg.Get(queue[i])
		if !ok {
			continue
		}
		for _, dep := range s.Deps {
			if !included[dep] {
				if ds, dok := reg.Get(dep); dok {
					if ds.Setup != "" || ds.Helpers != "" {
						moduleNames = append(moduleNames, dep)
						queue = append(queue, dep)
						included[dep] = true
					}
				}
			}
		}
	}

	if len(moduleNames) == 0 {
		return nil
	}

	selected, err := reg.ForModules(moduleNames)
	if err != nil {
		fmt.Fprintf(os.Stderr, "luthier: snippet warning: %v\n", err)
		return nil
	}

	mainSrc, err := snippets.Compose(spec.Name, selected)
	if err != nil {
		fmt.Fprintf(os.Stderr, "luthier: compose warning: %v\n", err)
		return nil
	}

	// Collect deduplicated requires
	seen := map[string]bool{}
	var requires []string
	for _, s := range selected {
		for _, r := range s.Requires {
			if !seen[r] {
				seen[r] = true
				requires = append(requires, r)
			}
		}
	}

	return &writer.ComposedOutput{
		MainGo:   mainSrc,
		Requires: requires,
	}
}

func verifyBuild(dir string) error {
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	tidy.Stderr = os.Stderr
	if out, err := tidy.Output(); err != nil {
		return fmt.Errorf("go mod tidy: %w\n%s", err, out)
	}

	build := exec.Command("go", "build", "./...")
	build.Dir = dir
	build.Stderr = os.Stderr
	if out, err := build.Output(); err != nil {
		return fmt.Errorf("go build: %w\n%s", err, out)
	}
	return nil
}

func newClient(provider string) talk.LLMClient {
	switch provider {
	case "local":
		localURL := os.Getenv("LUTHIER_LOCAL_URL")
		if localURL == "" {
			localURL = "https://models.hestia.internal"
		}
		return openai.NewClient(localURL, "")

	default: // anthropic
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		baseURL := "https://api.anthropic.com"

		accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
		gatewayToken := os.Getenv("CLOUDFLARE_AI_GATEWAY_TOKEN")
		gatewayName := os.Getenv("CLOUDFLARE_AI_GATEWAY_NAME")
		if gatewayName == "" {
			gatewayName = "axon"
		}

		var opts []anthropic.Option
		if accountID != "" && gatewayToken != "" {
			baseURL = fmt.Sprintf(
				"https://gateway.ai.cloudflare.com/v1/%s/%s/anthropic",
				strings.TrimSpace(accountID),
				gatewayName,
			)
			opts = append(opts, anthropic.WithGatewayToken(gatewayToken))
		}

		return anthropic.NewClient(baseURL, apiKey, opts...)
	}
}
