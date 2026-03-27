package writer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benaskins/luthier/internal/analysis"
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

	if err := Write(testSpec(), outDir); err != nil {
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
	if err := Write(testSpec(), outDir); err != nil {
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
	if err := Write(testSpec(), outDir); err != nil {
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
	if err := Write(testSpec(), outDir); err != nil {
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
