# maestro

Create main.go with cobra CLI setup that accepts a project directory path. Implement flag parsing for --dry-run, --verbose, --resume-from, and --coder (coding agent choice). Set up basic error handling and usage documentation. Test by running `maestro --help` and verifying flag parsing with a test directory.

## Prerequisites

- Go 1.24+
- [just](https://github.com/casey/just)

## Build & Run

```bash
just build
just install
maestro --help
```

## Development

```bash
just test   # run tests
just vet    # run go vet
```
