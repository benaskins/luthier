# Maestro — Product Requirements Document

## Problem

Luthier produces a scaffolded project with a step-by-step implementation plan, but someone still needs to write the code. Each plan step describes what to build and how to test it, but a developer (or coding agent) must manually interpret and execute each step. This is the bottleneck — the architecture is decided, the plan is written, but the building is manual.

## Solution

A CLI orchestrator that takes a scaffolded project directory (produced by luthier) and conducts the implementation plan step by step. For each step, maestro delegates to a coding agent (Claude Code, or any tool that can read/write files and run commands), verifies the result mechanically and semantically, and commits if it passes. Maestro manages the plan lifecycle; the coding agent does the building.

## Target User

A developer who has used luthier to generate a project scaffold and wants to go from scaffold to working application without manually implementing each plan step. They review the commits after maestro finishes, not during.

## User Stories

**As a developer**, I want to point maestro at my scaffolded project and have it work through the plan so that I get a working application with a clean commit history.

**As a developer**, I want maestro to stop and report when a step fails so that I can intervene on the hard parts while it handles the straightforward ones.

**As a developer**, I want each plan step to be a separate commit so that I can review, revert, or modify individual steps.

**As a developer**, I want maestro to verify that each step was implemented correctly, not just that it compiles.

## Requirements

### Must Have

1. Accept a project directory path as input via the command line
2. Read the implementation plan from the project's plans/ directory
3. For each plan step, in order:
   a. Read the step title, description, and commit message
   b. Delegate to a coding agent with the step description and project context
   c. Run the project's verification command (build, test)
   d. Review the git diff against the step description using an LLM conversation loop to verify the implementation matches what was asked for
   e. If both mechanical verification and semantic review pass, commit with the step's commit message
   f. If either fails, retry up to 3 times with the error/review feedback as context
   g. If all retries fail, stop and report which step failed and why
4. The coding agent must be configurable (default: Claude Code)
5. The review step must use an LLM conversation loop to assess whether the diff satisfies the step description
6. Support any project produced by luthier, regardless of language or framework

### Should Have

7. Resume from a specific step (skip already-completed steps)
8. Dry-run mode that shows what it would do without writing files
9. Verbose mode showing the agent's output for each step
10. Detect the verification command from the project's build tooling (justfile, Makefile, mix.exs, Gemfile)

### Won't Have (for now)

11. Interactive mode where the developer collaborates on each step
12. Parallel execution of independent steps
13. Automatic PR creation
14. Its own file/shell tools (delegates to the coding agent for all code changes)

## Architecture

Maestro is a conductor with two collaborators: a coder and a reviewer.

```
maestro (conductor)
    |
    +-- reads plan from plans/*.md
    |
    +-- for each step:
    |       |
    |       +-- delegates to coding agent ("implement step N...")
    |       |
    |       +-- runs verification command (build, test, compile)
    |       |
    |       +-- reviews diff via LLM conversation loop:
    |       |     "does this diff implement what the step asked for?"
    |       |
    |       +-- if both pass: git commit
    |       +-- if either fails: retry with feedback
    |
    +-- reports final status
```

The coding agent is swappable (Claude Code, a human, any tool that modifies files). The review judge uses an LLM conversation loop for semantic verification of the diff against the step description.

## Success Metrics

- Steps completed without human intervention: target 80%+
- Each committed step compiles and passes tests: 100%
- Time per step: under 5 minutes for typical steps

## Constraints

- Must work with any project produced by luthier, regardless of catalogue/framework
- Must not modify the implementation plan itself
- Must use conventional commits matching the plan's commit messages
- The coding agent handles all file operations; maestro only handles plan parsing, verification, review, and git
