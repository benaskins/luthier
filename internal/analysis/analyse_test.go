package analysis

import (
	"context"
	"testing"

	talk "github.com/benaskins/axon-talk"
)

// mockClient satisfies talk.LLMClient. It calls fn once with a canned Response.
type mockClient struct {
	resp talk.Response
}

func (m *mockClient) Chat(_ context.Context, _ *talk.Request, fn func(talk.Response) error) error {
	return fn(m.resp)
}

func TestAnalyse_ParsesStructuredResponse(t *testing.T) {
	mc := &mockClient{
		resp: talk.Response{
			Done: true,
			ToolCalls: []talk.ToolCall{
				{
					Name: "structured_response",
					Arguments: map[string]any{
						"name": "my-service",
						"modules": []any{
							map[string]any{
								"name":             "axon",
								"reason":           "HTTP server",
								"is_deterministic": true,
							},
							map[string]any{
								"name":             "axon-loop",
								"reason":           "LLM conversation",
								"is_deterministic": false,
							},
						},
						"boundaries": []any{
							map[string]any{"from": "handler", "to": "llm", "type": "non-det"},
						},
						"files": []any{
							map[string]any{
								"path":     "cmd/my-service/main.go",
								"template": "main",
								"vars":     map[string]any{"Name": "my-service"},
							},
						},
						"plan_steps": []any{
							map[string]any{
								"title":          "Scaffold repo",
								"description":    "Create initial structure.",
								"commit_message": "feat: scaffold my-service",
							},
						},
						"gaps": []any{},
					},
				},
			},
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
		t.Errorf("len(Modules) = %d, want 2", len(spec.Modules))
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
	if len(spec.Files) != 1 {
		t.Errorf("len(Files) = %d, want 1", len(spec.Files))
	}
	if len(spec.PlanSteps) != 1 {
		t.Errorf("len(PlanSteps) = %d, want 1", len(spec.PlanSteps))
	}
	if len(spec.Gaps) != 0 {
		t.Errorf("len(Gaps) = %d, want 0", len(spec.Gaps))
	}
}

func TestAnalyse_NoToolCall(t *testing.T) {
	mc := &mockClient{resp: talk.Response{Done: true}}
	_, err := Analyse(context.Background(), "some prd", mc, "claude-sonnet-4-6")
	if err == nil {
		t.Error("expected error when no tool call in response")
	}
}
