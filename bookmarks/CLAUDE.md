# CLAUDE.md

## What This Is

Create new Rails application with PostgreSQL database. Configure database.yml for development/production/tenancy. Run initial setup. Test by running rails server and accessing root path.

## Build & Run

```bash
just build     # builds to bin/bookmarks
just install   # installs to ~/.local/bin/bookmarks
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
