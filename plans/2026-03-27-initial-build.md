# Luthier — Initial Build Plan
# 2026-03-27

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Scaffold repo ✅

Create: AGENTS.md, CLAUDE.md, go.mod, justfile, cmd/luthier/main.go (stub), plans/

## Step 2 — Analysis call with structured output

Implement `internal/analysis/` package:
- `ScaffoldSpec` type and sub-types (`ModuleSelection`, `Boundary`, `FileSpec`, `PlanStep`, `Gap`)
- `Analyse(ctx, prd string, client loop.LLMClient) (*ScaffoldSpec, error)` (structured output call via axon-loop)
- System prompt in `internal/patterns/` (embedded, static — enables prompt caching)
- Module catalog embedded as text in the system prompt

Tests: unit test with a mock LLM client returning a known JSON spec; verify parsing.

## Step 3 — File writer

Implement `internal/writer/` package:
- `Write(spec *ScaffoldSpec, outDir string) error` — deterministically writes all files from the spec
- Templates for: AGENTS.md, CLAUDE.md, README.md, go.mod, justfile, cmd/main.go, plans/initial-build.md
- Template vars: `spec.Name`, `spec.Modules`, `spec.PlanSteps`, etc.

Tests: unit test Write() against a known ScaffoldSpec; verify file contents.

## Step 4 — Gap-resolution conversational loop

Implement `internal/gaps/` package:
- `Resolve(ctx, spec *ScaffoldSpec, client loop.LLMClient) (*ScaffoldSpec, error)`
- If `spec.Gaps` is empty, return immediately (no-op)
- Otherwise, open conversational loop to resolve each gap; update spec
- Plain readline (not Bubble Tea) for simplicity — luthier is a one-shot tool

Tests: unit test with pre-populated gaps and mock client; verify gaps cleared and spec updated.

## Step 5 — Wire together and test against a real PRD

Wire `main.go`:
1. Parse `os.Args[1]` as PRD path, read file
2. Call `analysis.Analyse()`
3. Call `gaps.Resolve()` if needed
4. Call `writer.Write()`
5. Print output directory path

Smoke test: run `luthier` against the luthier PRD itself (`plans/2026-03-27-luthier-design.md`).
Verify output is a buildable project scaffold.

## Step 6 — Update getlamina.ai

Update `getlamina.ai` site to feature luthier as the primary "getting started" entry point.
