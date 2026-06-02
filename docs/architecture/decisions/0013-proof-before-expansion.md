# Proof Before Expansion

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright now has enough implemented surface to be treated as a serious public proof line rather than a brainstorm.

The product vision remains broad: typed plans, importers, graph validation, reports, deployment packs, local web review, future TUI, CI review, policy work, live read-only scans and plugin adapters.

The immediate risk is sprawl. Adding another importer, generator or interface before a new user can understand the existing engine would make Planwright look larger but not necessarily more credible.

The strongest next step is to make one supported path easy to understand, run and review.

## Decision

Planwright will prioritise proof density before feature expansion.

The next release gate after the first published release is a usability and proof release. Its job is to make the existing engine understandable in a few minutes through a polished walkthrough, a canonical example, real command output and fixture-backed evidence.

The preferred proof path is:
- one canonical AWS web application example
- validation output from real commands
- security and cost notes from real commands
- Mermaid output from real commands
- Terraform/OpenTofu-oriented generated review files
- a deployment pack walkthrough
- a companion bad-example path where it improves learning
- golden tests or fixture checks for the generated evidence that public docs rely on

This release must avoid broad feature growth unless that work directly improves the proof path.

## Consequences

This decision means that:
- the next release should be judged by clarity, runnable examples and reviewable evidence rather than new surface count
- the README and docs should make the first useful workflow obvious before listing every implemented command
- compatibility claims must remain tied to fixtures and loss reports
- broad work such as new import families, live AWS scans, GUI canvas editing, TUI work, custom policy packs, OPA/Rego integration and plugin adapters should stay deferred
- one excellent example is more valuable than several shallow examples
- screenshots, transcripts and generated outputs should be maintained only when there is a clear verification process
- future attempts to pull expansion work ahead of proof work should update this ADR or create a superseding ADR
