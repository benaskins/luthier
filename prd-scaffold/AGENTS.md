# prd-scaffold

Initialize Go module, create directory structure (cmd/, internal/), add go.mod with dependencies for axon-loop, axon-talk, axon-tool. Create justfile with build, install, test targets. Test: `just build` succeeds, `go mod tidy` has no changes.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/prd-scaffold
just install   # copies to ~/.local/bin/prd-scaffold
```

## Module Selections

- **axon-loop**: Core conversation loop for analyzing PRDs and making architectural decisions. The agent needs to process the input document, determine component needs, ask clarifying questions when ambiguous, and generate structured output. (non-deterministic)
- **axon-talk**: Required to connect axon-loop to an LLM provider. Supports Anthropic, Ollama, or Cloudflare Workers AI for the PRD analysis and decision-making. (deterministic)
- **axon-tool**: Required for axon-loop to define tools that perform file operations (writing generated files), shell execution (verifying compilation), and potentially other structured actions needed to scaffold projects. (non-deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| main.go | axon-loop | non-det |
| axon-loop | axon-talk | det |
| axon-loop | axon-tool | non-det |
| axon-tool | filesystem | det |

