// Package review provides semantic review of git diffs using an LLM.
// It determines whether an implementation matches a plan step's requirements.
package review

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	loop "github.com/benaskins/axon-loop"
	talk "github.com/benaskins/axon-talk"
	"github.com/benaskins/axon-talk/openai"

	"github.com/benaskins/maestro/internal/plan"
)

const (
	defaultModel     = "llama3.2"
	defaultOllamaURL = "http://localhost:11434"
)

const systemPrompt = `You are a code reviewer. Given a plan step description and a git diff, determine whether the implementation satisfies the requirements described in the step.

Respond with a JSON object in this exact format:
{"passed": true, "reason": "brief explanation"}

Set "passed" to true only if the diff clearly implements what the step requires. Set it to false if the diff is missing required changes, implements something unrelated, or contains obvious errors.`

// Config configures a Reviewer.
type Config struct {
	// Client is the LLM client to use. Defaults to Ollama via OpenAI-compatible API.
	Client loop.LLMClient
	// Model is the model name. Defaults to "llama3.2".
	Model string
	// OllamaURL is the base URL for Ollama when using the default client.
	// Defaults to "http://localhost:11434".
	OllamaURL string
}

// Result is the outcome of a semantic review.
type Result struct {
	Passed bool
	Reason string
}

// Reviewer uses an LLM to assess whether a git diff satisfies a plan step.
type Reviewer struct {
	client loop.LLMClient
	model  string
}

// New creates a Reviewer. If cfg.Client is nil, an Ollama client is created
// pointing at cfg.OllamaURL (defaulting to http://localhost:11434).
func New(cfg Config) *Reviewer {
	client := cfg.Client
	if client == nil {
		ollamaURL := cfg.OllamaURL
		if ollamaURL == "" {
			ollamaURL = defaultOllamaURL
		}
		client = openai.NewClient(ollamaURL, "ollama")
	}
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}
	return &Reviewer{client: client, model: model}
}

// Review sends diff and step to the LLM and returns a pass/fail assessment.
func (r *Reviewer) Review(ctx context.Context, diff string, step plan.Step) (*Result, error) {
	userContent := fmt.Sprintf(
		"## Step %d: %s\n\n%s\n\n## Git Diff\n\n```diff\n%s\n```\n\nDoes this diff satisfy the step requirements?",
		step.Number, step.Title, step.Description, diff,
	)

	req := talk.NewRequest(r.model, []talk.Message{
		{Role: talk.RoleSystem, Content: systemPrompt},
		{Role: talk.RoleUser, Content: userContent},
	})

	result, err := loop.Run(ctx, loop.RunConfig{
		Client:  r.client,
		Request: req,
	})
	if err != nil {
		return nil, fmt.Errorf("review: LLM call failed: %w", err)
	}

	return parseResult(result.Content)
}

// parseResult extracts a pass/fail result from LLM response content.
// It first tries to parse JSON, then falls back to keyword detection.
func parseResult(content string) (*Result, error) {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		var parsed struct {
			Passed bool   `json:"passed"`
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal([]byte(content[start:end+1]), &parsed); err == nil {
			return &Result{Passed: parsed.Passed, Reason: parsed.Reason}, nil
		}
	}

	// Fallback: keyword detection when JSON parsing fails.
	lower := strings.ToLower(content)
	passed := strings.Contains(lower, "pass") && !strings.Contains(lower, "not pass") && !strings.Contains(lower, "fail")
	return &Result{Passed: passed, Reason: content}, nil
}
