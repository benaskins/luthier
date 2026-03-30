// Package verify detects and runs project verification commands.
package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// detector maps a filename to the command used to verify that project type.
type detector struct {
	file string
	cmd  string
}

// detectors is the ordered list of project file detectors. First match wins.
var detectors = []detector{
	{"justfile", "just test"},
	{"Justfile", "just test"},
	{"Makefile", "make test"},
	{"GNUmakefile", "make test"},
	{"mix.exs", "mix compile && mix test"},
	{"Gemfile", "bundle exec rake test"},
	{"Cargo.toml", "cargo test"},
	{"pyproject.toml", "python -m pytest"},
	{"package.json", "npm test"},
	{"go.mod", "go vet ./... && go test ./..."},
}

// DetectCommand finds the verification command for a project by scanning for
// well-known build tool files. Returns an error if no recognisable build tool
// is found.
func DetectCommand(projectDir string) (string, error) {
	for _, d := range detectors {
		if _, err := os.Stat(filepath.Join(projectDir, d.file)); err == nil {
			return d.cmd, nil
		}
	}

	return "", fmt.Errorf("no build tool detected in %s", projectDir)
}

// Run executes the verification command and returns the output.
func Run(projectDir string, command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = projectDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("verification failed: %w\n%s", err, out)
	}
	return string(out), nil
}
