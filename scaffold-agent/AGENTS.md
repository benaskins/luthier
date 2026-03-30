# scaffold-agent

Create the basic project structure: main.go entry point that parses CLI args (PRD file path), justfile with build/install/test targets, and README.md. Test by running 'just build' and verifying the binary compiles.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/scaffold-agent
just install   # copies to ~/.local/bin/scaffold-agent
```

## Module Selections

- **axon-loop**: This CLI tool uses an LLM to analyze PRDs and make architectural decisions. The conversation loop orchestrates the analysis process, handles clarifying questions, and coordinates with tools to produce structured scaffold output. (non-deterministic)
- **axon-talk**: Required with axon-loop to connect to LLM providers (Anthropic, Ollama, etc.). The CLI needs to send PRD content to an LLM for analysis and receive architectural decisions. (deterministic)
- **axon-tool**: Required with axon-loop. The LLM needs tools to call to produce structured output: selecting modules, defining boundaries, creating plan steps, and raising gaps. Without tools, the loop cannot produce the scaffold spec. (non-deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| main.go | axon-loop | det |
| axon-loop | axon-talk | det |
| axon-loop | axon-tool | non-det |
| axon-tool | tool-implementations | det |

