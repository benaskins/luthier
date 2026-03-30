package orchestrate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
