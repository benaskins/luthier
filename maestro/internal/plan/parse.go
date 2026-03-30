// Package plan parses luthier-generated implementation plans.
package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Step represents one commit-sized implementation step from a plan.
type Step struct {
	Number      int
	Title       string
	Description string
	Commit      string
}

var (
	stepHeader = regexp.MustCompile(`^## Step (\d+)\s*[—–-]\s*(.+)$`)
	commitLine = regexp.MustCompile(`^Commit:\s*` + "`" + `(.+)` + "`")
)

// ReadFromDir finds and parses the plan file in a project's plans/ directory.
func ReadFromDir(projectDir string) ([]Step, error) {
	planDir := filepath.Join(projectDir, "plans")
	entries, err := os.ReadDir(planDir)
	if err != nil {
		return nil, fmt.Errorf("read plans dir: %w", err)
	}

	var planFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			planFiles = append(planFiles, filepath.Join(planDir, e.Name()))
		}
	}
	if len(planFiles) == 0 {
		return nil, fmt.Errorf("no plan files found in %s", planDir)
	}

	sort.Strings(planFiles)
	// Use the most recent plan file (last alphabetically, since names are date-prefixed)
	return ParseFile(planFiles[len(planFiles)-1])
}

// ParseFile parses a markdown plan file into ordered steps.
func ParseFile(path string) ([]Step, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}
	return Parse(string(data))
}

// Parse extracts steps from plan markdown content.
func Parse(content string) ([]Step, error) {
	lines := strings.Split(content, "\n")
	var steps []Step
	var current *Step
	var descLines []string

	flush := func() {
		if current != nil {
			current.Description = strings.TrimSpace(strings.Join(descLines, "\n"))
			steps = append(steps, *current)
			current = nil
			descLines = nil
		}
	}

	for _, line := range lines {
		if m := stepHeader.FindStringSubmatch(line); m != nil {
			flush()
			num := 0
			fmt.Sscanf(m[1], "%d", &num)
			current = &Step{
				Number: num,
				Title:  strings.TrimSpace(m[2]),
			}
			continue
		}

		if current == nil {
			continue
		}

		if m := commitLine.FindStringSubmatch(line); m != nil {
			current.Commit = strings.TrimSpace(m[1])
			continue
		}

		descLines = append(descLines, line)
	}
	flush()

	if len(steps) == 0 {
		return nil, fmt.Errorf("no steps found in plan")
	}
	return steps, nil
}
