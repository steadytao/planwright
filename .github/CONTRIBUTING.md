# Contributing

Planwright is early-stage infrastructure tooling. Contributions should keep the engine small, reviewable and honest about compatibility limits.
- [Before Opening An Issue](#before-opening-an-issue)
- [Bug Reports](#bug-reports)
- [Feature Requests](#feature-requests)
- [Before Opening A Pull Request](#before-opening-a-pull-request)
- [Pull Request Expectations](#pull-request-expectations)
- [Development](#development)
- [Documentation Style](#documentation-style)
- [Compatibility Changes](#compatibility-changes)
- [File Headers And Licensing Notices](#file-headers-and-licensing-notices)
- [AI-Assisted Contributions](#ai-assisted-contributions)
- [DCO](#dco)
- [Security Issues](#security-issues)
- [Authors And Contributors](#authors-and-contributors)
- [Project Direction](#project-direction)

## Before Opening An Issue

Before opening an issue, please make sure you have:
- read [`README.md`](../README.md)
- read [`docs/architecture/design.md`](../docs/architecture/design.md)
- read [`docs/architecture/threat-model.md`](../docs/architecture/threat-model.md)
- read [`docs/compatibility.md`](../docs/compatibility.md)
- checked whether the issue already exists
- confirmed that the issue is actually about Planwright

Issues that are too vague to act on may be closed.

## Bug Reports

Bug reports should include:
- what happened
- what you expected to happen
- how to reproduce it
- relevant operating system, Go version and command details
- logs or command output in plain text, not screenshots where text would be clearer
- whether the issue affects a local file, generated output, import, report, schema or local web server path

Do not include secrets, credentials, private cloud account IDs or private infrastructure data in public issues.

## Feature Requests

Feature requests should include:
- the problem being solved
- why the current behaviour is insufficient
- the proposed direction
- expected trade-offs or risks, if known
- whether the feature changes a public command, schema, report or compatibility claim

Not every feature request will be accepted.

Planwright is intentionally scoped. Requests that conflict with the project's scope, security model, design principles or local-first posture may be declined even if technically feasible.

## Before Opening A Pull Request

Before opening a pull request, make sure you have:
- read the relevant documentation
- scoped the change clearly
- added or updated tests where appropriate
- updated documentation where behaviour changes
- checked compatibility and security implications
- signed off every commit under the DCO

Large, architectural or security-sensitive changes should usually begin with an issue or recorded decision first.

## Pull Request Expectations

Pull requests should be:
- well-scoped
- understandable
- justified by the problem they solve
- accompanied by tests when practical
- accompanied by documentation updates when behaviour, policy or interfaces change
- honest about unsupported cases and remaining limitations

Pull requests should not:
- add live cloud mutation without explicit maintainer approval
- add telemetry, analytics, remote fonts or remote scripts
- read cloud credentials without explicit design review
- imply lossless conversion without fixture-backed proof
- weaken local web server security boundaries
- make broad security, compliance, cost or deployability claims without evidence

## Development

Run the relevant checks before proposing a change: *(CI does this but you **must** resolve issues)*
```bash
go test ./...
go vet ./...
go build ./cmd/planwright
go mod verify
go mod tidy -diff
golangci-lint config verify --config .github/config/golangci.yml
golangci-lint fmt --config .github/config/golangci.yml --diff
golangci-lint run --config .github/config/golangci.yml
staticcheck ./...
gosec ./...
govulncheck ./...
go run ./cmd/planwright docs check .
npx --yes cspell@10.0.1 lint "**/*.md" ".github/**/*.yml" ".github/**/*.yaml" --config .github/config/cspell.json --no-progress --no-summary --no-must-find-files
python3 .github/scripts/ci/check_action_pins.py
python3 .github/scripts/ci/check_file_headers.py
python3 .github/scripts/release/generate_contributors.py --check
actionlint
```

The optional local pre-commit suite runs the same main formatting, lint, documentation, spelling and file-header checks:
```bash
pre-commit install
pre-commit run --all-files
```

If a command cannot be run, state that clearly in the change notes.

## Documentation Style

Public documentation should:
- use British English
- keep bullet lists directly attached to the text above them
- keep fenced code blocks directly attached to the text above them
- avoid commas before `but`
- avoid commas before `including`, `however` and `or`
- avoid simple comma-before-`and` phrasing unless the comma separates larger topics
- avoid hype and maturity inflation
- state limitations clearly

## Compatibility Changes

Changes to `planwright.v1`, `planwright.graph.v1`, diagnostics, exit codes or command output are public contract changes. They require tests and documentation updates.

Compatibility-impacting changes should explain:
- what changed
- who is affected
- whether old input still works
- whether generated output changes
- whether a loss report or compatibility matrix entry needs updating

## File Headers And Licensing Notices

Planwright source files, scripts and other copyright-affected files that support normal comments should include:
```text
Copyright <year> The Planwright Authors
SPDX-License-Identifier: Apache-2.0
```

Use the comment syntax appropriate to the file type.

For example:
```go
// Copyright <year> The Planwright Authors
// SPDX-License-Identifier: Apache-2.0
```
```python
# Copyright <year> The Planwright Authors
# SPDX-License-Identifier: Apache-2.0
```

This rule is intended for Planwright-owned source files and similar project files where a notice is practical.

It does not require contributors to rewrite:
- vendored third-party material
- generated files where the header would be unstable or misleading
- files whose format makes a normal comment header impractical

Contributors should preserve existing valid headers and should not remove or weaken per-file licensing notices casually.

## AI-Assisted Contributions

AI tools may be used to assist with research, drafting, refactoring, testing or documentation, provided their use is disclosed clearly in the pull request where the template asks for it.

The human contributor remains fully responsible for the contribution. This includes correctness, security, licensing, originality and fitness for inclusion in Planwright.

AI systems cannot sign off commits under the DCO. Every commit must be signed off by a human author who understands the change and has the legal right to submit it.

## DCO

Contributions are expected to follow the Developer Certificate of Origin. See [docs/project/DCO.md](../docs/project/DCO.md).

All commits should be signed off:
```bash
git commit -s
```

By signing off a commit, you certify the contribution under the Developer Certificate of Origin. Pull requests containing unsigned commits will be declined.

## Security Issues

Do not open public issues for suspected vulnerabilities.

Use the process in [`SECURITY.md`](SECURITY.md) instead.

## Authors And Contributors

`AUTHORS` is the curated copyright-purpose author list.

`CONTRIBUTORS` is the recognition list generated from reachable non-bot Git history. In an unborn repository, it is seeded from `AUTHORS` until commit history exists.

## Project Direction

Contributions that conflict with the project's intended direction may be declined.

See:
- [`docs/architecture/design.md`](../docs/architecture/design.md)
- [`docs/architecture/threat-model.md`](../docs/architecture/threat-model.md)
- [`docs/compatibility.md`](../docs/compatibility.md)
- [`docs/project/GOVERNANCE.md`](../docs/project/GOVERNANCE.md)
