# Architecture Decision Records

This directory contains Planwright's Architecture Decision Records (ADRs). ADRs document material technical and project-boundary decisions in a form that should remain understandable after the surrounding implementation changes.
- [Process](#process)
- [Records](#records)
- [Status](#status)

# Process

The purpose of ADRs is to:
- record important decisions and their rationale
- reduce undocumented architectural drift
- preserve context for future maintainers
- distinguish deliberate choices from accidental behaviour
- keep compatibility, security and release claims tied to explicit decisions

# Records

The intended reading order is numerical.

ADR | Decision | Status
:--- | :--- | :---
[`0000`](0000-adr-process.md) | ADR process | accepted
[`0001`](0001-engine-first-local-cli.md) | Engine-first local CLI | accepted
[`0002`](0002-project-scope.md) | Project scope | accepted
[`0003`](0003-language-choice.md) | Language choice | accepted
[`0004`](0004-canonical-graph-contract.md) | Canonical graph contract | accepted
[`0005`](0005-import-adapter-boundary.md) | Import adapter boundary | accepted
[`0006`](0006-local-web-workbench-boundary.md) | Local web workbench boundary | accepted
[`0007`](0007-release-artifact-boundary.md) | Release artefact boundary | accepted
[`0008`](0008-fixture-backed-compatibility-claims.md) | Fixture-backed compatibility claims | accepted
[`0009`](0009-stable-core-before-expansion.md) | Stable core before expansion | accepted
[`0010`](0010-policy-profile-boundary.md) | Policy profile boundary | accepted
[`0011`](0011-diagnostics-sarif-and-exit-code-stability.md) | Diagnostics, SARIF and exit code stability | accepted
[`0012`](0012-human-controlled-release-signing.md) | Human-controlled release signing | accepted
[`0013`](0013-proof-before-expansion.md) | Proof before expansion | accepted

# Status

Each ADR keeps the full hidden status badge set and enables exactly one active status badge.

An ADR is not required for every change. Routine refactors, small implementation details and ordinary bug fixes should usually stay in code review or release notes.

For the rules governing when an ADR is required and how ADRs should be written, see [`0000-adr-process.md`](0000-adr-process.md).
