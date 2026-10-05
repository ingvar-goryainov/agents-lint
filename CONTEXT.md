---
status: Living
updated_at: "2026-10-04"
---

# Domain Context — agents-lint

<!--
CONTEXT.md is the domain glossary — not a spec and not a scratch pad. NO implementation
detail here (no datastore/broker/framework names, no API contracts) — only domain words
and the boundaries between them. Implementation choices live in the SAD and ADRs; behaviour
lives in spec.md.
-->

## Glossary

- file kind — the category agents-lint assigns to an instruction file from a fixed, published set (agents-md, claude-md, skill, agent, command, plugin-manifest); rules declare which kinds they apply to. NOT file format (a skill and an agent are both markdown but different kinds; a plugin manifest is a different format).
- finding — one problem a rule reports about a specific location in an instruction file, carrying the rule's ID, a severity and a plain-language message. NOT a tool error (an invalid config, an unreadable target or a bad flag stops the run instead of producing findings).
- instruction file — a file in the scanned repository that agents-lint recognizes as configuring an AI agent and therefore assigns exactly one file kind. NOT any markdown file (a README or a docs page is not an instruction file unless it matches a file kind).
- integrity rule — a built-in rule that reports an instruction file agents-lint cannot read or parse; a file with an integrity finding is not checked by any other rule. NOT a content rule (which judges what a well-formed instruction file says).
- inventory — the summary of one run that states how many instruction files of each file kind were found, how many paths were ignored, and how many instruction files could not be checked; always reported, even when there are zero findings. NOT the findings list (an inventory says what was looked at, findings say what is wrong).
- repo maintainer — a person responsible for a repository's agent instruction files who runs agents-lint over that repository to see what it recognizes and what is wrong. NOT a rule author (someone who contributes rules to agents-lint itself).
- severity — how serious a finding is: off, info, warning or error; configurable per rule. NOT exit status (which severities fail a run is decided separately by the fail threshold).
