# scaffold-agent — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — Initialize project structure

Create the basic project structure: main.go entry point that parses CLI args (PRD file path), justfile with build/install/test targets, and README.md. Test by running 'just build' and verifying the binary compiles.

Commit: `feat: initialize project structure with main.go and justfile`

## Step 2 — Integrate axon-talk provider connection

Add axon-talk dependency and configure LLM provider selection (via env var for Anthropic/Ollama/Cloudflare). Create talk.go that wraps axon-talk's provider selection. Test by verifying the talk package initializes correctly with a mock provider config.

Commit: `feat: integrate axon-talk for LLM provider connection`

## Step 3 — Integrate axon-loop orchestration

Add axon-loop dependency and create the main conversation loop that orchestrates PRD analysis. The loop will read the PRD, send it to the LLM via axon-talk, and wait for tool calls. Test by running the loop with a simple test PRD.

Commit: `feat: integrate axon-loop for conversation orchestration`

## Step 4 — Implement tool definitions and handlers

Create tool definitions for select_module, define_boundary, add_plan_step, raise_gap, and finalize. Implement the tool handlers that execute these operations and collect the results into a ScaffoldSpec. Test by verifying each tool can be called and returns expected results.

Commit: `feat: implement axon-tool integration with tool definitions`

## Step 5 — Implement PRD analysis prompt

Create the system prompt that instructs the LLM to analyze PRDs and use the available tools to produce a scaffold spec. Include the module catalog and patterns as context. Test by running the loop with a sample PRD and verifying the LLM uses the tools correctly.

Commit: `feat: implement PRD analysis prompt and context`

## Step 6 — Generate scaffold spec files

Create code that takes the collected ScaffoldSpec (modules, boundaries, plan steps, gaps) and writes it to output files in the target directory. Generate: AGENTS.md (architecture doc), plans/YYYY-MM-DD-initial-build.md (plan steps), and a stub main.go for the generated project. Test by verifying the generated files are valid and contain expected content.

Commit: `feat: implement scaffold spec file generation`

## Step 7 — Add project compilation verification

Add code to verify the generated project compiles by running 'go build' in the target directory and checking for errors. Report success or failure with appropriate error messages. Test by generating a project and verifying the compile check passes.

Commit: `feat: implement project verification (compile check)`

## Step 8 — Add context caching

Add caching for the module catalog and patterns to avoid re-reading them on every invocation. Cache to disk and check for updates. Test by running the tool twice and verifying the second run uses cached context.

Commit: `feat: implement caching for static analysis context`

