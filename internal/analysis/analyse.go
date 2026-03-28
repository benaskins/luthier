package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/luthier/internal/patterns"
)

// scaffoldSpecSchema is the JSON Schema passed to WithStructuredOutput.
// The anthropic adapter's toPropertyDefsFromSchema expects a "properties"
// wrapper at the top level.
var scaffoldSpecSchema = map[string]any{
	"properties": map[string]any{
		"name": map[string]any{
			"type":        "string",
			"description": "Short kebab-case project name derived from the PRD.",
		},
		"modules": map[string]any{
			"type":        "array",
			"description": "Axon modules selected for this project.",
			"items": map[string]any{
				"type":        "object",
				"description": "A module selection with its name, rationale, and determinism classification.",
				"properties": map[string]any{
					"name":             map[string]any{"type": "string", "description": "Exact module name from the catalog (e.g. axon-loop)."},
					"reason":           map[string]any{"type": "string", "description": "Why this module was selected for this project."},
					"is_deterministic": map[string]any{"type": "boolean", "description": "True if this module's contribution is deterministic."},
				},
				"required": []string{"name", "reason", "is_deterministic"},
			},
		},
		"boundaries": map[string]any{
			"type":        "array",
			"description": "Interfaces between components, labelled det or non-det.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"from": map[string]any{"type": "string"},
					"to":   map[string]any{"type": "string"},
					"type": map[string]any{"type": "string", "enum": []string{"det", "non-det"}},
				},
				"required": []string{"from", "to", "type"},
			},
		},
		"plan_steps": map[string]any{
			"type":        "array",
			"description": "Commit-sized implementation steps.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":          map[string]any{"type": "string"},
					"description":    map[string]any{"type": "string"},
					"commit_message": map[string]any{"type": "string"},
				},
				"required": []string{"title", "description", "commit_message"},
			},
		},
		"gaps": map[string]any{
			"type":        "array",
			"description": "Ambiguities requiring conversational resolution. Empty if none.",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"question": map[string]any{"type": "string"},
					"context":  map[string]any{"type": "string"},
				},
				"required": []string{"question", "context"},
			},
		},
	},
}

// Analyse sends the PRD to Claude and returns a structured ScaffoldSpec.
// It uses prompt caching (static system prompt) and structured output
// (constrained tool use). The caller should pass anthropic.WithPromptCaching()
// as a request option where possible.
func Analyse(ctx context.Context, prd string, client talk.LLMClient, model string) (*ScaffoldSpec, error) {
	req := talk.NewRequest(
		model,
		[]talk.Message{
			{Role: talk.RoleSystem, Content: patterns.SystemPrompt},
			{Role: talk.RoleUser, Content: "Analyse this PRD and produce a ScaffoldSpec:\n\n" + prd},
		},
		talk.WithStructuredOutput(scaffoldSpecSchema),
	)

	var toolArgs map[string]any
	err := client.Chat(ctx, req, func(resp talk.Response) error {
		if len(resp.ToolCalls) > 0 {
			toolArgs = resp.ToolCalls[0].Arguments
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("analysis: chat: %w", err)
	}
	if toolArgs == nil {
		return nil, fmt.Errorf("analysis: no structured_response tool call in response")
	}

	return specFromArgs(toolArgs)
}

// specFromArgs converts the tool call arguments map into a ScaffoldSpec by
// round-tripping through JSON. This handles the map[string]any→struct
// conversion cleanly without manual field extraction.
func specFromArgs(args map[string]any) (*ScaffoldSpec, error) {
	data, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("analysis: marshal args: %w", err)
	}
	var spec ScaffoldSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("analysis: unmarshal spec: %w", err)
	}
	return &spec, nil
}
