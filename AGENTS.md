# luthier

CLI tool that takes a PRD and produces a scaffolded project directory ready for Claude Code to execute. The "getting started" experience for building axon-based applications.

Name: a luthier builds the instrument you then play.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/luthier
just install   # copies to ~/.local/bin/luthier
```

## Guiding Principle

> If the outcome can be determined from the set of inputs, write code to get that outcome. Only establish a conversational loop when the output is genuinely non-deterministic.

Luthier applies this principle to project scaffolding: most design work is deterministic analysis. The conversational loop is minimal — only for genuine ambiguity in the PRD or trade-offs with no clear right answer.

## Invocation

```bash
luthier my-prd.md
```

Produces a scaffolded directory `./my-app/` (name derived from PRD).

## Flow

1. **Read PRD**: parse the input document (deterministic)
2. **Analysis call**: structured output request to Claude with the PRD and lamina pattern library as context. Returns a machine-readable `ScaffoldSpec`.
   - Uses prompt caching: the system prompt (patterns + module catalog) is identical across invocations
   - May use extended thinking for complex designs
3. **Gap resolution**: if the spec contains unresolved ambiguities, open a conversational loop with Claude to resolve them. Loop exits when the spec is complete.
4. **Write scaffold**: deterministically write files from the finalised spec.

## Output Structure

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

## Structure

```
cmd/luthier/main.go     entry point, flag parsing, orchestration
internal/analysis/      analysis call: structured output, ScaffoldSpec type
internal/gaps/          gap-resolution conversational loop
internal/writer/        deterministic file writer from ScaffoldSpec
internal/patterns/      lamina pattern library (embedded text)
internal/prompt/        system prompt assembly with caching headers
```

## Deterministic / Non-deterministic Boundary

| Phase | Nature |
|-------|--------|
| Read PRD | Deterministic |
| Analysis call | Non-deterministic (Claude) — returns structured `ScaffoldSpec` |
| Gap resolution | Non-deterministic (Claude) — only if `spec.Gaps` non-empty |
| File writing | Deterministic — pure function of `ScaffoldSpec` |

## Key Dependencies

| Module | Role |
|--------|------|
| `axon-loop` | Conversation loop for gap-resolution phase |
| `axon-talk/anthropic` | Claude provider (direct or via Cloudflare AI Gateway) |
| `axon-tool` | Tool definitions for any resolution tools |

## ScaffoldSpec

```go
type ScaffoldSpec struct {
    Name       string
    Modules    []ModuleSelection
    Boundaries []Boundary
    Files      []FileSpec
    PlanSteps  []PlanStep
    Gaps       []Gap  // non-empty triggers conversational loop
}
```

## Environment

| Var | Purpose |
|-----|---------|
| `ANTHROPIC_API_KEY` | Direct Anthropic API |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account (AI Gateway) |
| `CLOUDFLARE_AI_GATEWAY_TOKEN` | AI Gateway auth token |

## Lamina Patterns

Luthier's system prompt encodes patterns for common app shapes:

| Requirement | Pattern |
|-------------|---------|
| HTTP service | `axon.ListenAndServe`, config via `axon.MustLoadConfig` |
| LLM conversation | `axon-loop` + `axon-talk` + `axon-tool` |
| Async work | `axon-task` + `axon-fact` — never block HTTP handlers |
| Authentication | `axon-auth` (WebAuthn/passkeys) |
| Event audit trail / replay | `axon-fact` projectors |
| Cross-session memory | `axon-memo` |
| Cross-instance fan-out | `axon-nats` |
| Process supervision | aurelia service YAML |
| Deterministic output | Go code, no LLM loop |
| Non-deterministic output | `axon-loop` conversation, not ad-hoc LLM calls |
