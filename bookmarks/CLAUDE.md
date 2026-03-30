# CLAUDE.md

## What This Is

Create new Rails application with PostgreSQL database, set up basic structure. Run rails new bookmarks --database=postgresql. Verify database connection and basic Rails setup.

## Build & Run

```bash
just build     # builds to bin/bookmarks
just install   # installs to ~/.local/bin/bookmarks
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
