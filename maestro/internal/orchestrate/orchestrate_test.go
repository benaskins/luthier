package orchestrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	fact "github.com/benaskins/axon-fact"
	"github.com/benaskins/maestro/internal/plan"
	"github.com/benaskins/maestro/internal/review"
)

// --- test doubles ---

// countdownAgent succeeds after failCount failures, recording the last feedback received.
type countdownAgent struct {
	failCount    int
	calls        int
	lastFeedback string
}

func (a *countdownAgent) Implement(_ string, _ plan.Step, feedback string) (string, error) {
	a.calls++
	a.lastFeedback = feedback
	if a.calls <= a.failCount {
		return "agent output on failure", fmt.Errorf("agent error on attempt %d", a.calls)
	}
	return "agent succeeded", nil
}

// stubReviewer returns pre-configured results in sequence.
type stubReviewer struct {
	results []*review.Result
	errors  []error
	calls   int
}

func (r *stubReviewer) Review(_ context.Context, _ string, _ plan.Step) (*review.Result, error) {
	i := r.calls
	r.calls++
	if i < len(r.errors) && r.errors[i] != nil {
		return nil, r.errors[i]
	}
	if i < len(r.results) {
		return r.results[i], nil
	}
	return &review.Result{Passed: true, Reason: "ok"}, nil
}

// --- helpers ---

func runOsExec(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w\n%s", name, args, err, out)
	}
	return nil
}

// newTestProject creates a temp dir with a plans/ directory, a plan file, and
// a git repository with an initial commit.
func newTestProject(t *testing.T, planContent string) string {
	t.Helper()
	dir := t.TempDir()

	planDir := filepath.Join(dir, "plans")
	if err := os.MkdirAll(planDir, 0755); err != nil {
		t.Fatalf("mkdir plans: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "2026-01-01-plan.md"), []byte(planContent), 0644); err != nil {
		t.Fatalf("write plan: %v", err)
	}

	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
	} {
		if err := runOsExec(dir, "git", args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}

	placeholder := filepath.Join(dir, ".gitkeep")
	if err := os.WriteFile(placeholder, []byte(""), 0644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}
	if err := runOsExec(dir, "git", "add", "."); err != nil {
		t.Fatalf("git add: %v", err)
	}
	if err := runOsExec(dir, "git", "commit", "-m", "feat: initial commit"); err != nil {
		t.Fatalf("git initial commit: %v", err)
	}

	return dir
}

// --- tests ---

const singleStepPlan = `## Step 1 — add greeting

Create a simple greeting.

Commit: ` + "`feat: add greeting`"

func TestRun_SingleStepSucceeds(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{failCount: 0},
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if result.Failed != 0 {
		t.Errorf("Failed = %d, want 0", result.Failed)
	}
}

func TestRun_AgentFailsThenSucceeds(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	a := &countdownAgent{failCount: 2} // fails on attempts 1 and 2, succeeds on 3
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if a.calls != 3 {
		t.Errorf("agent called %d times, want 3", a.calls)
	}
}

func TestRun_FeedbackPassedOnRetry(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	a := &countdownAgent{failCount: 1} // fails once, then succeeds
	if _, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		MaxRetries: 3,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// On the second call (retry), feedback should mention the first attempt's error.
	if !strings.Contains(a.lastFeedback, "Agent error") {
		t.Errorf("feedback on retry = %q, want it to contain 'Agent error'", a.lastFeedback)
	}
}

func TestRun_FailsAfterMaxRetries(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{failCount: 99},
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err == nil {
		t.Fatal("Run: expected error, got nil")
	}
	if result.Failed != 1 {
		t.Errorf("Failed = %d, want 1", result.Failed)
	}
	if !strings.Contains(err.Error(), "failed after 3 attempts") {
		t.Errorf("error = %q, want 'failed after 3 attempts'", err.Error())
	}
}

func TestRun_VerificationFailureTriggerRetry(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	// false exits with non-zero; true exits with zero.
	// Use a counter to alternate: fail first then pass.
	callCount := 0
	// We can't easily intercept verify.Run, so instead we use a shell script that
	// counts calls via a temp file.
	countFile := filepath.Join(dir, ".verify_count")
	verifyScript := fmt.Sprintf(
		`COUNT=$(cat %s 2>/dev/null || echo 0); echo $((COUNT+1)) > %s; [ $((COUNT+1)) -gt 1 ]`,
		countFile, countFile,
	)
	_ = callCount

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  fmt.Sprintf("sh -c '%s'", verifyScript),
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	// Agent called twice: once for the attempt that failed verify, once for success.
	if a.calls != 2 {
		t.Errorf("agent called %d times, want 2", a.calls)
	}
}

func TestRun_ReviewFailureTriggerRetry(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	reviewer := &stubReviewer{
		results: []*review.Result{
			{Passed: false, Reason: "missing tests"},
			{Passed: true, Reason: "looks good"},
		},
	}

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		Reviewer:   reviewer,
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if a.calls != 2 {
		t.Errorf("agent called %d times, want 2 (one retry after review failure)", a.calls)
	}
	if reviewer.calls != 2 {
		t.Errorf("reviewer called %d times, want 2", reviewer.calls)
	}
}

func TestRun_ReviewFeedbackAccumulated(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	reviewer := &stubReviewer{
		results: []*review.Result{
			{Passed: false, Reason: "missing error handling"},
			{Passed: true, Reason: "ok"},
		},
	}

	a := &countdownAgent{failCount: 0}
	if _, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		Reviewer:   reviewer,
		VerifyCmd:  "true",
		MaxRetries: 3,
	}); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Second agent call should receive the review failure reason as feedback.
	if !strings.Contains(a.lastFeedback, "missing error handling") {
		t.Errorf("feedback = %q, want it to contain review reason", a.lastFeedback)
	}
}

