# Changelog

All notable project changes are tracked here. The project is pre-1.0; compatibility promises are limited to the documented version gate and compatibility matrix for each release.

## Unreleased

No unreleased changes yet.

<details>
<summary><h2>v0.12.0</h2></summary>

`v0.12.0` is the usability and proof release. It focuses on making the current engine easier to understand, run and review through one polished proof path rather than expanding compatibility.

### Added
- Added the canonical AWS web application proof walkthrough under `examples/aws-webapp-basic/README.md`.
- Added a README quick proof path before the full command catalogue.
- Added a README non-goals section clarifying that Planwright is not a one-click deployer, universal converter, compliance tool, live cloud scanner or drag-and-drop diagrammer in the current implementation.
- Added ADR 0013 for the proof-before-expansion decision.
- Added `docs/releases/v0.12.0.md`.
- Added CLI regression coverage for the full AWS web application proof path.
- Added CLI regression coverage for public database risk reporting and Terraform/OpenTofu generator refusal.

### Changed
- Reframed roadmap `v0.12` as a usability and proof release.
- Updated agent guidance to prioritise one polished proof path over broad feature growth.
- Linked the canonical example walkthrough from the documentation index.

</details>

<details>
<summary><h2>v0.11.0</h2></summary>

`v0.11.0` is the first planned published release. Earlier version sections below describe internal project-history milestones, not published release tags.

### Changed
- Reworked the roadmap so the full engine, importer, generator, UI, policy, release and security plan is expressed as version gates.
- Centralised runtime version metadata so CLI output and pack manifests use the same source.
- Hardened local file reads through scoped root access and reduced `gosec` suppressions.
- Changed generated output files to private `0600` permissions on platforms that support POSIX file modes.
- Clarified that `planwright serve` is currently a text-and-table workbench rather than the future drag-and-drop visual planner.
- Updated documentation style rules to cover commas before `including`, `however`, `or` and `then`.
- Removed fake Kubernetes Secret data from the Gateway API example.

### Added
- Added this changelog.
- Added a full v0.11 hardening gate to the roadmap.
- Added `golangci-lint` formatting and lint configuration with `gofumpt`, `goimports` and `gci`.
- Added pre-commit hooks for secret scanning, Go linting, shell checks, documentation style, spelling and file-header checks.
- Added native Go fuzz targets for parser, importer and review boundaries.
- Added a bounded fuzzing workflow.
- Added Sponsoring information to the README.

</details>

<details>
<summary><h2>v0.10.0</h2></summary>

### Added
- Added broad GitHub Actions CI for build, tests, lint, static analysis, vulnerability checks, module hygiene, documentation style, workflow validation and release verification.
- Added British-English documentation style checks and spell checking.
- Added repository metadata for ownership, authorship, contributor recognition, governance, support, security and release process.

</details>

<details>
<summary><h2>v0.9.0</h2></summary>

### Added
- Added built-in `lab`, `small-business` and `production` policy profiles for local graph JSON.
- Added Markdown and SARIF policy finding output.

</details>

<details>
<summary><h2>v0.8.0</h2></summary>

### Added
- Added embedded JSON Schema 2020-12 for the graph representation.
- Added graph schema export and graph JSON validation commands.

</details>

<details>
<summary><h2>v0.7.0</h2></summary>

### Added
- Added local graph diff review for two graph JSON files.
- Added findings for new public database exposure and new internet-facing network paths.

</details>

<details>
<summary><h2>v0.6.0</h2></summary>

### Added
- Added local AWS scan bundle import for selected AWS CLI JSON artefacts.
- Added inventory extraction for selected VPC, subnet, security group, EC2, RDS, S3, Lambda and ELBv2 data.

</details>

<details>
<summary><h2>v0.5.0</h2></summary>

### Added
- Added rendered Kubernetes manifest import.
- Added Gateway API route and backend relationship inference for the supported subset.
- Added Cilium policy inventory with explicit semantic loss notes.

</details>

<details>
<summary><h2>v0.4.0</h2></summary>

### Added
- Added loopback-only local web workbench through `planwright serve`.
- Added Host allowlist enforcement and static/API security headers.

</details>

<details>
<summary><h2>v0.3.0</h2></summary>

### Added
- Added Terraform plan JSON review.
- Added Markdown and SARIF output for Terraform plan findings.

</details>

<details>
<summary><h2>v0.2.0</h2></summary>

### Added
- Added CloudFormation and SAM subset import.
- Added graph JSON and Markdown loss report output for imported templates.

</details>

<details>
<summary><h2>v0.1.0</h2></summary>

### Added
- Added Terraform/OpenTofu-oriented generated files for the first AWS web application pattern.
- Added Mermaid diagram output and Markdown review reports.
- Added directory-based Planwright pack output.

</details>

<details>
<summary><h2>v0.0.1</h2></summary>

### Added
- Added the initial Go CLI, typed plan parser, graph model, validation diagnostics and AWS web application example.

</details>
