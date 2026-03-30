#!/bin/bash
# Run luthier autoresearch as short-lived Claude sessions in a dedicated worktree.

set -euo pipefail

ROUNDS=${1:-5}
EXPERIMENTS=${2:-10}
WORKTREE="/Users/benaskins/dev/luthier-autoresearch"
DATE=$(date +%Y-%m-%d)

if [ ! -d "$WORKTREE" ]; then
    echo "ERROR: Worktree not found at $WORKTREE"
    echo "Create it: cd /Users/benaskins/dev/luthier && git worktree add $WORKTREE -b autoresearch/phase1"
    exit 1
fi

RESULTS="$WORKTREE/autoresearch/results.tsv"
SESSION_LOG="$WORKTREE/autoresearch/session-${DATE}.log"

echo "=== Luthier autoresearch: ${ROUNDS} rounds x ${EXPERIMENTS} experiments ==="
echo "Date: $DATE"
echo "Worktree: $WORKTREE"
echo ""

for round in $(seq 1 "$ROUNDS"); do
    echo "=========================================="
    echo "Round ${round}/${ROUNDS} ($(date))"
    echo "=========================================="

    prior_count=0
    if [ -f "$RESULTS" ]; then
        prior_count=$(tail -n +2 "$RESULTS" | wc -l | tr -d ' ')
    fi
    echo "Prior experiments: $prior_count"

    timeout 20m claude --cwd "$WORKTREE" -p "Read autoresearch/program.md and follow the instructions.
Read autoresearch/results.tsv to see what has already been tried (${prior_count} prior experiments).
Today's date is ${DATE}. This is round ${round} of ${ROUNDS}.

IMPORTANT: You are working in a dedicated worktree at ${WORKTREE}.
All git operations (commit, reset) happen here without affecting the main tree.

Run exactly ${EXPERIMENTS} experiments, then stop. Do NOT loop forever.
After ${EXPERIMENTS} experiments, print 'ROUND COMPLETE' and stop." \
        --allowedTools "Bash,Read,Write,Edit,Grep,Glob" \
        >> "$SESSION_LOG" 2>&1 || true

    new_count=0
    if [ -f "$RESULTS" ]; then
        new_count=$(tail -n +2 "$RESULTS" | wc -l | tr -d ' ')
    fi
    added=$(( new_count - prior_count ))
    echo "Experiments this round: $added (total: $new_count)"
    echo ""
done

echo "=== Luthier autoresearch complete ==="
echo "Results: $RESULTS"
echo ""
echo "To review changes: cd $WORKTREE && git log --oneline autoresearch/phase1 ^main"
echo "To merge wins:     cd /Users/benaskins/dev/luthier && git merge autoresearch/phase1"
