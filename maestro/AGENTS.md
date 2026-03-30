# maestro

Create main.go with cobra CLI setup that accepts a project directory path. Implement flag parsing for --dry-run, --verbose, --resume-from, and --coder (coding agent choice). Set up basic error handling and usage documentation. Test by running `maestro --help` and verifying flag parsing with a test directory.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/maestro
just install   # copies to ~/.local/bin/maestro
```

## Module Selections

- **axon-loop**: Required for the LLM conversation loop that semantically reviews whether the git diff matches the plan step description. This is the core non-deterministic component that assesses implementation correctness. (non-deterministic)
- **axon-talk**: Provides LLM provider adapters (Anthropic, Ollama, Cloudflare Workers AI) needed for the axon-loop review step. Enables the orchestrator to communicate with different LLM backends for semantic verification. (non-deterministic)
- **axon-fact**: Event sourcing primitives to track orchestration state, step execution results, and retry attempts. Creates an audit trail of the implementation process that can be replayed or inspected. (deterministic)
- **axon-task**: Generic async task runner for executing verification commands (build, test) and delegating to the coding agent. Separates I/O-bound operations from the main orchestration loop. (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| maestro-conductor | coding-agent | non-det |
| maestro-conductor | reviewer-loop | non-det |
| maestro-conductor | project-verification | det |
| maestro-conductor | plan-parser | det |

