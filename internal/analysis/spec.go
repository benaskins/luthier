// Package analysis implements the structured output call that converts a PRD
// into a ScaffoldSpec.
package analysis

// ScaffoldSpec is the machine-readable output of the analysis call.
type ScaffoldSpec struct {
	Name       string            `json:"name"`
	Modules    []ModuleSelection `json:"modules"`
	Boundaries []Boundary        `json:"boundaries"`
	Files      []FileSpec        `json:"files"`
	PlanSteps  []PlanStep        `json:"plan_steps"`
	Gaps       []Gap             `json:"gaps"`
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

// FileSpec describes a file to be written in the scaffold.
type FileSpec struct {
	Path     string            `json:"path"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars"`
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
