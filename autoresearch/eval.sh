#!/bin/bash
# Fixed evaluation script for luthier self-improvement. DO NOT MODIFY.
# Runs luthier against its own PRD, scores the output.
#
# Output format:
#   compiles:    1
#   modules:     0.750
#   coherence:   0.857
#   ordering:    1.000
#   score:       0.827
#
# Score = compiles * 0.4 + modules * 0.3 + coherence * 0.2 + ordering * 0.1

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LUTHIER_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PORT="${LUTHIER_PORT:-8091}"
MODEL="${LUTHIER_MODEL:-/Users/benaskins/models/mlx/Qwen3.5-122B-A10B-5bit}"
OUTPUT_DIR="${LUTHIER_DIR}/coding-agent"

# Reference modules (what a correct analysis should select)
# Required: primitives needed for a CLI tool with LLM analysis
# Accepted: additional primitives that add value without being strictly required
REQUIRED_MODULES="axon-loop axon-talk axon-tool"
ACCEPTED_MODULES="axon-fact"

# Reference phases for plan ordering
REFERENCE_PHASES="scaffold|setup|repo analysis|structured|spec writer|template|file|generation gap|resolution|conversation wire|integrate|main"

# Clean previous output
rm -rf "$OUTPUT_DIR"

# Build luthier
echo ">>> Building luthier..." >&2
(cd "$LUTHIER_DIR" && go build -o bin/luthier ./cmd/luthier 2>&1) >&2
if [ $? -ne 0 ]; then
    echo ">>> Luthier build failed" >&2
    echo "compiles:    0"
    echo "modules:     0.000"
    echo "coherence:   0.000"
    echo "ordering:    0.000"
    echo "score:       0.000"
    exit 1
fi

# Run luthier
echo ">>> Running luthier against PRD..." >&2
run_output=$(cd "$LUTHIER_DIR" && LUTHIER_PROVIDER=local \
    LUTHIER_LOCAL_URL="http://localhost:${PORT}" \
    LUTHIER_MODEL="$MODEL" \
    bin/luthier fixtures/coding-agent-prd.md 2>&1) || true

# Check if output directory exists and compiles
compiles=0
if [ -d "$OUTPUT_DIR" ]; then
    echo ">>> Scaffold generated, checking compilation..." >&2
    if (cd "$OUTPUT_DIR" && go build ./... 2>&1) >&2; then
        compiles=1
        echo ">>> Compiles: YES" >&2
    else
        echo ">>> Compiles: NO" >&2
    fi
else
    echo ">>> No output directory generated" >&2
    echo "compiles:    0"
    echo "modules:     0.000"
    echo "coherence:   0.000"
    echo "ordering:    0.000"
    echo "score:       0.000"
    exit 1
fi

# Parse module selections from debug output
selected_modules=$(echo "$run_output" | grep "tool: select_module" | sed 's/.*select_module(\(.*\))/\1/' | sort)

# Module accuracy: all required modules must be selected.
# Accepted modules don't penalise. Other extras do penalise.
req_count=0
req_matched=0
for ref in $REQUIRED_MODULES; do
    req_count=$((req_count + 1))
    if echo "$selected_modules" | grep -q "^${ref}$"; then
        req_matched=$((req_matched + 1))
    fi
done

# Count extras that are neither required nor accepted
total_selected=$(echo "$selected_modules" | grep -c . || echo 0)
valid=$req_matched
for acc in $ACCEPTED_MODULES; do
    if echo "$selected_modules" | grep -q "^${acc}$"; then
        valid=$((valid + 1))
    fi
done
unwanted=$((total_selected - valid))
penalty=0
if [ "$unwanted" -gt 0 ] && [ "$total_selected" -gt 0 ]; then
    penalty=$(awk "BEGIN {printf \"%.3f\", $unwanted / ($total_selected + $req_count)}")
fi
modules_score=$(awk "BEGIN {s = $req_matched / $req_count - $penalty; if (s < 0) s = 0; printf \"%.3f\", s}")

echo ">>> Modules: selected=[$selected_modules] required=[$REQUIRED_MODULES] accepted=[$ACCEPTED_MODULES] unwanted=$unwanted score=$modules_score" >&2

# Plan coherence: do plan steps only reference selected modules?
plan_steps=$(echo "$run_output" | grep "tool: add_plan_step" | sed 's/.*add_plan_step(\(.*\))/\1/')
all_modules="axon axon-auth axon-eval axon-fact axon-gate axon-lens axon-look axon-loop axon-memo axon-mind axon-nats axon-synd axon-talk axon-task axon-tool axon-wire"

total_steps=0
coherent_steps=0
while IFS= read -r step; do
    [ -z "$step" ] && continue
    total_steps=$((total_steps + 1))
    step_lower=$(echo "$step" | tr '[:upper:]' '[:lower:]')
    is_coherent=1
    for mod in $all_modules; do
        if echo "$step_lower" | grep -q "$mod"; then
            if ! echo "$selected_modules" | grep -q "^${mod}$"; then
                is_coherent=0
                echo ">>> Incoherent step: '$step' references unselected module '$mod'" >&2
                break
            fi
        fi
    done
    coherent_steps=$((coherent_steps + is_coherent))
done <<< "$plan_steps"

coherence=1.000
if [ "$total_steps" -gt 0 ]; then
    coherence=$(awk "BEGIN {printf \"%.3f\", $coherent_steps / $total_steps}")
fi

# Plan ordering: are phases in the right sequence?
IFS=$'\n'
phase_indices=()
phase_num=0
for phase_pattern in $REFERENCE_PHASES; do
    phase_num=$((phase_num + 1))
    step_num=0
    found=false
    while IFS= read -r step; do
        [ -z "$step" ] && continue
        step_num=$((step_num + 1))
        step_lower=$(echo "$step" | tr '[:upper:]' '[:lower:]')
        # Check each keyword in the phase pattern (separated by |)
        for kw in $(echo "$phase_pattern" | tr '|' '\n'); do
            if echo "$step_lower" | grep -q "$kw"; then
                phase_indices+=("$step_num")
                found=true
                break 2
            fi
        done
    done <<< "$plan_steps"
done
unset IFS

ordering=1.000
if [ ${#phase_indices[@]} -ge 2 ]; then
    correct_pairs=0
    total_pairs=$(( ${#phase_indices[@]} - 1 ))
    for i in $(seq 0 $((total_pairs - 1))); do
        if [ "${phase_indices[$i]}" -le "${phase_indices[$((i+1))]}" ]; then
            correct_pairs=$((correct_pairs + 1))
        fi
    done
    ordering=$(awk "BEGIN {printf \"%.3f\", $correct_pairs / $total_pairs}")
fi

# Composite score
score=$(awk "BEGIN {printf \"%.3f\", $compiles * 0.4 + $modules_score * 0.3 + $coherence * 0.2 + $ordering * 0.1}")

echo "compiles:    $compiles"
echo "modules:     $modules_score"
echo "coherence:   $coherence"
echo "ordering:    $ordering"
echo "score:       $score"
