# coding-agent

Create initial project structure with go.mod, main.go importing axon, and justfile with build/test targets. Verify 'go build' succeeds.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
coding-agent --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
