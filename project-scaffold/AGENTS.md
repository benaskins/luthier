# project-scaffold

Create the initial project structure: go.mod with module name, main.go with CLI argument parsing (accepts PRD file path), and basic error handling. Test: `go build` succeeds and running `./project-scaffold --help` shows usage.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/project-scaffold
just install   # copies to ~/.local/bin/project-scaffold
```

## Module Selections

- **axon-loop**: Required for conversational LLM interaction to analyze PRDs, determine component selections, and ask clarifying questions about ambiguities. This is the core mechanism for non-deterministic PRD analysis. (non-deterministic)
- **axon-talk**: Required to connect axon-loop to LLM providers (Anthropic, Ollama, etc.). The CLI needs to call an LLM to analyze PRD content and make architectural decisions. (deterministic)
- **axon-tool**: Required alongside axon-loop to define tools for structured output. The LLM needs tools to call for component selection, boundary definition, gap raising, and plan generation. Without axon-tool, the loop cannot produce structured scaffolding output. (non-deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| main | axon-loop | non-det |
| axon-loop | axon-talk | det |
| axon-loop | axon-tool | det |

