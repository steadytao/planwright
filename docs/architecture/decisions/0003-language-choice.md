# Language Choice

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright is a local infrastructure tool with parser boundaries, filesystem boundaries, report generation, static binaries and cross-platform release expectations.

The project needs:
- a language suited to local systems tooling
- straightforward static binary distribution
- mature testing, fuzzing and static analysis support
- reliable JSON, YAML, HTTP and filesystem libraries
- maintainability for a small maintainer-led project

## Decision

The Planwright core, CLI and local web server are implemented in Go.

Web assets for the current local workbench remain embedded static files served by the Go binary.

Future richer browser interfaces may use TypeScript or React if they earn the complexity; engine behaviour must stay in the Go core rather than moving into browser-only code.

## Consequences

This decision means that:
- repository structure, testing and release workflows should continue to align with Go tooling
- parser, importer, generator and policy behaviour should be tested in Go packages
- native Go fuzzing is the first fuzzing path for untrusted input boundaries
- local web work should avoid adding a separate frontend build pipeline until the interface needs it
- major non-Go runtime dependencies should be justified deliberately rather than introduced casually
