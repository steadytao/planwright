# Human-Controlled Release Signing

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright needs a release integrity model before the first published release.

The project is local-first infrastructure tooling. Users may run Planwright on machines that also hold infrastructure definitions, review evidence and generated deployment packs. Release verification must therefore be understandable without depending only on a hosted service identity.

Direct binaries make installation simpler than archives for the current single-binary CLI. They also keep README links stable through GitHub's latest-release download route.

## Decision

Planwright uses a maintainer-controlled OpenPGP release key as the release trust root.

The release workflow publishes:
- direct platform binaries
- release-level SPDX and CycloneDX SBOMs
- `SHA2-256SUMS`
- `SHA2-512SUMS`
- detached OpenPGP signatures for both checksum manifests
- a convenience public-key export named `public.key`

The signature scope is the checksum manifests rather than each binary. Users verify the signing key fingerprint first, verify the checksum manifest signature and then verify the downloaded binary against the signed manifest.

The GitHub release public key asset is not the trust root by itself. The trust root is the fingerprint published in `docs/releases/signing.md` or another maintainer-controlled channel.

Release signing secrets must be available only to the release workflow, with the `release` environment providing a protection boundary before the workflow can use them. The workflow must fail closed if the private key material is missing, the passphrase is missing or the fingerprint variable is missing.

## Consequences

This decision means that:
- release verification documentation must explain fingerprint verification rather than just key import
- the first signed release cannot be considered ready until `docs/releases/signing.md` contains the real release key fingerprint
- release assets stay small because signatures cover manifests instead of every binary separately
- changing the release signing key is a security-sensitive event that requires a dated documentation update
- OpenPGP signing does not prove reproducible builds, deployability or absence of vulnerabilities
- SBOMs and future provenance attestations are supplementary evidence; they do not replace the human-controlled trust root
