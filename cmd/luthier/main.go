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
	outFlag := flag.String("out", "", "output directory (default: ./<project-name>)")
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
	var (
		systemPrompt string
		cat          *catalogue.Catalogue
	)
	if *cataloguePath != "" {
		cat, err = catalogue.Load(*cataloguePath)
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
	if *outFlag != "" {
		outDir = *outFlag
	}
	fmt.Fprintf(os.Stderr, "luthier: writing scaffold to %s/\n", outDir)
	opts := &writer.Options{
		Composed:      composed,
		Catalogue:     cat,
		CataloguePath: *cataloguePath,
	}
	if err := writer.Write(spec, outDir, opts); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Fprintln(os.Stderr, "luthier: verifying scaffold builds...")
	if err := verifyBuild(outDir); err != nil {
		fmt.Fprintf(os.Stderr, "luthier: build failed, attempting self-heal...\n")
		if healErr := selfHeal(context.Background(), outDir, err.Error(), client, model); healErr != nil {
			return fmt.Errorf("build verification: %w (self-heal failed: %v)", err, healErr)
		}
		fmt.Fprintln(os.Stderr, "luthier: re-verifying after self-heal...")
		if err := verifyBuild(outDir); err != nil {
			return fmt.Errorf("build verification (after self-heal): %w", err)
		}
		fmt.Fprintln(os.Stderr, "luthier: self-heal succeeded")
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

// selfHeal reads Go source files from the scaffold, sends them with the
// compile error to the LLM, and writes back corrected files. One attempt only.
func selfHeal(ctx context.Context, dir, buildErr string, client talk.LLMClient, model string) error {
	// Collect all .go files in the scaffold.
	var files []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		return fmt.Errorf("no Go files found in %s", dir)
	}

	// Build a prompt with each file's contents and the error.
	var prompt strings.Builder
	prompt.WriteString("The following Go project failed to compile. Fix the error and return ALL corrected files.\n\n")
	prompt.WriteString("BUILD ERROR:\n```\n")
	prompt.WriteString(buildErr)
	prompt.WriteString("\n```\n\n")

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(dir, f)
		fmt.Fprintf(&prompt, "FILE: %s\n```go\n%s\n```\n\n", rel, src)
	}

	prompt.WriteString("Return each corrected file in this exact format:\n")
	prompt.WriteString("FILE: <path>\n```go\n<corrected source>\n```\n\n")
	prompt.WriteString("Only return files that need changes. Preserve the exact file paths.")

	think := false
	req := &talk.Request{
		Model: model,
		Messages: []talk.Message{
			{Role: "user", Content: prompt.String()},
		},
		Think:   &think,
		Options: map[string]any{"max_tokens": 4096, "temperature": float64(0)},
	}

	var resp strings.Builder
	if err := client.Chat(ctx, req, func(r talk.Response) error {
		resp.WriteString(r.Content)
		return nil
	}); err != nil {
		return fmt.Errorf("llm call: %w", err)
	}

	// Parse corrected files from the response.
	return applyFixes(dir, resp.String())
}

// applyFixes parses "FILE: path\n```go\n...\n```" blocks and writes them.
func applyFixes(dir, response string) error {
	applied := 0
	remaining := response
	for {
		idx := strings.Index(remaining, "FILE: ")
		if idx < 0 {
			break
		}
		remaining = remaining[idx+len("FILE: "):]

		newline := strings.Index(remaining, "\n")
		if newline < 0 {
			break
		}
		path := strings.TrimSpace(remaining[:newline])
		remaining = remaining[newline+1:]

		// Find the code block.
		codeStart := strings.Index(remaining, "```go\n")
		if codeStart < 0 {
			codeStart = strings.Index(remaining, "```\n")
			if codeStart < 0 {
				continue
			}
			codeStart += len("```\n")
		} else {
			codeStart += len("```go\n")
		}
		remaining = remaining[codeStart:]

		codeEnd := strings.Index(remaining, "\n```")
		if codeEnd < 0 {
			continue
		}
		source := remaining[:codeEnd]
		remaining = remaining[codeEnd:]

		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir for %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(source), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Fprintf(os.Stderr, "luthier: self-heal fixed %s\n", path)
		applied++
	}

	if applied == 0 {
		return fmt.Errorf("no fixes parsed from LLM response")
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
		return openai.NewClient(localURL, os.Getenv("LUTHIER_LOCAL_TOKEN"))

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
