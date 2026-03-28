build:
    go build -o bin/luthier ./cmd/luthier
    go build -o bin/luthier-eval ./cmd/luthier-eval

install: build
    cp bin/luthier ~/.local/bin/luthier

test:
    go test ./...

vet:
    go vet ./...
