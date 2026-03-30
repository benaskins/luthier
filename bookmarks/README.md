# bookmarks

Create new Rails application with PostgreSQL database, configure database.yml for local and production environments. Run initial setup to verify database connection. Test: rails db:create and rails db:migrate should succeed.

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
