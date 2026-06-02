# Import Adapter Boundary

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright imports selected local artefacts from CloudFormation, SAM, Kubernetes manifests and AWS scan bundles. Future import paths may include Terraform state, provider schemas, Docker Compose, live read-only scans and plugin adapters.

Importers sit on an important trust boundary: they read user-selected files that may be malformed, ambiguous, lossy, overly large or hostile.

The project needs importers that keep review evidence clear and make loss visible without executing infrastructure definitions.

## Decision

Planwright importers are adapters from explicit local input into graph data, diagnostics and loss reports.

Importers must:
- parse source as data, not executable content
- reject unsafe paths, symlinked inputs and oversized inputs where the boundary requires it
- lower only understood constructs into `planwright.graph.v1`
- preserve source metadata and loss information where practical
- report unsupported, ambiguous, normalised, inferred and redacted constructs explicitly
- avoid cloud credentials, live API calls and external command execution unless a future ADR approves that boundary

Importers must not:
- silently drop source constructs that matter to compatibility or safety
- imply lossless conversion without fixture-backed round-trip proof
- run Terraform, OpenTofu, the AWS CLI, `kubectl`, Helm, Kustomize or provider plugins as part of the current import boundary
- decode or expose secret values by default

## Consequences

This decision means that:
- importers may produce partial graphs with loss reports rather than pretending to understand every source construct
- compatibility levels must describe the actual supported subset
- importer tests should cover malformed input, duplicate keys, unsupported constructs, path safety and redaction
- future live read-only scan work requires a separate decision because it crosses from local artefact parsing into credentialed API interaction
- plugin adapters require a deliberate trust and execution model before they are accepted
