package snippets

import (
	"strings"
	"testing"
)

func TestTopoSortRespectsDeps(t *testing.T) {
	snippets := []Snippet{
		{Module: "axon-loop", Deps: []string{"axon-talk", "axon-tool"}},
		{Module: "axon-talk"},
		{Module: "axon-tool"},
	}

	sorted, err := topoSort(snippets)
	if err != nil {
		t.Fatal(err)
	}

	idx := map[string]int{}
	for i, s := range sorted {
		idx[s.Module] = i
	}

	if idx["axon-talk"] >= idx["axon-loop"] {
		t.Error("axon-talk should come before axon-loop")
	}
	if idx["axon-tool"] >= idx["axon-loop"] {
		t.Error("axon-tool should come before axon-loop")
	}
}

func TestTopoSortCycleDetection(t *testing.T) {
	snippets := []Snippet{
		{Module: "a", Deps: []string{"b"}},
		{Module: "b", Deps: []string{"a"}},
	}

	_, err := topoSort(snippets)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention cycle: %v", err)
	}
}

func TestTopoSortMissingDep(t *testing.T) {
	// If a dep is not in the selected set, it's skipped (not an error)
	snippets := []Snippet{
		{Module: "axon-loop", Deps: []string{"axon-talk"}},
	}

	sorted, err := topoSort(snippets)
	if err != nil {
		t.Fatal(err)
	}
	if len(sorted) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(sorted))
	}
}

func TestComposeMinimal(t *testing.T) {
	snippets := []Snippet{
		{
			Module:  "axon-tool",
			Imports: []Import{{Path: "github.com/benaskins/axon-tool", Alias: "tool"}},
			Setup:   "\tallTools := tool.NewRegistry()",
		},
	}

	src, err := Compose("myapp", snippets)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(src, "package main") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(src, `tool "github.com/benaskins/axon-tool"`) {
		t.Error("missing aliased import")
	}
	if !strings.Contains(src, "allTools := tool.NewRegistry()") {
		t.Error("missing setup code")
	}
	if !strings.Contains(src, "// TODO: wire myapp business logic here") {
		t.Error("missing TODO placeholder")
	}
}

func TestComposeWithDeps(t *testing.T) {
	snippets := []Snippet{
		{
			Module:  "axon-loop",
			Imports: []Import{{Path: "github.com/benaskins/axon-loop", Alias: "loop"}},
			Setup:   "\tloopCfg := loop.Config{Client: client}",
			Deps:    []string{"axon-talk"},
		},
		{
			Module:  "axon-talk",
			Imports: []Import{{Path: "os"}},
			Setup:   "\tclient := selectLLMClient()",
			Helpers: "func selectLLMClient() any { return nil }\n",
		},
	}

	src, err := Compose("myapp", snippets)
	if err != nil {
		t.Fatal(err)
	}

	// axon-talk setup should appear before axon-loop setup
	talkIdx := strings.Index(src, "selectLLMClient()")
	loopIdx := strings.Index(src, "loop.Config")
	if talkIdx >= loopIdx {
		t.Error("axon-talk setup should appear before axon-loop setup")
	}

	// Helpers should appear after main()
	mainIdx := strings.Index(src, "func main()")
	helperIdx := strings.Index(src, "func selectLLMClient()")
	if helperIdx <= mainIdx {
		t.Error("helpers should appear after main()")
	}
}

func TestComposeImportGrouping(t *testing.T) {
	snippets := []Snippet{
		{
			Module: "test",
			Imports: []Import{
				{Path: "os"},
				{Path: "fmt"},
				{Path: "github.com/benaskins/axon"},
			},
			Setup: "\t_ = os.Getenv(\"X\")",
		},
	}

	src, err := Compose("myapp", snippets)
	if err != nil {
		t.Fatal(err)
	}

	// stdlib should come before external
	fmtIdx := strings.Index(src, `"fmt"`)
	axonIdx := strings.Index(src, `"github.com/benaskins/axon"`)
	if fmtIdx >= axonIdx {
		t.Error("stdlib imports should come before external imports")
	}
}

func TestComposeWithCoreSnippets(t *testing.T) {
	// Integration test: compose using real core snippets
	r := NewRegistry()
	for _, s := range CoreSnippets() {
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

	// Should have all three modules wired in dependency order
	if !strings.Contains(src, "selectLLMClient()") {
		t.Error("missing axon-talk setup")
	}
	if !strings.Contains(src, "tool.NewRegistry()") {
		t.Error("missing axon-tool setup")
	}
	if !strings.Contains(src, "loop.Config") {
		t.Error("missing axon-loop setup")
	}

	// Verify ordering
	talkIdx := strings.Index(src, "selectLLMClient()")
	toolIdx := strings.Index(src, "tool.NewRegistry()")
	loopIdx := strings.Index(src, "loop.Config")
	if talkIdx >= loopIdx {
		t.Error("axon-talk should be before axon-loop")
	}
	if toolIdx >= loopIdx {
		t.Error("axon-tool should be before axon-loop")
	}
}
