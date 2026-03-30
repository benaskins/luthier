// Package orchestrate runs the maestro plan execution loop.
package orchestrate

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	fact "github.com/benaskins/axon-fact"
	"github.com/benaskins/maestro/internal/agent"
	gitpkg "github.com/benaskins/maestro/internal/git"
	"github.com/benaskins/maestro/internal/logger"
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
	// Logger controls progress output. Defaults to an Info-level logger on
	// os.Stderr when nil. Use logger.New(logger.LevelSilent, nil) to suppress.
	Logger *logger.Logger
}

// StepResult records what happened during a single step's execution.
type StepResult struct {
	Number   int
	Title    string
	Status   StepStatus
	Attempts int
	Duration time.Duration
}

// Result summarises a completed orchestration run.
type Result struct {
	Total     int
	Completed int
	Skipped   int
	Failed    int
	FailedAt  *plan.Step
	Steps     []StepResult
	Duration  time.Duration
}

// Run executes the plan steps in order.
func Run(cfg Config) (*Result, error) {
	log := cfg.Logger
	if log == nil {
		log = logger.Default()
	}

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
	runStart := time.Now()

	for i := range steps {
		step := &steps[i]

		// Skip already-committed steps
		if gitpkg.IsStepCommitted(cfg.ProjectDir, step.Commit) {
			result.Skipped++
			result.Steps = append(result.Steps, StepResult{
				Number: step.Number,
				Title:  step.Title,
				Status: StatusSkipped,
			})
			log.Info("  [%d/%d] %s (already committed, skipping)\n", step.Number, result.Total, step.Title)
			continue
		}

		// Handle resume-from
		if resuming {
			if step.Title != cfg.ResumeFrom && fmt.Sprintf("%d", step.Number) != cfg.ResumeFrom {
				result.Skipped++
				result.Steps = append(result.Steps, StepResult{
					Number: step.Number,
					Title:  step.Title,
					Status: StatusSkipped,
				})
				log.Info("  [%d/%d] %s (skipping, resuming from %s)\n", step.Number, result.Total, step.Title, cfg.ResumeFrom)
				continue
			}
			resuming = false
		}

		log.Info("\n  [%d/%d] %s\n", step.Number, result.Total, step.Title)

		if cfg.DryRun {
			log.Info("    dry-run: would delegate to coding agent\n")
			log.Info("    dry-run: would run: %s\n", verifyCmd)
			log.Info("    dry-run: would commit: %s\n", step.Commit)
			result.Completed++
			result.Steps = append(result.Steps, StepResult{
				Number:   step.Number,
				Title:    step.Title,
				Status:   StatusCompleted,
				Attempts: 1,
			})
			continue
		}

		stepStart := time.Now()
		attempts, err := executeStep(cfg, log, *step, verifyCmd)
		stepDur := time.Since(stepStart)

		if err != nil {
			result.Failed++
			result.FailedAt = step
			result.Steps = append(result.Steps, StepResult{
				Number:   step.Number,
				Title:    step.Title,
				Status:   StatusFailed,
				Attempts: attempts,
				Duration: stepDur,
			})
			result.Duration = time.Since(runStart)
			return result, fmt.Errorf("step %d (%s): %w", step.Number, step.Title, err)
		}

		result.Completed++
		result.Steps = append(result.Steps, StepResult{
			Number:   step.Number,
			Title:    step.Title,
			Status:   StatusCompleted,
			Attempts: attempts,
			Duration: stepDur,
		})
	}

	result.Duration = time.Since(runStart)
	return result, nil
}

