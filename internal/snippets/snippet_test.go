package snippets

import "testing"

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	s := Snippet{
		Module:  "axon-loop",
		Imports: []Import{{Path: "github.com/benaskins/axon-loop", Alias: "loop"}},
		Setup:   `loopCfg := loop.DefaultConfig()`,
		Deps:    []string{"axon-talk"},
	}
	r.Register(s)

	got, ok := r.Get("axon-loop")
	if !ok {
		t.Fatal("expected snippet for axon-loop")
	}
	if got.Module != "axon-loop" {
		t.Errorf("got module %q, want %q", got.Module, "axon-loop")
	}
	if len(got.Deps) != 1 || got.Deps[0] != "axon-talk" {
		t.Errorf("got deps %v, want [axon-talk]", got.Deps)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nope")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestRegistryDuplicatePanics(t *testing.T) {
	r := NewRegistry()
	s := Snippet{Module: "axon-loop"}
	r.Register(s)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate register")
		}
	}()
	r.Register(s)
}

func TestForModules(t *testing.T) {
	r := NewRegistry()
	r.Register(Snippet{Module: "axon-talk"})
	r.Register(Snippet{Module: "axon-loop", Deps: []string{"axon-talk"}})

	got, err := r.ForModules([]string{"axon-talk", "axon-loop"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d snippets, want 2", len(got))
	}
	if got[0].Module != "axon-talk" || got[1].Module != "axon-loop" {
		t.Errorf("got modules [%s, %s], want [axon-talk, axon-loop]", got[0].Module, got[1].Module)
	}
}

func TestForModulesUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.ForModules([]string{"nope"})
	if err == nil {
		t.Fatal("expected error for unknown module")
	}
}
