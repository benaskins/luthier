# maestro — Initial Build Plan
# 2026-03-30

Each step is commit-sized. Execute via `/iterate`.

## Step 1 — set up CLI argument parsing

Create main.go with cobra CLI setup that accepts a project directory path. Implement flag parsing for --dry-run, --verbose, --resume-from, and --coder (coding agent choice). Set up basic error handling and usage documentation. Test by running `maestro --help` and verifying flag parsing with a test directory.

Commit: `feat: initialize maestro CLI with project directory argument`

## Step 2 — build plan file parser

Create plan parser that reads plans/YYYY-MM-DD-*.md files from the project directory. Parse each step's title, description, and commit message using regex or markdown parsing. Return ordered list of steps. Test with sample plan files from a luthier-generated project, verifying correct parsing of all step metadata.

Commit: `feat: implement plan parser for luthier scaffold format`

## Step 3 — wire coding agent delegation

Create coding agent interface with implementations for Claude Code (default) and placeholder for others. Agent receives step description and project context, executes commands to implement the step. Implement command execution wrapper that captures stdout/stderr. Test by invoking Claude Code with a simple file creation task and verifying the file was created.

Commit: `feat: implement coding agent delegation interface`

## Step 4 — add verification command detection

Create verification command detector that scans project root for justfile, Makefile, mix.exs, Gemfile, etc. and extracts the appropriate build/test command. Implement fallback to common commands (make test, just test, go test, etc.). Test with sample projects using different build tools, verifying correct command detection.

Commit: `feat: implement project verification command detection`

## Step 5 — integrate async task runner

Integrate axon-task module to run verification commands asynchronously. Create task workers for build and test operations. Ensure verification runs don't block the main orchestration thread. Test by running verification tasks concurrently and confirming results are collected correctly.

Commit: `feat: integrate axon-task for async verification execution`

## Step 6 — build semantic review loop

Integrate axon-loop and axon-talk to create the semantic review component. The reviewer receives the git diff and step description, then uses an LLM to determine if the implementation matches the requirements. Configure LLM provider (default to Ollama for local testing). Test by reviewing a known-good diff and a known-bad diff, verifying correct pass/fail assessment.

Commit: `feat: implement LLM review loop using axon-loop and axon-talk`

## Step 7 — implement orchestration loop

Create the main orchestration loop that: (1) reads next step, (2) delegates to coding agent, (3) runs verification, (4) runs review, (5) commits if both pass, (6) retries up to 3 times on failure with feedback, (7) stops and reports on persistent failure. Implement retry counter and feedback accumulation. Test with a step that initially fails then succeeds on retry.

Commit: `feat: implement orchestration loop with retry logic`

## Step 8 — add git commit functionality

Create git wrapper that commits changes with the step's commit message. Verify the working directory is clean before committing. Handle git errors (no changes, merge conflicts). Test by committing a sample change and verifying the commit message matches the step's commit_message field.

Commit: `feat: implement git commit with conventional commit messages`

## Step 9 — add dry-run and verbose modes

Add dry-run mode that logs what would happen without executing commands. Add verbose mode that streams coding agent output in real-time. Ensure dry-run skips actual file modifications and git commits. Test dry-run with a full plan, verifying no changes are made. Test verbose mode with a simple step, confirming output visibility.

Commit: `feat: implement dry-run and verbose modes`

## Step 10 — add resume capability

Add logic to detect already-completed steps by checking git commit history for matching commit messages. Implement --resume-from flag to skip steps before a specified step title or index. Test by running a partial plan, then resuming from a middle step, verifying skipped steps are not re-executed.

Commit: `feat: implement resume-from functionality`

## Step 11 — add event audit trail

Integrate axon-fact to create an event log of orchestration activities: step started, agent invoked, verification run, review result, commit result, retry attempts. Implement projector to reconstruct orchestration state. Test by running a plan and verifying events are recorded in the event store.

Commit: `feat: integrate axon-fact for orchestration audit trail`

## Step 12 — add final status reporting

Create final status reporter that summarizes: total steps, completed steps, failed steps, retry statistics, and time per step. Output to stdout and optionally to a summary file. Test by running a plan with mixed success/failure and verifying accurate summary.

Commit: `feat: implement final status reporting`

## Step 13 — add error handling and logging

Add structured error handling throughout the codebase. Implement logging with configurable levels. Ensure all external calls (agent, verification, LLM, git) have proper error wrapping and context. Test by triggering various error conditions and verifying clear error messages.

Commit: `refactor: add comprehensive error handling and logging`

## Step 14 — write documentation

Create README.md with project overview, installation instructions, usage examples, and configuration options. Document all CLI flags, environment variables, and coding agent configuration. Include examples of running maestro against a luthier scaffold. Test by reviewing generated docs for clarity and completeness.

Commit: `docs: write README.md and CLI documentation`

## Step 15 — add comprehensive tests

Write comprehensive tests: unit tests for plan parser, verification detector, git wrapper; integration tests for orchestration loop with mocked agent and LLM; end-to-end test with a simple scaffold project. Ensure test coverage for retry logic and error paths. Run `just test` and verify all tests pass.

Commit: `test: add unit and integration tests for all components`

