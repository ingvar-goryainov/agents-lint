---
status: Accepted
owner: "ingvar-goryainov"
reviewers: []
updated_at: "2026-10-04"
feature_size: ""
ticket: ""
---

# 0002 — Structure the linter as a discover → parse → check → report pipeline with an explicit rule registry

- **Status:** Accepted
- **Date:** 2026-10-04
- **Deciders:** ingvar-goryainov (project owner), during the `survey` greenfield foundation session

## Context

Most future work on agents-lint is "add rule X" or "support file kind Y". The code layout decides
how many files a new rule touches and whether rules can be tested in isolation. Changing the style
later amounts to a rewrite, so it is fixed in the foundation.

## Decision drivers

- Adding a rule should be one package + one registry line.
- Rules must be unit-testable from fixtures without the CLI or a real filesystem.
- Deterministic output (CI diffs, golden tests).

## Considered options

1. **Pipeline + rule registry** — `discover` (fs.FS) → `document` (typed model) → `rules` (interface + explicit registry) → `report`, orchestrated by `engine`, fronted by `cli`.
2. **Hexagonal (ports & adapters)** — domain/app/infra/ports layers with interfaces at each boundary.
3. **Flat single package** — everything in `internal/lint`.

## Decision outcome

**Chosen:** Option 1. It is the idiomatic shape for Go linters (cf. go/analysis), keeps rules pure
over a parsed model, and gives a one-package-per-rule growth path — without the extra interface
layers hexagonal would add to a stateless CLI.

## Consequences

**Positive**
- New rule = `internal/rules/<ruleid>/` + one line in `registry.go`; tested via `testdata/` + golden files.
- `fs.FS` at the discovery boundary makes the whole engine testable with `fstest.MapFS`.

**Negative**
- Discipline required: rules must not read files or globals — only the `Document` they're given.
- Explicit registry is a hand-maintained list (chosen over `init()` magic for greppability).

**Neutral**
- A future HTTP/service front-end can reuse `engine` directly; exposing rules to third parties would need a `pkg/` API and a new ADR.

## Links

- Architecture map: [[../architecture-map.md]] §Module inventory, §Conventions
- Related ADR: [[0001-go-cobra-single-binary-cli]], [[0003-yaml-config-no-datastore]]
