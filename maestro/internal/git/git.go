// Package git manages commit operations for maestro.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit stages all changes and commits with the given message.
func Commit(projectDir string, message string) error {
	if err := run(projectDir, "git", "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are staged changes
	out, err := output(projectDir, "git", "diff", "--cached", "--stat")
	if err != nil {
		return fmt.Errorf("git diff: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return fmt.Errorf("no changes to commit")
	}

	if err := run(projectDir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Diff returns the staged + unstaged diff.
func Diff(projectDir string) (string, error) {
	return output(projectDir, "git", "diff", "HEAD")
}

// IsStepCommitted checks if a commit message already exists in the git log.
func IsStepCommitted(projectDir string, commitMsg string) bool {
	out, err := output(projectDir, "git", "log", "--oneline", "--all")
	if err != nil {
		return false
	}
	return strings.Contains(out, commitMsg)
}

// InitIfNeeded initialises a git repo if one doesn't exist.
func InitIfNeeded(projectDir string) error {
	if _, err := output(projectDir, "git", "rev-parse", "--git-dir"); err != nil {
		if err := run(projectDir, "git", "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
		if err := run(projectDir, "git", "add", "-A"); err != nil {
			return err
		}
		if err := run(projectDir, "git", "commit", "-m", "feat: initial scaffold from luthier"); err != nil {
			return err
		}
	}
	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, out)
	}
	return nil
}

func output(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
