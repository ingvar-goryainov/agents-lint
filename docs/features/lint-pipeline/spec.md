---
status: Draft
owner: "ingvar-goryainov"
reviewers: ["Tech Lead", "Security Lead"]
updated_at: "2026-10-08"
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

Contract terms used throughout this spec:

- **Scan target.** Each run checks exactly one directory — the current directory unless the maintainer names another. Configuration is read only from the target's root, or from a file the maintainer names explicitly at invocation; there is no search in parent directories. An explicitly named configuration file that does not exist is a tool error.
- **Paths.** Every path in both reports, and every ignore pattern, is relative to the scan target and uses `/` as the separator on every platform. The scanned location is shown as the maintainer gave it (`.` by default), never as an absolute path. A path's value is the same in both reports; the human report only renders control characters as visible escapes, the machine-readable report uses its format's standard escaping.
- **Locations.** Line and column are 1-based. A column counts Unicode characters, not bytes; a tab is one column; a leading byte-order mark and the carriage return of a Windows line ending are not counted. Text means valid UTF-8 (a leading byte-order mark is allowed); any other encoding, UTF-16 included, is "not valid text".
- **Fail threshold.** Set only at invocation, to one of info, warning or error (default error); any other value is a tool error.
- **Report vs diagnostics.** On a tool error the report stream stays empty in both human and machine-readable modes; the explanation goes to the diagnostics stream, and the run outcome is the signal scripts act on. The "no instruction files were found" statement is part of the report, not a diagnostic.
- **Integrity rules.** Three built-in rules, all defaulting to error severity; their IDs are part of the configuration contract:
  - frontmatter integrity rule — `frontmatter-unparseable`: a markdown kind's opening metadata block is never closed, is not valid structured data, or its top level is not a set of key-value pairs. A file with no metadata block at all is not an integrity problem (whether a kind requires one is a content rule's concern).
  - manifest integrity rule — `manifest-unparseable`: a plugin manifest is not valid structured data.
  - unreadable-file integrity rule — `file-unreadable`: a file or a directory inside the scan target cannot be read, or a file is binary, larger than 1 MiB, or not valid text.

  A file gets at most one integrity finding: the first problem found, with the unreadable-file rule taking precedence over the other two. The integrity check always runs — a rule's severity, including off and info, only decides whether and how its finding is reported; the file is counted as not checked and shielded from every other rule either way. Integrity messages are fixed plain-language reasons plus a location; they never quote file content and never pass a parser's raw error text through.
- **Inventory counts.** For each of the six file kinds: instruction files found (including those not checked) and, of those, how many were not checked. Plus totals not split by kind: paths excluded by default, paths ignored by configuration, skipped links, unreadable directories. An excluded or ignored directory counts as one path and is not walked into. A skipped link is any symbolic link the walk reaches (to a file or a directory, outside excluded and ignored paths); the paths of skipped links are listed in both reports.
- **Default exclusions.** Directories named `.git`, `.hg`, `.svn`, `node_modules` or `vendor`, at any depth.
- **Ignore patterns.** Matched against the target-relative path: `*` matches within one path segment, `**` matches any number of segments, `?` and `[...]` as usual; a pattern that matches a directory excludes everything beneath it. Negation is not supported — a pattern starting with `!` is a configuration error, as is a pattern that does not parse.

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
- **Several scan targets, or single files as a target, in one run** (for example a pre-commit hook passing the changed files) — version 1 checks one directory per run.
- **Fail threshold in the configuration** — version 1 accepts it only at invocation; a configuration setting for it can be added later as a backward-compatible change.
- **Turning default exclusions off, negated ignore patterns, and configurable input limits** — none is accepted in configuration version 1; each can be added later as a backward-compatible change.

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
**Then** AGENTS.md is counted once, the link is counted in the inventory's skipped-links total and listed by its path rather than counted as a second instruction file, and no finding is duplicated across the two paths

