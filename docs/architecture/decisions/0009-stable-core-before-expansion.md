# Stable Core Before Expansion

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

The full Planwright vision includes many valuable surfaces: richer importers, more generators, a TUI, a hosted static demo, a GitHub Action, a VS Code extension, live read-only AWS scans, policy packs, plugin adapters and broader infrastructure research.

If all of those surfaces are treated as pre-v1.0 requirements, the project risks delaying a stable core until the roadmap becomes too large to verify.

The stronger strategic path is to make a smaller core stable first then expand from that tested base.

## Decision

Planwright's pre-v1.0 roadmap is narrowed to the capabilities needed for a credible local-first compatibility and evidence engine.

The pre-v1.0 sequence is:
- first-release usability and proof
- golden compatibility fixture suite
- Terraform/OpenTofu state and provider schema import
- Terraform/OpenTofu graph lowering for a declared AWS subset
- deployment pack v1
- examples gallery and documentation site
- CI review Action and SARIF hardening
- security and accessibility hardening
- v1.0 stable core

The v1.0 stable core excludes broad expansion features unless a later decision explicitly moves them forward.

Post-v1.0 expansion gates include:
- CloudFormation and SAM generator expansion
- Docker Compose import and generation
- Kubernetes, Gateway API and Cilium generation
- TUI review interface
- hosted static demo
- VS Code extension
- live read-only AWS scan
- drift and diff expansion
- policy packs
- OPA/Rego and Conftest integration
- plugin SDK
- operational, resilience and cost analysis expansion
- multi-cloud and other IaC research gates

## Consequences

This decision means that:
- v1.0 can be assessed against a smaller and more defensible scope
- the project can publish a stable local engine without waiting for every planned interface
- hosted, live-scan, plugin and editor work remain important but no longer block the first stable core
- roadmap gates after v1.0 must still preserve the engine-first, local-first and fixture-backed compatibility decisions
- future attempts to pull broad expansion back before v1.0 should update this ADR or create a superseding ADR
