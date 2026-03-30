# project-scaffold — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize project structure

Create the initial project structure: go.mod with module name, main.go with CLI argument parsing (accepts PRD file path), and basic error handling. Test: `go build` succeeds and running `./project-scaffold --help` shows usage.

Commit: `feat: initialize project structure with go.mod and basic main.go`

## Step 2 — Add LLM provider integration

Import axon-talk and configure LLM provider selection (read from AXON_TALK_PROVIDER env var: anthropic, ollama, cloudflare). Create TalkConfig struct and initialize provider adapter. Test: Verify provider selection works and returns error for invalid provider.

Commit: `feat: add axon-talk integration for LLM provider connection`

## Step 3 — Define scaffold generation tools

Define tool schemas for: select_module (component selection), define_boundary (interface boundaries), raise_gap (clarification questions), add_plan_step (implementation steps), finalize (completion signal). Each tool returns structured output. Test: Tools parse input and return valid JSON responses.

Commit: `feat: implement axon-tool definitions for scaffold generation`

## Step 4 — Wire conversation loop for PRD analysis

Integrate axon-loop with axon-talk and axon-tool. Create conversation flow: load PRD file → analyze requirements → use tools to record decisions → ask clarifying questions for gaps → generate output. Test: Run with sample PRD and verify loop executes through all phases.

Commit: `feat: wire axon-loop for PRD analysis conversation`

## Step 5 — Add PRD parsing and context preparation

Add file I/O to read markdown PRD file, extract sections (problem, solution, requirements), and prepare context for LLM analysis. Include axon module catalog and patterns as static context. Test: Parse sample PRD file and verify context is properly formatted.

Commit: `feat: implement PRD file parsing and context preparation`

## Step 6 — Implement project file generation

Create generator that writes: go.mod, main.go (stub), justfile, AGENTS.md (architecture doc), README.md, and plans/ directory with initial plan step. Use collected module selections, boundaries, and plan steps to populate files. Test: Run generator and verify all files are created with correct content.

Commit: `feat: implement generated project file writer`

## Step 7 — Add compilation verification

After generating project files, run `go build` in the output directory to verify it compiles. Report success or failure. Test: Generate a project and verify build succeeds; generate with invalid config and verify failure is reported.

Commit: `feat: add verification step to check generated project compiles`

## Step 8 — Add interactive gap resolution

When axon-loop determines a gap exists, pause and prompt the user for clarification. Collect answer and resume conversation with the clarification. Test: Run with ambiguous PRD and verify tool pauses for user input, then continues.

Commit: `feat: implement gap raising with user interaction`

## Step 9 — Add unit tests

Write tests for: PRD parsing, tool schema validation, context preparation, and file generation. Ensure all edge cases are covered. Test: Run `go test ./...` and verify all tests pass.

Commit: `test: add unit tests for tool definitions and parsing`

## Step 10 — Add integration test

Create an integration test that runs the full pipeline with a sample PRD file and verifies the generated project structure is correct. Test: Run integration test and verify generated project matches expected structure.

Commit: `test: add integration test with sample PRD`

## Step 11 — Add architecture documentation

Create AGENTS.md documenting the tool's architecture, module selections, boundaries, and usage instructions. This file will be included in generated projects explaining the design decisions. Test: Verify AGENTS.md is created with all required sections.

Commit: `docs: add AGENTS.md with architecture documentation`

