package orchestrate

import (
	"context"
	"encoding/json"
	"testing"

	fact "github.com/benaskins/axon-fact"
)

// mustMarshal serialises v to JSON or panics — for test convenience only.
func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestProjector_InitialStateIsEmpty(t *testing.T) {
	state := NewOrchestrationState()
	steps := state.Steps()
	if len(steps) != 0 {
		t.Errorf("expected empty state, got %d steps", len(steps))
	}
}

func TestProjector_StepStarted(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 1, StepTitle: "init project"}),
	}
	if err := state.Handle(context.Background(), e); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	steps := state.Steps()
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Number != 1 {
		t.Errorf("Number = %d, want 1", steps[0].Number)
	}
	if steps[0].Title != "init project" {
		t.Errorf("Title = %q, want %q", steps[0].Title, "init project")
	}
	if steps[0].Status != StatusRunning {
		t.Errorf("Status = %q, want %q", steps[0].Status, StatusRunning)
	}
	if steps[0].Attempts != 1 {
		t.Errorf("Attempts = %d, want 1", steps[0].Attempts)
	}
}

func TestProjector_RetryAttemptUpdatesCount(t *testing.T) {
	state := NewOrchestrationState()

	state.Handle(context.Background(), fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 2, StepTitle: "add auth"}),
	})
	state.Handle(context.Background(), fact.Event{
		Type: EventRetryAttempt,
		Data: mustMarshal(RetryAttemptData{StepNumber: 2, Attempt: 2}),
	})

	steps := state.Steps()
	if steps[0].Attempts != 2 {
		t.Errorf("Attempts = %d after retry, want 2", steps[0].Attempts)
	}
}

func TestProjector_StepCompleted(t *testing.T) {
	state := NewOrchestrationState()

	state.Handle(context.Background(), fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 1, StepTitle: "setup"}),
	})
	state.Handle(context.Background(), fact.Event{
		Type: EventStepCompleted,
		Data: mustMarshal(StepCompletedData{StepNumber: 1, StepTitle: "setup"}),
	})

	steps := state.Steps()
	if steps[0].Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", steps[0].Status, StatusCompleted)
	}
}

func TestProjector_StepFailed(t *testing.T) {
	state := NewOrchestrationState()

	state.Handle(context.Background(), fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 3, StepTitle: "deploy"}),
	})
	state.Handle(context.Background(), fact.Event{
		Type: EventStepFailed,
		Data: mustMarshal(StepFailedData{StepNumber: 3, StepTitle: "deploy", Attempts: 3, LastError: "timeout"}),
	})

	steps := state.Steps()
	if steps[0].Status != StatusFailed {
		t.Errorf("Status = %q, want %q", steps[0].Status, StatusFailed)
	}
	if steps[0].Attempts != 3 {
		t.Errorf("Attempts = %d, want 3", steps[0].Attempts)
	}
}

func TestProjector_MultipleStepsSortedByNumber(t *testing.T) {
	state := NewOrchestrationState()

	// Emit events in reverse order.
	for _, n := range []int{3, 1, 2} {
		state.Handle(context.Background(), fact.Event{
			Type: EventStepStarted,
			Data: mustMarshal(StepStartedData{StepNumber: n, StepTitle: "step"}),
		})
		state.Handle(context.Background(), fact.Event{
			Type: EventStepCompleted,
			Data: mustMarshal(StepCompletedData{StepNumber: n, StepTitle: "step"}),
		})
	}

	steps := state.Steps()
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	for i, s := range steps {
		if s.Number != i+1 {
			t.Errorf("steps[%d].Number = %d, want %d", i, s.Number, i+1)
		}
	}
}

func TestProjector_UnknownEventTypeIgnored(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: "UnknownEventType",
		Data: []byte(`{}`),
	}
	if err := state.Handle(context.Background(), e); err != nil {
		t.Errorf("Handle with unknown event type returned error: %v", err)
	}
	if len(state.Steps()) != 0 {
		t.Error("unexpected steps from unknown event type")
	}
}

func TestProjector_MalformedStepStartedData(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: EventStepStarted,
		Data: []byte(`not valid json`),
	}
	if err := state.Handle(context.Background(), e); err == nil {
		t.Error("expected error for malformed StepStarted data, got nil")
	}
}

func TestProjector_MalformedRetryAttemptData(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: EventRetryAttempt,
		Data: []byte(`not valid json`),
	}
	if err := state.Handle(context.Background(), e); err == nil {
		t.Error("expected error for malformed RetryAttempt data, got nil")
	}
}

func TestProjector_MalformedStepCompletedData(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: EventStepCompleted,
		Data: []byte(`not valid json`),
	}
	if err := state.Handle(context.Background(), e); err == nil {
		t.Error("expected error for malformed StepCompleted data, got nil")
	}
}

func TestProjector_MalformedStepFailedData(t *testing.T) {
	state := NewOrchestrationState()
	e := fact.Event{
		Type: EventStepFailed,
		Data: []byte(`not valid json`),
	}
	if err := state.Handle(context.Background(), e); err == nil {
		t.Error("expected error for malformed StepFailed data, got nil")
	}
}

func TestProjector_GetOrCreateIdempotent(t *testing.T) {
	state := NewOrchestrationState()

	// Two StepStarted events for the same step number — second overwrites title/status.
	state.Handle(context.Background(), fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 1, StepTitle: "original title"}),
	})
	state.Handle(context.Background(), fact.Event{
		Type: EventStepStarted,
		Data: mustMarshal(StepStartedData{StepNumber: 1, StepTitle: "updated title"}),
	})

	steps := state.Steps()
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Title != "updated title" {
		t.Errorf("Title = %q, want 'updated title'", steps[0].Title)
	}
}
