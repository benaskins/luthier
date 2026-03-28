package snippets

import (
	"fmt"
	"sort"
	"strings"
)

// Compose produces a complete main.go source from the given snippets.
// Snippets are topologically sorted by Deps before composing.
func Compose(name string, snippets []Snippet) (string, error) {
	sorted, err := topoSort(snippets)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString("package main\n\n")

	// Imports
	imports := collectImports(sorted)
	if len(imports) > 0 {
		b.WriteString("import (\n")
		for _, imp := range imports {
			if imp.Path == "" {
				b.WriteString("\n") // blank line separator between stdlib and external
				continue
			}
			if imp.Alias != "" {
				fmt.Fprintf(&b, "\t%s %q\n", imp.Alias, imp.Path)
			} else {
				fmt.Fprintf(&b, "\t%q\n", imp.Path)
			}
		}
		b.WriteString(")\n\n")
	}

	// main()
	b.WriteString("func main() {\n")
	for _, s := range sorted {
		if s.Setup == "" {
			continue
		}
		fmt.Fprintf(&b, "%s\n\n", s.Setup)
	}
	fmt.Fprintf(&b, "\t// TODO: wire %s business logic here\n", name)
	fmt.Fprintf(&b, "\tfmt.Fprintln(os.Stderr, %q, \"ready\")\n", name+":")
	b.WriteString("}\n")

	// Helpers
	for _, s := range sorted {
		if s.Helpers == "" {
			continue
		}
		b.WriteString("\n")
		b.WriteString(s.Helpers)
		b.WriteString("\n")
	}

	return b.String(), nil
}

// collectImports deduplicates and sorts imports from all snippets.
// Standard library imports come first, then third-party.
func collectImports(snippets []Snippet) []Import {
	seen := map[string]Import{}
	for _, s := range snippets {
		for _, imp := range s.Imports {
			if _, exists := seen[imp.Path]; !exists {
				seen[imp.Path] = imp
			}
		}
	}

	var stdlib, external []Import
	for _, imp := range seen {
		if isStdlib(imp.Path) {
			stdlib = append(stdlib, imp)
		} else {
			external = append(external, imp)
		}
	}
	sort.Slice(stdlib, func(i, j int) bool { return stdlib[i].Path < stdlib[j].Path })
	sort.Slice(external, func(i, j int) bool { return external[i].Path < external[j].Path })

	if len(stdlib) > 0 && len(external) > 0 {
		// Add a blank separator import between stdlib and external
		return append(append(stdlib, Import{}), external...)
	}
	return append(stdlib, external...)
}

func isStdlib(path string) bool {
	return !strings.Contains(path, ".")
}

// topoSort orders snippets so that each snippet's Deps appear before it.
// Returns an error on cycles.
func topoSort(snippets []Snippet) ([]Snippet, error) {
	byModule := map[string]Snippet{}
	for _, s := range snippets {
		byModule[s.Module] = s
	}

	var order []Snippet
	state := map[string]int{} // 0=unvisited, 1=visiting, 2=visited

	var visit func(string) error
	visit = func(module string) error {
		switch state[module] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("snippets: dependency cycle involving %q", module)
		}
		state[module] = 1

		s, ok := byModule[module]
		if !ok {
			// Dependency not in the selected set — skip (it's not needed)
			state[module] = 2
			return nil
		}

		for _, dep := range s.Deps {
			if err := visit(dep); err != nil {
				return err
			}
		}

		state[module] = 2
		order = append(order, s)
		return nil
	}

	for _, s := range snippets {
		if err := visit(s.Module); err != nil {
			return nil, err
		}
	}
	return order, nil
}
