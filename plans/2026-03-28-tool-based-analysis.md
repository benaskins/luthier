# Tool-Based Analysis — Design Document
# 2026-03-28

## Problem

The current analysis call uses `WithStructuredOutput` to get the model to fill
in a complete ScaffoldSpec JSON in one shot. This works for Sonnet but fails
for Opus (renames fields) and will likely fail for non-Anthropic models. The
one-shot template approach fights the model instead of using its strengths.

## Insight

Models are excellent at tool use — they never rename tool parameters. Each
tool call is a tight, well-constrained invocation. Instead of asking for one
giant JSON, let the model reason through the PRD and call small tools as it goes.

## Design

Replace the single `client.Chat()` + `WithStructuredOutput` call with an
`axon-loop` conversation that provides five tools:

| Tool | Parameters | Called |
|------|-----------|--------|
| `select_module` | name, reason, is_deterministic | Once per module |
| `define_boundary` | from, to, type (det/non-det) | Once per boundary |
| `add_plan_step` | title, description, commit_message | Once per step |
| `raise_gap` | question, context | When ambiguous |
| `finalize` | name | Once, signals completion |

The loop runs until `finalize` is called. Each tool call appends to the
ScaffoldSpec being built up incrementally.

## Trade-offs

- **More API calls**: multi-turn vs one-shot. Higher latency and cost.
- **More reliable**: tool parameter names are never renamed by any model.
- **Model-agnostic**: works with any provider that supports tool use.
- **Observable**: each tool call is a discrete, loggable event.

## Build Order

1. Define the five tools using axon-tool
2. Build a SpecBuilder that accumulates tool calls into a ScaffoldSpec
3. Replace the analysis call with an axon-loop conversation
4. Remove WithStructuredOutput and the field normalization hack
5. Run eval on Sonnet and Opus, compare to baseline
