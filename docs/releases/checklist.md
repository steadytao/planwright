# Release Checklist

This checklist keeps Planwright releases tied to actual shipped behaviour.
- [Functional Baseline](#functional-baseline)
- [Documentation](#documentation)
- [Verification](#verification)
- [Supply Chain Integrity](#supply-chain-integrity)
- [Release Signing](#release-signing)
- [Scope Discipline](#scope-discipline)
- [Final Release Preparation](#final-release-preparation)

# Functional Baseline

Before a release, confirm that Planwright can:
- run `planwright version`
- run `planwright validate examples/aws-webapp-basic/planwright.yaml`
- run `planwright explain examples/aws-webapp-basic/planwright.yaml`
- run `planwright generate terraform examples/aws-webapp-basic/planwright.yaml --out <dir>`
- run `planwright generate mermaid examples/aws-webapp-basic/planwright.yaml --out <dir>`
- run `planwright pack examples/aws-webapp-basic/planwright.yaml --out <dir>`
- run `planwright import cloudformation examples/cloudformation-basic/template.yaml --out <graph.json> --loss-report <loss.md>`
- run `planwright import sam examples/sam-basic/template.yaml --out <graph.json> --loss-report <loss.md>`
- run `planwright import k8s examples/kubernetes-gateway-basic/manifests.yaml --out <graph.json> --loss-report <loss.md>`
- run `planwright import awsscan examples/aws-scan-bundle-basic --out <graph.json> --loss-report <loss.md>`
- run `planwright review terraform-plan examples/terraform-plan-risk-review/tfplan.json --out <review.md> --sarif <planwright.sarif>`
- run `planwright schema graph --out <schema.json>`
- run `planwright validate-graph <planwright.graph.json>`
- run `planwright policy profiles`
- run `planwright policy graph <planwright.graph.json> --profile lab --out <policy.md> --sarif <policy.sarif>`
- run `planwright diff <old.graph.json> <new.graph.json> --out <review.md>`
- start `planwright serve . --addr 127.0.0.1:5786` and confirm it remains a loopback-only text-and-table workbench

For `v0.11.0`, this means the first formal release-tracked baseline: the current local engine, importer, review, policy, local web, CI, governance and release-readiness surface. Earlier version gates are project-history milestones, not published release tags.

# Documentation

Before a release, confirm that:
- `README.md` describes the shipped state rather than planned behaviour
- `docs/README.md` maps the current documentation surface
- `docs/compatibility.md` matches implemented compatibility levels
- `docs/architecture/design.md` matches the current pipeline boundaries
- `docs/architecture/threat-model.md` matches the current safety posture
- `docs/architecture/decisions/` records current project-boundary decisions
- `docs/releases/<version>.md` follows `docs/releases/template.md`
- `docs/releases/<version>.md` contains the changelog marker
- `CHANGELOG.md` describes project-history milestones and the release line honestly
- `CONTRIBUTORS` matches the reachable non-bot commit history for the release commit

# Verification

Before a release, confirm that:
- `go build ./cmd/planwright` passes
- `go vet ./...` passes
- `staticcheck ./...` passes
- `gosec ./...` passes
- `govulncheck ./...` passes
- `go test ./...` passes
- `go test -race ./...` passes where runtime is reasonable
- `go mod verify` passes
- `go mod tidy -diff` passes
- file header checks pass
- documentation style checks pass
- spelling checks pass
- workflow validation passes
- action pin checks pass
- contributor generation checks pass
- release verification builds a snapshot release surface
- CI is green across all required runners

If a release changes behaviour without updating tests or documentation, it is not ready.

# Supply Chain Integrity

Before closing a release, confirm that GoReleaser and the release asset preparation script generate the expected user install assets:
- `planwright_windows_amd64.exe`
- `planwright_windows_arm64.exe`
- `planwright_linux_amd64`
- `planwright_linux_arm64`
- `planwright_darwin_amd64`
- `planwright_darwin_arm64`
- `planwright_linux_amd64_desktop.zip`
- `planwright_linux_arm64_desktop.zip`
- `planwright_darwin_amd64_app.zip`
- `planwright_darwin_arm64_app.zip`

Then confirm that:
- `SHA2-256SUMS` is generated
- `SHA2-512SUMS` is generated
- `SHA2-256SUMS.sig` verifies against `SHA2-256SUMS`
- `SHA2-512SUMS.sig` verifies against `SHA2-512SUMS`
- published binaries verify cleanly against `SHA2-256SUMS`
- `public.key` is attached to the release
- `planwright_sbom.spdx.json` is attached to the release
- `planwright_sbom.cdx.json` is attached to the release
- both SBOM files are valid JSON
- both SBOM files are listed in `SHA2-256SUMS` and `SHA2-512SUMS`
- the verification commands documented in [the release docs index](README.md) were tested against the release assets

# Release Signing

Before publishing a signed release, confirm that:
- `docs/releases/signing.md` contains the release key fingerprint
- the release key fingerprint is published through a maintainer-controlled channel
- the `release` environment exists in GitHub
- the `release` environment requires maintainer approval before secrets are exposed
- `RELEASE_SIGNING_PRIVATE_KEY` is configured as a GitHub Actions secret available to the release workflow
- `RELEASE_SIGNING_PASSPHRASE` is configured as a GitHub Actions secret available to the release workflow
- `RELEASE_SIGNING_KEY_FINGERPRINT` is configured as a GitHub Actions variable available to the release workflow
- the release workflow fails closed if signing secrets are missing
- private key material is not committed or logged and is not copied into release notes

# Scope Discipline

Before a release, confirm that the release has not silently drifted into:
- hosted SaaS behaviour
- live cloud scans
- cloud account mutation
- browser-triggered credential use
- Terraform/OpenTofu execution
- Kubernetes cluster mutation
- custom policy execution
- vulnerability scanning
- compliance certification
- lossless conversion claims
- security claims the implementation cannot support

Planwright releases should stay narrow, local-first and defensible.

# Final Release Preparation

Before tagging:
- review open milestone items
- confirm branch protection and required CI checks on `main`
- confirm GitHub DCO app enforcement is active if the repository relies on signed-off commits
- regenerate `CONTRIBUTORS` and commit any real changes before tagging
- update release notes
- update `CHANGELOG.md`
- confirm the version to tag
- confirm docs and release scripts from a clean checkout

If the README still needs to explain away missing core behaviour, the release is not ready.
