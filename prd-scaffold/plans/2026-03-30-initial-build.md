# prd-scaffold — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Set up project structure and dependencies

Initialize Go module, create directory structure (cmd/, internal/), add go.mod with dependencies for axon-loop, axon-talk, axon-tool. Create justfile with build, install, test targets. Test: `just build` succeeds, `go mod tidy` has no changes.

Commit: `feat: initialize project structure with axon dependencies`

## Step 2 — Implement CLI argument parsing

Create main.go that parses command-line arguments to accept a PRD file path. Handle --help and error cases for missing file. Test: `prd-scaffold --help` shows usage, `prd-scaffold nonexistent.md` returns error.

Commit: `feat: add CLI argument parsing for PRD file input`

## Step 3 — Configure axon-talk provider adapter

Create internal/config package that loads LLM provider configuration (provider type, API key, model). Support Anthropic, Ollama, Cloudflare Workers AI. Create MustLoadConfig function. Test: Config loads from env vars, invalid provider returns error.

Commit: `feat: add LLM provider configuration and adapter setup`

## Step 4 — Implement axon-tool tool definitions

Define tools for: file writing (write_file), directory creation (create_dir), shell execution (run_command for verification), and read_file for context. Implement tool handlers that execute these operations. Test: Tools can be registered and called with valid parameters.

Commit: `feat: implement tool definitions for file and shell operations`

## Step 5 — Wire axon-loop with talk and tool integrations

Create internal/agent package that initializes axon-loop with axon-talk provider and axon-tool definitions. Configure the loop with system prompt for PRD analysis. Test: Loop initializes without errors, can start a conversation.

Commit: `feat: wire axon-loop with talk provider and tool definitions`

## Step 6 — Implement PRD analysis prompt and context

Create internal/prompt package with system prompt that instructs the LLM to: analyze PRD, identify required components from catalog, detect ambiguities, ask clarifying questions. Include axon module catalog and patterns as context. Test: Prompt constructs correctly, includes all required context.

Commit: `feat: add PRD analysis system prompt with module catalog context`

## Step 7 — Implement file generation for scaffolded projects

Create internal/generator package that generates: go.mod, main.go stub, justfile, README.md, AGENTS.md (architecture), plans/ directory with initial plan. Templates use selected module info. Test: Generator produces valid Go project structure.

Commit: `feat: implement project file generation with templates`

## Step 8 — Add compilation verification step

Implement verification that runs `go build` on generated project and checks exit code. If build fails, report errors. Test: Verification passes for valid generated project, fails for invalid.

Commit: `feat: add compilation verification for generated projects`

## Step 9 — Integrate full pipeline in main.go

Wire everything together: read PRD file, run analysis loop, collect tool calls (file writes), execute file operations, run verification, report success/failure. Handle clarifying questions by pausing loop for user input. Test: End-to-end flow from PRD to verified project.

Commit: `feat: integrate full PRD-to-scaffold pipeline`

## Step 10 — Add unit tests for core components

Write unit tests for: config loading, tool definitions, prompt construction, file generation templates. Mock axon-loop and axon-talk for isolated testing. Test: All unit tests pass with `just test`.

Commit: `test: add unit tests for core components`

