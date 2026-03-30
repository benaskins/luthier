# project-scaffold

Create the initial project structure: go.mod with module name, main.go with CLI argument parsing (accepts PRD file path), and basic error handling. Test: `go build` succeeds and running `./project-scaffold --help` shows usage.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
project-scaffold --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
