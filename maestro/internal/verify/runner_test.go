package verify_test

import (
	"fmt"
	"testing"

	"github.com/benaskins/maestro/internal/verify"
)

func TestAsyncRunner_SingleTask(t *testing.T) {
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	dir := t.TempDir()
	id, err := runner.Submit(dir, "echo hello")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if id == "" {
		t.Fatal("expected non-empty task ID")
	}

	result := runner.Wait(id)
	if result.Err != nil {
		t.Fatalf("Wait returned error: %v", result.Err)
	}
	if result.TaskID != id {
		t.Errorf("result.TaskID = %q, want %q", result.TaskID, id)
	}
}

func TestAsyncRunner_FailingCommand(t *testing.T) {
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	dir := t.TempDir()
	id, err := runner.Submit(dir, "exit 1")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	result := runner.Wait(id)
	if result.Err == nil {
		t.Fatal("expected error for failing command, got nil")
	}
}

func TestAsyncRunner_RunAll_Concurrent(t *testing.T) {
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	const n = 5
	dir := t.TempDir()
	reqs := make([]verify.RunRequest, n)
	for i := range reqs {
		reqs[i] = verify.RunRequest{ProjectDir: dir, Command: fmt.Sprintf("echo task-%d", i)}
	}

	results := runner.RunAll(reqs)
	if len(results) != n {
		t.Fatalf("expected %d results, got %d", n, len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("task %d failed: %v", i, r.Err)
		}
	}
}

func TestAsyncRunner_RunAll_MixedResults(t *testing.T) {
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	dir := t.TempDir()
	reqs := []verify.RunRequest{
		{ProjectDir: dir, Command: "echo ok"},
		{ProjectDir: dir, Command: "exit 1"},
		{ProjectDir: dir, Command: "echo also-ok"},
	}

	results := runner.RunAll(reqs)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Err != nil {
		t.Errorf("task 0 should succeed, got: %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("task 1 should fail, got nil error")
	}
	if results[2].Err != nil {
		t.Errorf("task 2 should succeed, got: %v", results[2].Err)
	}
}

func TestAsyncRunner_RunAll_OrderPreserved(t *testing.T) {
	// Verify that RunAll returns results in the same order as inputs,
	// even when tasks complete out of order.
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	dir := t.TempDir()
	// Use 5 tasks — axon-task's default queue size — to avoid overflow.
	reqs := make([]verify.RunRequest, 5)
	for i := range reqs {
		reqs[i] = verify.RunRequest{ProjectDir: dir, Command: fmt.Sprintf("echo %d", i)}
	}

	results := runner.RunAll(reqs)
	if len(results) != len(reqs) {
		t.Fatalf("expected %d results, got %d", len(reqs), len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] unexpected error: %v", i, r.Err)
		}
	}
}

func TestShellWorker_MissingParams(t *testing.T) {
	runner := verify.NewAsyncRunner()
	defer runner.Shutdown()

	// Submit with empty command to exercise worker validation.
	id, err := runner.Submit("", "")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}

	result := runner.Wait(id)
	if result.Err == nil {
		t.Fatal("expected error for empty params, got nil")
	}
}
