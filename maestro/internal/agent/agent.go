// Package agent delegates plan step implementation to a coding agent.
package agent

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/benaskins/maestro/internal/plan"
)

// Agent can implement a plan step by modifying files in a project.
type Agent interface {
	Implement(projectDir string, step plan.Step, feedback string) (string, error)
}

// ExecResult holds the captured output from a command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// runCommand executes a command in the given directory and captures stdout/stderr separately.
func runCommand(dir string, name string, args ...string) (ExecResult, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	result := ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: elapsed,
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	return result, err
}

// Claude delegates to Claude Code via `claude -p`.
type Claude struct {
	Verbose bool
}

// Implement sends the step to Claude Code and returns the output.
func (c *Claude) Implement(projectDir string, step plan.Step, feedback string) (string, error) {
	prompt := buildPrompt(step, feedback)

	result, err := runCommand(projectDir, "claude",
		"-p", prompt,
		"--allowedTools", "Bash,Read,Write,Edit,Grep,Glob",
	)

	combined := result.Stdout
	if result.Stderr != "" {
		combined += "\n" + result.Stderr
	}

	if err != nil {
		return combined, fmt.Errorf("claude exited %d: %w", result.ExitCode, err)
	}
	return combined, nil
}

// Noop is a placeholder agent that records calls without executing anything.
// Useful for dry-run mode and as a template for other agent implementations.
type Noop struct {
	Calls []NoopCall
}

// NoopCall records the arguments passed to a single Implement call.
type NoopCall struct {
	ProjectDir string
	Step       plan.Step
	Feedback   string
}

// Implement records the call and returns a no-op success message.
func (n *Noop) Implement(projectDir string, step plan.Step, feedback string) (string, error) {
	n.Calls = append(n.Calls, NoopCall{
		ProjectDir: projectDir,
		Step:       step,
		Feedback:   feedback,
	})
	return fmt.Sprintf("[noop] would implement step %d: %s", step.Number, step.Title), nil
}

// New returns an Agent for the given coder name.
// Valid values: "claude" (default), "noop".
func New(coder string, verbose bool) (Agent, error) {
	switch coder {
	case "", "claude":
		return &Claude{Verbose: verbose}, nil
	case "noop":
		return &Noop{}, nil
	default:
		return nil, fmt.Errorf("unknown coder %q: supported values are claude, noop", coder)
	}
}

func buildPrompt(step plan.Step, feedback string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Implement this plan step. Write the code, not a plan.\n\n")
	fmt.Fprintf(&b, "## Step %d: %s\n\n", step.Number, step.Title)
	fmt.Fprintf(&b, "%s\n\n", step.Description)
	fmt.Fprintf(&b, "Commit message when done: %s\n", step.Commit)
	fmt.Fprintf(&b, "\nDo NOT commit. Just write the code and tests described above.\n")
	fmt.Fprintf(&b, "Today's date is %s.\n", time.Now().Format("2006-01-02"))

	if feedback != "" {
		fmt.Fprintf(&b, "\n## Previous attempt failed\n\n%s\n", feedback)
		fmt.Fprintf(&b, "\nFix the issues described above.\n")
	}

	return b.String()
}
