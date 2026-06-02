# Engine-First Local CLI

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright could become a visual planning tool, an importer collection, a report generator or a deployment helper.

If the first implementation put compatibility logic inside a UI, script or command handler, later interfaces would be harder to trust and harder to test consistently.

The project also needs a serious interface that works in version control, CI, terminal sessions and local review workflows before broader UI surfaces exist.

## Decision

Planwright starts with a local CLI over reusable engine packages.

The CLI is responsible for argument parsing, user-facing command output and exit status.

Core behaviour belongs outside the command layer:
- typed plan parsing lives in `internal/plan`
- graph types, validation and diffing live in `internal/graph`
- importer logic lives in `internal/importers`
- generator logic lives in `internal/generators`
- review, policy and report rendering live in dedicated internal packages
- local web serving calls the same engine packages rather than owning private infrastructure logic

The initial implementation exposed validation and explanation through the CLI first. Later local web, CI and release workflows are interfaces over the same implementation boundary.

## Consequences

This decision means that:
- CLI behaviour remains a first-class product surface
- new interfaces must reuse engine packages rather than growing private compatibility logic
- generated reports, SARIF, local web previews and future TUI output should agree because they share the same graph and analyser paths
- test coverage should target the engine packages as well as CLI integration
- GUI-first feature work may be deferred when it would bypass or distort the engine contract
- future hosted demo work must remain an interface over sample data and local-compatible engine behaviour
