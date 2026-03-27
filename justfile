build:
    go build -o bin/luthier ./cmd/luthier

install: build
    cp bin/luthier ~/.local/bin/luthier

test:
    go test ./...

vet:
    go vet ./...
