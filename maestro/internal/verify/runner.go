package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	task "github.com/benaskins/axon-task"
)

const (
	taskType     = "verify"
	pollInterval = 50 * time.Millisecond
	pollTimeout  = 10 * time.Minute
)

// TaskResult holds the outcome of an async verification task.
type TaskResult struct {
	TaskID string
	Output string
	Err    error
}

// AsyncRunner submits verification commands as axon-task tasks and collects
// their results. Multiple verifications can be in flight concurrently.
type AsyncRunner struct {
	executor *task.Executor
}

// NewAsyncRunner creates an AsyncRunner backed by an in-memory task store.
func NewAsyncRunner() *AsyncRunner {
	store := newMemStore()
	executor := task.NewExecutor("", "", "", store)
	executor.RegisterWorker(taskType, &ShellWorker{})
	return &AsyncRunner{executor: executor}
}

// Submit enqueues an async verification task and returns the task ID.
func (r *AsyncRunner) Submit(projectDir, command string) (string, error) {
	params := Params{ProjectDir: projectDir, Command: command}
	raw, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("marshal params: %w", err)
	}

	t, err := r.executor.SubmitTask(taskType, command, "", "", "", raw)
	if err != nil {
		return "", fmt.Errorf("submit task: %w", err)
	}
	return t.ID, nil
}

// Wait blocks until the task with the given ID reaches a terminal state,
// then returns the result. Returns an error if the task is not found or
// times out.
func (r *AsyncRunner) Wait(taskID string) TaskResult {
	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		t, ok := r.executor.Get(taskID)
		if !ok {
			return TaskResult{TaskID: taskID, Err: fmt.Errorf("task %s not found", taskID)}
		}
		switch t.Status {
		case task.StatusCompleted:
			return TaskResult{TaskID: taskID, Output: t.Summary}
		case task.StatusFailed:
			return TaskResult{TaskID: taskID, Output: t.Summary, Err: fmt.Errorf("%s", t.Error)}
		}
		time.Sleep(pollInterval)
	}
	return TaskResult{TaskID: taskID, Err: fmt.Errorf("timed out waiting for task %s", taskID)}
}

// RunAll submits all (projectDir, command) pairs concurrently and waits for
// all results. The returned slice preserves the order of the inputs.
func (r *AsyncRunner) RunAll(tasks []RunRequest) []TaskResult {
	ids := make([]string, len(tasks))
	for i, req := range tasks {
		id, err := r.Submit(req.ProjectDir, req.Command)
		if err != nil {
			ids[i] = ""
			// Placeholder; collect inline below.
			_ = err
		} else {
			ids[i] = id
		}
	}

	results := make([]TaskResult, len(tasks))
	var wg sync.WaitGroup
	for i, id := range ids {
		if id == "" {
			results[i] = TaskResult{Err: fmt.Errorf("failed to submit task for %q", tasks[i].Command)}
			continue
		}
		wg.Add(1)
		go func(idx int, taskID string) {
			defer wg.Done()
			results[idx] = r.Wait(taskID)
		}(i, id)
	}
	wg.Wait()
	return results
}

// Shutdown stops the underlying executor, failing any queued tasks.
func (r *AsyncRunner) Shutdown() {
	r.executor.Shutdown()
}

// RunRequest holds the parameters for a single verification run.
type RunRequest struct {
	ProjectDir string
	Command    string
}

// memStore is a simple in-memory implementation of task.Store.
type memStore struct {
	mu    sync.Mutex
	items map[string]*task.Task
}

func newMemStore() *memStore {
	return &memStore{items: make(map[string]*task.Task)}
}

func (s *memStore) Save(_ context.Context, t *task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	copy := *t
	s.items[t.ID] = &copy
	return nil
}

func (s *memStore) Get(_ context.Context, id string) (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[id]
	if !ok {
		return nil, nil
	}
	copy := *t
	return &copy, nil
}

func (s *memStore) ListByAgent(_ context.Context, _ string, _, _ int) ([]task.Task, error) {
	return nil, nil
}
