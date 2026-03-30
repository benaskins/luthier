package verify_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benaskins/maestro/internal/verify"
)

// touch creates an empty file at path, failing the test on error.
func touch(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("touch %s: %v", path, err)
	}
	f.Close()
}

func TestDetectCommand(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantCmd string
	}{
		{"justfile lowercase", "justfile", "just test"},
		{"justfile titlecase", "Justfile", "just test"},
		{"Makefile", "Makefile", "make test"},
		{"GNUmakefile", "GNUmakefile", "make test"},
		{"mix.exs (Elixir)", "mix.exs", "mix compile && mix test"},
		{"Gemfile (Ruby)", "Gemfile", "bundle exec rake test"},
		{"Cargo.toml (Rust)", "Cargo.toml", "cargo test"},
		{"pyproject.toml (Python)", "pyproject.toml", "python -m pytest"},
		{"package.json (Node)", "package.json", "npm test"},
		{"go.mod (Go)", "go.mod", "go vet ./... && go test ./..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			touch(t, filepath.Join(dir, tt.file))

			got, err := verify.DetectCommand(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantCmd {
				t.Errorf("DetectCommand() = %q, want %q", got, tt.wantCmd)
			}
		})
	}
}

func TestDetectCommand_NoBuildTool(t *testing.T) {
	dir := t.TempDir()
	// Write an unrecognised file so the directory is not empty.
	touch(t, filepath.Join(dir, "README.md"))

	_, err := verify.DetectCommand(dir)
	if err == nil {
		t.Fatal("expected error for project with no recognised build tool, got nil")
	}
}

func TestDetectCommand_PrecedenceJustfileOverMakefile(t *testing.T) {
	// When both a justfile and a Makefile exist, justfile should win because
	// it appears first in the detector list.
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "justfile"))
	touch(t, filepath.Join(dir, "Makefile"))

	got, err := verify.DetectCommand(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "just test" {
		t.Errorf("expected justfile to take precedence, got %q", got)
	}
}

func TestDetectCommand_PrecedenceJustfileOverGoMod(t *testing.T) {
	// A Go project with a justfile should use just test, not go test.
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "justfile"))
	touch(t, filepath.Join(dir, "go.mod"))

	got, err := verify.DetectCommand(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "just test" {
		t.Errorf("expected justfile to take precedence over go.mod, got %q", got)
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()

	out, err := verify.Run(dir, "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Errorf("Run() output = %q, want %q", out, "hello\n")
	}
}

func TestRun_FailingCommand(t *testing.T) {
	dir := t.TempDir()

	_, err := verify.Run(dir, "exit 1")
	if err == nil {
		t.Fatal("expected error for failing command, got nil")
	}
}

func TestRun_CommandInProjectDir(t *testing.T) {
	// Verify that the command runs with the project directory as CWD.
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "sentinel"))

	out, err := verify.Run(dir, "ls sentinel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output listing sentinel file")
	}
}
