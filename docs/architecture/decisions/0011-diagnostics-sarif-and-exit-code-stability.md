# Diagnostics, SARIF and Exit Code Stability

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright produces diagnostics, Markdown findings, SARIF results and command exit statuses.

Those outputs are currently used by local users. They will become more important when Planwright adds a CI review Action, code-scanning integration and a stable v1.0 diagnostics policy.

Machine-readable output can become an accidental public contract. Changing a SARIF rule ID, exit status meaning or diagnostic code after users automate against it can break workflows even when the human text still looks acceptable.

At the same time, Planwright is still pre-v1.0 and should not freeze every message string too early.

## Decision

Planwright treats diagnostics, SARIF rule IDs and exit status meanings as compatibility-sensitive surfaces.

The current stability boundary is:
- human-readable diagnostic text may evolve before v1.0 when it improves clarity
- machine-readable rule IDs should be stable once documented or used in release artefacts
- exit status meanings should be documented before the GitHub Action is treated as stable
- SARIF output must validate against the supported SARIF schema when the relevant tests or release gates claim it
- source locations should be included where an importer or review path can provide them safely
- findings should include severity, subject, explanation and fix guidance where the source model supports it

Planwright should not treat:
- incidental message wording as a stable machine interface
- undocumented exit status behaviour as a public compatibility promise
- SARIF output as proof of deployability, compliance or live infrastructure state
- suppression handling as designed until a suppression and baseline policy exists

## Consequences

This decision means that:
- new rule IDs need deliberate names and tests before they are exposed in SARIF
- rule ID changes after release need compatibility notes
- CLI exit code changes require documentation and tests
- report wording can improve without a compatibility break when rule IDs and structured fields remain stable
- the CI review Action must document token permissions, supported inputs and exit behaviour before it is considered stable
- suppression, baseline and generated-location behaviour should be designed before Planwright encourages long-term CI automation
