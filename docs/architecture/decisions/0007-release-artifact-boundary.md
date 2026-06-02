# Release Artefact Boundary

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright has release workflows, GoReleaser configuration, release-level SBOM generation, checksum manifests and OpenPGP signing steps planned for published releases.

The repository has no published release artefacts yet. Checked-in release documentation should not imply that a release has been produced or that source-tree files prove release integrity by themselves.

The project also needs a clear distinction between source metadata, generated release artefacts and checksum evidence such as `go.sum`.

## Decision

Planwright treats release artefacts as outputs of the release workflow rather than canonical source files.

The source repository owns:
- release workflow configuration
- checked-in release-note source files
- release process documentation
- module metadata such as `go.mod` and `go.sum`
- source-level licence, governance and security policy files

The release workflow owns generated release outputs:
- platform binaries
- release-level SPDX and CycloneDX SBOMs
- checksum manifests
- OpenPGP signatures for checksum manifests
- provenance or attestations where configured
- generated release notes assembled from checked-in notes and commit history

Root-level generated SBOMs are not checked in by default because they drift quickly and can imply a source tree has been built or released when it has not.

## Consequences

This decision means that:
- docs should say Planwright has no published release artefacts until a release workflow actually publishes them
- direct binaries and checksum manifests belong in release assets
- SBOMs belong beside the release artefacts they describe
- `go.mod` remains the Go module requirement contract
- `go.sum` remains checksum evidence for module authentication and may contain checksum-only modules
- release workflow changes are security-sensitive because they affect supply-chain evidence
- release readiness requires maintainer judgement in addition to passing automation
