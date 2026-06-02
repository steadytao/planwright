# Canonical Graph Contract

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright has multiple inputs and outputs: typed plans, imported infrastructure artefacts, generated Terraform/OpenTofu-oriented files, Mermaid diagrams, Markdown reports, SARIF and policy findings.

If those surfaces each define their own infrastructure model, compatibility claims will drift and review findings will become difficult to trace back to source evidence.

The project needs one canonical representation for understood infrastructure intent while preserving unsupported source material through loss reporting.

## Decision

Planwright treats `planwright.graph.v1` as the canonical internal architecture graph contract for understood infrastructure intent.

The graph records:
- nodes and edges
- provider, region and profile metadata
- resource properties that Planwright understands
- relationship semantics such as network, dependency and generated-from edges
- diagnostics from structural and semantic validation

The graph is not a deployment format and not a promise that imported source can be converted without loss.

`planwright.graph.v1` is validated in two layers:
- JSON Schema validation checks structural shape
- semantic graph validation checks rules that JSON Schema cannot express cleanly

The compatibility matrix remains the public explanation for what each input or output path can do with the graph.

## Consequences

This decision means that:
- importers should lower only understood constructs into the graph
- unsupported, ambiguous or unsafe constructs must be reported rather than silently dropped
- generators should consume the graph rather than re-parsing source-specific formats
- Markdown reports, SARIF and local web previews should derive from the graph or from explicit review models
- schema changes are compatibility-sensitive and require tests plus documentation updates
- a quiet graph validation result does not prove deployability, compliance or live infrastructure state
