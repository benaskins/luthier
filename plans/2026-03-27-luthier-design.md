# Luthier — Design Document
# 2026-03-27

## What

Luthier is a CLI tool that takes a PRD and produces a scaffolded project directory ready for Claude Code to execute. It is the "getting started" experience for building axon-based applications.

Name: a luthier builds the instrument you then play.

## Guiding Principle

> If the outcome can be determined from the set of inputs, write code to get that outcome. Only establish a conversational loop when the output is genuinely non-deterministic.

Luthier applies this principle to itself: most of the design work is deterministic analysis. The conversational loop is minimal — only for genuine ambiguity in the PRD or trade-offs with no clear right answer.

## Invocation

```bash
luthier my-prd.md
```

Produces a scaffolded directory `./my-app/` (name derived from PRD).

## Flow

1. **Read PRD**: parse the input document (deterministic)
2. **Analysis call**: structured output request to Claude with the PRD and lamina pattern library as context. Returns a machine-readable scaffold spec: modules selected, service boundaries, deterministic/non-deterministic boundaries, plan steps.
   - Uses `WithPromptCaching()`: the system prompt (patterns + module catalog) is identical across sessions
   - May use extended thinking (`req.Think`) for complex designs
3. **Gap resolution**: if the spec contains unresolved ambiguities, open a conversational loop with Claude to resolve them. Loop exits when the spec is complete.
4. **Write scaffold**: deterministically write files from the finalised spec.

## Output

```
my-app/
├── AGENTS.md       # Architecture, module selections, det/non-det boundaries, dep graph
├── CLAUDE.md       # Working instructions for Claude Code
├── README.md       # What it is, how to run it
├── plans/
│   └── 2026-03-27-initial-build.md   # Commit-sized implementation steps
├── go.mod          # axon-* dependencies declared
├── justfile        # build, test, install targets
└── cmd/my-app/
    └── main.go     # axon wiring stub — structure only, no business logic
```

AGENTS.md carries the design decisions: which modules, why, where the deterministic/non-deterministic boundary sits, the internal dependency graph. These are the decisions a developer would otherwise have to re-derive. The plan is what Claude Code iterates through via `/iterate`.

## Architecture

**Repo**: `benaskins/luthier` (standalone, like Imago)
**Language**: Go
**Task runner**: `just`

### Axon stack

| Module | Role |
|--------|------|
| `axon-loop` | Conversation loop for the gap-resolution phase |
| `axon-talk/anthropic` | Claude provider |
| `axon-tool` | Tool definitions for any resolution tools |

### Provider

Default: Claude via `anthropic.NewClient()` (direct to `api.anthropic.com` or via Cloudflare AI Gateway).

Model: `claude-sonnet-4-6` for analysis; may escalate to Opus for complex designs.

Evaluation: swap to `openai.NewClient()` for any OpenAI-compatible provider (Groq, Gemini, Ollama with OpenAI shim) — no other code changes needed. Evaluations via `axon-eval`.

### Analysis call design

Structured output via constrained tool use (`WithStructuredOutput()`). The response schema captures:

```
ScaffoldSpec {
  name        string
  modules     []ModuleSelection { name, reason, deterministic bool }
  boundaries  []Boundary { from, to, type: "det"|"non-det" }
  files       []FileSpec { path, template, vars }
  plan_steps  []PlanStep { title, description, commit_message }
  gaps        []Gap { question, context }  // triggers conversational loop if non-empty
}
```

Gaps drive the conversational loop. When gaps is empty, write the scaffold.

### System prompt

Luthier's system prompt encodes:
- The deterministic/non-deterministic principle
- The axon module catalog (name, purpose, deps, when to use)
- Established patterns for common app shapes (chat service, async pipeline, auth-gated API, etc.)
- The scaffold output format and conventions (AGENTS.md structure, plan step format, etc.)

This prompt is static per luthier version — prompt caching makes repeated invocations cheap.

## Lamina Patterns (initial set)

These are the opinions luthier applies. To be expanded as patterns are formalised.

| Requirement | Pattern |
|-------------|---------|
| HTTP service | `axon.ListenAndServe`, config via `axon.MustLoadConfig` |
| LLM conversation | `axon-loop` + `axon-talk` + `axon-tool` |
| Async work | `axon-task` + `axon-fact` — never block HTTP handlers on heavy work |
| Authentication | `axon-auth` (WebAuthn/passkeys) |
| Event audit trail / replay | `axon-fact` projectors |
| Cross-session memory | `axon-memo` |
| Cross-instance fan-out | `axon-nats` |
| Process supervision | aurelia service YAML |
| Deterministic output | Go code, no LLM loop |
| Non-deterministic output | `axon-loop` conversation, not ad-hoc LLM calls |

## Site integration

Luthier is the answer to "getting started with lamina." Once built, `getlamina.ai` should feature it as the primary entry point — not a documentation of manual steps. The interactive "getting started" section on the site points to `luthier` as the tool.

## Open questions

- Should luthier register the new app in a catalog (lamina's repo list)? Or is that a separate `lamina init` step?
- Should `main.go` stub include aurelia service YAML generation, or leave that for the plan steps?
- Conversational loop UI: Bubble Tea (like Imago) or plain readline?

## Build order

1. Scaffold `benaskins/luthier` repo (AGENTS.md, CLAUDE.md, go.mod, justfile)
2. Implement analysis call with structured output
3. Implement file writer (deterministic, template-driven)
4. Implement gap-resolution conversational loop
5. Wire together, test against a real PRD
6. Update `getlamina.ai` to feature luthier
