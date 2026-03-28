package snippets

import (
	"strings"
	"testing"
)

func TestGeneratedSnippetsRegister(t *testing.T) {
	r := NewRegistry()
	for _, s := range GeneratedSnippets() {
		r.Register(s)
	}

	// Core modules must be present with setup code
	coreModules := []string{"axon", "axon-talk", "axon-tool", "axon-loop"}
	for _, m := range coreModules {
		s, ok := r.Get(m)
		if !ok {
			t.Errorf("missing snippet for %q", m)
			continue
		}
		if s.Setup == "" {
			t.Errorf("snippet %q has empty Setup", m)
		}
		if len(s.Imports) == 0 {
			t.Errorf("snippet %q has no imports", m)
		}
	}
}

func TestGeneratedSnippetDeps(t *testing.T) {
	r := NewRegistry()
	for _, s := range GeneratedSnippets() {
		r.Register(s)
	}

	loopSnippet, _ := r.Get("axon-loop")
	depSet := map[string]bool{}
	for _, d := range loopSnippet.Deps {
		depSet[d] = true
	}
	if !depSet["axon-talk"] {
		t.Error("axon-loop should depend on axon-talk")
	}
	if !depSet["axon-tool"] {
		t.Error("axon-loop should depend on axon-tool")
	}
}

func TestGeneratedSnippetsAllHaveRequires(t *testing.T) {
	for _, s := range GeneratedSnippets() {
		if len(s.Requires) == 0 {
			t.Errorf("snippet %q has no require paths", s.Module)
		}
	}
}

func TestGeneratedSnippetsNoDuplicates(t *testing.T) {
	r := NewRegistry()
	for _, s := range GeneratedSnippets() {
		r.Register(s) // panics on duplicate
	}
}

func TestGeneratedSnippetsComposeWithCore(t *testing.T) {
	r := NewRegistry()
	for _, s := range GeneratedSnippets() {
		r.Register(s)
	}

	selected, err := r.ForModules([]string{"axon-talk", "axon-tool", "axon-loop"})
	if err != nil {
		t.Fatal(err)
	}

	src, err := Compose("testapp", selected)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(src, "client := selectLLMClient()") {
		t.Error("missing axon-talk setup")
	}
	if !strings.Contains(src, "tool.ToolDef") {
		t.Error("missing axon-tool setup")
	}
	if !strings.Contains(src, "loop.RunConfig") {
		t.Error("missing axon-loop setup")
	}
}
