package analysis

import (
	"context"
	"testing"

	talk "github.com/benaskins/axon-talk"
)

// sequenceClient returns a sequence of responses, one per Chat() call.
type sequenceClient struct {
	responses []talk.Response
	call      int
}

func (s *sequenceClient) Chat(_ context.Context, _ *talk.Request, fn func(talk.Response) error) error {
	if s.call >= len(s.responses) {
		// No more tool calls — model is done
		return fn(talk.Response{Content: "Done.", Done: true})
	}
	resp := s.responses[s.call]
	s.call++
	return fn(resp)
}

func TestAnalyse_ToolBasedFlow(t *testing.T) {
	mc := &sequenceClient{
		responses: []talk.Response{
			// Turn 1: model selects two modules
			{ToolCalls: []talk.ToolCall{
				{ID: "1", Name: "select_module", Arguments: map[string]any{
					"name": "axon", "reason": "HTTP server", "is_deterministic": true,
				}},
				{ID: "2", Name: "select_module", Arguments: map[string]any{
					"name": "axon-loop", "reason": "LLM conversation", "is_deterministic": false,
				}},
			}},
			// Turn 2: boundary + plan step
			{ToolCalls: []talk.ToolCall{
				{ID: "3", Name: "define_boundary", Arguments: map[string]any{
					"from": "handler", "to": "llm", "type": "non-det",
				}},
				{ID: "4", Name: "add_plan_step", Arguments: map[string]any{
					"title": "Scaffold repo", "description": "Create initial structure.", "commit_message": "feat: scaffold my-service",
				}},
			}},
			// Turn 3: finalize
			{ToolCalls: []talk.ToolCall{
				{ID: "5", Name: "finalize", Arguments: map[string]any{
					"name": "my-service",
				}},
			}},
		},
	}

	spec, err := Analyse(context.Background(), "Build a chat service.", mc, "claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	if spec.Name != "my-service" {
		t.Errorf("Name = %q, want %q", spec.Name, "my-service")
	}
	if len(spec.Modules) != 2 {
		t.Fatalf("len(Modules) = %d, want 2", len(spec.Modules))
	}
	if spec.Modules[0].Name != "axon" {
		t.Errorf("Modules[0].Name = %q, want %q", spec.Modules[0].Name, "axon")
	}
	if !spec.Modules[0].IsDeterministic {
		t.Error("Modules[0].IsDeterministic = false, want true")
	}
	if spec.Modules[1].IsDeterministic {
		t.Error("Modules[1].IsDeterministic = true, want false")
	}
	if len(spec.Boundaries) != 1 || spec.Boundaries[0].Type != "non-det" {
		t.Errorf("Boundaries = %+v, want one non-det boundary", spec.Boundaries)
	}
	if len(spec.PlanSteps) != 1 {
		t.Errorf("len(PlanSteps) = %d, want 1", len(spec.PlanSteps))
	}
	if len(spec.Gaps) != 0 {
		t.Errorf("len(Gaps) = %d, want 0", len(spec.Gaps))
	}
}

func TestAnalyse_WithGaps(t *testing.T) {
	mc := &sequenceClient{
		responses: []talk.Response{
			{ToolCalls: []talk.ToolCall{
				{ID: "1", Name: "select_module", Arguments: map[string]any{
					"name": "axon", "reason": "server", "is_deterministic": true,
				}},
				{ID: "2", Name: "raise_gap", Arguments: map[string]any{
					"question": "Which provider?", "context": "PRD is ambiguous",
				}},
			}},
			{ToolCalls: []talk.ToolCall{
				{ID: "3", Name: "finalize", Arguments: map[string]any{
					"name": "my-app",
				}},
			}},
		},
	}

	spec, err := Analyse(context.Background(), "Build something.", mc, "claude-sonnet-4-6")
	if err != nil {
		t.Fatalf("Analyse: %v", err)
	}

	if len(spec.Gaps) != 1 {
		t.Fatalf("len(Gaps) = %d, want 1", len(spec.Gaps))
	}
	if spec.Gaps[0].Question != "Which provider?" {
		t.Errorf("Gap question = %q", spec.Gaps[0].Question)
	}
}
