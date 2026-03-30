# autoresearch: luthier self-improvement

You are improving luthier until it can produce a working copy of itself from a PRD.

Luthier is a CLI tool that reads a markdown PRD, analyses it with an LLM, and generates a buildable Go project scaffold. The goal is to improve luthier's system prompt until the generated scaffold reliably compiles with correct module selection, coherent plan steps, and proper ordering.

## Phase 1: prompt tuning

You may only edit `internal/patterns/system_prompt.txt`. This is the system prompt sent to the LLM during analysis. It contains the module catalog, established patterns, file conventions, and instructions for the model.

## Setup

1. Read this file and `autoresearch/eval.sh` for full context.
2. Read `internal/patterns/system_prompt.txt` (the file you will edit).
3. Read `fixtures/coding-agent-prd.md` (the PRD used for evaluation).
4. Read `cmd/luthier/main.go` to understand the pipeline.
5. Create a branch: `git checkout -b autoresearch/<tag>` from current HEAD.
6. Run `bash autoresearch/eval.sh` to establish the baseline. Record the results.
7. Initialize `autoresearch/results.tsv` with the header and baseline row.
8. Begin the experiment loop.

## The experiment loop

LOOP:

1. Review `autoresearch/results.tsv` to understand what has been tried and what worked.
2. If a previous run failed, read the eval log to understand WHY. Common issues:
   - Model selected wrong modules: improve catalog descriptions or "Use when" column.
   - Model selected experimental modules: strengthen the classification guidance.
   - Plan steps reference unselected modules: add explicit instruction about coherence.
   - Output didn't compile: check if snippet dependencies are described in patterns.
   - Model didn't call finalize: check if instructions about tool calling order are clear.
3. Form a hypothesis about what prompt change would fix the issue.
4. Edit `internal/patterns/system_prompt.txt`.
5. `git add internal/patterns/system_prompt.txt && git commit -m "experiment: <description>"`
6. Run: `bash autoresearch/eval.sh > autoresearch/run.log 2>&1`
7. Read results: `grep "^score:\|^compiles:\|^modules:\|^coherence:\|^ordering:" autoresearch/run.log`
8. If grep output is empty, the run failed. Run `tail -50 autoresearch/run.log` to diagnose.
9. Record the results in `autoresearch/results.tsv`.
10. If score improved (higher), KEEP the commit. The branch advances.
11. If score is equal or worse, DISCARD: `git reset --hard HEAD~1`
12. Go to step 1.

## Metric

    score = compiles * 0.4 + modules * 0.3 + coherence * 0.2 + ordering * 0.1

- **compiles** (0 or 1): does the generated project pass `go build`?
- **modules** (0-1): fraction of reference modules selected, penalised for extras. Reference set: axon-loop, axon-talk, axon-tool (the PRD describes a CLI tool with LLM analysis).
- **coherence** (0-1): fraction of plan steps that only reference modules actually selected.
- **ordering** (0-1): are plan phases in the correct sequence (setup, analysis, writer, gaps, integration)?

A perfect score of 1.0 means: compiles, selects exactly the right modules, plan steps are coherent, and phases are ordered correctly.

## What you CAN edit

Only `internal/patterns/system_prompt.txt`. This file contains:
- Role description and task framing
- Core principle (deterministic vs non-deterministic)
- Module catalog (name, class, purpose, use-when)
- Established patterns table
- Boundary classification guidance
- File conventions
- Plan step format instructions
- Gap raising criteria

## What you CANNOT edit

- `autoresearch/eval.sh`: the evaluation is fixed.
- `cmd/luthier/main.go`: the pipeline is fixed.
- `internal/analysis/`: analysis loop and tools are fixed.
- `internal/snippets/`: snippet templates are fixed (Phase 2).
- `internal/writer/`: file writer is fixed (Phase 2).
- `fixtures/coding-agent-prd.md`: the test PRD is fixed.

## Key context

The system prompt is sent as the first message to the LLM. The LLM then receives the PRD as a user message and must call tools in order: select_module (multiple), define_boundary (multiple), add_plan_step (multiple), optionally raise_gap, then finalize.

The LLM is Qwen3.5-122B-A10B-5bit served via vllm-mlx on port 8091. It supports tool calling via the qwen3_coder parser. It has a thinking mode that generates `<think>` blocks before tool calls.

The PRD describes a CLI tool (not an HTTP service). The correct module set is axon-loop + axon-talk + axon-tool. The model should NOT select axon (HTTP only), axon-fact, axon-memo, axon-mind, or other modules not needed for a CLI with LLM analysis.

## Logging results

Tab-separated `autoresearch/results.tsv`:

```
commit	score	compiles	modules	coherence	ordering	status	description
a1b2c3d	0.827	1	0.750	0.857	1.000	keep	baseline
```

## NEVER STOP

Once the experiment loop has begun, do NOT pause to ask the human if you should continue. You are autonomous. If you run out of ideas, read the model's thinking trace in `autoresearch/run.log` for clues about what confused it. Try rewording, reordering, adding examples, or removing ambiguity. The loop runs until the human interrupts you.
