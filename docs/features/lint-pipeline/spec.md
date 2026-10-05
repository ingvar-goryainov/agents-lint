---
status: Draft
owner: "ingvar-goryainov"
reviewers: ["Tech Lead", "Security Lead"]
updated_at: "2026-10-05"
feature_size: "M"
---

# Spec — lint-pipeline

> **Glossary:** [CONTEXT](../../../CONTEXT.md) (repo-root; no feature-scoped CONTEXT.md)
> **Reference module / docs / channels used:** [`docs/architecture-map.md`](../../architecture-map.md), [ADR-0001](../../adr/0001-go-cobra-single-binary-cli.md), [ADR-0002](../../adr/0002-pipeline-with-rule-registry.md), [ADR-0003](../../adr/0003-yaml-config-no-datastore.md), the scaffolded skeleton (`internal/rules`, `internal/cli`). No external channels.

## 1. Context

A repo maintainer who keeps AGENTS.md, CLAUDE.md, skills, agents, commands and plugin manifests in a repository has no way to ask "which of these files does a linter actually see, and are any of them broken?". agents-lint today is a skeleton: it boots and prints its version, but it cannot find, read or report on a single instruction file. Every future rule depends on that missing frame, and a malformed instruction file (an unclosed frontmatter block, a manifest that does not parse) currently reaches the agent runtime unnoticed — the maintainer finds out only when a skill silently fails to load.

Why now: the foundation and skeleton were just laid down (survey + scaffold, 2026-10-04), and the frame is the first thing every content rule needs. Fixing its public contracts — the file kinds, the configuration format, the machine-readable output shape, the meaning of each run outcome — before any rule ships is far cheaper than changing them after users depend on them.

The committed approach is **inventory-first**: the pipeline's first user-visible value is an honest inventory — how many instruction files of each file kind were found, how many paths were ignored, how many files could not be checked — reported on every run in both human and machine-readable form, together with integrity rules that turn unreadable or unparseable instruction files into ordinary, configurable findings. Competitive research found that none of four comparable tools (cclint, skill-lint, claudelint, the official plugin validator) covers all six file kinds, reports a per-kind inventory, or documents malformed files as configurable findings; the adversarial review's sharpest failure mode was a false green — "zero findings" when nothing was actually checked — which an always-present inventory directly counters. Success, per the interview, is that the maintainer can confirm the tool sees their files correctly.

Run outcomes used throughout this spec: **clean** (no finding at or above the fail threshold), **failed** (at least one finding at or above the fail threshold), **tool error** (the run could not be performed — invalid configuration, invalid invocation, missing target — and no findings report is produced). The fail threshold defaults to error severity.

Traceability: the configuration schema in `docs/architecture-map.md` (Conventions → Configuration) lists per-rule `options`; this spec defers them (§3), so `design` reconciles the map with that deferral.

## 2. Goals

- A repo maintainer can tell, from any single run, exactly what agents-lint looked at — per file kind, including what was ignored and what could not be checked — so a clean result is never ambiguous.
- No broken instruction file passes silently: every unreadable or unparseable instruction file surfaces as a finding, or — when the maintainer has switched that integrity rule off — still shows up in the inventory as not checked.

## 3. Non-goals

- **Content rules** (for example "a skill must have a description") — each content rule is its own feature built on this frame; shipping them here would blur whether the frame itself is correct.
- **Honoring the repository's ignore file for version control** — its semantics (nesting, negation) are a feature in their own right, and local-only files such as a personal CLAUDE.local.md are usually listed there and would silently vanish; explicit `ignore` globs plus built-in default exclusions cover the need for now.
- **Following symbolic links** — following links invites double counting, loops and reads outside the scanned repository; each file is evaluated once at its real location.
- **Per-rule options in the configuration** — no rule has options yet; in configuration version 1 an `options` setting under a rule is refused as unrecognized, and accepting it later is a backward-compatible addition that needs no version bump.
- **Additional output formats, inline suppression comments and automatic fixes** — the static-analysis exchange format, per-line disables and autofix all build on a stable findings model and come later.

## 4. User stories

### US-01: See what was discovered
**As a** repo maintainer
**I want** every run to report how many instruction files of each file kind agents-lint found, ignored or could not check
**So that** I can confirm the tool sees my files the way my agent runtime does

### US-02: Be warned when nothing was found
**As a** repo maintainer
**I want** a run that found no instruction files to say so plainly
**So that** I never mistake "nothing checked" for "nothing wrong"

### US-03: Catch broken instruction files
**As a** repo maintainer
**I want** unreadable or unparseable instruction files reported as findings with a location and a plain-language reason
**So that** I fix them before my agent runtime silently fails to load them

### US-04: Control which paths are checked
**As a** repo maintainer
**I want** vendored and version-control directories skipped by default and my own ignore patterns honored
**So that** the inventory and findings cover only files I am responsible for

