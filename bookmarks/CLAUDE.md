# CLAUDE.md

## What This Is

Create new Rails application with PostgreSQL database, configure database.yml for local and production environments. Run initial setup to verify database connection. Test: rails db:create and rails db:migrate should succeed.

## Build & Run

```bash
just build     # builds to bin/bookmarks
just install   # installs to ~/.local/bin/bookmarks
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
