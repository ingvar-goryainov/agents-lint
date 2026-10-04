---
status: Accepted
owner: "ingvar-goryainov"
reviewers: []
updated_at: "2026-10-04"
feature_size: ""
ticket: ""
---

# 0003 — Keep agents-lint stateless with a versioned, strictly-decoded `.agents-lint.yaml` config

- **Status:** Accepted
- **Date:** 2026-10-04
- **Deciders:** ingvar-goryainov (project owner), during the `survey` greenfield foundation session

## Context

The linter reads files and writes a report; it has nothing to persist. The only durable contract
with users is the project config (which rules are on, severities, ignores). Once released, changing
its format breaks users, so the format and versioning are fixed now.

## Decision drivers

- No datastore, no migrations — keep the tool stateless.
- Config must be comfortable in CI and support comments.
- YAML is already required to parse skill/agent frontmatter — reuse one library.
- Config mistakes must fail loudly, not be silently ignored.

## Considered options

1. **YAML file `.agents-lint.yaml`, strict decode, `version` field** — golangci-lint-style.
2. **TOML file `agents-lint.toml`** — fewer syntax traps, but a second parsing dependency.
3. **No config file, flags only** — defers the decision; awkward in CI.

## Decision outcome

**Chosen:** Option 1. One YAML library (`go.yaml.in/yaml/v3`) serves config and frontmatter;
strict decoding (unknown keys = error) and an explicit `version: 1` guard against YAML pitfalls and
give a migration path for future format changes.

## Consequences

**Positive**
- Familiar to Go/CI users; comments allowed; single dependency.
- `version` makes breaking config changes explicit and detectable.

**Negative**
- YAML pitfalls (implicit typing such as `no` → false in YAML 1.1 parsers) — mitigated by typed structs + strict decode.
- Rule IDs and the schema become public contracts from the first release.

**Neutral**
- `migration_tool` stays empty; if a cache or datastore is ever needed it requires a new ADR.

## Links

- Architecture map: [[../architecture-map.md]] §Conventions (Configuration), §Datastores
- Related ADR: [[0002-pipeline-with-rule-registry]]
