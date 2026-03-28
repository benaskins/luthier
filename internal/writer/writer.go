// Package writer deterministically writes a project scaffold from a ScaffoldSpec.
package writer

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/benaskins/luthier/internal/analysis"
)

//go:embed templates
var templateFS embed.FS

// templateData is the common data passed to every template.
type templateData struct {
	Name        string
	Description string
	Modules     []analysis.ModuleSelection
	Boundaries  []analysis.Boundary
	PlanSteps   []analysis.PlanStep
	Requires    []string // go.mod require paths from snippets
	Date        string
}

// ComposedOutput holds the pre-composed source files from snippet composition.
type ComposedOutput struct {
	MainGo   string   // composed main.go source
	Requires []string // deduplicated go.mod require paths
}

// Write creates outDir and writes the scaffold files derived from spec.
// If composed is non-nil, main.go uses the composed source and go.mod
// includes the require lines. Otherwise falls back to the stub template.
func Write(spec *analysis.ScaffoldSpec, outDir string, composed *ComposedOutput) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("writer: create output dir: %w", err)
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return fmt.Errorf("writer: parse templates: %w", err)
	}

	var requires []string
	if composed != nil {
		requires = composed.Requires
	}

	data := templateData{
		Name:        spec.Name,
		Description: descriptionFromSpec(spec),
		Modules:     spec.Modules,
		Boundaries:  spec.Boundaries,
		PlanSteps:   spec.PlanSteps,
		Requires:    requires,
		Date:        time.Now().Format("2006-01-02"),
	}

	files := []struct {
		tmplName string
		path     string
	}{
		{"AGENTS.md.tmpl", "AGENTS.md"},
		{"CLAUDE.md.tmpl", "CLAUDE.md"},
		{"README.md.tmpl", "README.md"},
		{"go.mod.tmpl", "go.mod"},
		{"justfile.tmpl", "justfile"},
		{"plan.md.tmpl", filepath.Join("plans", data.Date+"-initial-build.md")},
	}

	// main.go: use composed source if available, otherwise fall back to template
	mainPath := filepath.Join(outDir, "cmd", spec.Name, "main.go")
	if composed != nil && composed.MainGo != "" {
		if err := os.MkdirAll(filepath.Dir(mainPath), 0o755); err != nil {
			return fmt.Errorf("writer: mkdir %s: %w", filepath.Dir(mainPath), err)
		}
		if err := os.WriteFile(mainPath, []byte(composed.MainGo), 0o644); err != nil {
			return fmt.Errorf("writer: write main.go: %w", err)
		}
	} else {
		files = append(files, struct {
			tmplName string
			path     string
		}{"main.go.tmpl", filepath.Join("cmd", spec.Name, "main.go")})
	}

	for _, f := range files {
		if err := writeFile(tmpl, f.tmplName, filepath.Join(outDir, f.path), data); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(tmpl *template.Template, tmplName, destPath string, data templateData) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("writer: mkdir %s: %w", filepath.Dir(destPath), err)
	}
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("writer: create %s: %w", destPath, err)
	}
	defer f.Close()
	if err := tmpl.ExecuteTemplate(f, tmplName, data); err != nil {
		return fmt.Errorf("writer: execute template %s: %w", tmplName, err)
	}
	return nil
}

// descriptionFromSpec derives a one-line description from the plan steps or
// falls back to a generic phrase.
func descriptionFromSpec(spec *analysis.ScaffoldSpec) string {
	if len(spec.PlanSteps) > 0 {
		return spec.PlanSteps[0].Description
	}
	return spec.Name + " — axon-based service"
}
