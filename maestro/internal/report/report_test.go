package report

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/benaskins/maestro/internal/orchestrate"
	"github.com/benaskins/maestro/internal/plan"
)

func TestPrint_AllCompleted(t *testing.T) {
	result := &orchestrate.Result{
		Total:     2,
		Completed: 2,
		Duration:  5 * time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "add greeting", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 2 * time.Second},
			{Number: 2, Title: "add tests", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 3 * time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertContains(t, out, "add greeting")
	assertContains(t, out, "add tests")
	assertContains(t, out, "completed")
	assertContains(t, out, "Total:     2 step")
	assertContains(t, out, "Completed: 2 step")
	assertNotContains(t, out, "Failed")
	assertNotContains(t, out, "Skipped")
}

func TestPrint_WithFailure(t *testing.T) {
	failedStep := plan.Step{Number: 2, Title: "add tests"}
	result := &orchestrate.Result{
		Total:     2,
		Completed: 1,
		Failed:    1,
		FailedAt:  &failedStep,
		Duration:  10 * time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "add greeting", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 3 * time.Second},
			{Number: 2, Title: "add tests", Status: orchestrate.StatusFailed, Attempts: 3, Duration: 7 * time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertContains(t, out, "Failed:    1 step")
	assertContains(t, out, "step 2: add tests")
	assertContains(t, out, "!")
}

func TestPrint_WithSkipped(t *testing.T) {
	result := &orchestrate.Result{
		Total:     3,
		Completed: 2,
		Skipped:   1,
		Duration:  8 * time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "first", Status: orchestrate.StatusSkipped},
			{Number: 2, Title: "second", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 4 * time.Second},
			{Number: 3, Title: "third", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 4 * time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertContains(t, out, "Skipped:   1 step")
	assertContains(t, out, "-")
}

func TestPrint_WithRetries(t *testing.T) {
	result := &orchestrate.Result{
		Total:     1,
		Completed: 1,
		Duration:  15 * time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "add auth", Status: orchestrate.StatusCompleted, Attempts: 3, Duration: 15 * time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertContains(t, out, "3 attempts")
	assertContains(t, out, "Retries:")
	assertContains(t, out, "1 step")
	assertContains(t, out, "2 extra attempt")
}

func TestPrint_NilResult(t *testing.T) {
	var w strings.Builder
	Print(&w, nil) // must not panic
	if w.Len() != 0 {
		t.Errorf("expected empty output for nil result, got %q", w.String())
	}
}

func TestPrint_Duration(t *testing.T) {
	result := &orchestrate.Result{
		Total:     1,
		Completed: 1,
		Duration:  2*time.Minute + 30*time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "step", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 2*time.Minute + 30*time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertContains(t, out, "Time:")
	assertContains(t, out, "2m")
}

func TestPrint_NoRetryStatsWhenNoneNeeded(t *testing.T) {
	result := &orchestrate.Result{
		Total:     2,
		Completed: 2,
		Duration:  4 * time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "a", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 2 * time.Second},
			{Number: 2, Title: "b", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: 2 * time.Second},
		},
	}

	var w strings.Builder
	Print(&w, result)
	out := w.String()

	assertNotContains(t, out, "Retries:")
}

func TestWriteFile(t *testing.T) {
	result := &orchestrate.Result{
		Total:     1,
		Completed: 1,
		Duration:  time.Second,
		Steps: []orchestrate.StepResult{
			{Number: 1, Title: "step one", Status: orchestrate.StatusCompleted, Attempts: 1, Duration: time.Second},
		},
	}

	path := t.TempDir() + "/summary.txt"
	if err := WriteFile(path, result); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Re-parse to verify content was written
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "step one") {
		t.Errorf("summary file missing expected content; got: %s", out)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{90 * time.Second, "1m30s"},
		{120 * time.Second, "2m"},
		{3 * time.Minute, "3m"},
	}
	for _, tc := range cases {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected output to contain %q\ngot:\n%s", sub, s)
	}
}

func assertNotContains(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", sub, s)
	}
}
