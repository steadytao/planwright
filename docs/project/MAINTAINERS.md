# Maintainers

This document provides a practical overview of current Planwright maintainers.

The canonical governance model including project roles, decision-making and maintainer authority is defined in [`GOVERNANCE.md`](GOVERNANCE.md).
- [Core Maintainer](#core-maintainer)
- [Maintainer Responsibilities](#maintainer-responsibilities)
- [Review Expectations](#review-expectations)
- [Sensitive Areas](#sensitive-areas)
- [Contact And Responsibility](#contact-and-responsibility)
- [Maintainer Changes](#maintainer-changes)

## Core Maintainer

### [steadytao](https://github.com/steadytao)
[![GitHub-Sponsor](https://img.shields.io/badge/_-Github-white.svg?logo=github&labelColor=555555&style=for-the-badge)](https://github.com/sponsors/steadytao)

Zen Dodd (@steadytao) is the current core maintainer and final decision-maker for:
- project direction
- scope and roadmap decisions
- merge authority
- release readiness
- security response decisions
- maintainer appointment or removal
- governance changes

## Maintainer Responsibilities

Maintainers are expected to protect Planwright's product and trust boundaries.

That means maintainers should:
- keep Planwright local-first by default
- keep the engine as the source of product behaviour
- reject compatibility claims that are not supported by fixtures or tests
- reject security, compliance, cost or deployability claims that exceed the evidence
- require clear loss reports for partial or ambiguous imports
- require tests for behaviour changes
- require documentation updates for user-facing changes
- keep release notes and examples aligned with shipped behaviour
- treat security reports carefully and privately where appropriate

## Review Expectations

Maintainer review should focus on:
- correctness
- security and privacy impact
- compatibility impact
- test coverage
- documentation accuracy
- maintainability
- consistency with the current roadmap stage

Maintainers may ask for changes even when code compiles and tests pass.

## Sensitive Areas

The following areas need particular care:
- graph and typed-plan schemas
- importer and generator behaviour
- policy findings and SARIF output
- local web server security boundaries
- file-system reads and writes
- release, signing and supply-chain metadata
- GitHub Actions workflow permissions
- governance, security, licence and contribution documents

Changes in these areas may require additional review or recorded reasoning.

## Contact And Responsibility

Project communication should generally take place through the repository unless a private security report is required.

Security issues should follow the process in [`.github/SECURITY.md`](../../.github/SECURITY.md).

Conduct concerns should follow the process in [`.github/CODE_OF_CONDUCT.md`](../../.github/CODE_OF_CONDUCT.md).

## Maintainer Changes

Maintainer appointments and removals are governed by [`GOVERNANCE.md`](GOVERNANCE.md).

Changes to this file should reflect current authority accurately. They should not create or remove authority without the corresponding governance decision.
