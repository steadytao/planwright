<div align="center">

[![Planwright](./.github/banner.svg)](#readme)

[![Release version](https://img.shields.io/badge/Download-latest-2064FC.svg?style=for-the-badge&labelColor=031835)](#installation "Installation")
[![Sponsor](https://img.shields.io/badge/_-Sponsor-495977.svg?logo=githubsponsors&logoColor=white&labelColor=031835&style=for-the-badge)](https://github.com/sponsors/steadytao "Sponsor")
[![License: Apache 2.0](https://img.shields.io/badge/-Apache_2.0-2064FC.svg?style=for-the-badge&labelColor=031835)](LICENSE "Licence")
[![CI Status](https://img.shields.io/github/actions/workflow/status/steadytao/planwright/ci-main.yml?branch=main&label=Tests&style=for-the-badge&labelColor=031835)](https://github.com/steadytao/planwright/actions/workflows/ci-main.yml "CI Status")
[![Go Report Card](https://img.shields.io/badge/Go_Report-A%2B-2064FC.svg?style=for-the-badge&labelColor=031835)](https://goreportcard.com/report/github.com/steadytao/planwright "Go Report Card")
[![OpenSSF Best Practices](https://img.shields.io/badge/OpenSSF_Best_Practices-passing-2064FC.svg?style=for-the-badge&labelColor=031835)](https://www.bestpractices.dev/en/projects/13072 "OpenSSF Best Practices")
[![OpenSSF Scorecard](https://img.shields.io/ossf-scorecard/github.com/steadytao/planwright?label=Scorecard&style=for-the-badge&labelColor=031835)](https://scorecard.dev/viewer/?uri=github.com/steadytao/planwright "OpenSSF Scorecard")
[![Commits](https://img.shields.io/github/commit-activity/m/steadytao/planwright?label=commits&style=for-the-badge&labelColor=031835)](https://github.com/steadytao/planwright/commits "Commit History")

</div>

Planwright is a local-first infrastructure planning engine. It turns typed plans and selected infrastructure-as-code artefacts into a reviewable architecture graph then generates evidence: validation results, security notes, cost notes, loss reports, diagrams and deployment packs.

Planwright treats infrastructure conversion as a migration and evidence problem, not a syntax conversion problem.
- [INSTALLATION](#installation)
- [Release Files](#release-files)
- [Update](#update)
- [Dependencies](#dependencies)
- [Compile](#compile)
- [Quick Proof Path](#quick-proof-path)
- [Usage and Options](#usage-and-options)
- [Examples](#examples)
- [What Planwright Is Not](#what-planwright-is-not)
- [Current Scope](#current-scope)
- [Compatibility](#compatibility)
- [Safety Boundaries](#safety-boundaries)
- [Documentation](#documentation)
- [Development](#development)
- [Contributing](#contributing)
- [Governance](#governance)
- [Support](#support)
- [Security](#security)
- [Sponsoring](#sponsoring)
- [Licence](#licence)
- [Changelog](#changelog)

# INSTALLATION

[![Windows](https://img.shields.io/badge/-Windows_x64-2064FC.svg?style=for-the-badge&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAyNCAyNCI%2BPHBhdGggZmlsbD0id2hpdGUiIGQ9Ik0zIDQuNiAxMC42IDMuNnY3LjJIM1Y0LjZabTguNi0xLjFMMjEgMi4ydjguNmgtOS40VjMuNVpNMyAxMmg3LjZ2Ny4yTDMgMTguMlYxMlptOC42IDBIMjF2OC42bC05LjQtMS4zVjEyWiIvPjwvc3ZnPg%3D%3D&logoColor=white&labelColor=031835)](https://github.com/steadytao/planwright/releases/latest/download/planwright_windows_amd64.exe "Windows x64")
[![Linux](https://img.shields.io/badge/-Linux_x64-2064FC.svg?style=for-the-badge&logo=linux&logoColor=white&labelColor=031835)](https://github.com/steadytao/planwright/releases/latest/download/planwright_linux_amd64 "Linux x64")
[![macOS](https://img.shields.io/badge/-macOS-2064FC.svg?style=for-the-badge&logo=apple&logoColor=white&labelColor=031835)](https://github.com/steadytao/planwright/releases/latest/download/planwright_darwin_arm64 "macOS Arm64")
[![Source TAR](https://img.shields.io/badge/-Source_TAR-495977.svg?style=for-the-badge&labelColor=031835)](https://github.com/steadytao/planwright/tarball/main "Source TAR")
[![Other variants](https://img.shields.io/badge/-Other-364457.svg?style=for-the-badge&labelColor=031835)](#alternatives "Alternative Downloads")
[![All versions](https://img.shields.io/badge/-All_Versions-031835.svg?style=for-the-badge&labelColor=020F22)](https://github.com/steadytao/planwright/releases "All Versions")

You can install Planwright using release binaries or by compiling from source.

## RELEASE FILES

#### Recommended

File | Description
:--- | :---
[planwright_windows_amd64.exe](https://github.com/steadytao/planwright/releases/latest/download/planwright_windows_amd64.exe) | Windows x64 standalone executable, recommended for Windows x64
[planwright_linux_amd64](https://github.com/steadytao/planwright/releases/latest/download/planwright_linux_amd64) | Linux x64 standalone executable, recommended for Linux x64
[planwright_darwin_arm64](https://github.com/steadytao/planwright/releases/latest/download/planwright_darwin_arm64) | macOS Apple Silicon standalone executable, recommended for Apple Silicon Macs

#### Alternatives

File | Description
:--- | :---
[planwright_windows_arm64.exe](https://github.com/steadytao/planwright/releases/latest/download/planwright_windows_arm64.exe) | Windows Arm64 standalone executable
[planwright_linux_arm64](https://github.com/steadytao/planwright/releases/latest/download/planwright_linux_arm64) | Linux Arm64 standalone executable
[planwright_darwin_amd64](https://github.com/steadytao/planwright/releases/latest/download/planwright_darwin_amd64) | macOS Intel standalone executable
[planwright_linux_amd64_desktop.zip](https://github.com/steadytao/planwright/releases/latest/download/planwright_linux_amd64_desktop.zip) | Linux x64 desktop metadata package with the Planwright icon
[planwright_linux_arm64_desktop.zip](https://github.com/steadytao/planwright/releases/latest/download/planwright_linux_arm64_desktop.zip) | Linux Arm64 desktop metadata package with the Planwright icon
[planwright_darwin_amd64_app.zip](https://github.com/steadytao/planwright/releases/latest/download/planwright_darwin_amd64_app.zip) | macOS Intel `.app` bundle with the Planwright icon
[planwright_darwin_arm64_app.zip](https://github.com/steadytao/planwright/releases/latest/download/planwright_darwin_arm64_app.zip) | macOS Apple Silicon `.app` bundle with the Planwright icon

#### Misc

File | Description
:--- | :---
[SHA2-256SUMS](https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS) | SHA-256 checksum manifest
[SHA2-256SUMS.sig](https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS.sig) | OpenPGP signature for `SHA2-256SUMS`
[SHA2-512SUMS](https://github.com/steadytao/planwright/releases/latest/download/SHA2-512SUMS) | SHA-512 checksum manifest
[SHA2-512SUMS.sig](https://github.com/steadytao/planwright/releases/latest/download/SHA2-512SUMS.sig) | OpenPGP signature for `SHA2-512SUMS`
[public.key](https://github.com/steadytao/planwright/releases/latest/download/public.key) | Convenience copy of the release public key
[planwright_sbom.spdx.json](https://github.com/steadytao/planwright/releases/latest/download/planwright_sbom.spdx.json) | SPDX JSON SBOM
[planwright_sbom.cdx.json](https://github.com/steadytao/planwright/releases/latest/download/planwright_sbom.cdx.json) | CycloneDX JSON SBOM

Planwright signs checksum manifests with a maintainer-controlled OpenPGP release key. Treat the documented fingerprint in [docs/releases/signing.md](docs/releases/signing.md), a maintainer-controlled announcement or another trusted channel as the trust root; `public.key` is only a convenience release asset.

Example verification on Linux:
```bash
curl -LO https://github.com/steadytao/planwright/releases/latest/download/public.key
curl -LO https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS
curl -LO https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS.sig
gpg --import ./public.key
gpg --verify ./SHA2-256SUMS.sig ./SHA2-256SUMS
sha256sum -c ./SHA2-256SUMS --ignore-missing
gh attestation verify ./planwright_linux_amd64 -R steadytao/planwright
```

Example verification on Windows:
```powershell
Invoke-WebRequest -Uri "https://github.com/steadytao/planwright/releases/latest/download/public.key" -OutFile "public.key"
Invoke-WebRequest -Uri "https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS" -OutFile "SHA2-256SUMS"
Invoke-WebRequest -Uri "https://github.com/steadytao/planwright/releases/latest/download/SHA2-256SUMS.sig" -OutFile "SHA2-256SUMS.sig"
gpg --import .\public.key
gpg --verify .\SHA2-256SUMS.sig .\SHA2-256SUMS
(Get-FileHash .\planwright_windows_amd64.exe -Algorithm SHA256).Hash.ToLower()
Select-String -Path .\SHA2-256SUMS -Pattern "planwright_windows_amd64.exe"
gh attestation verify .\planwright_windows_amd64.exe -R steadytao/planwright
```

# UPDATE

Planwright does not currently include a self-update command. Download the new release binary and replace the old executable after verifying the checksum manifest.

# DEPENDENCIES

Planwright release binaries are intended to run without Terraform, OpenTofu, AWS CLI, kubectl, Helm, Kustomize, OPA or Rego installed. Those tools are not invoked by Planwright currently; they may still be useful around Planwright for preparing input files or reviewing generated output.

For source builds, install:
- Go as declared by [go.mod](go.mod)
- Git
- OpenPGP tooling such as GnuPG if you want to verify release manifests

# COMPILE

Build the CLI from source:
```bash
git clone https://github.com/steadytao/planwright.git
cd planwright
go build -o planwright ./cmd/planwright
./planwright version
```

On Windows:
```powershell
git clone https://github.com/steadytao/planwright.git
cd planwright
go build -o planwright.exe ./cmd/planwright
.\planwright.exe version
```

# QUICK PROOF PATH

The quickest useful Planwright path is the AWS web application example:
```bash
go run ./cmd/planwright validate examples/aws-webapp-basic/planwright.yaml
go run ./cmd/planwright risks examples/aws-webapp-basic/planwright.yaml
go run ./cmd/planwright cost-notes examples/aws-webapp-basic/planwright.yaml
go run ./cmd/planwright generate terraform examples/aws-webapp-basic/planwright.yaml --out ./generated/terraform
go run ./cmd/planwright generate mermaid examples/aws-webapp-basic/planwright.yaml --out ./generated/diagrams
go run ./cmd/planwright pack examples/aws-webapp-basic/planwright.yaml --out ./planwright-pack
```

That path produces:
- validation output
- Markdown security and cost notes
- Terraform/OpenTofu-oriented review files
- a Mermaid architecture diagram
- a directory-based Planwright pack with a manifest, graph, reports, generated files and diagrams

Read the full walkthrough in [examples/aws-webapp-basic/README.md](examples/aws-webapp-basic/README.md).

# USAGE AND OPTIONS

Planwright is CLI-first. The current command surface is:
```text
planwright validate <planwright.yaml>
planwright validate-graph <planwright.graph.json>
planwright explain <planwright.yaml>
planwright generate terraform <planwright.yaml> --out <dir>
planwright generate mermaid <planwright.yaml> --out <dir>
planwright risks <planwright.yaml>
planwright cost-notes <planwright.yaml>
planwright docs check [path ...]
planwright import cloudformation <template.yaml> --out <graph.json> --loss-report <loss.md>
planwright import sam <template.yaml> --out <graph.json> --loss-report <loss.md>
planwright import k8s <manifest-path-or-dir> --out <graph.json> --loss-report <loss.md>
planwright import awsscan <bundle-dir> --out <graph.json> --loss-report <loss.md>
planwright diff <old.graph.json> <new.graph.json> --out <review.md>
planwright schema graph --out <schema.json>
planwright policy profiles
planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>
planwright pack <planwright.yaml> --out <dir>
planwright review terraform-plan <tfplan.json> --out <review.md> --sarif <planwright.sarif>
planwright serve [project-dir] [--addr 127.0.0.1:5786]
planwright version
```

Planned commands such as `tui`, `scan aws`, `generate kubernetes`, custom policy packs and OPA/Rego policy execution are not part of v0.11.

# EXAMPLES

Canonical walkthrough:
- [examples/aws-webapp-basic/README.md](examples/aws-webapp-basic/README.md), the current v0.12 proof path

Validate the example plan:
```bash
go run ./cmd/planwright validate examples/aws-webapp-basic/planwright.yaml
```

Generate Terraform/OpenTofu-oriented review files:
```bash
go run ./cmd/planwright generate terraform examples/aws-webapp-basic/planwright.yaml --out ./generated/terraform
```

Generate a Mermaid diagram:
```bash
go run ./cmd/planwright generate mermaid examples/aws-webapp-basic/planwright.yaml --out ./generated/diagrams
```

Create a local Planwright pack directory:
```bash
go run ./cmd/planwright pack examples/aws-webapp-basic/planwright.yaml --out ./planwright-pack
```

Review a Terraform plan JSON fixture:
```bash
go run ./cmd/planwright review terraform-plan examples/terraform-plan-risk-review/tfplan.json --out ./generated/terraform-review.md --sarif ./generated/planwright.sarif
```

Start the current local browser workbench:
```bash
go run ./cmd/planwright serve . --addr 127.0.0.1:5786
```

# WHAT PLANWRIGHT IS NOT

Planwright is not:
- a one-click deployment console
- a live cloud account scanner in the current implementation
- a Terraform or OpenTofu runner
- a Kubernetes cluster client
- a compliance certification tool
- a lossless universal IaC converter
- a drag-and-drop cloud diagrammer in the current local web workbench
- a replacement for infrastructure review, threat modelling, cost review or deployment planning

The current value is local evidence generation. Planwright should make infrastructure easier to inspect before deployment; it should not hide uncertainty behind compatibility claims.

# CURRENT SCOPE

Planwright is in the v0.12 usability and proof stage. The immediate focus is making the existing engine easy to understand, run and review through one polished proof path rather than expanding the feature surface.

Current implemented surfaces:
- Go CLI
- `planwright.v1` typed plan parser
- `planwright.graph.v1` architecture graph model
- JSON Schema 2020-12 export and local graph validation
- graph lowering for the first AWS web application pattern
- Terraform/OpenTofu-oriented review output for that first AWS pattern
- Mermaid architecture diagram output
- Markdown reports for security, cost, deployability, cleanup and assumptions
- directory-based Planwright pack output
- CloudFormation and SAM subset import with loss reports
- rendered Kubernetes manifest import with Gateway API and Cilium inventory support
- local AWS scan bundle import from selected AWS CLI JSON artefacts
- Terraform plan JSON review with Markdown and SARIF output
- local graph JSON diff review
- built-in `lab`, `small-business` and `production` policy profile review
- text-and-table local browser workbench through `planwright serve`
- GitHub Actions CI for tests, linting, static analysis, vulnerabilities, docs, workflow validation and supply-chain hygiene

Important deferred surfaces:
- Terraform state import
- Terraform HCL or module evaluation
- Terraform provider schema ingestion
- Pulumi import
- Kubernetes generation
- Kubernetes live cluster scans
- live AWS account scans
- AWS SDK integration
- AWS credential loading
- complete CloudFormation or SAM import
- lossless conversion or round-trip import
- zip archive pack output
- hosted demo
- drag-and-drop visual canvas editing
- resource-card graph planning
- Terraform/OpenTofu execution
- live drift proof
- custom policy packs
- OPA/Rego integration
- compliance certification
- reproducible build guarantees

# COMPATIBILITY
Level | Meaning
:--- | :---
0 | Preserved only as original source
1 | Syntax parsed
2 | Resource inventory extracted
3 | Relationships inferred
4 | Lowered into Planwright graph
5 | Generated into another format
6 | Round-trip tested with fixtures
7 | Deployability tested in a sandbox
8 | Production-profile validated

Current support is intentionally partial. The typed plan path reaches level 5 for the first built-in AWS web application pattern. CloudFormation, SAM, rendered Kubernetes manifests and local AWS scan bundles reach level 4 for their supported subsets. Terraform plan JSON review is level 2 plus selected review findings because it does not lower Terraform plans into the Planwright graph yet.

See [docs/compatibility.md](docs/compatibility.md) for the matrix and current limitations.

# SAFETY BOUNDARIES

Planwright defaults to local analysis.

In the current implementation it:
- reads explicit local plan, graph, template, manifest and scan-bundle files
- emits local validation, review and explanation output
- writes local generated artefacts only when explicitly asked
- runs a loopback-only local browser workbench when explicitly started
- rejects unexpected Host headers in the local web server
- does not read cloud credentials
- does not contact cloud APIs
- does not contact Kubernetes clusters
- does not deploy or destroy infrastructure
- does not execute imported content
- does not run the AWS CLI, AWS SDK, `kubectl`, Helm, Kustomize, Terraform, OpenTofu, OPA or Rego
- does not certify compliance

Future live scan features must remain read-only by default and require explicit account, region and identity confirmation.

# DOCUMENTATION

Start with [docs/README.md](docs/README.md).

Core documents:
- [docs/roadmap.md](docs/roadmap.md), full version-gated roadmap
- [docs/compatibility.md](docs/compatibility.md), compatibility matrix and import boundaries
- [docs/architecture/design.md](docs/architecture/design.md), engine design and product boundary
- [docs/architecture/threat-model.md](docs/architecture/threat-model.md), current local-first threat model
- [docs/architecture/decisions/](docs/architecture/decisions/), ADRs
- [docs/releases/README.md](docs/releases/README.md), release process
- [docs/releases/signing.md](docs/releases/signing.md), release signing model

# DEVELOPMENT

Useful local checks:
```bash
go test ./...
go vet ./...
go build ./cmd/planwright
golangci-lint config verify --config .github/config/golangci.yml
golangci-lint fmt --config .github/config/golangci.yml --diff
golangci-lint run --config .github/config/golangci.yml
go test -race ./...
staticcheck ./...
gosec ./...
govulncheck ./...
go mod verify
go mod tidy -diff
go run ./cmd/planwright docs check .
npx --yes cspell@10.0.1 lint "**/*.md" ".github/**/*.yml" ".github/**/*.yaml" --config .github/config/cspell.json --no-progress --no-summary --no-must-find-files
python3 .github/scripts/ci/check_action_pins.py
python3 .github/scripts/ci/check_file_headers.py
python3 .github/scripts/release/generate_contributors.py --check
actionlint
pre-commit run --all-files
```

Each roadmap version must finish with a full codebase pass for security, correctness, compatibility, documentation consistency and accidental generated artefacts before the next version starts.

# CONTRIBUTING

Before opening an issue or pull request, read [.github/CONTRIBUTING.md](.github/CONTRIBUTING.md).

All commits must be signed off in accordance with the Developer Certificate of Origin. See [docs/project/DCO.md](docs/project/DCO.md).

# GOVERNANCE

Governance is documented in [docs/project/GOVERNANCE.md](docs/project/GOVERNANCE.md), [docs/project/MAINTAINERS.md](docs/project/MAINTAINERS.md) and [.github/CODEOWNERS](.github/CODEOWNERS).

# SUPPORT

Support expectations are documented in [.github/SUPPORT.md](.github/SUPPORT.md).

# SECURITY

Please do not report security vulnerabilities in public issues.

See [.github/SECURITY.md](.github/SECURITY.md) for reporting guidance.

# SPONSORING

Planwright is independent open-source infrastructure tooling.

If Planwright is useful to your work or organisation, sponsorship helps fund maintenance, compatibility fixtures, security review and release infrastructure.

Sponsor the project through [GitHub Sponsors](https://github.com/sponsors/steadytao).

# LICENCE

Planwright is released under the Apache License 2.0. See [LICENSE](LICENSE).

[AUTHORS](AUTHORS) is the curated copyright-purpose author list.

[CONTRIBUTORS](CONTRIBUTORS) is the contributor-recognition list.

# CHANGELOG

See [CHANGELOG.md](CHANGELOG.md).
