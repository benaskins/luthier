package snippets

import "testing"

func TestCoreSnippetsRegister(t *testing.T) {
	r := NewRegistry()
	for _, s := range CoreSnippets() {
		r.Register(s)
	}

	modules := []string{"axon", "axon-talk", "axon-tool", "axon-loop"}
	for _, m := range modules {
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

func TestCoreSnippetDeps(t *testing.T) {
	r := NewRegistry()
	for _, s := range CoreSnippets() {
		r.Register(s)
	}

	// axon-loop depends on axon-talk and axon-tool
	loopSnippet, _ := r.Get("axon-loop")
	if len(loopSnippet.Deps) != 2 {
		t.Fatalf("axon-loop deps: got %d, want 2", len(loopSnippet.Deps))
	}

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

	// axon-talk, axon-tool, axon have no deps
	for _, m := range []string{"axon", "axon-talk", "axon-tool"} {
		s, _ := r.Get(m)
		if len(s.Deps) != 0 {
			t.Errorf("%q should have no deps, got %v", m, s.Deps)
		}
	}
}

func TestCoreSnippetRequires(t *testing.T) {
	for _, s := range CoreSnippets() {
		if len(s.Requires) == 0 {
			t.Errorf("snippet %q has no require paths", s.Module)
		}
		for _, r := range s.Requires {
			if r == "" {
				t.Errorf("snippet %q has empty require path", s.Module)
			}
		}
	}
}
