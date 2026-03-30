// Package catalogue loads YAML catalogue files and renders them into system prompts.
package catalogue

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed system_prompt.tmpl
var promptTemplate string

// Catalogue represents a framework/stack catalogue loaded from YAML.
type Catalogue struct {
	Name        string      `yaml:"name"`
	Language    string      `yaml:"language"`
	Version     string      `yaml:"version"`
	Description string      `yaml:"description"`
	Components  []Component `yaml:"components"`
	Patterns    []Pattern   `yaml:"patterns"`
	FileConvs   []FileConv  `yaml:"file_conventions"`
	BoundaryN   string      `yaml:"boundary_notes"`
}

type Component struct {
	Name     string   `yaml:"name"`
	Class    string   `yaml:"class"`
	Purpose  string   `yaml:"purpose"`
	UseWhen  string   `yaml:"use_when"`
	Package  string   `yaml:"package"`
	Requires []string `yaml:"requires,omitempty"`
}

type Pattern struct {
	Requirement string `yaml:"requirement"`
	Pattern     string `yaml:"pattern"`
}

type FileConv struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
}

// Load reads a catalogue from a YAML file.
func Load(path string) (*Catalogue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read catalogue: %w", err)
	}
	var c Catalogue
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse catalogue: %w", err)
	}
	return &c, nil
}

// SystemPrompt renders the catalogue into a system prompt string.
func (c *Catalogue) SystemPrompt() (string, error) {
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := struct {
		Name           string
		Language       string
		Components     []Component
		Patterns       []Pattern
		FileConventions []FileConv
		BoundaryNotes  string
	}{
		Name:           c.Name,
		Language:       c.Language,
		Components:     c.Components,
		Patterns:       c.Patterns,
		FileConventions: c.FileConvs,
		BoundaryNotes:  c.BoundaryN,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}
	return buf.String(), nil
}
