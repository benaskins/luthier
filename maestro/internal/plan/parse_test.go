package plan

import (
	"os"
	"path/filepath"
	"strings"
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

func TestParse_EmptyContent(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty content, got nil")
	}
}

func TestParse_ContentWithNoSteps(t *testing.T) {
	content := "# Just a title\n\nSome prose but no step headers.\n"
	_, err := Parse(content)
	if err == nil {
		t.Fatal("expected error when no steps found, got nil")
	}
}

func TestParse_StepWithNoCommit(t *testing.T) {
	content := `## Step 1 — do something

Description without a commit line.
`
	steps, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 {
		t.Fatalf("got %d steps, want 1", len(steps))
	}
	if steps[0].Commit != "" {
		t.Errorf("expected empty commit for step without commit line, got %q", steps[0].Commit)
	}
}

func TestParse_DashVariants(t *testing.T) {
	// stepHeader regex accepts em-dash, en-dash, and hyphen-minus.
	cases := []struct {
		name    string
		header  string
		wantNum int
		wantTitle string
	}{
		{"em dash", "## Step 1 — em dash title", 1, "em dash title"},
		{"en dash", "## Step 2 – en dash title", 2, "en dash title"},
		{"hyphen", "## Step 3 - hyphen title", 3, "hyphen title"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			content := tc.header + "\n\nDesc.\n\nCommit: `feat: test`\n"
			steps, err := Parse(content)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if len(steps) != 1 {
				t.Fatalf("got %d steps, want 1", len(steps))
			}
			if steps[0].Number != tc.wantNum {
				t.Errorf("Number = %d, want %d", steps[0].Number, tc.wantNum)
			}
			if steps[0].Title != tc.wantTitle {
				t.Errorf("Title = %q, want %q", steps[0].Title, tc.wantTitle)
			}
		})
	}
}

func TestParse_DescriptionPreservedAcrossLines(t *testing.T) {
	content := `## Step 1 — multi line

Line one of description.
Line two of description.

Commit: ` + "`feat: multi line`" + `
`
	steps, err := Parse(content)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(steps[0].Description, "Line one") {
		t.Errorf("description missing 'Line one': %q", steps[0].Description)
	}
	if !strings.Contains(steps[0].Description, "Line two") {
		t.Errorf("description missing 'Line two': %q", steps[0].Description)
	}
}

func TestReadFromDir_MultiplePlanFilesPicksLast(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "plans")
	os.MkdirAll(planDir, 0o755)

	older := `## Step 1 — old step

Old description.

Commit: ` + "`feat: old`" + `
`
	newer := `## Step 1 — new step

New description.

Commit: ` + "`feat: new`" + `
`
	// Alphabetically, 2026-03-30 > 2026-01-01, so the newer file should be picked.
	os.WriteFile(filepath.Join(planDir, "2026-01-01-old.md"), []byte(older), 0o644)
	os.WriteFile(filepath.Join(planDir, "2026-03-30-new.md"), []byte(newer), 0o644)

	steps, err := ReadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].Title != "new step" {
		t.Errorf("expected most recent plan, got title %q", steps[0].Title)
	}
}

func TestReadFromDir_NoPlanFiles(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, "plans")
	os.MkdirAll(planDir, 0o755)
	// Write a non-markdown file — should be ignored.
	os.WriteFile(filepath.Join(planDir, "notes.txt"), []byte("not a plan"), 0o644)

	_, err := ReadFromDir(dir)
	if err == nil {
		t.Fatal("expected error when no .md plan files exist, got nil")
	}
}

func TestReadFromDir_MissingPlanDirectory(t *testing.T) {
	dir := t.TempDir()
	// Do not create plans/ subdirectory.
	_, err := ReadFromDir(dir)
	if err == nil {
		t.Fatal("expected error when plans/ directory is missing, got nil")
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
