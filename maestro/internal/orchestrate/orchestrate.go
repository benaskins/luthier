// Package orchestrate runs the maestro plan execution loop.
package orchestrate

import (
	"context"
	"fmt"
	"os"

	"github.com/benaskins/maestro/internal/agent"
	gitpkg "github.com/benaskins/maestro/internal/git"
	"github.com/benaskins/maestro/internal/plan"
	"github.com/benaskins/maestro/internal/review"
	"github.com/benaskins/maestro/internal/verify"
)

// Reviewer assesses whether an implementation satisfies a plan step.
// review.Reviewer satisfies this interface.
type Reviewer interface {
	Review(ctx context.Context, diff string, step plan.Step) (*review.Result, error)
}

// Config holds orchestration settings.
type Config struct {
	ProjectDir string
	Agent      agent.Agent
	Reviewer   Reviewer // optional; if nil, semantic review is skipped
	DryRun     bool
	Verbose    bool
	ResumeFrom string
	MaxRetries int
	// VerifyCmd overrides automatic verification command detection.
	// Useful in tests and when the caller already knows the command.
	VerifyCmd string
}

// Result summarises a completed orchestration run.
type Result struct {
	Total     int
	Completed int
	Skipped   int
	Failed    int
	FailedAt  *plan.Step
}

// Run executes the plan steps in order.
func Run(cfg Config) (*Result, error) {
	steps, err := plan.ReadFromDir(cfg.ProjectDir)
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}

	verifyCmd := cfg.VerifyCmd
	if verifyCmd == "" {
		verifyCmd, err = verify.DetectCommand(cfg.ProjectDir)
		if err != nil {
			return nil, fmt.Errorf("detect verification: %w", err)
		}
	}

	if err := gitpkg.InitIfNeeded(cfg.ProjectDir); err != nil {
		return nil, fmt.Errorf("git init: %w", err)
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	result := &Result{Total: len(steps)}
	resuming := cfg.ResumeFrom != ""

	for i := range steps {
		step := &steps[i]

		// Skip already-committed steps
		if gitpkg.IsStepCommitted(cfg.ProjectDir, step.Commit) {
			result.Skipped++
			fmt.Fprintf(os.Stderr, "  [%d/%d] %s (already committed, skipping)\n", step.Number, result.Total, step.Title)
			continue
		}

		// Handle resume-from
		if resuming {
			if step.Title != cfg.ResumeFrom && fmt.Sprintf("%d", step.Number) != cfg.ResumeFrom {
				result.Skipped++
				fmt.Fprintf(os.Stderr, "  [%d/%d] %s (skipping, resuming from %s)\n", step.Number, result.Total, step.Title, cfg.ResumeFrom)
				continue
			}
			resuming = false
		}

		fmt.Fprintf(os.Stderr, "\n  [%d/%d] %s\n", step.Number, result.Total, step.Title)

		if cfg.DryRun {
			fmt.Fprintf(os.Stderr, "    dry-run: would delegate to coding agent\n")
			fmt.Fprintf(os.Stderr, "    dry-run: would run: %s\n", verifyCmd)
			fmt.Fprintf(os.Stderr, "    dry-run: would commit: %s\n", step.Commit)
			result.Completed++
			continue
		}

		if err := executeStep(cfg, *step, verifyCmd); err != nil {
			result.Failed++
			result.FailedAt = step
			return result, fmt.Errorf("step %d (%s): %w", step.Number, step.Title, err)
		}

		result.Completed++
	}

	return result, nil
}

func executeStep(cfg Config, step plan.Step, verifyCmd string) error {
	var feedback string

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 1 {
			fmt.Fprintf(os.Stderr, "    retry %d/%d\n", attempt, cfg.MaxRetries)
		}

		// Delegate to coding agent
		fmt.Fprintf(os.Stderr, "    delegating to coding agent...\n")
		agentOut, err := cfg.Agent.Implement(cfg.ProjectDir, step, feedback)
		if cfg.Verbose && agentOut != "" {
			fmt.Fprintf(os.Stderr, "    agent output:\n%s\n", agentOut)
		}
		if err != nil {
			feedback = fmt.Sprintf("Agent error: %s\nOutput: %s", err, agentOut)
			fmt.Fprintf(os.Stderr, "    agent failed: %v\n", err)
			continue
		}

		// Run verification
		fmt.Fprintf(os.Stderr, "    verifying: %s\n", verifyCmd)
		verifyOut, err := verify.Run(cfg.ProjectDir, verifyCmd)
		if err != nil {
			feedback = fmt.Sprintf("Verification failed:\n%s", verifyOut)
			fmt.Fprintf(os.Stderr, "    verification failed\n")
			continue
		}
		fmt.Fprintf(os.Stderr, "    verification passed\n")

		// Run semantic review if a reviewer is configured
		if cfg.Reviewer != nil {
			diff, diffErr := gitpkg.Diff(cfg.ProjectDir)
			if diffErr != nil {
				feedback = fmt.Sprintf("Failed to get diff for review: %s", diffErr)
				fmt.Fprintf(os.Stderr, "    could not get diff: %v\n", diffErr)
				continue
			}

			reviewResult, reviewErr := cfg.Reviewer.Review(context.Background(), diff, step)
			if reviewErr != nil {
				// Review errors are non-fatal: log and proceed to commit.
				fmt.Fprintf(os.Stderr, "    review error (skipping): %v\n", reviewErr)
			} else if !reviewResult.Passed {
				feedback = fmt.Sprintf("Semantic review failed: %s", reviewResult.Reason)
				fmt.Fprintf(os.Stderr, "    review failed: %s\n", reviewResult.Reason)
				continue
			} else {
				fmt.Fprintf(os.Stderr, "    review passed: %s\n", reviewResult.Reason)
			}
		}

		// Commit
		if err := gitpkg.Commit(cfg.ProjectDir, step.Commit); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		fmt.Fprintf(os.Stderr, "    committed: %s\n", step.Commit)
		return nil
	}

	return fmt.Errorf("failed after %d attempts. Last error: %s", cfg.MaxRetries, feedback)
}
