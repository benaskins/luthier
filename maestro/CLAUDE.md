# CLAUDE.md

## What This Is

Create main.go with cobra CLI setup that accepts a project directory path. Implement flag parsing for --dry-run, --verbose, --resume-from, and --coder (coding agent choice). Set up basic error handling and usage documentation. Test by running `maestro --help` and verifying flag parsing with a test directory.

## Build & Run

```bash
just build     # builds to bin/maestro
just install   # installs to ~/.local/bin/maestro
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
