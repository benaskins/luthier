package gaps

import (
	"context"
	"strings"
	"testing"

	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/luthier/internal/analysis"
)

type mockClient struct {
	resp talk.Response
}

func (m *mockClient) Chat(_ context.Context, _ *talk.Request, fn func(talk.Response) error) error {
	return fn(m.resp)
}

func TestResolve_NoGaps_ReturnsSpecUnchanged(t *testing.T) {
	spec := &analysis.ScaffoldSpec{Name: "my-service"}
	r := New("claude-sonnet-4-6").WithIO(strings.NewReader(""), &strings.Builder{})

	got, err := r.Resolve(context.Background(), spec, &mockClient{})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if got != spec {
		t.Error("expected same spec pointer returned when no gaps")
	}
}

func TestResolve_WithGaps_CollectsAnswersAndUpdatesSpec(t *testing.T) {
	spec := &analysis.ScaffoldSpec{
		Name: "my-service",
		Gaps: []analysis.Gap{
			{Question: "Will this need auth?", Context: "The PRD does not mention it."},
		},
	}

	// The mock client returns an updated spec with gaps cleared.
	mc := &mockClient{
		resp: talk.Response{
			Done: true,
			ToolCalls: []talk.ToolCall{
				{
					Name: "structured_response",
					Arguments: map[string]any{
						"name":       "my-service",
						"modules":    []any{},
						"boundaries": []any{},
						"files":      []any{},
						"plan_steps": []any{},
						"gaps":       []any{},
					},
				},
			},
		},
	}

	in := strings.NewReader("yes, use axon-auth\n")
	var out strings.Builder
	r := New("claude-sonnet-4-6").WithIO(in, &out)

	updated, err := r.Resolve(context.Background(), spec, mc)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(updated.Gaps) != 0 {
		t.Errorf("expected gaps cleared, got %d", len(updated.Gaps))
	}
	if !strings.Contains(out.String(), "Will this need auth?") {
		t.Error("expected gap question to be printed")
	}
}
