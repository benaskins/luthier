// Package verify detects and runs project verification commands.
package verify

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DetectCommand finds the verification command for a project.
func DetectCommand(projectDir string) (string, error) {
	checks := []struct {
		file string
		cmd  string
	}{
		{"justfile", "just test"},
		{"Makefile", "make test"},
		{"mix.exs", "mix compile && mix test"},
		{"Gemfile", "bundle exec rails test"},
		{"package.json", "npm test"},
		{"go.mod", "go vet ./... && go test ./..."},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(projectDir, c.file)); err == nil {
			return c.cmd, nil
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
