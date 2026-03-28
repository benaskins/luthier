package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/benaskins/axon-talk/anthropic"
	"github.com/benaskins/luthier/internal/analysis"
)

const defaultModel = "claude-sonnet-4-6"
const defaultRuns = 3

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "luthier-eval:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: luthier-eval <prd.md> [runs]")
	}
	prdPath := os.Args[1]

	numRuns := defaultRuns
	if len(os.Args) >= 3 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			return fmt.Errorf("invalid run count: %w", err)
		}
		numRuns = n
	}

	prd, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("read PRD: %w", err)
	}

	client := newClient()
	model := defaultModel
	if m := os.Getenv("LUTHIER_MODEL"); m != "" {
		model = m
	}

	fmt.Fprintf(os.Stderr, "luthier-eval: running %d analyses against %s (model: %s)\n", numRuns, prdPath, model)

	var specs []*analysis.ScaffoldSpec
	for i := 0; i < numRuns; i++ {
		fmt.Fprintf(os.Stderr, "  run %d/%d… ", i+1, numRuns)
		spec, err := analysis.Analyse(context.Background(), string(prd), client, model)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAILED: %v\n", err)
			continue
		}
		specs = append(specs, spec)
		fmt.Fprintf(os.Stderr, "ok (%d modules, %d boundaries, %d gaps, %d steps)\n",
			len(spec.Modules), len(spec.Boundaries), len(spec.Gaps), len(spec.PlanSteps))
	}

	if len(specs) < 2 {
		return fmt.Errorf("need at least 2 successful runs to compare, got %d", len(specs))
	}

	report := evaluate(specs)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// Report is the eval output.
type Report struct {
	Runs           int            `json:"runs"`
	ModuleScore    float64        `json:"module_score"`
	BoundaryScore  float64        `json:"boundary_score"`
	GapScore       float64        `json:"gap_score"`
	PlanStepScore  float64        `json:"plan_step_score"`
	OverallScore   float64        `json:"overall_score"`
	ModuleDetails  ModuleDetails  `json:"module_details"`
	BoundaryDetail BoundaryDetail `json:"boundary_details"`
	GapDetail      GapDetail      `json:"gap_details"`
	PlanDetail     PlanDetail     `json:"plan_details"`
	RawSpecs       []SpecSummary  `json:"raw_specs"`
}

type ModuleDetails struct {
	Unanimous []string         `json:"unanimous"`
	Partial   []PartialModule  `json:"partial,omitempty"`
	Unique    []UniqueModule   `json:"unique,omitempty"`
	CountRange [2]int          `json:"count_range"`
}

type PartialModule struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Total int    `json:"total"`
}

type UniqueModule struct {
	Name string `json:"name"`
	Run  int    `json:"run"`
}

type BoundaryDetail struct {
	CountRange [2]int `json:"count_range"`
}

type GapDetail struct {
	CountRange    [2]int   `json:"count_range"`
	AllQuestions   []string `json:"all_questions"`
}

type PlanDetail struct {
	CountRange [2]int `json:"count_range"`
}

type SpecSummary struct {
	Name       string   `json:"name"`
	Modules    []string `json:"modules"`
	Boundaries int      `json:"boundaries"`
	Gaps       int      `json:"gaps"`
	PlanSteps  int      `json:"plan_steps"`
}

