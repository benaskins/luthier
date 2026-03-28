// Package snippets provides composable code fragments for axon modules.
// Each snippet declares its imports, setup code, dependencies, and helper
// functions. The Compose function topologically sorts snippets by dependency
// and produces a complete main.go source.
package snippets

import "fmt"

// Import is a single Go import line.
type Import struct {
	Path  string // e.g. "github.com/benaskins/axon-loop"
	Alias string // e.g. "loop", or "" for default
}

// Snippet is a composable code fragment for one axon module.
type Snippet struct {
	Module   string   // e.g. "axon-loop"
	Imports  []Import // Go imports this snippet needs
	Requires []string // go.mod require paths (may differ from imports)
	Setup    string   // variable initialisation code (inserted in dependency order)
	Deps     []string // modules whose Setup must run before this one
	Helpers  string   // standalone helper functions appended after main()
}

// Registry maps module names to their snippets.
type Registry struct {
	snippets map[string]Snippet
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{snippets: make(map[string]Snippet)}
}

// Register adds a snippet. Panics on duplicate module names.
func (r *Registry) Register(s Snippet) {
	if _, exists := r.snippets[s.Module]; exists {
		panic(fmt.Sprintf("snippets: duplicate module %q", s.Module))
	}
	r.snippets[s.Module] = s
}

// Get returns the snippet for a module, or false if not registered.
func (r *Registry) Get(module string) (Snippet, bool) {
	s, ok := r.snippets[module]
	return s, ok
}

// ForModules returns snippets for the given module names, in registration order.
// Returns an error if any module is not registered.
func (r *Registry) ForModules(modules []string) ([]Snippet, error) {
	var out []Snippet
	for _, m := range modules {
		s, ok := r.snippets[m]
		if !ok {
			return nil, fmt.Errorf("snippets: unknown module %q", m)
		}
		out = append(out, s)
	}
	return out, nil
}
