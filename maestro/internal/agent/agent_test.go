package agent

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benaskins/maestro/internal/plan"
)

func TestNew(t *testing.T) {
	tests := []struct {
		coder   string
		wantErr bool
		wantType string
	}{
		{"claude", false, "*agent.Claude"},
		{"", false, "*agent.Claude"},
		{"noop", false, "*agent.Noop"},
		{"unknown", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.coder, func(t *testing.T) {
			a, err := New(tt.coder, false)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("New(%q): expected error, got nil", tt.coder)
				}
				return
			}
			if err != nil {
				t.Fatalf("New(%q): unexpected error: %v", tt.coder, err)
			}
			if a == nil {
				t.Fatalf("New(%q): got nil agent", tt.coder)
			}
		})
	}
}

func TestNoopAgent(t *testing.T) {
	n := &Noop{}
	step := plan.Step{Number: 1, Title: "test step", Description: "do something", Commit: "feat: test"}

	out, err := n.Implement("/tmp", step, "")
	if err != nil {
		t.Fatalf("Implement: unexpected error: %v", err)
	}
	if !strings.Contains(out, "noop") {
		t.Errorf("output %q does not contain 'noop'", out)
	}
	if len(n.Calls) != 1 {
		t.Fatalf("expected 1 call recorded, got %d", len(n.Calls))
	}
	call := n.Calls[0]
	if call.ProjectDir != "/tmp" {
		t.Errorf("ProjectDir = %q, want /tmp", call.ProjectDir)
	}
	if call.Step.Title != "test step" {
		t.Errorf("Step.Title = %q, want 'test step'", call.Step.Title)
	}
}

func TestNoopAgentRecordsFeedback(t *testing.T) {
	n := &Noop{}
	step := plan.Step{Number: 2, Title: "retry step"}

	n.Implement("/proj", step, "previous error message")

	if n.Calls[0].Feedback != "previous error message" {
		t.Errorf("Feedback not recorded: got %q", n.Calls[0].Feedback)
	}
}

func TestBuildPrompt(t *testing.T) {
	step := plan.Step{
		Number:      3,
		Title:       "wire agent delegation",
		Description: "Create coding agent interface with implementations.",
		Commit:      "feat: implement coding agent delegation interface",
	}

	prompt := buildPrompt(step, "")

	checks := []string{
		"Step 3: wire agent delegation",
		"Create coding agent interface",
		"feat: implement coding agent delegation interface",
		"Do NOT commit",
	}
	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestBuildPromptWithFeedback(t *testing.T) {
	step := plan.Step{Number: 1, Title: "step"}
	prompt := buildPrompt(step, "tests failed: undefined: Foo")

	if !strings.Contains(prompt, "Previous attempt failed") {
		t.Error("prompt missing 'Previous attempt failed' section")
	}
	if !strings.Contains(prompt, "undefined: Foo") {
		t.Error("prompt missing feedback content")
	}
}

func TestRunCommand(t *testing.T) {
	result, err := runCommand("", "echo", "hello world")
	if err != nil {
		t.Fatalf("runCommand: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "hello world" {
		t.Errorf("stdout = %q, want 'hello world'", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Error("duration should be positive")
	}
}

func TestRunCommandCapturesStderr(t *testing.T) {
	// Use a shell command that writes to stderr
	result, err := runCommand("", "sh", "-c", "echo errout >&2; exit 1")
	if err == nil {
		t.Fatal("expected non-zero exit, got nil error")
	}
	if !strings.Contains(result.Stderr, "errout") {
		t.Errorf("stderr = %q, want 'errout'", result.Stderr)
	}
	if result.ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", result.ExitCode)
	}
}

func TestRunCommandSeparatesStreams(t *testing.T) {
	result, err := runCommand("", "sh", "-c", "echo stdout; echo stderr >&2")
	if err != nil {
		t.Fatalf("runCommand: %v", err)
	}
	if !strings.Contains(result.Stdout, "stdout") {
		t.Errorf("stdout %q missing 'stdout'", result.Stdout)
	}
	if strings.Contains(result.Stdout, "stderr") {
		t.Errorf("stderr leaked into stdout: %q", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "stderr") {
		t.Errorf("stderr %q missing 'stderr'", result.Stderr)
	}
}

func TestRunCommandWithWriter_StreamsOutput(t *testing.T) {
	var buf bytes.Buffer
	result, err := runCommandWithWriter("", &buf, "echo", "hello stream")
	if err != nil {
		t.Fatalf("runCommandWithWriter: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "hello stream" {
		t.Errorf("stdout = %q, want 'hello stream'", result.Stdout)
	}
	if strings.TrimSpace(buf.String()) != "hello stream" {
		t.Errorf("streamed output = %q, want 'hello stream'", buf.String())
	}
}

func TestRunCommandWithWriter_NilWriterBuffersOnly(t *testing.T) {
	result, err := runCommandWithWriter("", nil, "echo", "buffered only")
	if err != nil {
		t.Fatalf("runCommandWithWriter: %v", err)
	}
	if strings.TrimSpace(result.Stdout) != "buffered only" {
		t.Errorf("stdout = %q, want 'buffered only'", result.Stdout)
	}
}

func TestClaudeVerbose_StreamsToWriter(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not found in PATH, skipping integration test")
	}

	dir := t.TempDir()
	step := plan.Step{
		Number:      1,
		Title:       "create a file",
		Description: "Create a file named out.txt containing the text 'verbose test'.",
		Commit:      "feat: create out.txt",
	}

	var buf bytes.Buffer
	a := &Claude{Verbose: true, Out: &buf}
	_, err := a.Implement(dir, step, "")
	if err != nil {
		t.Fatalf("Implement: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("verbose mode produced no streaming output")
	}
}

// TestClaudeCodeIntegration is an integration test that invokes Claude Code
// with a simple file-creation task and verifies the file was created.
// It is skipped if `claude` is not found in PATH.
func TestClaudeCodeIntegration(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not found in PATH, skipping integration test")
	}

	dir := t.TempDir()

	step := plan.Step{
		Number:      1,
		Title:       "create hello file",
		Description: "Create a file named hello.txt in the current directory containing the text 'hello from maestro'.",
		Commit:      "feat: create hello file",
	}

	agent := &Claude{Verbose: false}
	out, err := agent.Implement(dir, step, "")
	if err != nil {
		t.Fatalf("Implement: %v\noutput: %s", err, out)
	}

	target := filepath.Join(dir, "hello.txt")
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("hello.txt not created: %v\nagent output:\n%s", err, out)
	}
	if !strings.Contains(string(data), "hello from maestro") {
		t.Errorf("hello.txt content = %q, want 'hello from maestro'", string(data))
	}
}
