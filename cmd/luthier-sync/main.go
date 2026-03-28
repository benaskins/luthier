// luthier-sync reads luthier.yaml manifests from axon modules in the lamina
// workspace and generates:
// 1. The module catalog section of internal/patterns/system_prompt.txt
// 2. internal/snippets/generated.go with snippet definitions
//
// It uses `lamina repo --json` to discover modules.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name          string  `yaml:"name"`
	Purpose       string  `yaml:"purpose"`
	UseWhen       string  `yaml:"use_when"`
	Deterministic bool    `yaml:"deterministic"`
	Snippet       Snippet `yaml:"snippet"`
}

type Snippet struct {
	Imports  []Import `yaml:"imports"`
	Requires []string `yaml:"requires"`
	Deps     []string `yaml:"deps"`
	Setup    string   `yaml:"setup"`
	Helpers  string   `yaml:"helpers"`
}

type Import struct {
	Path  string `yaml:"path"`
	Alias string `yaml:"alias"`
}

type laminaRepo struct {
	Name string `json:"name"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "luthier-sync:", err)
		os.Exit(1)
	}
}

func run() error {
	laminaDir, err := findLaminaDir()
	if err != nil {
		return err
	}

	repos, err := discoverRepos(laminaDir)
	if err != nil {
		return err
	}

	manifests, err := loadManifests(laminaDir, repos)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "luthier-sync: found %d manifests\n", len(manifests))

	luthierRoot, err := findLuthierRoot()
	if err != nil {
		return err
	}

	if err := generateCatalog(manifests, luthierRoot); err != nil {
		return fmt.Errorf("generate catalog: %w", err)
	}

	if err := generateSnippets(manifests, luthierRoot); err != nil {
		return fmt.Errorf("generate snippets: %w", err)
	}

	fmt.Fprintln(os.Stderr, "luthier-sync: done")
	return nil
}

func findLaminaDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, "dev", "lamina")
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("lamina dir not found: %w", err)
	}
	return dir, nil
}

func discoverRepos(laminaDir string) ([]string, error) {
	cmd := exec.Command("lamina", "repo", "--json")
	cmd.Dir = laminaDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lamina repo --json: %w", err)
	}

	var repos []laminaRepo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("parse lamina repo output: %w", err)
	}

	var names []string
	for _, r := range repos {
		names = append(names, r.Name)
	}
	return names, nil
}

func loadManifests(laminaDir string, repos []string) ([]Manifest, error) {
	var manifests []Manifest
	for _, name := range repos {
		yamlPath := filepath.Join(laminaDir, name, "luthier.yaml")
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			continue // no manifest, skip
		}
		var m Manifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse %s: %w", yamlPath, err)
		}
		if m.Name == "" {
			return nil, fmt.Errorf("manifest %s has no name", yamlPath)
		}
		manifests = append(manifests, m)
		fmt.Fprintf(os.Stderr, "  %s: %s\n", m.Name, m.Purpose)
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].Name < manifests[j].Name
	})
	return manifests, nil
}

func findLuthierRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		modPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modPath); err == nil {
			if strings.Contains(string(data), "github.com/benaskins/luthier") {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find luthier repo root")
		}
		dir = parent
	}
}

func generateCatalog(manifests []Manifest, root string) error {
	promptPath := filepath.Join(root, "internal", "patterns", "system_prompt.txt")
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("read system prompt: %w", err)
	}

	content := string(data)

	startMarker := "## Axon Module Catalog\n"
	endMarker := "\n## Established Patterns"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)
	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("could not find catalog markers in system prompt")
	}

	var table strings.Builder
	table.WriteString("## Axon Module Catalog\n\n")
	table.WriteString("| Module | Purpose | Use when |\n")
	table.WriteString("|--------|---------|----------|\n")
	for _, m := range manifests {
		fmt.Fprintf(&table, "| %s | %s | %s |\n", m.Name, m.Purpose, m.UseWhen)
	}

	newContent := content[:startIdx] + table.String() + content[endIdx:]

	if err := os.WriteFile(promptPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("write system prompt: %w", err)
	}

	fmt.Fprintf(os.Stderr, "luthier-sync: updated %s (%d modules)\n", promptPath, len(manifests))
	return nil
}

func generateSnippets(manifests []Manifest, root string) error {
	outPath := filepath.Join(root, "internal", "snippets", "generated.go")

	var b strings.Builder
	b.WriteString("// Code generated by luthier-sync. DO NOT EDIT.\n")
	b.WriteString("package snippets\n\n")
	b.WriteString("// GeneratedSnippets returns all snippets loaded from axon module luthier.yaml manifests.\n")
	b.WriteString("func GeneratedSnippets() []Snippet {\n")
	b.WriteString("\treturn []Snippet{\n")

	for _, m := range manifests {
		b.WriteString("\t\t{\n")
		fmt.Fprintf(&b, "\t\t\tModule: %q,\n", m.Name)

		if len(m.Snippet.Imports) > 0 {
			b.WriteString("\t\t\tImports: []Import{\n")
			for _, imp := range m.Snippet.Imports {
				if imp.Alias != "" {
					fmt.Fprintf(&b, "\t\t\t\t{Path: %q, Alias: %q},\n", imp.Path, imp.Alias)
				} else {
					fmt.Fprintf(&b, "\t\t\t\t{Path: %q},\n", imp.Path)
				}
			}
			b.WriteString("\t\t\t},\n")
		}

		if len(m.Snippet.Requires) > 0 {
			b.WriteString("\t\t\tRequires: []string{\n")
			for _, r := range m.Snippet.Requires {
				fmt.Fprintf(&b, "\t\t\t\t%q,\n", r)
			}
			b.WriteString("\t\t\t},\n")
		}

		if len(m.Snippet.Deps) > 0 {
			b.WriteString("\t\t\tDeps: []string{\n")
			for _, d := range m.Snippet.Deps {
				fmt.Fprintf(&b, "\t\t\t\t%q,\n", d)
			}
			b.WriteString("\t\t\t},\n")
		}

		if s := strings.TrimRight(m.Snippet.Setup, "\n"); s != "" {
			lines := strings.Split(s, "\n")
			for i, l := range lines {
				if l != "" {
					lines[i] = "\t" + l
				}
			}
			fmt.Fprintf(&b, "\t\t\tSetup: %q,\n", strings.Join(lines, "\n"))
		}

		if s := strings.TrimRight(m.Snippet.Helpers, "\n"); s != "" {
			fmt.Fprintf(&b, "\t\t\tHelpers: %q,\n", s)
		}

		b.WriteString("\t\t},\n")
	}

	b.WriteString("\t}\n")
	b.WriteString("}\n")

	if err := os.WriteFile(outPath, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write generated snippets: %w", err)
	}

	fmt.Fprintf(os.Stderr, "luthier-sync: wrote %s\n", outPath)
	return nil
}
