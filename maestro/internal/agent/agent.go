// Package agent delegates plan step implementation to a coding agent.
package agent

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/benaskins/maestro/internal/plan"
)

// Agent can implement a plan step by modifying files in a project.
type Agent interface {
	Implement(projectDir string, step plan.Step, feedback string) (string, error)
}

// Claude delegates to Claude Code via `claude -p`.
type Claude struct {
	Verbose bool
}

// Implement sends the step to Claude Code and returns the output.
func (c *Claude) Implement(projectDir string, step plan.Step, feedback string) (string, error) {
	prompt := buildPrompt(step, feedback)

	args := []string{
		"-p", prompt,
		"--allowedTools", "Bash,Read,Write,Edit,Grep,Glob",
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = projectDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("claude: %w\n%s", err, out)
	}
	return string(out), nil
}

func buildPrompt(step plan.Step, feedback string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Implement this plan step. Write the code, not a plan.\n\n")
	fmt.Fprintf(&b, "## Step %d: %s\n\n", step.Number, step.Title)
	fmt.Fprintf(&b, "%s\n\n", step.Description)
	fmt.Fprintf(&b, "Commit message when done: %s\n", step.Commit)
	fmt.Fprintf(&b, "\nDo NOT commit. Just write the code and tests described above.\n")
	fmt.Fprintf(&b, "Today's date is %s.\n", "2026-03-30")

	if feedback != "" {
		fmt.Fprintf(&b, "\n## Previous attempt failed\n\n%s\n", feedback)
		fmt.Fprintf(&b, "\nFix the issues described above.\n")
	}

	return b.String()
}
