# CLAUDE.md

## What This Is

Create the initial project structure: go.mod with module name, main.go with CLI argument parsing (accepts PRD file path), and basic error handling. Test: `go build` succeeds and running `./project-scaffold --help` shows usage.

## Build & Run

```bash
just build     # builds to bin/project-scaffold
just install   # installs to ~/.local/bin/project-scaffold
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
