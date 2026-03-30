# scaffold-agent

Create the basic project structure: main.go entry point that parses CLI args (PRD file path), justfile with build/install/test targets, and README.md. Test by running 'just build' and verifying the binary compiles.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
scaffold-agent --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
