# prd-scaffold

Initialize Go module, create directory structure (cmd/, internal/), add go.mod with dependencies for axon-loop, axon-talk, axon-tool. Create justfile with build, install, test targets. Test: `just build` succeeds, `go mod tidy` has no changes.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
prd-scaffold --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
