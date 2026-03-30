# bookmarks

Create new Rails application with PostgreSQL database, set up basic structure. Run rails new bookmarks --database=postgresql. Verify database connection and basic Rails setup.

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
