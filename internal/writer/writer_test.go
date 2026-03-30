package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benaskins/luthier/internal/analysis"
	"github.com/benaskins/luthier/internal/catalogue"
)

func testSpec() *analysis.ScaffoldSpec {
	return &analysis.ScaffoldSpec{
		Name: "my-service",
		Modules: []analysis.ModuleSelection{
			{Name: "axon", Reason: "HTTP server", IsDeterministic: true},
			{Name: "axon-loop", Reason: "LLM conversation", IsDeterministic: false},
		},
		Boundaries: []analysis.Boundary{
			{From: "handler", To: "llm", Type: "non-det"},
		},
		PlanSteps: []analysis.PlanStep{
			{
				Title:         "Scaffold repo",
				Description:   "Create the initial project structure.",
				CommitMessage: "feat: scaffold my-service",
			},
			{
				Title:         "Implement handler",
				Description:   "Add HTTP handler.",
				CommitMessage: "feat: add HTTP handler",
			},
		},
		Gaps: []analysis.Gap{},
	}
}

func TestWrite_CreatesExpectedFiles(t *testing.T) {
	outDir := t.TempDir()

	if err := Write(testSpec(), outDir, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	expected := []string{
		"AGENTS.md",
		"CLAUDE.md",
		"README.md",
		"go.mod",
		"justfile",
		filepath.Join("cmd", "my-service", "main.go"),
	}
	for _, f := range expected {
		path := filepath.Join(outDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file not found: %s", f)
		}
	}

	// plans/ directory should contain one file
	entries, err := os.ReadDir(filepath.Join(outDir, "plans"))
	if err != nil {
		t.Fatalf("ReadDir plans: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 plan file, got %d", len(entries))
	}
}

func TestWrite_AgentsMdContainsModules(t *testing.T) {
	outDir := t.TempDir()
	if err := Write(testSpec(), outDir, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("ReadFile AGENTS.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "my-service") {
		t.Error("AGENTS.md does not contain project name")
	}
	if !strings.Contains(content, "axon") {
		t.Error("AGENTS.md does not contain axon module")
	}
	if !strings.Contains(content, "axon-loop") {
		t.Error("AGENTS.md does not contain axon-loop module")
	}
	if !strings.Contains(content, "non-det") {
		t.Error("AGENTS.md does not contain boundary type")
	}
}

func TestWrite_PlanContainsSteps(t *testing.T) {
	outDir := t.TempDir()
	if err := Write(testSpec(), outDir, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	entries, _ := os.ReadDir(filepath.Join(outDir, "plans"))
	data, err := os.ReadFile(filepath.Join(outDir, "plans", entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile plan: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Scaffold repo") {
		t.Error("plan does not contain first step title")
	}
	if !strings.Contains(content, "Implement handler") {
		t.Error("plan does not contain second step title")
	}
	if !strings.Contains(content, "feat: scaffold my-service") {
		t.Error("plan does not contain commit message")
	}
}

func TestWrite_GoModContainsModuleName(t *testing.T) {
	outDir := t.TempDir()
	if err := Write(testSpec(), outDir, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "go.mod"))
	if err != nil {
		t.Fatalf("ReadFile go.mod: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "github.com/benaskins/my-service") {
		t.Errorf("go.mod missing module path, got:\n%s", content)
	}
}

func testCatalogue() *catalogue.Catalogue {
	return &catalogue.Catalogue{
		Name:     "Ruby on Rails",
		Language: "ruby",
		Version:  "8.0",
		Components: []catalogue.Component{
			{Name: "rails", Class: "platform", Purpose: "Full-stack web framework"},
			{Name: "devise", Class: "domain", Purpose: "Authentication"},
		},
		Patterns: []catalogue.Pattern{
			{Requirement: "Authentication", Pattern: "devise with default modules"},
		},
		FileConvs: []catalogue.FileConv{
			{Path: "app/models/", Description: "ActiveRecord models"},
		},
		BoundaryN: "Most boundaries are deterministic in Rails.",
	}
}

func TestWrite_WithCatalogue_CLAUDEMdContainsCatalogueContent(t *testing.T) {
	outDir := t.TempDir()
	opts := &Options{Catalogue: testCatalogue()}
	if err := Write(testSpec(), outDir, opts); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile CLAUDE.md: %v", err)
	}
	content := string(data)

	checks := []struct {
		desc string
		want string
	}{
		{"framework name", "Ruby on Rails"},
		{"component", "devise"},
		{"pattern", "devise with default modules"},
		{"file convention", "app/models/"},
		{"boundary notes", "Most boundaries are deterministic"},
		{"practice section", "## Practice"},
		{"TDD instruction", "Write a failing test first"},
	}
	for _, c := range checks {
		if !strings.Contains(content, c.want) {
			t.Errorf("CLAUDE.md missing %s (%q)", c.desc, c.want)
		}
	}
}

func TestWrite_WithCatalogue_CopiesCatalogueYAML(t *testing.T) {
	outDir := t.TempDir()

	// Write a temp catalogue file to copy
	catPath := filepath.Join(t.TempDir(), "test-catalogue.yaml")
	if err := os.WriteFile(catPath, []byte("name: test\n"), 0o644); err != nil {
		t.Fatalf("write test catalogue: %v", err)
	}

	opts := &Options{
		Catalogue:     testCatalogue(),
		CataloguePath: catPath,
	}
	if err := Write(testSpec(), outDir, opts); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "catalogue.yaml"))
	if err != nil {
		t.Fatalf("catalogue.yaml not written: %v", err)
	}
	if !strings.Contains(string(data), "name: test") {
		t.Error("catalogue.yaml content mismatch")
	}
}

func TestWrite_WithoutCatalogue_NoCatalogueSection(t *testing.T) {
	outDir := t.TempDir()
	if err := Write(testSpec(), outDir, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("ReadFile CLAUDE.md: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "## Framework") {
		t.Error("CLAUDE.md should not contain Framework section without catalogue")
	}
	// Practice section should still be present
	if !strings.Contains(content, "## Practice") {
		t.Error("CLAUDE.md should contain Practice section even without catalogue")
	}
}
