# AGENTS.md

**Human readers:** this file is primarily for coding agents. It is an agent entry point, not a replacement for the canonical project documents.

This file provides agent-focused instructions for work in Planwright.

## Mission

Planwright is a local-first infrastructure planning, compatibility and evidence engine.

It turns typed plans and, later, imported infrastructure sources into a typed architecture graph with provenance, diagnostics and reviewable outputs. It is not a one-click deployment console, a cloud account mutation tool, a generic diagrammer or a lossless universal converter.

## Canonical Authority

Agents must treat the following as authoritative:
- [`README.md`](README.md)
- [`AUTHORS`](AUTHORS)
- [`CONTRIBUTORS`](CONTRIBUTORS)
- [`.github/CONTRIBUTING.md`](.github/CONTRIBUTING.md)
- [`.github/CODEOWNERS`](.github/CODEOWNERS)
- [`docs/project/DCO.md`](docs/project/DCO.md)
- [`docs/project/GOVERNANCE.md`](docs/project/GOVERNANCE.md)
- [`.github/SECURITY.md`](.github/SECURITY.md)
- [`.github/SUPPORT.md`](.github/SUPPORT.md)
- [`docs/README.md`](docs/README.md)
- [`docs/architecture/design.md`](docs/architecture/design.md)
- [`docs/architecture/threat-model.md`](docs/architecture/threat-model.md)
- [`docs/compatibility.md`](docs/compatibility.md)
- [`docs/development/standards.md`](docs/development/standards.md)
- [`docs/development/testing.md`](docs/development/testing.md)
- [`docs/architecture/decisions/`](docs/architecture/decisions/)

If this file appears to conflict with those documents, follow the canonical documents.

## Project Stage

Planwright is in the v0.12 usability and proof stage after the first planned public release.

The next high-value work is to make the existing engine easy to understand, run and review through one polished proof path. Do not add broad features merely because they appear in the long-term roadmap.

Be especially careful not to:
- write documentation as if stable releases already exist
- imply that TUI, hosted demo, zip archives or live scans already exist unless the code proves it
- imply that CloudFormation/SAM import is complete or lossless beyond the v0.2 supported subset
- imply that Terraform review supports state import, provider schema ingestion, HCL evaluation or apply-time proof beyond the v0.3 plan JSON checks
- imply that Kubernetes import evaluates Helm templates, runs Kustomize, contacts a cluster, decodes Secret values or fully models NetworkPolicy/Cilium semantics beyond the v0.5 rendered-manifest subset
- imply that AWS scan import contacts AWS, loads credentials, runs the AWS CLI, uses the AWS SDK, verifies live account identity or proves drift beyond the v0.6 local bundle subset
- imply that graph diff review proves live infrastructure drift beyond the v0.7 local graph JSON comparison
- imply that the graph JSON Schema is a stable v1.0 compatibility guarantee beyond the v0.8 structural `planwright.graph.v1` schema
- imply that built-in policy profiles certify compliance, replace human review or support custom policy packs beyond the current local static checks
- imply that CI proves security, correctness, deployability or documentation quality beyond the checks that actually run
- imply that generated Terraform/OpenTofu output has been executed, validated by Terraform or proven deployable unless fresh evidence proves it
- imply that the local web UI deploys, writes project files from browser actions, reads credentials or contacts cloud APIs
- imply that the v0.12 proof work broadens compatibility beyond existing fixture-backed support
- claim security, compliance, cost or deployability guarantees beyond current diagnostics
- describe conversion as lossless unless a fixture proves the exact round trip

## Scope Discipline

Planwright is engine-first.

Agents must not:
- add GUI-only logic
- add live cloud mutation
- add telemetry, analytics, remote fonts or remote scripts
- add cloud credentials or default credential loading without explicit design review
- make outbound cloud calls on validation, explanation or project load
- treat imported infrastructure as executable content
- add broad multi-cloud scope before the graph contract is useful for the first AWS slice

Every version must finish with a full local security/correctness pass before the next version starts.

## Documentation and Language

Use British English in documentation, reports, examples and other user-facing prose.

Keep public text precise and conservative. Avoid hype, maturity inflation and vague safety claims.

Documentation lists and fenced code blocks must hug the text above them:
- write `Text:` followed immediately by `- item`
- do not insert a blank line between the lead-in text and the bullet list
- write `Text:` followed immediately by a fenced code block when showing commands or examples
- do not insert a blank line between the lead-in text and the code block
- avoid commas before `but`
- avoid commas before `including`, `however` and `or`
- avoid commas before `then`
- avoid simple comma-before-`and` constructions unless the comma separates larger topics

## Code Standards

When creating new Planwright-owned source files, scripts or other copyright-affected files that support normal comments:
- add a copyright notice near the top of the file
- add `SPDX-License-Identifier: Apache-2.0`
- preserve valid existing file headers unless there is a real reason to normalise them

Use the standard form:
```go
// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0
```

## Current Quality Gates

Current useful commands include:
```bash
go test ./...
go vet ./...
go build ./cmd/planwright
go mod verify
go mod tidy -diff
golangci-lint config verify --config .github/config/golangci.yml
golangci-lint fmt --config .github/config/golangci.yml --diff
golangci-lint run --config .github/config/golangci.yml
go run ./cmd/planwright docs check .
npx --yes cspell@10.0.1 lint "**/*.md" ".github/**/*.yml" ".github/**/*.yaml" --config .github/config/cspell.json --no-progress --no-summary --no-must-find-files
python3 .github/scripts/ci/check_action_pins.py
python3 .github/scripts/ci/check_file_headers.py
python3 .github/scripts/release/generate_contributors.py --check
actionlint
pre-commit run --all-files
```

Run a focused security/correctness review before advancing the roadmap version. Do not claim a version gate is complete without fresh verification evidence.