### AC-03b (US-01) — domain invariant: linked directories are not walked
**Given** the repository contains a symbolic link to a directory holding three skills
**When** the repo maintainer runs a check
**Then** the link is counted once in the skipped-links total and listed by its path, nothing beneath it is read or counted as an instruction file, and the run outcome is decided by the other files only

### AC-04 (US-01) — error: target does not exist
**Given** the repo maintainer names a location to check that does not exist or is not a directory, or explicitly names a configuration file that does not exist
**When** the repo maintainer runs a check
**Then** the system refuses to run, states on the diagnostics stream which location could not be used and why, leaves the report stream empty with no inventory or findings, and the run ends as a tool error

### AC-05 (US-02) — domain invariant: inventory is always reported
**Given** a repository with no instruction files, or one where every instruction file is excluded by ignore patterns
**When** the repo maintainer runs a check
**Then** the report states plainly that no instruction files were found, names the location that was scanned and how many paths were excluded by default and ignored by configuration, includes the same zero-count inventory in the machine-readable report, and the run ends clean

### AC-06 (US-03) — error: unparseable frontmatter
**Given** a skill whose opening metadata block is never closed, and a second skill that is well-formed
**When** the repo maintainer runs a check
**Then** the system reports one finding from the frontmatter integrity rule pointing at the line where the block opens, explains with a fixed plain-language reason that quotes no content of the file that the block is not closed, shows two skills found with one of them not checked, still checks the second skill, and the run ends failed

### AC-06b (US-03) — error: closed but malformed frontmatter
**Given** one skill whose metadata block is closed but is not valid structured data, a second skill whose metadata block is a list rather than key-value pairs, and a third skill with no metadata block at all
**When** the repo maintainer runs a check
**Then** the first two skills each get exactly one finding from the frontmatter integrity rule at the first problem found and are counted as not checked, the third skill gets no integrity finding and is counted as checked, and the run ends failed

### AC-07 (US-03) — error: unparseable manifest
**Given** a plugin manifest that is not valid structured data (for example a trailing comma)
**When** the repo maintainer runs a check
**Then** the system reports one finding from the manifest integrity rule with the line and column where parsing stopped and a fixed plain-language reason that quotes no content of the file, counts the manifest as not checked, and the run ends failed

### AC-08 (US-03) — authorization: file the maintainer cannot read
**Given** an instruction file that the repo maintainer's account has no permission to read
**When** the repo maintainer runs a check
**Then** the system reports a finding from the unreadable-file integrity rule naming the path and the reason (no permission) without disclosing any content, counts it as not checked, checks every other file, does not abort the run, and the run ends failed

### AC-08b (US-03) — domain invariant: content that cannot be read as text
**Given** an instruction file that is binary, larger than 1 MiB, or not valid text (for example saved as UTF-16)
**When** the repo maintainer runs a check
**Then** the system reports a finding from the unreadable-file integrity rule naming the path and which of those reasons applies, counts it as not checked, stays within the per-file processing bound (§6), checks every other file, and the run ends failed

### AC-08c (US-03) — authorization: directory the maintainer cannot read
**Given** a directory inside the scan target that the repo maintainer's account has no permission to open
**When** the repo maintainer runs a check
**Then** the system reports a finding from the unreadable-file integrity rule naming the directory's path and the reason, with no line or column, counts it in the inventory's unreadable-directories total, checks everything outside that directory, and the run ends failed

### AC-09 (US-03) — domain invariant: well-formed is independent of line endings
**Given** a well-formed skill saved with Windows line endings and a leading byte-order mark
**When** the repo maintainer runs a check
**Then** the system reports no integrity finding for it, and any location reported for that file has the same line and column as for the same file saved with Unix line endings and no byte-order mark

### AC-10 (US-03) — domain invariant: integrity findings shield other rules
**Given** an instruction file that fails an integrity check, whatever that integrity rule's severity (including off and info)
**When** the repo maintainer runs a check
**Then** no other rule reports on that file, and the inventory counts it as not checked