func evaluate(specs []*analysis.ScaffoldSpec) Report {
	n := len(specs)

	// Module consistency
	moduleCounts := map[string]int{}
	for _, spec := range specs {
		for _, m := range spec.Modules {
			moduleCounts[m.Name]++
		}
	}

	var unanimous, partialNames []string
	var partial []PartialModule
	var unique []UniqueModule
	for name, count := range moduleCounts {
		switch {
		case count == n:
			unanimous = append(unanimous, name)
		case count == 1:
			// Find which run
			for i, spec := range specs {
				for _, m := range spec.Modules {
					if m.Name == name {
						unique = append(unique, UniqueModule{Name: name, Run: i + 1})
					}
				}
			}
		default:
			partialNames = append(partialNames, name)
			partial = append(partial, PartialModule{Name: name, Count: count, Total: n})
		}
	}
	sort.Strings(unanimous)

	totalModules := len(moduleCounts)
	moduleScore := 0.0
	if totalModules > 0 {
		moduleScore = float64(len(unanimous)) / float64(totalModules)
	}

	// Module count range
	minMods, maxMods := len(specs[0].Modules), len(specs[0].Modules)
	for _, spec := range specs[1:] {
		if len(spec.Modules) < minMods {
			minMods = len(spec.Modules)
		}
		if len(spec.Modules) > maxMods {
			maxMods = len(spec.Modules)
		}
	}

	// Boundary count range
	minBound, maxBound := len(specs[0].Boundaries), len(specs[0].Boundaries)
	for _, spec := range specs[1:] {
		if len(spec.Boundaries) < minBound {
			minBound = len(spec.Boundaries)
		}
		if len(spec.Boundaries) > maxBound {
			maxBound = len(spec.Boundaries)
		}
	}
	boundaryScore := 1.0
	if maxBound > 0 {
		boundaryScore = float64(minBound) / float64(maxBound)
	}

	// Gap consistency
	gapQuestions := map[string]int{}
	minGaps, maxGaps := len(specs[0].Gaps), len(specs[0].Gaps)
	for _, spec := range specs {
		for _, g := range spec.Gaps {
			gapQuestions[g.Question]++
		}
		if len(spec.Gaps) < minGaps {
			minGaps = len(spec.Gaps)
		}
		if len(spec.Gaps) > maxGaps {
			maxGaps = len(spec.Gaps)
		}
	}
	gapScore := 1.0
	if maxGaps > 0 {
		gapScore = float64(minGaps) / float64(maxGaps)
	}
	var allQuestions []string
	for q := range gapQuestions {
		allQuestions = append(allQuestions, q)
	}
	sort.Strings(allQuestions)

	// Plan step count range
	minSteps, maxSteps := len(specs[0].PlanSteps), len(specs[0].PlanSteps)
	for _, spec := range specs[1:] {
		if len(spec.PlanSteps) < minSteps {
			minSteps = len(spec.PlanSteps)
		}
		if len(spec.PlanSteps) > maxSteps {
			maxSteps = len(spec.PlanSteps)
		}
	}
	planScore := 1.0
	if maxSteps > 0 {
		planScore = float64(minSteps) / float64(maxSteps)
	}

	overall := (moduleScore + boundaryScore + gapScore + planScore) / 4.0

	// Raw spec summaries
	var raw []SpecSummary
	for _, spec := range specs {
		var mods []string
		for _, m := range spec.Modules {
			mods = append(mods, m.Name)
		}
		sort.Strings(mods)
		raw = append(raw, SpecSummary{
			Name:       spec.Name,
			Modules:    mods,
			Boundaries: len(spec.Boundaries),
			Gaps:       len(spec.Gaps),
			PlanSteps:  len(spec.PlanSteps),
		})
	}

	return Report{
		Runs:          n,
		ModuleScore:   moduleScore,
		BoundaryScore: boundaryScore,
		GapScore:      gapScore,
		PlanStepScore: planScore,
		OverallScore:  overall,
		ModuleDetails: ModuleDetails{
			Unanimous:  unanimous,
			Partial:    partial,
			Unique:     unique,
			CountRange: [2]int{minMods, maxMods},
		},
		BoundaryDetail: BoundaryDetail{
			CountRange: [2]int{minBound, maxBound},
		},
		GapDetail: GapDetail{
			CountRange:   [2]int{minGaps, maxGaps},
			AllQuestions:  allQuestions,
		},
		PlanDetail: PlanDetail{
			CountRange: [2]int{minSteps, maxSteps},
		},
		RawSpecs: raw,
	}
}

func newClient() *anthropic.Client {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	baseURL := "https://api.anthropic.com"

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	gatewayToken := os.Getenv("CLOUDFLARE_AI_GATEWAY_TOKEN")
	gatewayName := os.Getenv("CLOUDFLARE_AI_GATEWAY_NAME")
	if gatewayName == "" {
		gatewayName = "axon"
	}

	var opts []anthropic.Option
	if accountID != "" && gatewayToken != "" {
		baseURL = fmt.Sprintf(
			"https://gateway.ai.cloudflare.com/v1/%s/%s/anthropic",
			strings.TrimSpace(accountID),
			gatewayName,
		)
		opts = append(opts, anthropic.WithGatewayToken(gatewayToken))
	}

	return anthropic.NewClient(baseURL, apiKey, opts...)
}
