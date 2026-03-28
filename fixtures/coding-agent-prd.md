# Coding Agent — Product Requirements Document

## Problem

Developers starting new projects spend hours on boilerplate: setting up project structure, choosing dependencies, wiring configuration, writing build scripts. For teams using a shared component library, this is worse — they need to know which components exist, when to use each one, and how they fit together. Most get it wrong on the first try.

## Solution

A CLI tool that takes a product requirements document and produces a fully scaffolded, buildable project. The developer describes *what* they want to build in plain English. The tool figures out the architecture, selects the right components, and writes the project files.

## Target User

A developer on the team who knows what they want to build but doesn't have deep familiarity with every component in the shared library. They have a markdown document describing their project and want to go from that to working code as fast as possible.

## User Stories

**As a developer**, I want to run a single command with my PRD so that I get a project directory I can immediately build and start iterating on.

**As a developer**, I want the tool to ask me about genuine ambiguities in my requirements rather than guessing, so that the output reflects my actual intent.

**As a developer**, I want the generated project to include a step-by-step implementation plan so that I (or an AI coding assistant) can build out the business logic incrementally.

**As a developer**, I want the generated project to document which components were selected and why, so I can understand the architectural decisions without reading the source.

## Requirements

### Must Have

1. Accept a markdown file as input via the command line
2. Analyse the document and determine which components and patterns are needed
3. If the requirements are ambiguous, ask clarifying questions before proceeding
4. Generate a buildable project directory with:
   - Working build and test configuration
   - Dependency declarations for selected components
   - A stub entry point that wires the components together (structure only, no business logic)
   - An architecture document explaining the component selections and design boundaries
   - A step-by-step implementation plan with commit-sized steps
5. Verify the generated project compiles before reporting success

### Should Have

6. Static analysis context (component catalog, established patterns) should be cached across invocations to reduce cost
7. Support for evaluating output quality across different LLM providers

### Won't Have (for now)

8. Generating business logic — the tool produces scaffolding only
9. Registering the new project in any central catalog
10. A graphical interface — CLI only

## Success Metrics

- Time from PRD to first successful build: under 2 minutes
- Generated projects compile on first try: 100%
- Developer can understand architecture decisions without asking someone: yes/no qualitative

## Constraints

- Must work offline except for the LLM API call
- Must not require any credentials beyond an LLM API key
- Generated projects must follow the team's existing conventions for file naming, build tooling, and documentation structure
