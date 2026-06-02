# Design

Planwright is a local-first infrastructure planning, compatibility and evidence engine.

The core design principle is that all interfaces talk to the same engine.
- [Product Boundary](#product-boundary)
- [Current Architecture](#current-architecture)
- [Later Architecture](#later-architecture)

## Product Boundary

Planwright should help users understand infrastructure before deploying, migrating or reviewing it. It should preserve source intent and make ambiguity visible.

It should not pretend that unrelated infrastructure formats can always be converted without loss.

## Current Architecture

The current implementation contains:
- a CLI entrypoint in `cmd/planwright`
- command handling in `internal/cli`
- typed plan parsing in `internal/plan`
- CloudFormation/SAM subset import in `internal/importers/cloudformation`
- Kubernetes/Gateway API/Cilium manifest subset import in `internal/importers/kubernetes`
- AWS scan bundle subset import in `internal/importers/awsscan`
- Terraform plan JSON review in `internal/review/terraformplan`
- graph types and validation in `internal/graph`
- graph diffing in `internal/graph`
- graph JSON Schema validation in `internal/graph`
- built-in policy profile review in `internal/policy`
- text report rendering in `internal/reports`
- Terraform/OpenTofu-oriented rendering in `internal/generators/terraform`
- Mermaid rendering in `internal/generators/mermaid`
- safe local artefact writing in `internal/project`
- local web workbench serving in `internal/server`

The current pipeline is:
```text
planwright.yaml
  -> plan parser
  -> graph lowering
  -> graph validation
  -> CLI diagnostics, explanation output, generators or reports
```

The current import pipeline is:
```text
CloudFormation/SAM template
  -> source parser
  -> supported resource inventory
  -> planwright.graph.v1
  -> graph validation
  -> graph JSON and loss report
```

The current Kubernetes import pipeline is:
```text
rendered Kubernetes YAML/JSON manifest file or direct manifest directory
  -> local manifest reader
  -> multi-document parser
  -> supported Kubernetes, Gateway API and Cilium resource inventory
  -> conservative relationship inference
  -> planwright.graph.v1
  -> graph validation
  -> graph JSON and loss report
```

The Kubernetes importer does not contact a cluster, run `kubectl`, evaluate Helm templates, run Kustomize or decode Secret values.

The current AWS scan bundle import pipeline is:
```text
local AWS CLI JSON bundle directory
  -> direct child JSON reader
  -> supported AWS service output parsers
  -> conservative relationship inference
  -> planwright.graph.v1
  -> graph validation
  -> graph JSON and loss report
```

The AWS scan bundle importer does not load credentials, call AWS APIs, run the AWS CLI, use the AWS SDK or verify live account identity.

The current graph diff pipeline is:
```text
old planwright.graph.v1 JSON + new planwright.graph.v1 JSON
  -> local graph JSON readers
  -> graph validation
  -> deterministic graph comparison
  -> Markdown graph diff review
```

The graph diff command compares local graph artefacts only. It does not contact cloud APIs, inspect Terraform state, contact Kubernetes clusters or prove live drift.

The current policy profile pipeline is:
```text
planwright.graph.v1 JSON
  -> local graph JSON reader
  -> JSON Schema and semantic graph validation
  -> built-in policy profile evaluation
  -> Markdown policy review
  -> SARIF policy findings
```

The policy profile command uses static built-in checks only. It does not run custom policy code, execute OPA/Rego, contact cloud APIs, inspect Kubernetes clusters or certify compliance.

The current graph schema pipeline is:
```text
planwright.graph.v1 JSON
  -> JSON Schema 2020-12 structural validation
  -> Planwright semantic graph validation
  -> CLI diagnostics
```

The schema is embedded in the CLI and can be written with `planwright schema graph --out <schema.json>`. Planwright does not fetch schemas from the network at validation time.

The current Terraform review pipeline is:
```text
terraform show -json plan output
  -> local JSON parser
  -> plan action and selected AWS attribute review
  -> Markdown review report
  -> SARIF results
```

The current local web pipeline is:
```text
browser-posted planwright.yaml text
  -> local loopback HTTP handler
  -> plan parser
  -> graph lowering and validation
  -> in-memory report and generator previews
  -> JSON response to the browser
```

## Later Architecture

Later versions can add:
- additional importers
- broader CloudFormation/SAM coverage
- Terraform state JSON import
- provider schema ingestion
- additional generators
- zip archive packs
- TUI
- live read-only scans

Those features must use the engine rather than private interface logic.
