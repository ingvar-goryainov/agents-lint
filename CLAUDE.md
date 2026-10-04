# agents-lint — working notes for agents

A Go CLI that lints AI-agent instruction files (`AGENTS.md`, `CLAUDE.md`, `.claude/skills/**/SKILL.md`,
`agents/*.md`, commands, plugin manifests). Architecture reference: `docs/architecture-map.md` + `docs/adr/`.

## Commands

| What  | Command                    | Make         |
|-------|----------------------------|--------------|
| Build | `go build ./...`           | `make build` |
| Test  | `go test -race ./...`      | `make test`  |
| Lint  | `golangci-lint run ./...`  | `make lint`  |

Run the CLI locally: `go run ./cmd/agents-lint <command>`. Go ≥ 1.26, golangci-lint v2 (`.golangci.yml`, gofumpt formatting).

## Package layout

Pipeline: discover → parse → check → report. Everything except the entry point is under `internal/`.

- `cmd/agents-lint/` — `main` only: `os.Exit(cli.Execute(ctx, args, stdout, stderr))`.
- `internal/cli/` — cobra root (`root.go`) + one file per subcommand; exit codes in `exit.go`.
- `internal/config/` — `.agents-lint.yaml` loading (strict decode, `version: 1`).
- `internal/discover/` — walks an `fs.FS`, classifies files by kind, applies ignore globs.
- `internal/document/` — typed `Document` model + markdown/frontmatter parser (pure).
- `internal/rules/` — `Rule` interface, `Finding`, `Severity`, explicit registry; one package per rule in `internal/rules/<ruleid>/`.
- `internal/engine/` — orchestrates discover → parse → rules → sorted findings.
- `internal/report/` — text / JSON formatters writing to an `io.Writer`.

Dependency direction: `cli → engine → {discover, document, rules}`, `cli → {config, report}`;
`document` imports nothing internal, `rules` imports only `document`. Packages not yet present are created by the first feature that needs them.

## Rules — the non-negotiables

- **Explicit registry.** Every rule is listed by hand in `internal/rules/registry.go` (`All()`). No `init()` registration.
- **Rule purity.** A rule sees only the parsed `*document.Document` (plus context passed by `engine`). No filesystem, network, env or globals. Filesystem access belongs to `discover`, via `fs.FS` (tests use `fstest.MapFS`).
- **Stable IDs.** Rule IDs are kebab-case (`skill-missing-description`) and are a public config contract — never rename or renumber without a deprecation path.
- **Config is a public contract.** Unknown keys are errors; breaking changes bump `version`.

## Errors, findings, exit codes, output

- Wrap errors with context: `fmt.Errorf("load config %s: %w", path, err)`; sentinel errors checked with `errors.Is`.
- Lint findings are **data, not errors**: `rules.Finding{RuleID, Severity, Path, Line, Col, Message}`.
- Exit codes (`internal/cli/exit.go`): `0` clean · `1` findings at/above `--fail-on` (default `error`) · `2` usage/config/runtime error.
- Findings → stdout via `internal/report`; diagnostics → stderr. Output is deterministic: sort by path, line, col, rule ID.

## Tests

- stdlib `testing`, table-driven, `t.Parallel()` where safe.
- **Rules:** fixtures in `internal/rules/<ruleid>/testdata/*.md` with golden expected findings (`*.golden`).
  Regenerate with `go test ./... -update`, then review the golden diff like code — never accept it blindly.
- **CLI e2e:** `testscript` scripts in `internal/cli/testdata/script/*.txtar`, driven by `internal/cli/cli_test.go`.
  `version.txtar` is the skeleton smoke test.

## Where new things go

- New rule → `internal/rules/<ruleid>/{rule.go,rule_test.go,testdata/}` + one line in `registry.go`.
- New file kind → classifier in `internal/discover` + a `document.Kind` constant.
- New subcommand → `internal/cli/<name>.go` + a `.txtar` script (precedent: `version.go`).
- New output format → `internal/report/<format>.go`.
