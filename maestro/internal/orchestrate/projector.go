package orchestrate

import (
	"context"
	"encoding/json"
	"sort"
	"sync"

	fact "github.com/benaskins/axon-fact"
)

// StepStatus describes the projected state of a plan step.
type StepStatus string

const (
	StatusPending   StepStatus = "pending"
	StatusRunning   StepStatus = "running"
	StatusCompleted StepStatus = "completed"
	StatusFailed    StepStatus = "failed"
	StatusSkipped   StepStatus = "skipped"
)

// StepState is the projected state for a single plan step.
type StepState struct {
	Number   int
	Title    string
	Status   StepStatus
	Attempts int
}

// OrchestrationState reconstructs orchestration progress from the event
// stream. It implements fact.Projector and is safe for concurrent use.
type OrchestrationState struct {
	mu    sync.Mutex
	steps map[int]*StepState
}

// NewOrchestrationState creates an empty OrchestrationState.
func NewOrchestrationState() *OrchestrationState {
	return &OrchestrationState{
		steps: make(map[int]*StepState),
	}
}

// Handle processes an orchestration event and updates the read model.
func (s *OrchestrationState) Handle(_ context.Context, e fact.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch e.Type {
	case EventStepStarted:
		var d StepStartedData
		if err := json.Unmarshal(e.Data, &d); err != nil {
			return err
		}
		step := s.getOrCreate(d.StepNumber)
		step.Title = d.StepTitle
		step.Status = StatusRunning
		step.Attempts = 1

	case EventRetryAttempt:
		var d RetryAttemptData
		if err := json.Unmarshal(e.Data, &d); err != nil {
			return err
		}
		step := s.getOrCreate(d.StepNumber)
		step.Attempts = d.Attempt

	case EventStepCompleted:
		var d StepCompletedData
		if err := json.Unmarshal(e.Data, &d); err != nil {
			return err
		}
		step := s.getOrCreate(d.StepNumber)
		step.Title = d.StepTitle
		step.Status = StatusCompleted

	case EventStepFailed:
		var d StepFailedData
		if err := json.Unmarshal(e.Data, &d); err != nil {
			return err
		}
		step := s.getOrCreate(d.StepNumber)
		step.Title = d.StepTitle
		step.Status = StatusFailed
		step.Attempts = d.Attempts
	}

	return nil
}

// Steps returns a snapshot of all tracked step states sorted by step number.
func (s *OrchestrationState) Steps() []StepState {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]StepState, 0, len(s.steps))
	for _, step := range s.steps {
		out = append(out, *step)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Number < out[j].Number
	})
	return out
}

func (s *OrchestrationState) getOrCreate(n int) *StepState {
	if step, ok := s.steps[n]; ok {
		return step
	}
	step := &StepState{Number: n, Status: StatusPending}
	s.steps[n] = step
	return step
}
