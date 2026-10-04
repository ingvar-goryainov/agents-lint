---
status: Accepted
owner: "ingvar-goryainov"
reviewers: []
updated_at: "2026-10-04"
feature_size: ""
ticket: ""
---

# 0001 — Build agents-lint as a single Go binary with cobra

- **Status:** Accepted
- **Date:** 2026-10-04
- **Deciders:** ingvar-goryainov (project owner), during the `survey` greenfield foundation session

## Context

agents-lint is a new, empty repository for a linter of AI-agent instruction files (AGENTS.md,
CLAUDE.md, skills, agents). It must run locally and in CI with no setup. The stack and CLI framework
shape every package and test that follows, so they are fixed before any code exists.

## Decision drivers

- Zero-runtime install in CI and on dev machines (one binary, no interpreter).
- Multiple subcommands expected (`check`, `rules`, `init`, `version`) with help + shell completion.
- The repo was initialised with the Go `.gitignore` template; Go 1.26 + golangci-lint v2 already on the owner's machine.

## Considered options

1. **Go + spf13/cobra** — mainstream Go CLI framework with subcommands, completion and docs generation.
2. **Go + urfave/cli v3** — lighter declarative CLI library; smaller ecosystem, recent v3 API churn.
3. **Go + stdlib `flag`** — no dependencies; subcommands, help and completion hand-written.
4. **TypeScript/Node CLI** — `npx`-friendly, but needs a Node runtime and replaces the Go setup.

## Decision outcome

**Chosen:** Option 1 (Go + cobra). A single static binary meets the zero-install driver, and cobra
gives subcommands/completion out of the box with the widest familiarity among Go CLIs.

## Consequences

**Positive**
- `go install` / one downloaded binary; trivially usable in CI.
- Subcommands, `--help`, shell completion and man/doc generation for free.

**Negative**
- cobra + pflag add dependencies and some boilerplate per command.
- JS-ecosystem users can't `npx` it without a wrapper package.

**Neutral**
- Swapping the CLI library later touches only `internal/cli` (commands are thin), not rules.

## Links

- Architecture map: [[../architecture-map.md]] §Stack
- Related ADR: [[0002-pipeline-with-rule-registry]]
