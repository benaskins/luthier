# coding-agent

Create initial project structure with go.mod, main.go importing axon, and justfile with build/test targets. Verify 'go build' succeeds.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/coding-agent
just install   # copies to ~/.local/bin/coding-agent
```

## Module Selections

- **axon**: Core HTTP service lifecycle, config loading, and server management for the CLI tool's API endpoint (deterministic)
- **axon-loop**: Provider-agnostic conversation loop for analyzing PRDs and asking clarifying questions about ambiguities (non-deterministic)
- **axon-talk**: LLM provider adapters (Anthropic, Ollama, Cloudflare Workers AI) required by axon-loop for PRD analysis (non-deterministic)
- **axon-tool**: Tool definitions for file operations, code generation, and PRD parsing during scaffolding (deterministic)
- **axon-fact**: Event sourcing primitives for audit trail of generated projects and replay capability (deterministic)
- **axon-task**: Async task runner for scaffolding generation without blocking HTTP handlers (deterministic)
- **axon-memo**: Long-term memory for caching static analysis context (component catalog, patterns) across invocations (deterministic)
- **axon-eval**: Evaluation framework for assessing output quality across different LLM providers (non-deterministic)
- **axon-auth**: WebAuthn/passkey authentication for securing the scaffolding API endpoint (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| axon | axon-auth | det |
| axon | axon-task | det |
| axon-task | axon-loop | non-det |
| axon-loop | axon-talk | non-det |
| axon-loop | axon-tool | det |
| axon-loop | axon-memo | det |
| axon-task | axon-fact | det |
| axon-loop | axon-eval | non-det |

