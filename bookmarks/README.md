# bookmarks

Create new Rails application with PostgreSQL database. Configure database.yml for development/production/tenancy. Run initial setup. Test by running rails server and accessing root path.

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
