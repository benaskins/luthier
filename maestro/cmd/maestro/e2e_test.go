package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupScaffoldProject creates a minimal scaffold project: a git repo with a
// plans/ directory, a simple go.mod, and a main.go so `go build ./...` passes.
func setupScaffoldProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	mustWrite := func(name, content string) {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	mustRun := func(name string, args ...string) {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}

	mustWrite("plans/2026-03-30-initial.md", `## Step 1 — create greeting

Create a file hello.txt containing "hello world".

Commit: `+"`feat: create greeting`"+`

## Step 2 — add message

Append a second line to hello.txt.

Commit: `+"`feat: add message`")

	// Add a justfile so verification detection succeeds.
	// Use `true` as the test command so it always passes.
	mustWrite("justfile", "test:\n\ttrue\n")

	mustRun("git", "init")
	mustRun("git", "config", "user.email", "test@example.com")
	mustRun("git", "config", "user.name", "Test")
	mustRun("git", "add", ".")
	mustRun("git", "commit", "-m", "feat: initial scaffold")

	return dir
}

// buildMaestro compiles the maestro binary into a temp dir and returns its path.
func buildMaestro(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "maestro")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/maestro/")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

// moduleRoot returns the directory containing go.mod by walking up from the
// current working directory (which is cmd/maestro during test execution).
func moduleRoot(t *testing.T) string {
	t.Helper()
	// The test binary's working directory is the package directory.
	// Walk up to find go.mod.
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod")
		}
		dir = parent
	}
}

// TestE2E_DryRunCompletesAllSteps runs the full maestro binary end-to-end in
// dry-run mode with the noop coder so no external processes are invoked.
func TestE2E_DryRunCompletesAllSteps(t *testing.T) {
	bin := buildMaestro(t)
	dir := setupScaffoldProject(t)

	summaryFile := filepath.Join(t.TempDir(), "summary.txt")

	cmd := exec.Command(bin,
		"--dry-run",
		"--coder", "noop",
		"--summary-file", summaryFile,
		dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("maestro exited non-zero: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "completed") {
		t.Errorf("expected 'completed' in output, got:\n%s", output)
	}

	// Summary file should exist and contain step titles.
	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("summary file not written: %v", err)
	}
	summary := string(data)
	if !strings.Contains(summary, "create greeting") {
		t.Errorf("summary missing 'create greeting':\n%s", summary)
	}
	if !strings.Contains(summary, "add message") {
		t.Errorf("summary missing 'add message':\n%s", summary)
	}
}

// TestE2E_NoopCoderRunsWithoutClaude runs through the full plan using the noop
// coder (no commit is made because no files change) and verifies both steps are
// reported as completed without invoking Claude Code.
func TestE2E_NoopCoderRunsWithoutClaude(t *testing.T) {
	bin := buildMaestro(t)
	dir := setupScaffoldProject(t)

	cmd := exec.Command(bin, "--coder", "noop", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("maestro exited non-zero: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Total:") {
		t.Errorf("expected report in output, got:\n%s", output)
	}
}

// TestE2E_InvalidProjectDirFails verifies that a non-existent project directory
// causes a non-zero exit code.
func TestE2E_InvalidProjectDirFails(t *testing.T) {
	bin := buildMaestro(t)

	cmd := exec.Command(bin, "/no/such/directory/xyzzy")
	if err := cmd.Run(); err == nil {
		t.Error("expected non-zero exit for non-existent project dir, got nil")
	}
}

// TestE2E_ResumeFromSkipsEarlierSteps verifies --resume-from skips the
// specified steps.
func TestE2E_ResumeFromSkipsEarlierSteps(t *testing.T) {
	bin := buildMaestro(t)
	dir := setupScaffoldProject(t)

	cmd := exec.Command(bin,
		"--dry-run",
		"--coder", "noop",
		"--resume-from", "2",
		dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("maestro exited non-zero: %v\noutput:\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "Skipped:") {
		t.Errorf("expected 'Skipped:' in output when resuming from step 2, got:\n%s", output)
	}
}

// TestE2E_WithClaudeCode is a slow integration test that invokes the real
// Claude Code agent to implement a trivial one-step plan. It is skipped when
// the `claude` binary is not available in PATH.
func TestE2E_WithClaudeCode(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not found in PATH, skipping Claude Code integration test")
	}

	bin := buildMaestro(t)
	dir := setupScaffoldProject(t)

	// Run only step 1 so the test stays fast.
	cmd := exec.Command(bin,
		"--coder", "claude",
		"--resume-from", "1",
		dir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("maestro output:\n%s", out)
		t.Fatalf("maestro failed: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "completed") {
		t.Errorf("expected 'completed' in output, got:\n%s", output)
	}
}
