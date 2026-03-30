# CLAUDE.md

## What This Is

Create initial project structure with go.mod, main.go importing axon, and justfile with build/test targets. Verify 'go build' succeeds.

## Build & Run

```bash
just build     # builds to bin/coding-agent
just install   # installs to ~/.local/bin/coding-agent
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
