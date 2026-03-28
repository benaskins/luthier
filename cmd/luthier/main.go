package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/benaskins/axon-talk/anthropic"
	"github.com/benaskins/luthier/internal/analysis"
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
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: luthier <prd.md>")
	}
	prdPath := os.Args[1]

	prd, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	client := newClient()

	model := defaultModel
	if m := os.Getenv("LUTHIER_MODEL"); m != "" {
		model = m
	}

	fmt.Fprintln(os.Stderr, "luthier: analysing PRD…")
	spec, err := analysis.Analyse(context.Background(), string(prd), client, model)
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

	// Compose glue code from snippets
	composed := composeFromSpec(spec)

	outDir := filepath.Join(".", spec.Name)
	fmt.Fprintf(os.Stderr, "luthier: writing scaffold to %s/\n", outDir)
	if err := writer.Write(spec, outDir, composed); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Fprintln(os.Stderr, "luthier: verifying scaffold builds…")
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

	var moduleNames []string
	for _, m := range spec.Modules {
		if s, ok := reg.Get(m.Name); ok {
			// Only include snippets that have wiring code
			if s.Setup != "" || s.Helpers != "" {
				moduleNames = append(moduleNames, m.Name)
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

func newClient() *anthropic.Client {
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
