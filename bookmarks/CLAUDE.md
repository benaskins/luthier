# CLAUDE.md

## What This Is

Create Phoenix project with mix phx.new bookmarks --database postgres. Add phoenix_live_view, phoenix_html, postgrex dependencies. Configure dev.exs with PostgreSQL connection. Verify server starts with mix phx.server.

## Build & Run

```bash
just build     # builds to bin/bookmarks
just install   # installs to ~/.local/bin/bookmarks
just test      # go test ./...
just vet       # go vet ./...
```

## Plan

See `plans/` for commit-sized implementation steps. Use `/iterate` to execute them.
