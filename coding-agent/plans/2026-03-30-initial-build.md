# coding-agent — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize project skeleton

Create initial project structure with go.mod, main.go importing axon, and justfile with build/test targets. Verify 'go build' succeeds.

Commit: `feat: scaffold project structure with axon dependencies`

## Step 2 — Setup config and authentication

Add axon config loading with LLM API key support, integrate axon-auth for WebAuthn. Test with mock auth flow.

Commit: `feat: implement axon config loading and auth setup`

## Step 3 — Configure async task runner

Configure axon-task with worker pool for scaffolding generation. Add endpoint to submit PRD analysis job. Test job submission and completion.

Commit: `feat: wire axon-task for async scaffolding jobs`

## Step 4 — Wire LLM conversation loop

Wire axon-loop with axon-talk adapters and axon-tool definitions. Implement PRD parsing logic that identifies ambiguities. Test with sample PRD.

Commit: `feat: integrate axon-loop for PRD analysis`

## Step 5 — Add context caching

Set up axon-memo for caching component catalog and patterns. Implement cache read/write during PRD analysis. Verify cache hits reduce analysis time.

Commit: `feat: implement axon-memo for context caching`

## Step 6 — Implement scaffolding generation

Use axon-fact to create event-based scaffolding output. Generate project files (main.go, justfile, AGENTS.md, CLAUDE.md, README.md, plans/). Test generated project builds.

Commit: `feat: implement scaffolding generation with axon-fact`

## Step 7 — Add output evaluation

Integrate axon-eval to assess generated project quality across LLM providers. Add evaluation metrics tracking. Test with different provider outputs.

Commit: `feat: add axon-eval for output quality assessment`

## Step 8 — E2E workflow testing

Full E2E test: submit PRD, handle clarifying questions, generate scaffolding, verify generated project builds and compiles. Measure time from PRD to build.

Commit: `test: end-to-end scaffolding workflow`

