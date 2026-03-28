package snippets

import "testing"

func TestCapabilitySnippetsRegister(t *testing.T) {
	r := NewRegistry()
	// Register core first (capabilities depend on them)
	for _, s := range CoreSnippets() {
		r.Register(s)
	}
	for _, s := range CapabilitySnippets() {
		r.Register(s)
	}

	modules := []string{"axon-fact", "axon-task", "axon-auth", "axon-memo"}
	for _, m := range modules {
		s, ok := r.Get(m)
		if !ok {
			t.Errorf("missing snippet for %q", m)
			continue
		}
		if s.Setup == "" {
			t.Errorf("snippet %q has empty Setup", m)
		}
	}
}

func TestCapabilitySnippetsDependOnAxon(t *testing.T) {
	for _, s := range CapabilitySnippets() {
		found := false
		for _, d := range s.Deps {
			if d == "axon" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("snippet %q should depend on axon", s.Module)
		}
	}
}

func TestAllSnippetsNoDuplicates(t *testing.T) {
	r := NewRegistry()
	all := append(CoreSnippets(), CapabilitySnippets()...)
	for _, s := range all {
		r.Register(s) // panics on duplicate
	}
	if len(all) != 8 {
		t.Errorf("expected 8 total snippets, got %d", len(all))
	}
}
