// Package git manages commit operations for maestro.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// ErrNoChanges is returned when there is nothing to commit.
var ErrNoChanges = fmt.Errorf("no changes to commit")

// ErrMergeConflict is returned when the working directory has unresolved merge conflicts.
var ErrMergeConflict = fmt.Errorf("unresolved merge conflicts")

// Commit stages all changes and commits with the given message.
// Returns ErrNoChanges if there is nothing to commit.
// Returns ErrMergeConflict if there are unresolved conflicts.
func Commit(projectDir string, message string) error {
	if err := HasConflicts(projectDir); err != nil {
		return err
	}

	// Stage only files within the project directory, not the entire repo.
	// Using "git add ." scopes to projectDir since cmd.Dir is set there.
	if err := run(projectDir, "git", "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are staged changes
	out, err := output(projectDir, "git", "diff", "--cached", "--stat")
	if err != nil {
		return fmt.Errorf("git diff: %w", err)
	}
	if strings.TrimSpace(out) == "" {
		return ErrNoChanges
	}

	if err := run(projectDir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// HasConflicts returns ErrMergeConflict if the working directory has unresolved merge conflicts.
func HasConflicts(projectDir string) error {
	out, err := output(projectDir, "git", "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil // not in a git repo or other non-conflict error
	}
	if strings.TrimSpace(out) != "" {
		return fmt.Errorf("%w: %s", ErrMergeConflict, strings.TrimSpace(out))
	}
	return nil
}

// IsClean returns true if the working directory has no uncommitted changes.
func IsClean(projectDir string) (bool, error) {
	out, err := output(projectDir, "git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(out) == "", nil
}

// Diff returns the staged + unstaged diff.
func Diff(projectDir string) (string, error) {
	out, err := output(projectDir, "git", "diff", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
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
		if err := run(projectDir, "git", "add", "."); err != nil {
			return fmt.Errorf("git add: %w", err)
		}
		if err := run(projectDir, "git", "commit", "-m", "feat: initial scaffold from luthier"); err != nil {
			return fmt.Errorf("git commit: %w", err)
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
