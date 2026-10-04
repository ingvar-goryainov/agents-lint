# agents-lint

A linter for AI-agent instruction files — `AGENTS.md`, `CLAUDE.md`, `.claude/skills/**/SKILL.md`,
`agents/*.md`, commands and plugin manifests. It finds these files in a repository, parses their
markdown and YAML frontmatter, runs a set of rules over them and reports findings with a
CI-friendly exit code.

> Status: early skeleton. Only `agents-lint version` exists so far; `check` and the first rules are next.

## Install

```sh
go install github.com/ingvar-goryainov/agents-lint/cmd/agents-lint@latest
```

## Build from source

Requires Go 1.26+ (and golangci-lint v2 for `make lint`).

```sh
make build   # go build ./...
make test    # go test -race ./...
make lint    # golangci-lint run ./...
go run ./cmd/agents-lint version
```

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | No findings at or above the `--fail-on` severity |
| 1 | Findings at or above `--fail-on` (default `error`) |
| 2 | Usage, config or runtime error |

## License

Apache-2.0 — see [LICENSE](LICENSE).