func TestRun_ReviewErrorIsNonFatal(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	reviewer := &stubReviewer{
		errors: []error{errors.New("LLM unavailable")},
	}

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{failCount: 0},
		Reviewer:   reviewer,
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error when reviewer errors: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1 (review error should be non-fatal)", result.Completed)
	}
}

func TestRun_SkipsAlreadyCommittedStep(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	if err := runOsExec(dir, "git", "commit", "--allow-empty", "-m", "feat: add greeting"); err != nil {
		t.Fatalf("pre-commit step: %v", err)
	}

	a := &countdownAgent{}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if a.calls != 0 {
		t.Errorf("agent should not be called for already-committed step, got %d calls", a.calls)
	}
}

func TestRun_DryRunDoesNotCallAgent(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)

	a := &countdownAgent{}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		DryRun:     true,
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1 (dry-run counts as completed)", result.Completed)
	}
	if a.calls != 0 {
		t.Errorf("agent should not be called in dry-run mode, got %d calls", a.calls)
	}
}

func gitCommitCount(t *testing.T, dir string) int {
	t.Helper()
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-list: %v", err)
	}
	count, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return count
}

func TestRun_DryRunMakesNoGitCommit(t *testing.T) {
	planContent := `## Step 1 — first step

First thing.

Commit: ` + "`feat: first step`" + `

## Step 2 — second step

Second thing.

Commit: ` + "`feat: second step`"

	dir := newTestProject(t, planContent)
	initialCount := gitCommitCount(t, dir)

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{},
		VerifyCmd:  "true",
		DryRun:     true,
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Completed != 2 {
		t.Errorf("Completed = %d, want 2", result.Completed)
	}

	afterCount := gitCommitCount(t, dir)
	if afterCount != initialCount {
		t.Errorf("dry-run made %d new git commits, want 0", afterCount-initialCount)
	}
}

func TestRun_MultipleSteps(t *testing.T) {
	planContent := `## Step 1 — first step

First thing.

Commit: ` + "`feat: first step`" + `

## Step 2 — second step

Second thing.

Commit: ` + "`feat: second step`"

	dir := newTestProject(t, planContent)

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{failCount: 0},
		VerifyCmd:  "true",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("Total = %d, want 2", result.Total)
	}
	if result.Completed != 2 {
		t.Errorf("Completed = %d, want 2", result.Completed)
	}
}

// threeStepPlan is a three-step plan used by resume-from tests.
const threeStepPlan = `## Step 1 — first step

First thing.

Commit: ` + "`feat: first step`" + `

## Step 2 — second step

Second thing.

Commit: ` + "`feat: second step`" + `

## Step 3 — third step

Third thing.

Commit: ` + "`feat: third step`"

