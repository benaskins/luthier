# CLAUDE.md

## What This Is

Initialize Go module, create directory structure (cmd/, internal/), add go.mod with dependencies for axon-loop, axon-talk, axon-tool. Create justfile with build, install, test targets. Test: `just build` succeeds, `go mod tidy` has no changes.

## Build & Run

```bash
just build     # builds to bin/prd-scaffold
just install   # installs to ~/.local/bin/prd-scaffold
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
