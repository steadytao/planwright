# ADR Process

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright depends on deliberate decisions about scope, graph semantics, local execution, importer trust boundaries, generated artefacts, release evidence and compatibility claims.

Without a compact decision record, those decisions can drift into scattered README text, roadmap notes, tests or implementation details that are hard to evaluate later.

The project needs a lightweight process for recording material decisions without turning routine changes into ceremony.

## Decision

Planwright uses Architecture Decision Records (ADRs) to record material technical and project-boundary decisions.

ADRs are stored under [`docs/architecture/decisions/`](README.md).

They are numbered in ascending order, starting at `0000`.

Each ADR should be concise, specific and written so a future maintainer can understand:
- the problem or context
- the decision that was made
- the main consequences of that decision
- the compatibility or security boundary affected by the decision

## When An ADR Is Required

An ADR is required for decisions that materially affect:
- project scope or product boundary
- local-first, credential or network trust assumptions
- graph, typed-plan, pack or report contracts
- importer, generator, analyser or policy semantics
- public command names, flag names, exit codes or diagnostic meanings
- release, signing, provenance, SBOM or verification model
- plugin, hosted, TUI or browser architecture

## When An ADR Is Not Required

An ADR is not required for:
- routine refactors
- small implementation details
- documentation-only edits that do not record a new project direction
- naming changes without architectural effect
- short-lived experiments that are not adopted
- ordinary bug fixes that do not change project direction or assumptions

## ADR Structure

Each ADR should contain:
- title
- status badge
- context
- decision
- consequences

Optional sections may be added when helpful but ADRs should remain compact.

## ADR Status Badges

Planwright ADRs should express status with a single badge rather than a plain text status line.

The standard status values are:
- `proposed`
- `accepted`
- `superseded`
- `deprecated`
- `denied`

Use these badge forms:
```md
<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->
```

Only one status badge should be active in an ADR at a time.

## ADR Status Meanings

`proposed` means a decision is being considered but is not yet in force.

`accepted` means the decision has been made and is the current project direction.

`superseded` means the decision was previously accepted but has been replaced by a later ADR that now governs.

`deprecated` means the decision is no longer preferred and should be phased out but has not yet been fully replaced or removed.

`denied` means a materially considered proposal was explicitly rejected and should not be treated as an undecided open question.

## ADR Lifecycle

An accepted ADR remains in force until it is replaced or superseded by another ADR.

ADRs should not be rewritten to hide historical decisions. If a decision changes, create a new ADR and mark the older ADR accordingly.

## Consequences

This process creates a stable record of important Planwright decisions and reduces the risk of undocumented architectural drift as the tool evolves.
