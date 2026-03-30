# bookmarks

Create Phoenix project with mix phx.new bookmarks --database postgres. Add phoenix_live_view, phoenix_html, postgrex dependencies. Configure dev.exs with PostgreSQL connection. Verify server starts with mix phx.server.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
bookmarks --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
