package analysis

import (
	"context"
	"fmt"

	loop "github.com/benaskins/axon-loop"
	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/luthier/internal/patterns"
)

// Analyse sends the PRD to the LLM via an axon-loop conversation.
// The model calls tools (select_module, define_boundary, add_plan_step,
// raise_gap, finalize) to incrementally build a ScaffoldSpec.
// The loop exits when the model stops calling tools after finalize.
func Analyse(ctx context.Context, prd string, client talk.LLMClient, model string) (*ScaffoldSpec, error) {
	builder := NewSpecBuilder()

	req := talk.NewRequest(
		model,
		[]talk.Message{
			{Role: talk.RoleSystem, Content: patterns.SystemPrompt},
			{Role: talk.RoleUser, Content: "Analyse this PRD and produce a scaffold spec.\n\nUse the tools to build the spec incrementally: select_module for each module, define_boundary for each boundary, add_plan_step for each implementation step, raise_gap for any ambiguities, then finalize when done.\n\n" + prd},
		},
	)

	_, err := loop.Run(ctx, loop.RunConfig{
		Client: client,
		Request: req,
		Tools:  builder.Tools(),
	})
	if err != nil {
		return nil, fmt.Errorf("analysis: %w", err)
	}

	if !builder.Finalized() {
		return nil, fmt.Errorf("analysis: model did not call finalize")
	}

	return builder.Spec(), nil
}
