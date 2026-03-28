# Glue Code Generation — Design Document
# 2026-03-28

## Problem

Luthier scaffolds a project directory with the right modules selected, but main.go is a dead stub. The developer (or Claude Code via /iterate) has to write all the wiring from scratch. The module selections and their dependencies are already known — the glue code is deterministic from the ScaffoldSpec.

## Design Principle

> If the outcome can be determined from the set of inputs, write code to get that outcome.

The ScaffoldSpec contains modules, boundaries, and dependency relationships. The wiring patterns are standardised across lamina apps. Therefore: generate the glue code deterministically from snippets. No second LLM call.

## Snippet Registry

Each axon module registers a snippet:

```go
type Snippet struct {
    Module   string   // "axon-loop"
    Imports  []Import // [{Path: "github.com/benaskins/axon-loop", Alias: "loop"}]
    Requires []string // go.mod require paths
    Setup    string   // variable initialisation code
    Deps     []string // other modules this snippet needs setup before it
    Helpers  string   // standalone helper functions appended to main.go
}
```

### No shapes — capabilities compose freely

There is no "HTTP service" vs "CLI tool" vs "TUI app" type. Apps in lamina freely combine these capabilities — aurelia has all three. HTTP listener, Bubble Tea program, and Cobra command tree are snippets just like axon-loop and axon-fact. They compose via the same dependency mechanism.

### Composition

1. Collect snippets for all modules in ScaffoldSpec.Modules
2. Topologically sort by Deps
3. Concatenate Imports (deduplicated)
4. Concatenate Setup lines in dependency order
5. Append Helpers as standalone functions
6. Emit require lines into go.mod template

### Variable conventions (observed, not imposed)

These names are already consistent across lamina apps:

| Snippet | Produces | Used by |
|---------|----------|---------|
| axon-talk | `client` | axon-loop |
| axon-tool | `allTools` | axon-loop |
| axon-fact | `events` | axon-task |
| axon | `db`, `mux` | axon-fact, axon-auth |

No formal naming convention needed — the existing practice is stable.

### go.mod and lamina deps

Snippets declare their require paths. The go.mod template includes these as require lines. After scaffolding, `lamina deps` validates the dependency graph — it reads go.mod files, so it works on scaffolded projects automatically.

## ScaffoldSpec Changes

Add a `shape` field? **No.** The model selects modules during analysis. Capabilities like HTTP, TUI, CLI are expressed as module selections, not a separate type field. The snippet registry handles composition.

## Scope

- Generate a wiring main.go that imports selected modules and composes them
- Business logic is stubbed (handlers return "not implemented", tools are registered but empty)
- go.mod includes require lines for selected modules
- Scaffold still compiles on first try (existing verify step)

## What This Does NOT Do

- Generate business logic
- Generate tests (left to plan steps)
- Replace the analysis call or gap resolution
- Add a second LLM call

## Build Order

1. Define the Snippet type and registry in `internal/snippets/`
2. Implement snippets for core modules: axon, axon-talk, axon-loop, axon-tool
3. Implement snippets for capability modules: axon-fact, axon-task, axon-auth, axon-memo
4. Implement topological sort and main.go composition
5. Update go.mod template to include require lines from snippets
6. Update writer to use composed main.go instead of dead stub
7. Test: run against fixtures/coding-agent-prd.md, verify output compiles and imports are correct
