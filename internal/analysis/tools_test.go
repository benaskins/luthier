package analysis

import "testing"

func TestSpecBuilderSelectModule(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	result := tools["select_module"].Execute(nil, map[string]any{
		"name":             "axon-loop",
		"reason":           "conversation loop",
		"is_deterministic": false,
	})

	if result.Content == "" {
		t.Error("expected non-empty result")
	}

	spec := b.Spec()
	if len(spec.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(spec.Modules))
	}
	if spec.Modules[0].Name != "axon-loop" {
		t.Errorf("got name %q, want axon-loop", spec.Modules[0].Name)
	}
	if spec.Modules[0].IsDeterministic {
		t.Error("expected non-deterministic")
	}
}

func TestSpecBuilderDefineBoundary(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	tools["define_boundary"].Execute(nil, map[string]any{
		"from": "CLI",
		"to":   "agent",
		"type": "non-det",
	})

	spec := b.Spec()
	if len(spec.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(spec.Boundaries))
	}
	if spec.Boundaries[0].Type != "non-det" {
		t.Errorf("got type %q, want non-det", spec.Boundaries[0].Type)
	}
}

func TestSpecBuilderAddPlanStep(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	tools["add_plan_step"].Execute(nil, map[string]any{
		"title":          "Scaffold repo",
		"description":    "Create initial structure",
		"commit_message": "feat: scaffold repo",
	})

	spec := b.Spec()
	if len(spec.PlanSteps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(spec.PlanSteps))
	}
	if spec.PlanSteps[0].CommitMessage != "feat: scaffold repo" {
		t.Errorf("got commit %q", spec.PlanSteps[0].CommitMessage)
	}
}

func TestSpecBuilderRaiseGap(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	tools["raise_gap"].Execute(nil, map[string]any{
		"question": "Which provider?",
		"context":  "PRD is ambiguous",
	})

	spec := b.Spec()
	if len(spec.Gaps) != 1 {
		t.Fatalf("expected 1 gap, got %d", len(spec.Gaps))
	}
}

func TestSpecBuilderFinalize(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	if b.Finalized() {
		t.Error("should not be finalized yet")
	}

	tools["finalize"].Execute(nil, map[string]any{
		"name": "my-service",
	})

	if !b.Finalized() {
		t.Error("should be finalized")
	}
	if b.Spec().Name != "my-service" {
		t.Errorf("got name %q, want my-service", b.Spec().Name)
	}
}

func TestSpecBuilderFullFlow(t *testing.T) {
	b := NewSpecBuilder()
	tools := b.Tools()

	tools["select_module"].Execute(nil, map[string]any{
		"name": "axon-loop", "reason": "conversation", "is_deterministic": false,
	})
	tools["select_module"].Execute(nil, map[string]any{
		"name": "axon-talk", "reason": "provider", "is_deterministic": false,
	})
	tools["define_boundary"].Execute(nil, map[string]any{
		"from": "CLI", "to": "agent", "type": "non-det",
	})
	tools["add_plan_step"].Execute(nil, map[string]any{
		"title": "Scaffold", "description": "Create repo", "commit_message": "feat: scaffold",
	})
	tools["raise_gap"].Execute(nil, map[string]any{
		"question": "Which provider?", "context": "Ambiguous",
	})
	tools["finalize"].Execute(nil, map[string]any{
		"name": "my-app",
	})

	spec := b.Spec()
	if len(spec.Modules) != 2 {
		t.Errorf("modules: got %d, want 2", len(spec.Modules))
	}
	if len(spec.Boundaries) != 1 {
		t.Errorf("boundaries: got %d, want 1", len(spec.Boundaries))
	}
	if len(spec.PlanSteps) != 1 {
		t.Errorf("plan steps: got %d, want 1", len(spec.PlanSteps))
	}
	if len(spec.Gaps) != 1 {
		t.Errorf("gaps: got %d, want 1", len(spec.Gaps))
	}
	if spec.Name != "my-app" {
		t.Errorf("name: got %q, want my-app", spec.Name)
	}
	if !b.Finalized() {
		t.Error("should be finalized")
	}
}