### US-05: Tune rule severity
**As a** repo maintainer
**I want** to set each rule's severity (including off) in the repository's configuration, and have configuration mistakes rejected loudly
**So that** the tool matches my project's tolerance and a typo never silently changes behavior

### US-06: Script against the results
**As a** repo maintainer
**I want** a machine-readable report carrying the same inventory and findings, and a run outcome that reflects a fail threshold I choose
**So that** my own scripts and hooks can act on the result without parsing human text

## 5. Acceptance criteria

### AC-01 (US-01) — happy
**Given** a repo maintainer's repository contains one AGENTS.md, one CLAUDE.md, two skills, one agent, one command and one plugin manifest, all well-formed
**When** the repo maintainer runs a check over the repository
**Then** the system shows an inventory listing all six file kinds with counts 1, 1, 2, 1, 1, 1, reports zero findings, and the run ends clean

### AC-02 (US-01) — domain invariant: one file, one kind
**Given** the repository contains a README, a documentation page and a design note under a directory named like an agent folder but not at a recognized agent location
**When** the repo maintainer runs a check
**Then** none of those files appears in the inventory or in any finding, because only files matching a file kind are instruction files, and each instruction file is counted under exactly one kind

### AC-03 (US-01) — domain invariant: links are not followed
**Given** the repository's CLAUDE.md is a symbolic link to its AGENTS.md
**When** the repo maintainer runs a check
**Then** AGENTS.md is counted once, the link is reported in the inventory as a skipped link rather than as a second instruction file, and no finding is duplicated across the two paths

### AC-04 (US-01) — error: target does not exist
**Given** the repo maintainer names a location to check that does not exist or is not a directory
**When** the repo maintainer runs a check
**Then** the system refuses to run, states which location could not be used and why, produces no inventory or findings, and the run ends as a tool error

### AC-05 (US-02) — domain invariant: inventory is always reported
**Given** a repository with no instruction files, or one where every instruction file is excluded by ignore patterns
**When** the repo maintainer runs a check
**Then** the system states plainly that no instruction files were found, names the location that was scanned and how many paths were ignored, includes the same zero-count inventory in the machine-readable report, and the run ends clean

### AC-06 (US-03) — error: unparseable frontmatter
**Given** a skill whose opening metadata block is never closed, and a second skill that is well-formed
**When** the repo maintainer runs a check
**Then** the system reports one finding from the frontmatter integrity rule pointing at the line where the block opens, explains in plain language that the block is not closed, counts that skill as not checked in the inventory, still checks the second skill, and the run ends failed

### AC-07 (US-03) — error: unparseable manifest
**Given** a plugin manifest that is not valid structured data (for example a trailing comma)
**When** the repo maintainer runs a check
**Then** the system reports one finding from the manifest integrity rule with the line and column where parsing stopped and a plain-language reason, counts the manifest as not checked, and the run ends failed

### AC-08 (US-03) — authorization: file the maintainer cannot read
**Given** an instruction file that the repo maintainer's account has no permission to read
**When** the repo maintainer runs a check
**Then** the system reports a finding from the unreadable-file integrity rule naming the path and the reason (no permission) without disclosing any content, counts it as not checked, checks every other file, and does not abort the run

### AC-08b (US-03) — domain invariant: content that cannot be read as text
**Given** an instruction file that is binary, larger than the read limit, or not valid text
**When** the repo maintainer runs a check
**Then** the system reports a finding from the unreadable-file integrity rule naming the path and which of those reasons applies, counts it as not checked, finishes the run in its usual time, and checks every other file

### AC-09 (US-03) — domain invariant: well-formed is independent of line endings
**Given** a well-formed skill saved with Windows line endings and a leading byte-order mark
**When** the repo maintainer runs a check
**Then** the system reports no integrity finding for it, and any location reported for that file matches the line and column the repo maintainer's editor shows

### AC-10 (US-03) — domain invariant: integrity findings shield other rules
**Given** an instruction file with an integrity finding
**When** the repo maintainer runs a check
**Then** no other rule reports on that file, and the inventory counts it as not checked

### AC-11 (US-04) — happy: default exclusions
**Given** the repository has instruction files inside its version-control metadata directory and inside third-party dependency directories, and no configuration
**When** the repo maintainer runs a check
**Then** those files are neither counted as instruction files nor reported in findings, and the inventory shows how many paths were excluded by default

### AC-12 (US-04) — cross-context: configured ignore patterns
**Given** the repository configuration lists an ignore pattern matching a test-fixtures directory that contains broken skills
**When** the repo maintainer runs a check
**Then** no finding is reported for files under that directory, the inventory counts them as ignored, and files outside it are checked as usual

### AC-13 (US-05) — cross-context: severity override keeps the inventory honest
**Given** the repository configuration sets the frontmatter integrity rule to off, and one skill has unclosed frontmatter
**When** the repo maintainer runs a check
**Then** no finding is reported for that skill, the run ends clean, and the inventory still counts that skill as not checked

