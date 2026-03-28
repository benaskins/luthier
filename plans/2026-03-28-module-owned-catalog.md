# Module-Owned Catalog — Build Plan
# 2026-03-28

Each axon module owns its luthier metadata (purpose, use_when, snippet).
`just sync-catalog` reads these manifests and generates both the system
prompt catalog and the snippet registry in luthier.

## Step 1 — Define luthier.yaml schema and add manifests to core modules

Add `luthier.yaml` to axon, axon-loop, axon-talk, axon-tool in ~/dev/lamina/.
Each manifest declares: name, purpose, use_when, deterministic, and snippet
(imports, requires, setup, deps, helpers).

Test: `cat ~/dev/lamina/axon-loop/luthier.yaml` is valid YAML.

Commit: `feat: add luthier.yaml manifests to core axon modules`

## Step 2 — Add manifests to capability modules

Add `luthier.yaml` to axon-fact, axon-task, axon-auth, axon-memo,
axon-eval, axon-nats, axon-lens, axon-mind, axon-wire, axon-synd,
axon-gate, axon-look in ~/dev/lamina/.

Test: all modules with a luthier.yaml parse without error.

Commit: `feat: add luthier.yaml manifests to capability axon modules`

## Step 3 — Build the sync-catalog generator in luthier

Add `cmd/luthier-sync/main.go` that:
1. Walks ~/dev/lamina/ for luthier.yaml files
2. Parses each manifest
3. Generates the module catalog table for system_prompt.txt
4. Generates `internal/snippets/generated.go` from the snippet fields
5. Writes both files

Add a `just sync-catalog` target that runs it.

Test: `just sync-catalog` succeeds, generated files compile.

Commit: `feat: implement sync-catalog generator`

## Step 4 — Remove hand-maintained snippets and system prompt table

Delete `internal/snippets/core.go` and `internal/snippets/capability.go`.
Replace the hand-written module catalog in `system_prompt.txt` with the
generated version. Update imports.

Test: `go test ./...` passes, `just build` succeeds.

Commit: `refactor: replace hand-maintained catalog with generated source`

## Step 5 — Verify end-to-end

Run `just sync-catalog && just build` then run luthier against the PM PRD.
Verify the scaffold compiles and module selections match.

Run `luthier-eval fixtures/coding-agent-prd.md 3` to confirm eval still works.

Commit: `test: verify generated catalog produces equivalent scaffolds`