func TestRun_ResumeFromTitle(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		ResumeFrom: "second step",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if result.Completed != 2 {
		t.Errorf("Completed = %d, want 2", result.Completed)
	}
	// Agent should only be called for steps 2 and 3.
	if a.calls != 2 {
		t.Errorf("agent called %d times, want 2", a.calls)
	}
}

func TestRun_ResumeFromNumber(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		ResumeFrom: "2",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", result.Skipped)
	}
	if result.Completed != 2 {
		t.Errorf("Completed = %d, want 2", result.Completed)
	}
	if a.calls != 2 {
		t.Errorf("agent called %d times, want 2", a.calls)
	}
}

func TestRun_ResumeFromFirstStep(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		ResumeFrom: "1",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if result.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0", result.Skipped)
	}
	if result.Completed != 3 {
		t.Errorf("Completed = %d, want 3", result.Completed)
	}
	if a.calls != 3 {
		t.Errorf("agent called %d times, want 3", a.calls)
	}
}

func TestRun_ResumeFromInvalidValue(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	_, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{},
		VerifyCmd:  "true",
		ResumeFrom: "nonexistent step",
		MaxRetries: 3,
	})
	if err == nil {
		t.Fatal("Run: expected error for unknown resume-from value, got nil")
	}
	if !strings.Contains(err.Error(), "--resume-from") {
		t.Errorf("error = %q, want it to mention --resume-from", err.Error())
	}
}

func TestRun_ResumeFromAfterPartialRun(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	// Simulate steps 1 and 2 already committed (partial prior run).
	if err := runOsExec(dir, "git", "commit", "--allow-empty", "-m", "feat: first step"); err != nil {
		t.Fatalf("pre-commit step 1: %v", err)
	}
	if err := runOsExec(dir, "git", "commit", "--allow-empty", "-m", "feat: second step"); err != nil {
		t.Fatalf("pre-commit step 2: %v", err)
	}

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		ResumeFrom: "3",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	// Steps 1 and 2 are skipped (already committed); step 3 is executed.
	if result.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", result.Skipped)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if a.calls != 1 {
		t.Errorf("agent called %d times, want 1 (only step 3)", a.calls)
	}
}

func TestRun_ResumeFromSkippedStepsNotExecuted(t *testing.T) {
	dir := newTestProject(t, threeStepPlan)

	a := &countdownAgent{failCount: 0}
	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      a,
		VerifyCmd:  "true",
		ResumeFrom: "third step",
		MaxRetries: 3,
	})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	// Steps 1 and 2 skipped; only step 3 executed.
	if result.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", result.Skipped)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
	if a.calls != 1 {
		t.Errorf("agent called %d times, want 1 (only step 3)", a.calls)
	}
}

// --- event audit trail tests ---

// eventTypesFrom loads events from the store and returns their types in order.
func eventTypesFrom(t *testing.T, store fact.EventStore, stream string) []string {
	t.Helper()
	events, err := store.Load(context.Background(), stream)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	types := make([]string, len(events))
	for i, e := range events {
		types[i] = e.Type
	}
	return types
}

// findEvent returns the first event of the given type, or fails the test.
func findEvent(t *testing.T, store fact.EventStore, stream, eventType string) fact.Event {
	t.Helper()
	events, err := store.Load(context.Background(), stream)
	if err != nil {
		t.Fatalf("load events: %v", err)
	}
	for _, e := range events {
		if e.Type == eventType {
			return e
		}
	}
	t.Fatalf("event %q not found in stream %q", eventType, stream)
	return fact.Event{}
}

func TestRun_EventsRecordedOnSuccess(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	store := fact.NewMemoryStore()

	_, err := Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{failCount: 0},
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	types := eventTypesFrom(t, store, "test-run")

	// Must contain the lifecycle events. The noop agent creates no files so
	// the commit step emits CommitSkipped rather than CommitSucceeded.
	wantContains := []string{
		EventStepStarted,
		EventAgentInvoked,
		EventAgentSucceeded,
		EventVerificationRun,
		EventVerificationPassed,
		EventCommitSkipped,
		EventStepCompleted,
	}
	for _, want := range wantContains {
		found := false
		for _, got := range types {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("event %q not found; got events: %v", want, types)
		}
	}
}