### AC-14 (US-05) — error: unknown rule in configuration
**Given** the repository configuration sets a severity for a rule name agents-lint does not know (for example a misspelling)
**When** the repo maintainer runs a check
**Then** the system refuses to run, names the unknown rule and, when one is close, the known rule it most resembles, produces no findings report, and the run ends as a tool error

### AC-15 (US-05) — error: invalid configuration
**Given** the repository configuration contains an unrecognized setting, an unsupported configuration version, or a severity that is not one of off, info, warning or error
**When** the repo maintainer runs a check
**Then** the system refuses to run, names the offending setting or value and its location in the configuration, and the run ends as a tool error

### AC-16 (US-06) — happy: machine-readable report
**Given** a repository with one unparseable skill and three well-formed instruction files
**When** the repo maintainer asks for the machine-readable report
**Then** the report carries the same inventory and the same single finding (rule, severity, path, line, column, message) as the human report, declares the version of its own format, and lists findings in a stable order

### AC-17 (US-06) — happy: fail threshold
**Given** a run whose only finding has warning severity
**When** the repo maintainer runs a check with the default fail threshold, and again with the fail threshold set to warning
**Then** the first run ends clean and the second ends failed, while both show the same finding

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Run time p95, typical repository (≤ 1,000 instruction files, ≤ 50,000 paths after exclusions) | ≤ 1 s | benchmark fixture in CI, p95 over 20 runs |
| Determinism | 100% identical output (human and machine-readable) across repeated runs on the same tree | e2e test runs each fixture twice and compares byte-for-byte |
| Cross-platform consistency | identical inventory and findings for the same tree on Linux, macOS and Windows | e2e fixtures run on all three in CI |
| Outcome taxonomy | 100% of configuration/invocation problems end as tool error, never as findings; 100% of per-file read/parse problems end as findings, never as tool error | e2e fixture suite covering each problem class |
| Network access | 0 network calls during a run | review + test with network disabled |
| Extensibility | adding a new rule changes exactly 1 file outside that rule's own package (the registry line) | measured on the first content rule's change (see §7) |

## 6.1 Security / privacy

- **Data classification:** internal — the tool reads the scanned repository's files locally and emits paths and short messages; nothing leaves the machine.
- **Personal data touched:** none — no new stored fields; file content is read only in memory and never echoed beyond the location of a problem.
- **AuthZ/AuthN impact:** no new capabilities; the run has exactly the file-read permissions of the account that starts it, and a file it cannot read becomes an integrity finding, not a crash.
- **Abuse cases:**
  - Link escaping the repository (a link to a private file elsewhere on the machine): links are never followed, so content outside the scanned tree is never read or reported.
  - Hostile metadata (deeply nested or self-expanding structured data in frontmatter or a manifest): parsing is bounded; the file gets a finding from the frontmatter integrity rule (markdown kinds) or the manifest integrity rule (plugin manifests) instead of exhausting memory or hanging the run.
  - Control characters in file names or content injected into messages: terminal control sequences are neutralized in human output so a crafted file cannot rewrite the maintainer's terminal.
  - Very large or binary files with an instruction-file name: read size is bounded; oversized, binary or non-text files get a finding from the unreadable-file integrity rule.
- **Security review:** Required — M feature that parses untrusted input from arbitrary repositories.

## 7. Metrics / KPIs

- **Kind-recognition accuracy** on a curated corpus of real-world repositories with hand-labeled instruction files — baseline: 0 (no discovery exists), target: 100% correctly classified, 0 non-instruction files counted, at feature completion.
- **False integrity findings** on the same corpus's well-formed files — baseline: n/a, target: 0 at feature completion.
- **Cost of the next rule** — files changed outside the new rule's own package when the first content rule is added — baseline: n/a, target: exactly 1 (the registry line), measured on that rule's change.
- **Dogfooding** — agents-lint checks its own repository in its CI — baseline: not running, target: running and clean within 7 days of merge.

## 8. Open questions

- [ ] Which exact locations and names map to each file kind (nested CLAUDE.md in subdirectories, CLAUDE.local.md, project-level vs plugin-level agents and commands folders, marketplace manifests, case variants such as claude.md), and does a leading `---` in AGENTS.md / CLAUDE.md count as frontmatter? Default now: follow the agent runtimes' documented layouts, names are case-sensitive, AGENTS.md / CLAUDE.md have no frontmatter (a leading `---` is body). — owner: Tech Lead, due: before `sdd:design`
- [ ] Should a run that found no instruction files be able to end failed (opt-in), so scripts that look only at the outcome can catch a false green? Default now: always ends clean, with the explicit "nothing found" statement. — owner: Tech Lead, due: before the first public release
- [ ] How should a configuration written for a newer agents-lint (naming rules this binary does not know) be handled by an older pinned binary? Default now: refused as a tool error (AC-14). — owner: Tech Lead, due: before the first public release
- [ ] Should shared skill directories linked into the repository be checkable via an opt-in? Default now: links are never followed (§3). — owner: Tech Lead, due: before the first public release