### AC-11 (US-04) — happy: default exclusions
**Given** the repository has instruction files inside its version-control metadata directory and inside third-party dependency directories, and no configuration
**When** the repo maintainer runs a check
**Then** those files are neither counted as instruction files nor reported in findings, and the inventory shows how many paths were excluded by default, each excluded directory counted once

### AC-12 (US-04) — cross-context: configured ignore patterns
**Given** the repository configuration lists an ignore pattern matching a test-fixtures directory that contains broken skills
**When** the repo maintainer runs a check
**Then** no finding is reported for files under that directory, the inventory counts the directory once as a path ignored by configuration, and files outside it are checked as usual

### AC-13 (US-05) — cross-context: severity override keeps the inventory honest
**Given** the repository configuration sets the frontmatter integrity rule to off, and one skill has unclosed frontmatter
**When** the repo maintainer runs a check
**Then** no finding is reported for that skill, the run ends clean, and the inventory still counts that skill as not checked

### AC-14 (US-05) — error: unknown rule in configuration
**Given** the repository configuration sets a severity for a rule name agents-lint does not know (for example a misspelling)
**When** the repo maintainer runs a check
**Then** the system refuses to run, names the unknown rule and, when a known rule is within two character edits of it, the single closest known rule (the alphabetically first on a tie), produces no findings report, and the run ends as a tool error

### AC-15 (US-05) — error: invalid configuration
**Given** the repository configuration contains an unrecognized setting, an unsupported configuration version, a severity that is not one of off, info, warning or error, or an ignore pattern that is negated or does not parse
**When** the repo maintainer runs a check
**Then** the system refuses to run, names the offending setting or value and its location in the configuration, and the run ends as a tool error

### AC-16 (US-06) — happy: machine-readable report
**Given** a repository with one unparseable skill and three well-formed instruction files
**When** the repo maintainer asks for the machine-readable report
**Then** the report carries the same inventory and the same single finding (rule, severity, path, line, column, message) as the human report, with the path relative to the scanned location and separated by `/`, declares the version of its own format, and lists findings in a stable order

### AC-17 (US-06) — happy: fail threshold
**Given** a run whose only finding has warning severity
**When** the repo maintainer runs a check with the default fail threshold, and again with the fail threshold set to warning
**Then** the first run ends clean and the second ends failed, while both show the same finding

### AC-17b (US-06) — error: invalid fail threshold
**Given** a repository with well-formed instruction files
**When** the repo maintainer runs a check with a fail threshold other than info, warning or error
**Then** the system refuses to run, names the value and the accepted values, leaves the report stream empty, and the run ends as a tool error

### AC-18 (US-06) — domain invariant: same tree, same output on every platform
**Given** the same repository checked out on Linux, macOS and Windows, at different absolute locations
**When** the repo maintainer runs a check in each, naming the target the same way
**Then** the human and machine-readable reports are identical across the three, every path relative to the scanned location with `/` separators and no absolute path anywhere

### AC-19 (US-03) — abuse case: hostile metadata
**Given** one skill whose metadata nests deeper than 100 levels and a second whose metadata expands to more than 10,000 nodes through self-references
**When** the repo maintainer runs a check
**Then** each gets exactly one finding from the frontmatter integrity rule and is counted as not checked, each stays within the per-file processing bound (§6), every other file is checked, and the run ends failed

## 6. Non-functional requirements

