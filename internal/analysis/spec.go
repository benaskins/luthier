// Package analysis implements the structured output call that converts a PRD
// into a ScaffoldSpec.
package analysis

// ProjectType indicates whether the scaffold is a library or a runnable service/CLI.
type ProjectType string

const (
	ProjectLibrary ProjectType = "library"
	ProjectService ProjectType = "service"
	ProjectCLI     ProjectType = "cli"
)

// ScaffoldSpec is the machine-readable output of the analysis call.
type ScaffoldSpec struct {
	Name        string            `json:"name"`
	Type        ProjectType       `json:"type"`
	Modules     []ModuleSelection `json:"modules"`
	Boundaries  []Boundary        `json:"boundaries"`
	PlanSteps   []PlanStep        `json:"plan_steps"`
	Constraints []string          `json:"constraints"`
	Gaps        []Gap             `json:"gaps"`
}

// ModuleSelection records which axon module was selected and why.
type ModuleSelection struct {
	Name            string `json:"name"`
	Reason          string `json:"reason"`
	IsDeterministic bool   `json:"is_deterministic"`
}

// Boundary describes the interface between two components.
type Boundary struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "det" or "non-det"
}

// PlanStep is one commit-sized implementation step for the generated plan.
type PlanStep struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	CommitMessage string `json:"commit_message"`
}

// Gap is an ambiguity in the PRD that requires conversational resolution.
type Gap struct {
	Question string `json:"question"`
	Context  string `json:"context"`
}
