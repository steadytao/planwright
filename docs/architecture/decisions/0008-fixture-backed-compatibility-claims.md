# Fixture-Backed Compatibility Claims

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright's value depends on honest compatibility claims.

The project supports multiple source and output surfaces: typed plans, graph JSON, CloudFormation, SAM, Terraform plan JSON, Kubernetes manifests, local AWS scan bundles, Markdown reports, Mermaid diagrams and SARIF.

Without fixtures, the compatibility matrix can drift into aspiration. A table row can look authoritative even when there is no checked input, expected output or loss-report evidence behind it.

The highest-risk product mistake is claiming broad conversion support before Planwright can prove what it preserved, normalised, inferred, rejected or lost.

## Decision

Planwright treats compatibility claims as fixture-backed claims.

The roadmap places the golden compatibility fixture suite before further broad compatibility expansion.

Each current compatibility matrix row must be backed by at least one fixture or clearly marked as documentation-only.

A fixture should record:
- the source format or surface
- the expected compatibility level
- supported resources or constructs
- unsupported, ambiguous, redacted and lossy constructs
- generated output where that output is part of the claim
- expected diagnostics, findings or loss-report categories

Round-trip support may only be claimed when round-trip fixtures exist for that exact source and output path.

Deployability support may only be claimed when a sandbox deployment test exists for that exact generated path.

## Consequences

This decision means that:
- compatibility work must improve the fixture suite alongside importer, generator and report code
- documentation cannot raise a support level without corresponding fixture evidence
- broad service coverage is lower priority than narrow, tested and explainable coverage
- loss reports are part of the compatibility contract rather than optional documentation
- future plugin adapters must provide fixtures before their compatibility claims are accepted
- fixture metadata can become the source for public compatibility tables