| Aspect | Target | Measurement |
|---|---|---|
| Run time p95, typical repository (≤ 1,000 instruction files, ≤ 50,000 paths after exclusions) | ≤ 1 s | benchmark fixture in CI, p95 over 20 runs |
| Per-file processing bound | ≤ 100 ms to read, decode and integrity-check any single file, including hostile inputs at the bounds below | CI benchmark over boundary and hostile fixtures (AC-08b, AC-19) |
| Input bounds | file read ≤ 1 MiB; structured data ≤ 100 nesting levels and ≤ 10,000 nodes after alias expansion; fixed, not configurable in version 1 | fixtures exactly at each bound pass with no finding; fixtures one step over each bound produce exactly one integrity finding |
| Determinism | 100% identical output (human and machine-readable) across repeated runs on the same tree | e2e test runs each fixture twice and compares byte-for-byte |
| Cross-platform consistency | byte-identical inventory and findings for the same tree on Linux, macOS and Windows (paths target-relative with `/`) | e2e fixtures run on all three in CI and compared byte-for-byte |
| Outcome taxonomy | 100% of configuration/invocation problems end as tool error, never as findings; 100% of per-file and per-directory read/parse problems end as findings, never as tool error | e2e fixture suite covering each problem class |
| Network access | 0 network calls during a run | review + test with network disabled |
| Extensibility | adding a new rule changes exactly 1 file outside that rule's own package (the registry line) | measured on the first content rule's change (see §7) |

## 6.1 Security / privacy

- **Data classification:** internal — the tool reads the scanned repository's files locally and emits paths and short messages; nothing leaves the machine.
- **Personal data touched:** none — no new stored fields; file content is read only in memory and never echoed beyond the location of a problem: integrity messages are fixed plain-language reasons and never quote file content or a parser's raw error text.
- **AuthZ/AuthN impact:** no new capabilities; the run has exactly the file-read permissions of the account that starts it, and a file or directory it cannot read becomes an integrity finding, not a crash.
- **Abuse cases:**
  - Link escaping the repository (a link to a private file elsewhere on the machine): links are never followed, so content outside the scanned tree is never read or reported.
  - Hostile metadata (deeply nested or self-expanding structured data in frontmatter or a manifest): parsing is bounded (§6 Input bounds); the file gets a finding from the frontmatter integrity rule (markdown kinds) or the manifest integrity rule (plugin manifests) instead of exhausting memory or hanging the run.
  - Control characters in file names or content injected into messages: terminal control sequences are rendered as visible escapes in human output so a crafted file cannot rewrite the maintainer's terminal.
  - Very large or binary files with an instruction-file name: reads are bounded at 1 MiB; oversized, binary or non-text files get a finding from the unreadable-file integrity rule.
- **Security review:** Required — M feature that parses untrusted input from arbitrary repositories.

## 7. Metrics / KPIs

- **Kind-recognition accuracy** on a curated corpus of at least 20 public repositories that together cover all six file kinds — a snapshot of each repository's layout and instruction files with hand-assigned kinds is kept in this repository and runs as a test, curated by the Tech Lead — baseline: 0 (no discovery exists), target: 100% correctly classified, 0 non-instruction files counted, at feature completion.
- **False integrity findings** on the same corpus's well-formed files — baseline: n/a, target: 0 at feature completion.
- **Cost of the next rule** — files changed outside the new rule's own package when the first content rule is added — baseline: n/a, target: exactly 1 (the registry line), measured on that rule's change.
- **Dogfooding** — agents-lint checks its own repository in its CI — baseline: not running, target: running and clean within 7 days of merge.

## 8. Open questions

- [ ] Which exact locations and names map to each file kind (nested CLAUDE.md in subdirectories, CLAUDE.local.md, project-level vs plugin-level agents and commands folders, marketplace manifests, case variants such as claude.md), and does a leading `---` in AGENTS.md / CLAUDE.md count as frontmatter? Default now: follow the agent runtimes' documented layouts, names are case-sensitive, AGENTS.md / CLAUDE.md have no frontmatter (a leading `---` is body). — owner: Tech Lead, due: before `sdd:design`
- [ ] Should a run that found no instruction files be able to end failed (opt-in), so scripts that look only at the outcome can catch a false green? Default now: always ends clean, with the explicit "nothing found" statement. — owner: Tech Lead, due: before the first public release
- [ ] How should a configuration written for a newer agents-lint (naming rules this binary does not know) be handled by an older pinned binary? Default now: refused as a tool error (AC-14). — owner: Tech Lead, due: before the first public release
- [ ] Should shared skill directories linked into the repository be checkable via an opt-in? Default now: links are never followed (§3). — owner: Tech Lead, due: before the first public release
