build:
    go build -o bin/luthier ./cmd/luthier
    go build -o bin/luthier-eval ./cmd/luthier-eval
    go build -o bin/luthier-sync ./cmd/luthier-sync

install: build
    cp bin/luthier ~/.local/bin/luthier

sync-catalog:
    go run ./cmd/luthier-sync

test:
    go test ./...

vet:
    go vet ./...
