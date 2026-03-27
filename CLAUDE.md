# CLAUDE.md

## What This Is

Luthier scaffolds axon-based projects from a PRD. It reads a markdown PRD, calls Claude for structured analysis, resolves any gaps conversationally, then writes a complete project scaffold deterministically.

## Build & Run

```bash
just build          # builds to bin/luthier
just install        # installs to ~/.local/bin/luthier
just test           # go test ./...
just vet            # go vet ./...
```

## Key Conventions

- **No ad-hoc LLM calls** — all Claude interaction goes through `axon-loop`
- **Structured output via constrained tool use** — analysis call uses `WithStructuredOutput()` on `axon-loop`
- **Prompt caching** — system prompt is static per luthier version; use `WithPromptCaching()` on the client
- **Local replace directives** for axon-* development — `go.mod` has replace directives pointing to `~/dev/lamina/`

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
