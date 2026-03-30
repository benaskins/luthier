# CLAUDE.md

## What This Is

Create the basic project structure: main.go entry point that parses CLI args (PRD file path), justfile with build/install/test targets, and README.md. Test by running 'just build' and verifying the binary compiles.

## Build & Run

```bash
just build     # builds to bin/scaffold-agent
just install   # installs to ~/.local/bin/scaffold-agent
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
