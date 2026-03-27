// Package gaps resolves ambiguities in a ScaffoldSpec via a conversational
// loop with Claude. If the spec has no gaps, Resolve is a no-op.
package gaps

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/luthier/internal/analysis"
)

// Resolver handles gap resolution. The in/out fields allow tests to inject
// a reader/writer instead of os.Stdin/os.Stdout.
type Resolver struct {
	in    io.Reader
	out   io.Writer
	model string
}

// New creates a Resolver using the given model. in/out default to nil;
// call WithIO to override for testing.
func New(model string) *Resolver {
	return &Resolver{model: model}
}

// WithIO sets the reader and writer used for prompts and answers.
func (r *Resolver) WithIO(in io.Reader, out io.Writer) *Resolver {
	r.in = in
	r.out = out
	return r
}

// Resolve resolves any Gaps in spec. If Gaps is empty it returns spec
// unchanged. Otherwise it asks Claude to update the spec based on each
// human answer collected from the terminal.
func (r *Resolver) Resolve(ctx context.Context, spec *analysis.ScaffoldSpec, client talk.LLMClient) (*analysis.ScaffoldSpec, error) {
	if len(spec.Gaps) == 0 {
		return spec, nil
	}

	answers, err := r.collectAnswers(spec)
	if err != nil {
		return nil, err
	}

	return r.updateSpec(ctx, spec, answers, client)
}

// collectAnswers prompts the user for an answer to each gap question.
func (r *Resolver) collectAnswers(spec *analysis.ScaffoldSpec) (map[string]string, error) {
	out := r.out
	if out == nil {
		return nil, fmt.Errorf("gaps: no output writer configured")
	}
	in := r.in
	if in == nil {
		return nil, fmt.Errorf("gaps: no input reader configured")
	}

	answers := make(map[string]string, len(spec.Gaps))
	scanner := bufio.NewScanner(in)

	fmt.Fprintf(out, "\nLuthier needs a few more details (%d question(s)):\n\n", len(spec.Gaps))
	for i, gap := range spec.Gaps {
		fmt.Fprintf(out, "[%d/%d] %s\n", i+1, len(spec.Gaps), gap.Question)
		if gap.Context != "" {
			fmt.Fprintf(out, "      Context: %s\n", gap.Context)
		}
		fmt.Fprint(out, "> ")

		if !scanner.Scan() {
			return nil, fmt.Errorf("gaps: input ended unexpectedly")
		}
		answers[gap.Question] = strings.TrimSpace(scanner.Text())
		fmt.Fprintln(out)
	}

	return answers, nil
}

// updateSpec asks Claude to produce an updated ScaffoldSpec given the answers.
func (r *Resolver) updateSpec(ctx context.Context, spec *analysis.ScaffoldSpec, answers map[string]string, client talk.LLMClient) (*analysis.ScaffoldSpec, error) {
	specJSON, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("gaps: marshal spec: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("Here is the current ScaffoldSpec:\n\n")
	sb.WriteString("```json\n")
	sb.Write(specJSON)
	sb.WriteString("\n```\n\n")
	sb.WriteString("Here are the answers to the gaps:\n\n")
	for q, a := range answers {
		fmt.Fprintf(&sb, "Q: %s\nA: %s\n\n", q, a)
	}
	sb.WriteString("Please return an updated ScaffoldSpec with the gaps resolved. Use the structured_response tool.")

	req := talk.NewRequest(
		r.model,
		[]talk.Message{
			{
				Role:    talk.RoleSystem,
				Content: "You are Luthier. Update the ScaffoldSpec by incorporating the provided answers. Clear the gaps field in the result.",
			},
			{Role: talk.RoleUser, Content: sb.String()},
		},
		talk.WithStructuredOutput(resolvedSpecSchema),
	)

	var toolArgs map[string]any
	err = client.Chat(ctx, req, func(resp talk.Response) error {
		if len(resp.ToolCalls) > 0 {
			toolArgs = resp.ToolCalls[0].Arguments
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("gaps: chat: %w", err)
	}
	if toolArgs == nil {
		return nil, fmt.Errorf("gaps: no structured_response in gap-resolution reply")
	}

	data, err := json.Marshal(toolArgs)
	if err != nil {
		return nil, fmt.Errorf("gaps: marshal updated spec args: %w", err)
	}
	var updated analysis.ScaffoldSpec
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("gaps: unmarshal updated spec: %w", err)
	}
	return &updated, nil
}

// resolvedSpecSchema is a minimal schema for the gap-resolution response.
// It mirrors the analysis schema but is defined here to avoid import cycles.
var resolvedSpecSchema = map[string]any{
	"name":    map[string]any{"type": "string"},
	"modules": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
	"boundaries": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
	"files":      map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
	"plan_steps": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
	"gaps":       map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
}
