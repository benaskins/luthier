# bookmarks

Create Phoenix project with mix phx.new bookmarks --database postgres. Add phoenix_live_view, phoenix_html, postgrex dependencies. Configure dev.exs with PostgreSQL connection. Verify server starts with mix phx.server.

## Build & Test

```bash
go test ./...
go vet ./...
just build     # builds to bin/bookmarks
just install   # copies to ~/.local/bin/bookmarks
```

## Module Selections

- **phoenix**: Web framework for the bookmarking application with routing, controllers, and endpoint. Required for any Phoenix web application. (deterministic)
- **ecto**: Database toolkit for schemas, migrations, and queries. Required for PostgreSQL-backed bookmark storage. (deterministic)
- **postgrex**: PostgreSQL driver for Ecto. Required as specified in constraints. (deterministic)
- **phoenix_html**: HTML helpers and form builders for server-rendered views. Required for the web interface. (deterministic)
- **phoenix_live_view**: Interactive real-time UI for bookmark management. Enables keyboard shortcuts and fast page updates without JavaScript framework. (deterministic)
- **phoenix_live_dashboard**: Production monitoring dashboard for memory, processes, and performance metrics. Required for observability. (deterministic)
- **phoenix_pubsub**: Distributed pub/sub for broadcasting bookmark events (new bookmarks, deletions) to connected LiveViews. Included with Phoenix. (deterministic)
- **telemetry**: Metrics and instrumentation for tracking performance (page load times, search latency). Included with Phoenix. (deterministic)
- **pow**: User authentication with email/password. Required for secure user accounts with hashed passwords and CSRF protection. (deterministic)
- **oban**: Background job processing for async URL metadata fetching (title/description extraction) and bookmark import parsing. PostgreSQL-backed, no Redis required. (deterministic)

## Deterministic / Non-deterministic Boundary

| From | To | Type |
|------|----|------|
| BookmarksWeb | Bookmarks | det |
| Bookmarks | Accounts | det |
| Bookmarks | ExternalWeb | non-det |
| Bookmarks | Oban | det |
| Accounts | Pow | det |

