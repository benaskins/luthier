package verify

import (
	"context"
	"encoding/json"
	"fmt"
)

// Params holds the parameters for a shell verification task.
type Params struct {
	ProjectDir string `json:"project_dir"`
	Command    string `json:"command"`
}

// ShellWorker implements task.Worker by running a shell command in a project directory.
type ShellWorker struct{}

// Execute runs the shell command specified in params. The params argument must
// be JSON-encoded Params.
func (w *ShellWorker) Execute(ctx context.Context, params json.RawMessage) error {
	var p Params
	if err := json.Unmarshal(params, &p); err != nil {
		return fmt.Errorf("decode params: %w", err)
	}
	if p.ProjectDir == "" {
		return fmt.Errorf("project_dir is required")
	}
	if p.Command == "" {
		return fmt.Errorf("command is required")
	}

	_, err := Run(p.ProjectDir, p.Command)
	return err
}