func TestRun_EventsRecordedOnAgentRetry(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	store := fact.NewMemoryStore()

	_, err := Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{failCount: 1}, // fails once then succeeds
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	types := eventTypesFrom(t, store, "test-run")

	// Expect an AgentFailed followed by a RetryAttempt.
	mustContain := []string{EventAgentFailed, EventRetryAttempt}
	for _, want := range mustContain {
		found := false
		for _, got := range types {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("event %q not found after retry; got: %v", want, types)
		}
	}
}

func TestRun_EventsRecordedOnStepFailed(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	store := fact.NewMemoryStore()

	_, _ = Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{failCount: 99},
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})

	types := eventTypesFrom(t, store, "test-run")

	found := false
	for _, got := range types {
		if got == EventStepFailed {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("EventStepFailed not recorded; got: %v", types)
	}
}

func TestRun_StepStartedDataIsCorrect(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	store := fact.NewMemoryStore()

	_, err := Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{},
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	e := findEvent(t, store, "test-run", EventStepStarted)
	var d StepStartedData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		t.Fatalf("unmarshal StepStartedData: %v", err)
	}
	if d.StepNumber != 1 {
		t.Errorf("StepNumber = %d, want 1", d.StepNumber)
	}
	if d.StepTitle != "add greeting" {
		t.Errorf("StepTitle = %q, want %q", d.StepTitle, "add greeting")
	}
}

func TestRun_NoEventsWithoutStore(t *testing.T) {
	// Ensure orchestration works correctly when no EventStore is configured.
	dir := newTestProject(t, singleStepPlan)

	result, err := Run(Config{
		ProjectDir: dir,
		Agent:      &countdownAgent{},
		VerifyCmd:  "true",
		MaxRetries: 3,
		// EventStore intentionally omitted
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Completed != 1 {
		t.Errorf("Completed = %d, want 1", result.Completed)
	}
}

func TestRun_ProjectorReconstructsState(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	state := NewOrchestrationState()
	store := fact.NewMemoryStore(fact.WithProjector(state))

	_, err := Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{},
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	steps := state.Steps()
	if len(steps) != 1 {
		t.Fatalf("projected %d steps, want 1", len(steps))
	}
	if steps[0].Status != StatusCompleted {
		t.Errorf("step status = %q, want %q", steps[0].Status, StatusCompleted)
	}
	if steps[0].Title != "add greeting" {
		t.Errorf("step title = %q, want %q", steps[0].Title, "add greeting")
	}
}

func TestRun_ProjectorReconstructsFailedState(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	state := NewOrchestrationState()
	store := fact.NewMemoryStore(fact.WithProjector(state))

	_, _ = Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{failCount: 99},
		VerifyCmd:   "true",
		MaxRetries:  2,
		EventStore:  store,
		EventStream: "test-run",
	})

	steps := state.Steps()
	if len(steps) != 1 {
		t.Fatalf("projected %d steps, want 1", len(steps))
	}
	if steps[0].Status != StatusFailed {
		t.Errorf("step status = %q, want %q", steps[0].Status, StatusFailed)
	}
	if steps[0].Attempts != 2 {
		t.Errorf("step attempts = %d, want 2", steps[0].Attempts)
	}
}

func TestRun_ReviewEventsRecorded(t *testing.T) {
	dir := newTestProject(t, singleStepPlan)
	store := fact.NewMemoryStore()

	reviewer := &stubReviewer{
		results: []*review.Result{
			{Passed: false, Reason: "needs more tests"},
			{Passed: true, Reason: "looks good"},
		},
	}

	_, err := Run(Config{
		ProjectDir:  dir,
		Agent:       &countdownAgent{},
		Reviewer:    reviewer,
		VerifyCmd:   "true",
		MaxRetries:  3,
		EventStore:  store,
		EventStream: "test-run",
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	types := eventTypesFrom(t, store, "test-run")

	for _, want := range []string{EventReviewRun, EventReviewFailed, EventReviewPassed} {
		found := false
		for _, got := range types {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("event %q not found; got: %v", want, types)
		}
	}
}