func executeStep(cfg Config, log *logger.Logger, step plan.Step, verifyCmd string) (int, error) {
	ctx := context.Background()
	stream := cfg.EventStream

	emit(ctx, cfg.EventStore, stream, EventStepStarted, StepStartedData{
		StepNumber: step.Number,
		StepTitle:  step.Title,
	})

	var feedback string

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 1 {
			log.Info("    retry %d/%d\n", attempt, cfg.MaxRetries)
			emit(ctx, cfg.EventStore, stream, EventRetryAttempt, RetryAttemptData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
		}

		// Delegate to coding agent
		log.Info("    delegating to coding agent...\n")
		emit(ctx, cfg.EventStore, stream, EventAgentInvoked, AgentInvokedData{
			StepNumber:  step.Number,
			Attempt:     attempt,
			HasFeedback: feedback != "",
		})
		agentOut, err := cfg.Agent.Implement(cfg.ProjectDir, step, feedback)
		if err != nil {
			feedback = fmt.Sprintf("Agent error: %s\nOutput: %s", err, agentOut)
			log.Warn("agent failed: %v\n", err)
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
		log.Debug("    verifying: %s\n", verifyCmd)
		emit(ctx, cfg.EventStore, stream, EventVerificationRun, VerificationRunData{
			StepNumber: step.Number,
			Attempt:    attempt,
			Command:    verifyCmd,
		})
		verifyOut, err := verify.Run(cfg.ProjectDir, verifyCmd)
		if err != nil {
			feedback = fmt.Sprintf("Verification failed:\n%s", verifyOut)
			log.Warn("verification failed: %v\n", err)
			emit(ctx, cfg.EventStore, stream, EventVerificationFailed, VerificationFailedData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
			continue
		}
		log.Info("    verification passed\n")
		emit(ctx, cfg.EventStore, stream, EventVerificationPassed, VerificationPassedData{
			StepNumber: step.Number,
			Attempt:    attempt,
		})

		// Run semantic review if a reviewer is configured
		if cfg.Reviewer != nil {
			diff, diffErr := gitpkg.Diff(cfg.ProjectDir)
			if diffErr != nil {
				feedback = fmt.Sprintf("Failed to get diff for review: %s", diffErr)
				log.Warn("could not get diff: %v\n", diffErr)
				continue
			}

			emit(ctx, cfg.EventStore, stream, EventReviewRun, ReviewRunData{
				StepNumber: step.Number,
				Attempt:    attempt,
			})
			reviewResult, reviewErr := cfg.Reviewer.Review(ctx, diff, step)
			if reviewErr != nil {
				// Review errors are non-fatal: log and proceed to commit.
				log.Warn("review error (skipping): %v\n", reviewErr)
				emit(ctx, cfg.EventStore, stream, EventReviewErrored, ReviewErroredData{
					StepNumber: step.Number,
					Attempt:    attempt,
					Error:      reviewErr.Error(),
				})
			} else if !reviewResult.Passed {
				feedback = fmt.Sprintf("Semantic review failed: %s", reviewResult.Reason)
				log.Warn("review failed: %s\n", reviewResult.Reason)
				emit(ctx, cfg.EventStore, stream, EventReviewFailed, ReviewFailedData{
					StepNumber: step.Number,
					Attempt:    attempt,
					Reason:     reviewResult.Reason,
				})
				continue
			} else {
				log.Info("    review passed: %s\n", reviewResult.Reason)
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
				log.Info("    nothing to commit for: %s\n", step.Commit)
				emit(ctx, cfg.EventStore, stream, EventCommitSkipped, CommitSkippedData{
					StepNumber: step.Number,
					Message:    step.Commit,
				})
			} else {
				log.Error("commit failed: %v\n", err)
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
				return attempt, fmt.Errorf("commit: %w", err)
			}
		} else {
			log.Info("    committed: %s\n", step.Commit)
			emit(ctx, cfg.EventStore, stream, EventCommitSucceeded, CommitSucceededData{
				StepNumber: step.Number,
				Message:    step.Commit,
			})
		}

		emit(ctx, cfg.EventStore, stream, EventStepCompleted, StepCompletedData{
			StepNumber: step.Number,
			StepTitle:  step.Title,
		})
		return attempt, nil
	}

	lastErr := fmt.Sprintf("failed after %d attempts. Last error: %s", cfg.MaxRetries, feedback)
	emit(ctx, cfg.EventStore, stream, EventStepFailed, StepFailedData{
		StepNumber: step.Number,
		StepTitle:  step.Title,
		Attempts:   cfg.MaxRetries,
		LastError:  feedback,
	})
	return cfg.MaxRetries, fmt.Errorf("%s", lastErr)
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
