// Package orchestrate runs the maestro plan execution loop.
package orchestrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	fact "github.com/benaskins/axon-fact"
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
	// EventStore records orchestration activities. If nil, events are dropped.
	EventStore fact.EventStore
	// EventStream is the stream name for the event store.
	// Defaults to "orchestration" if empty.
	EventStream string
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

	if cfg.EventStream == "" {
		cfg.EventStream = "orchestration"
	}

	if cfg.ResumeFrom != "" {
		found := false
		for _, s := range steps {
			if s.Title == cfg.ResumeFrom || fmt.Sprintf("%d", s.Number) == cfg.ResumeFrom {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("--resume-from %q does not match any step title or number", cfg.ResumeFrom)
		}
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
	ctx := context.Background()
	stream := cfg.EventStream

	emit(ctx, cfg.EventStore, stream, EventStepStarted, StepStartedData{
		StepNumber: step.Number,
		StepTitle:  step.Title,
	})

	var feedback string

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 1 {
			fmt.Fprintf(os.Stderr, "    retry %d/%d\n", attempt, cfg.MaxRetries)
			emit(ctx, cfg.EventStore, stream, EventRetryAttempt, RetryAttemptData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
		}

		// Delegate to coding agent
		fmt.Fprintf(os.Stderr, "    delegating to coding agent...\n")
		emit(ctx, cfg.EventStore, stream, EventAgentInvoked, AgentInvokedData{
			StepNumber:  step.Number,
			Attempt:     attempt,
			HasFeedback: feedback != "",
		})
		agentOut, err := cfg.Agent.Implement(cfg.ProjectDir, step, feedback)
		if err != nil {
			feedback = fmt.Sprintf("Agent error: %s\nOutput: %s", err, agentOut)
			fmt.Fprintf(os.Stderr, "    agent failed: %v\n", err)
			emit(ctx, cfg.EventStore, stream, EventAgentFailed, AgentFailedData{
				StepNumber: step.Number,
				Attempt:    attempt,
				Error:      err.Error(),
			})
			continue
		}
		emit(ctx, cfg.EventStore, stream, EventAgentSucceeded, AgentSucceededData{
			StepNumber: step.Number,
			Attempt:    attempt,
		})

		// Run verification
		fmt.Fprintf(os.Stderr, "    verifying: %s\n", verifyCmd)
		emit(ctx, cfg.EventStore, stream, EventVerificationRun, VerificationRunData{
			StepNumber: step.Number,
			Attempt:    attempt,
			Command:    verifyCmd,
		})
		verifyOut, err := verify.Run(cfg.ProjectDir, verifyCmd)
		if err != nil {
			feedback = fmt.Sprintf("Verification failed:\n%s", verifyOut)
			fmt.Fprintf(os.Stderr, "    verification failed\n")
			emit(ctx, cfg.EventStore, stream, EventVerificationFailed, VerificationFailedData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
			continue
		}
		fmt.Fprintf(os.Stderr, "    verification passed\n")
		emit(ctx, cfg.EventStore, stream, EventVerificationPassed, VerificationPassedData{
			StepNumber: step.Number,
			Attempt:    attempt,
		})

		// Run semantic review if a reviewer is configured
		if cfg.Reviewer != nil {
			diff, diffErr := gitpkg.Diff(cfg.ProjectDir)
			if diffErr != nil {
				feedback = fmt.Sprintf("Failed to get diff for review: %s", diffErr)
				fmt.Fprintf(os.Stderr, "    could not get diff: %v\n", diffErr)
				continue
			}

			emit(ctx, cfg.EventStore, stream, EventReviewRun, ReviewRunData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
			reviewResult, reviewErr := cfg.Reviewer.Review(ctx, diff, step)
			if reviewErr != nil {
				// Review errors are non-fatal: log and proceed to commit.
				fmt.Fprintf(os.Stderr, "    review error (skipping): %v\n", reviewErr)
				emit(ctx, cfg.EventStore, stream, EventReviewErrored, ReviewErroredData{
					StepNumber: step.Number,
					Attempt:    attempt,
					Error:      reviewErr.Error(),
				})
			} else if !reviewResult.Passed {
				feedback = fmt.Sprintf("Semantic review failed: %s", reviewResult.Reason)
				fmt.Fprintf(os.Stderr, "    review failed: %s\n", reviewResult.Reason)
				emit(ctx, cfg.EventStore, stream, EventReviewFailed, ReviewFailedData{
					StepNumber: step.Number,
					Attempt:    attempt,
					Reason:     reviewResult.Reason,
				})
				continue
			} else {
				fmt.Fprintf(os.Stderr, "    review passed: %s\n", reviewResult.Reason)
				emit(ctx, cfg.EventStore, stream, EventReviewPassed, ReviewPassedData{
					StepNumber: step.Number,
					Attempt:    attempt,
					Reason:     reviewResult.Reason,
				})
			}
		}

		// Commit
		if err := gitpkg.Commit(cfg.ProjectDir, step.Commit); err != nil {
			if err == gitpkg.ErrNoChanges {
				fmt.Fprintf(os.Stderr, "    nothing to commit for: %s\n", step.Commit)
				emit(ctx, cfg.EventStore, stream, EventCommitSkipped, CommitSkippedData{
					StepNumber: step.Number,
					Message:    step.Commit,
				})
			} else {
				emit(ctx, cfg.EventStore, stream, EventCommitFailed, CommitFailedData{
					StepNumber: step.Number,
					Message:    step.Commit,
					Error:      err.Error(),
				})
				emit(ctx, cfg.EventStore, stream, EventStepFailed, StepFailedData{
					StepNumber: step.Number,
					StepTitle:  step.Title,
					Attempts:   attempt,
					LastError:  err.Error(),
				})
				return fmt.Errorf("commit: %w", err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "    committed: %s\n", step.Commit)
			emit(ctx, cfg.EventStore, stream, EventCommitSucceeded, CommitSucceededData{
				StepNumber: step.Number,
				Message:    step.Commit,
			})
		}

		emit(ctx, cfg.EventStore, stream, EventStepCompleted, StepCompletedData{
			StepNumber: step.Number,
			StepTitle:  step.Title,
		})
		return nil
	}

	lastErr := fmt.Sprintf("failed after %d attempts. Last error: %s", cfg.MaxRetries, feedback)
	emit(ctx, cfg.EventStore, stream, EventStepFailed, StepFailedData{
		StepNumber: step.Number,
		StepTitle:  step.Title,
		Attempts:   cfg.MaxRetries,
		LastError:  feedback,
	})
	return fmt.Errorf("%s", lastErr)
}

// emit appends a single event to the store. Errors are silently dropped so
// that audit trail failures never abort orchestration.
func emit(ctx context.Context, store fact.EventStore, stream, eventType string, data any) {
	if store == nil {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	_ = store.Append(ctx, stream, []fact.Event{{
		Type: eventType,
		Data: json.RawMessage(b),
	}})
}
