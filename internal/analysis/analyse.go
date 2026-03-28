package analysis

import (
	"context"
	"fmt"
	"os"

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
	verbose := os.Getenv("LUTHIER_DEBUG") != ""

	req := talk.NewRequest(
		model,
		[]talk.Message{
			{Role: talk.RoleSystem, Content: patterns.SystemPrompt},
			{Role: talk.RoleUser, Content: `Analyse this PRD and produce a scaffold spec by calling the provided tools.

You MUST call tools in this order:
1. Call select_module once for each axon module the project needs
2. Call define_boundary for each interface between components
3. Call add_plan_step for each commit-sized implementation step
4. Call raise_gap ONLY for genuine ambiguities you cannot resolve from the catalog
5. Call finalize with the project name when you are done

Do NOT respond with text between tool calls. Call multiple tools per turn when possible.
You MUST call finalize when your analysis is complete.

PRD:

` + prd},
		},
	)
	req.MaxIterations = 50

	cfg := loop.RunConfig{
		Client:  client,
		Request: req,
		Tools:   builder.Tools(),
	}

	if verbose {
		cfg.Callbacks = loop.Callbacks{
			OnToolUse: func(name string, args map[string]any) {
				switch name {
				case "select_module":
					fmt.Fprintf(os.Stderr, "  tool: select_module(%s)\n", args["name"])
				case "define_boundary":
					fmt.Fprintf(os.Stderr, "  tool: define_boundary(%s → %s, %s)\n", args["from"], args["to"], args["type"])
				case "add_plan_step":
					fmt.Fprintf(os.Stderr, "  tool: add_plan_step(%s)\n", args["title"])
				case "raise_gap":
					fmt.Fprintf(os.Stderr, "  tool: raise_gap(%s)\n", args["question"])
				case "finalize":
					fmt.Fprintf(os.Stderr, "  tool: finalize(%s)\n", args["name"])
				default:
					fmt.Fprintf(os.Stderr, "  tool: %s(...)\n", name)
				}
			},
			OnDone: func(durationMs int64) {
				spec := builder.Spec()
				fmt.Fprintf(os.Stderr, "  done: %dms, %d modules, %d boundaries, %d steps, %d gaps, finalized=%v\n",
					durationMs, len(spec.Modules), len(spec.Boundaries), len(spec.PlanSteps), len(spec.Gaps), builder.Finalized())
			},
		}
	}

	_, err := loop.Run(ctx, cfg)
	if err != nil {
		if verbose {
			spec := builder.Spec()
			fmt.Fprintf(os.Stderr, "  error state: %d modules, %d boundaries, %d steps, %d gaps, finalized=%v\n",
				len(spec.Modules), len(spec.Boundaries), len(spec.PlanSteps), len(spec.Gaps), builder.Finalized())
		}
		return nil, fmt.Errorf("analysis: %w", err)
	}

	if !builder.Finalized() {
		if verbose {
			spec := builder.Spec()
			fmt.Fprintf(os.Stderr, "  no finalize: %d modules, %d boundaries, %d steps, %d gaps\n",
				len(spec.Modules), len(spec.Boundaries), len(spec.PlanSteps), len(spec.Gaps))
		}
		return nil, fmt.Errorf("analysis: model did not call finalize")
	}

	return builder.Spec(), nil
}
