package analysis

import (
	"fmt"
	"sync"

	"github.com/benaskins/axon-tool"
)

// SpecBuilder accumulates tool calls into a ScaffoldSpec.
type SpecBuilder struct {
	mu   sync.Mutex
	spec ScaffoldSpec
	done bool
}

// NewSpecBuilder returns an empty builder.
func NewSpecBuilder() *SpecBuilder {
	return &SpecBuilder{}
}

// Spec returns the accumulated spec. Only valid after Finalized() returns true.
func (b *SpecBuilder) Spec() *ScaffoldSpec {
	b.mu.Lock()
	defer b.mu.Unlock()
	return &b.spec
}

// Finalized returns true if the finalize tool has been called.
func (b *SpecBuilder) Finalized() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.done
}

// Tools returns the five analysis tools wired to this builder.
func (b *SpecBuilder) Tools() map[string]tool.ToolDef {
	return map[string]tool.ToolDef{
		"select_module":   b.selectModuleTool(),
		"define_boundary": b.defineBoundaryTool(),
		"add_plan_step":   b.addPlanStepTool(),
		"raise_gap":       b.raiseGapTool(),
		"finalize":        b.finalizeTool(),
	}
}

func (b *SpecBuilder) selectModuleTool() tool.ToolDef {
	return tool.ToolDef{
		Name:        "select_module",
		Description: "Select an axon module for this project. Call once per module.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"name", "reason", "is_deterministic"},
			Properties: map[string]tool.PropertySchema{
				"name": {
					Type:        "string",
					Description: "Exact module name from the catalog (e.g. axon-loop, axon-fact).",
				},
				"reason": {
					Type:        "string",
					Description: "Why this module is needed for the project described in the PRD.",
				},
				"is_deterministic": {
					Type:        "boolean",
					Description: "True if this module's contribution to the project is deterministic.",
				},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			name, _ := args["name"].(string)
			reason, _ := args["reason"].(string)
			isDet, _ := args["is_deterministic"].(bool)

			b.mu.Lock()
			b.spec.Modules = append(b.spec.Modules, ModuleSelection{
				Name:            name,
				Reason:          reason,
				IsDeterministic: isDet,
			})
			count := len(b.spec.Modules)
			b.mu.Unlock()

			return tool.ToolResult{Content: fmt.Sprintf("Module %q selected (%d total).", name, count)}
		},
	}
}

func (b *SpecBuilder) defineBoundaryTool() tool.ToolDef {
	return tool.ToolDef{
		Name:        "define_boundary",
		Description: "Define an interface boundary between two components. Label it deterministic or non-deterministic.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"from", "to", "type"},
			Properties: map[string]tool.PropertySchema{
				"from": {
					Type:        "string",
					Description: "Source component or module.",
				},
				"to": {
					Type:        "string",
					Description: "Target component or module.",
				},
				"type": {
					Type:        "string",
					Description: "Boundary type.",
					Enum:        []any{"det", "non-det"},
				},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			from, _ := args["from"].(string)
			to, _ := args["to"].(string)
			typ, _ := args["type"].(string)

			b.mu.Lock()
			b.spec.Boundaries = append(b.spec.Boundaries, Boundary{
				From: from,
				To:   to,
				Type: typ,
			})
			count := len(b.spec.Boundaries)
			b.mu.Unlock()

			return tool.ToolResult{Content: fmt.Sprintf("Boundary %s → %s (%s) defined (%d total).", from, to, typ, count)}
		},
	}
}

func (b *SpecBuilder) addPlanStepTool() tool.ToolDef {
	return tool.ToolDef{
		Name:        "add_plan_step",
		Description: "Add a commit-sized implementation step to the build plan.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"title", "description", "commit_message"},
			Properties: map[string]tool.PropertySchema{
				"title": {
					Type:        "string",
					Description: "Short verb-phrase title for this step.",
				},
				"description": {
					Type:        "string",
					Description: "What to build and how to test it.",
				},
				"commit_message": {
					Type:        "string",
					Description: "Conventional commit message (feat:/fix:/refactor: prefix).",
				},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			title, _ := args["title"].(string)
			desc, _ := args["description"].(string)
			commit, _ := args["commit_message"].(string)

			b.mu.Lock()
			b.spec.PlanSteps = append(b.spec.PlanSteps, PlanStep{
				Title:         title,
				Description:   desc,
				CommitMessage: commit,
			})
			count := len(b.spec.PlanSteps)
			b.mu.Unlock()

			return tool.ToolResult{Content: fmt.Sprintf("Plan step %d: %q added.", count, title)}
		},
	}
}

func (b *SpecBuilder) raiseGapTool() tool.ToolDef {
	return tool.ToolDef{
		Name:        "raise_gap",
		Description: "Raise an ambiguity in the PRD that needs human clarification before the design can be completed. Only use when there is genuine ambiguity — do not raise gaps for decisions you can make from the module catalog and patterns.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"question", "context"},
			Properties: map[string]tool.PropertySchema{
				"question": {
					Type:        "string",
					Description: "The clarifying question to ask the developer.",
				},
				"context": {
					Type:        "string",
					Description: "Why this question matters and what trade-offs are involved.",
				},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			question, _ := args["question"].(string)
			context, _ := args["context"].(string)

			b.mu.Lock()
			b.spec.Gaps = append(b.spec.Gaps, Gap{
				Question: question,
				Context:  context,
			})
			count := len(b.spec.Gaps)
			b.mu.Unlock()

			return tool.ToolResult{Content: fmt.Sprintf("Gap %d raised: %q", count, question)}
		},
	}
}

func (b *SpecBuilder) finalizeTool() tool.ToolDef {
	return tool.ToolDef{
		Name:        "finalize",
		Description: "Signal that the analysis is complete. Call this after all modules, boundaries, plan steps, and gaps have been defined.",
		Parameters: tool.ParameterSchema{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]tool.PropertySchema{
				"name": {
					Type:        "string",
					Description: "Short kebab-case project name derived from the PRD.",
				},
			},
		},
		Execute: func(ctx *tool.ToolContext, args map[string]any) tool.ToolResult {
			name, _ := args["name"].(string)

			b.mu.Lock()
			b.spec.Name = name
			b.done = true
			b.mu.Unlock()

			return tool.ToolResult{Content: fmt.Sprintf("Scaffold spec for %q finalized.", name)}
		},
	}
}
