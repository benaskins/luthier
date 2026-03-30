package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	content := `# test — Initial Build Plan
# 2026-03-30

Each step is commit-sized.

## Step 1 — set up project

Create main.go and go.mod. Test by running go build.

Commit: ` + "`feat: initialize project`" + `

## Step 2 — add parser

Implement the parser module. Test with sample files.

Commit: ` + "`feat: add parser`" + `
`
	steps, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(steps))
	}

	if steps[0].Number != 1 {
		t.Errorf("step 0 number = %d, want 1", steps[0].Number)
	}
	if steps[0].Title != "set up project" {
		t.Errorf("step 0 title = %q, want %q", steps[0].Title, "set up project")
	}
	if steps[0].Commit != "feat: initialize project" {
		t.Errorf("step 0 commit = %q, want %q", steps[0].Commit, "feat: initialize project")
	}
	if steps[0].Description == "" {
		t.Error("step 0 description is empty")
	}

	if steps[1].Number != 2 {
		t.Errorf("step 1 number = %d, want 2", steps[1].Number)
	}
	if steps[1].Title != "add parser" {
		t.Errorf("step 1 title = %q, want %q", steps[1].Title, "add parser")
	}
}

func TestReadFromDir(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "plans")
	os.MkdirAll(planDir, 0o755)

	content := `## Step 1 — do something

Description here.

Commit: ` + "`feat: do something`" + `
`
	os.WriteFile(filepath.Join(planDir, "2026-03-30-initial-build.md"), []byte(content), 0o644)

	steps, err := ReadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(steps))
	}
	if steps[0].Title != "do something" {
		t.Errorf("title = %q, want %q", steps[0].Title, "do something")
	}
}

func TestParseMaestroOwnPlan(t *testing.T) {
	// Parse maestro's own plan as a real-world test
	steps, err := ReadFromDir("../..")
	if err != nil {
		t.Skip("maestro plan not found:", err)
	}
	if len(steps) == 0 {
		t.Fatal("no steps found in maestro's own plan")
	}
	t.Logf("parsed %d steps from maestro's plan", len(steps))
	for _, s := range steps {
		if s.Title == "" {
			t.Errorf("step %d has empty title", s.Number)
		}
		if s.Commit == "" {
			t.Errorf("step %d (%s) has empty commit message", s.Number, s.Title)
		}
	}
}
