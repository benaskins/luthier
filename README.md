# Luthier

A luthier builds the instrument you then play.

Luthier takes a product requirements document and produces a scaffolded, compilable Go project wired with [axon](https://github.com/benaskins/axon) components. You describe what you want to build. Luthier figures out the architecture, selects modules, writes the glue code, and produces a step-by-step implementation plan.

## Install

```bash
just install
```

## Usage

```bash
luthier my-prd.md
```

Luthier reads your PRD, calls an LLM to analyse the requirements, asks clarifying questions if anything is ambiguous, then writes a project directory with:

- `cmd/<name>/main.go` : glue code wiring selected axon modules together
- `go.mod` : dependencies with local replace directives
- `AGENTS.md` : architecture decisions, module selections, det/non-det boundaries
- `CLAUDE.md` : working instructions for Claude Code
- `plans/<date>-initial-build.md` : commit-sized implementation steps
- `justfile` : build, test, install targets

The generated project compiles immediately. Business logic is stubbed; the plan tells you (or Claude Code via `/iterate`) what to build next.

## How it works

1. **Analyse**: an LLM reads your PRD alongside the axon module catalog and calls tools (`select_module`, `define_boundary`, `add_plan_step`, `raise_gap`, `finalize`) to build a scaffold spec incrementally
2. **Resolve gaps**: if the spec contains ambiguities, luthier asks you to clarify
3. **Compose**: snippets for each selected module are topologically sorted by dependency and composed into a working `main.go`
4. **Write**: templates render the scaffold deterministically from the spec
5. **Verify**: `go build ./...` confirms the scaffold compiles

## Module catalog

Each axon module owns a `luthier.yaml` manifest declaring its purpose, when to use it, and its scaffold snippet. Run `just sync-catalog` to regenerate the catalog from the module manifests in the lamina workspace.

```bash
just sync-catalog    # regenerate system prompt + snippets from luthier.yaml files
```

## Eval

Luthier includes an eval tool that runs the analysis N times against a PRD and scores consistency across runs.

```bash
just build
bin/luthier-eval fixtures/coding-agent-prd.md 5           # Sonnet (default)
LUTHIER_MODEL=claude-opus-4-6 bin/luthier-eval fixtures/coding-agent-prd.md 5  # Opus
LUTHIER_PROVIDER=local bin/luthier-eval fixtures/coding-agent-prd.md 3         # local llama-server
```

Scores: module selection consistency, boundary count stability, gap detection stability, plan step count stability.

## Guiding principle

> If the outcome can be determined from the set of inputs, write code to get that outcome. Only establish a conversational loop when the output is genuinely non-deterministic.

The module catalog and wiring patterns are known inputs, so snippet composition is deterministic. The LLM is only involved where genuine reasoning is needed: which modules fit this PRD, where the det/non-det boundaries are, and what the implementation plan should be.
