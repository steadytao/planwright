# Governance

This document defines the governance model for Planwright.

Planwright is intended to be a serious local-first infrastructure planning, compatibility and evidence engine. Governance should therefore be clear, practical and proportionate to the current stage of the project.
- [Purpose](#purpose)
- [Current Model](#current-model)
- [Roles](#roles)
- [Decision-Making](#decision-making)
- [Project Boundaries](#project-boundaries)
- [Merge Authority](#merge-authority)
- [Release Authority](#release-authority)
- [Security Decisions](#security-decisions)
- [Recorded Decisions](#recorded-decisions)
- [Maintainer Changes](#maintainer-changes)
- [Conflicts Of Interest](#conflicts-of-interest)
- [Governance Changes](#governance-changes)

## Purpose

The purpose of governance in Planwright is to:
- define who is responsible for project direction and repository decisions
- make merge and release authority clear
- protect the engine-first product boundary
- provide a stable basis for review, contribution and maintenance
- reduce ambiguity around compatibility claims, security posture and release readiness
- avoid project drift caused by unclear ownership or informal process

## Current Model

Planwright currently uses a maintainer-led governance model.

At this stage, the project is controlled by its maintainer or maintainers. Decisions are made by the maintainers, with primary authority resting with the core maintainers unless and until this governance model is changed.

This document does not create a foundation, committee, voting body or consensus governance model. Input from contributors is welcome; final responsibility remains with maintainers.

## Roles

Planwright currently recognises the following roles.

### Core Maintainer

The core maintainer is responsible for:
- defining and protecting project direction
- deciding whether proposed work fits Planwright's scope
- approving or declining architectural changes
- setting repository standards and contribution expectations
- deciding release readiness
- managing security response and disclosure decisions
- appointing or removing maintainers
- amending governance when necessary

### Maintainers

Maintainers are responsible for:
- reviewing contributions
- protecting the project's local-first defaults
- protecting the engine-first architecture
- protecting the documented security posture and quality standards
- merging changes within their authority
- helping manage issues, pull requests and documentation
- escalating major architectural, compatibility, release or security decisions when appropriate

### Contributors

Contributors may:
- report bugs
- suggest improvements
- submit pull requests
- improve documentation and examples
- participate in project discussion

Contributors do not have merge, security advisory or release authority unless they are also maintainers.

## Decision-Making

Planwright prefers clear maintainer decisions over vague consensus language.

Input from contributors is welcome. Maintainers are responsible for making final decisions about:
- whether a change fits the project
- whether a pull request will be merged
- whether a release is ready
- whether a proposed change requires a recorded decision
- whether a contribution conflicts with Planwright's scope, security model or standards
- whether a compatibility claim is justified by tests and fixtures

For major changes, maintainers should prefer recorded reasoning over ad hoc judgement.

## Project Boundaries

Maintainers must protect Planwright's current boundaries. In particular, Planwright should remain:
- local-first by default
- engine-first rather than GUI-first
- explicit about compatibility limits
- explicit about provenance, loss and unsupported constructs
- cautious about cloud credentials, live scans and mutation
- honest about security, cost, compliance and deployability claims

Changes that move Planwright towards live infrastructure mutation, hosted account management, broad multi-cloud scope, AI-generated deployment decisions or lossless conversion claims require explicit maintainer review and documented reasoning.

## Merge Authority

No pull request should be merged solely because it is technically functional.

A pull request may be merged only if it:
- fits Planwright's scope and direction
- meets the documented standards
- is understandable and reviewable
- includes tests where the behaviour changes
- updates documentation where user-facing behaviour, contracts or policy change
- does not introduce unacceptable security, privacy, compatibility or maintenance risk
- has been reviewed by an authorised maintainer

Sensitive areas may require stricter review.

Sensitive areas include:
- graph and typed-plan schemas
- importer and generator behaviour
- policy and security findings
- SARIF output
- local web server behaviour
- release, signing and supply-chain metadata
- governance, security and licence documents
- workflows that publish artefacts or receive elevated permissions

## Release Authority

Releases are authorised by the core maintainer.

A release should not be made unless maintainers are satisfied that it meets the project's stated release, security and quality expectations.

Release readiness should consider:
- tests and static analysis results
- vulnerability checks
- documentation accuracy
- compatibility matrix accuracy
- known limitations and non-goals
- generated artefact integrity
- whether release notes describe the real shipped surface

Release tooling can automate packaging; it does not replace maintainer judgement.

## Security Decisions

Security-sensitive decisions are handled by maintainers.

This includes:
- coordinated disclosure timing
- remediation direction
- advisory publication
- whether an issue is security-sensitive
- whether a finding affects a released version
- whether a public discussion should be moved to private reporting channels

Where necessary, the core maintainer has final authority.

## Recorded Decisions

Material architectural and project-boundary decisions should be recorded under [`docs/architecture/decisions/`](../architecture/decisions/).

Governance changes, scope changes and other significant changes should not rely only on chat, memory or issue discussion.

Examples of decisions that usually deserve a record:
- changing the graph or typed-plan compatibility contract
- adding live cloud calls or credential handling
- changing local web server trust boundaries
- adding release signing or publishing changes
- changing maintainer authority or contribution rules
- making a previously experimental feature part of the supported surface

## Adding Maintainers

A maintainer may be added when the core maintainer judges that the person has demonstrated:
- consistent good judgement
- understanding of Planwright's direction
- review quality
- technical competence
- reliable and constructive participation
- respect for the project's security and compatibility boundaries

Maintainer status is not automatic and is not granted based only on contribution count.

## Removing Maintainers

A maintainer may be removed if they:
- act against the project's interests
- repeatedly undermine the documented direction of the project
- fail to meet expected standards of judgement or conduct
- mishandle security-sensitive information
- become inactive for an extended period where that creates project risk
- no longer wish to serve in the role

## Conflicts Of Interest

Maintainers should disclose material conflicts when they affect project decisions.

Examples include:
- reviewing a change where they have a direct outside obligation
- making a release decision tied to an undisclosed commercial commitment
- handling a security report that affects another project they control

Disclosure does not automatically prevent participation. It gives the project enough context to decide whether another maintainer should review the matter.

## Governance Changes

This document is the canonical governance document for Planwright.

Changes to the governance model should be made deliberately and should reflect the actual needs of the project rather than process for its own sake.

Governance changes require maintainer approval and should update related documents where needed including [`MAINTAINERS.md`](MAINTAINERS.md), [`.github/CONTRIBUTING.md`](../../.github/CONTRIBUTING.md), [`.github/SECURITY.md`](../../.github/SECURITY.md) and [`.github/CODE_OF_CONDUCT.md`](../../.github/CODE_OF_CONDUCT.md).
