# Threat Model

This document describes the current Planwright v0.11 threat model.
- [Assets](#assets)
- [Trust Boundaries](#trust-boundaries)
- [Non-Goals](#non-goals)
- [Current Controls](#current-controls)

## Assets

Current assets:
- user-authored plan files
- user-selected CloudFormation/SAM template files
- user-selected Terraform plan JSON files
- user-selected Kubernetes manifest files or direct manifest directories
- user-selected AWS scan bundle directories containing local JSON files
- user-selected Planwright graph JSON files for diff review
- user-selected Planwright graph JSON files for schema and semantic validation
- user-selected Planwright graph JSON files for policy profile review
- embedded Planwright graph JSON Schema
- built-in policy profile metadata
- generated diagnostics
- graph data derived from plan files
- graph data derived from imported templates
- graph data derived from imported Kubernetes manifests
- graph data derived from imported AWS scan bundle JSON
- review findings derived from Terraform plan JSON
- local filesystem paths selected by the user
- local generated Terraform/OpenTofu-oriented files
- local generated reports, diagrams and pack manifests
- local imported-template loss reports
- local Kubernetes manifest loss reports
- local AWS scan bundle loss reports
- local Terraform plan review Markdown and SARIF reports
- local graph diff Markdown reports
- local graph schema JSON output
- local policy profile Markdown and SARIF reports
- local HTTP requests to the `planwright serve` workbench
- plan text posted to the local workbench for in-memory validation

Cloud credentials, live account state, Kubernetes cluster state and applied deployment state are not in the current implementation.

## Trust Boundaries

Current trust boundaries:
- plan files are user input and must be validated before use
- imported CloudFormation/SAM templates are user input and must be parsed without execution
- Terraform plan JSON files are user input and must be parsed without execution
- Kubernetes manifests are user input and must be parsed without execution
- AWS scan bundle JSON files are user input and must be parsed without execution
- Planwright graph JSON files are user input and must be parsed without execution
- policy profile IDs are user input and must resolve only to built-in profiles
- graph schema output paths are user input
- local filesystem paths are user input
- browser-posted plan text is user input and must be validated before use
- browser requests to `planwright serve` cross an HTTP boundary even when loopback-only
- terminal output may be copied into issues, reviews or documentation
- generated Terraform/OpenTofu-oriented files may be reviewed or executed by a human outside Planwright

## Non-Goals

Planwright v0.11 does not defend against:
- a compromised operating system account
- malicious local users with write access to the same repository
- malicious local users who can connect to the same loopback service while it is running
- terminal history capture
- supply-chain compromise of the Go toolchain or dependencies

## Current Controls

The current controls are:
- no network listener unless `planwright serve` is explicitly started
- no cloud API calls
- no Kubernetes API calls
- no AWS API calls
- no credential loading
- no infrastructure mutation
- no `kubectl`, Helm or Kustomize execution
- no AWS CLI execution
- local web server defaults to loopback and rejects non-loopback bind addresses
- local web server rejects unexpected Host headers
- local web server sets CSP, `nosniff`, no-referrer and frame-denial headers
- local web server does not set permissive CORS headers
- local web API applies a request body limit
- local web API validates browser-posted plan text in memory and does not write files
- plan files must be regular files, not directories or symlinks
- imported template files must be regular files, not directories or symlinks
- selected Kubernetes manifest files must be regular files, not directories or symlinks
- selected Kubernetes manifest directories and their direct manifest files must not be symlinks
- Kubernetes manifest import only reads direct `.yaml`, `.yml` and `.json` files from a selected directory
- Kubernetes manifest import applies a per-file size limit
- Kubernetes Secret values are not lowered into graph properties, diagnostics or loss reports
- Kubernetes import results keep source path and size metadata but do not expose raw manifest bytes
- NetworkPolicy and Cilium policy semantics are reported as partial rather than treated as fully modelled
- AWS scan bundle directories and their direct JSON files must not be symlinks
- AWS scan bundle import only reads direct `.json` files from the selected directory
- AWS scan bundle import applies a per-file size limit
- AWS scan bundle import does not expose raw source bytes through importer results
- AWS scan bundle import does not lower STS ARN or UserId values into graph properties or loss messages
- graph diff input files must be regular files, not directories or symlinks
- graph diff input files are size-limited and decoded as Planwright graph JSON
- graph diff validates both graph inputs before writing a report
- graph diff compares local graph artefacts only and does not prove live infrastructure drift
- graph validation input files must be regular files, not directories or symlinks
- graph validation uses the embedded schema and does not fetch schemas from the network
- graph schema export writes only the embedded schema to an explicit local output path
- policy profile review input files must be regular files, not directories or symlinks
- policy profile review validates graph JSON before evaluation
- policy profile review uses built-in checks only and does not execute custom policy code
- policy profile review writes Markdown and SARIF only to explicit local output paths
- policy profile review does not certify compliance or decide deployment safety
- Terraform plan JSON files must be regular files, not directories or symlinks
- generated output paths are checked for traversal, symlinked output roots and symlinked output ancestors
- explicit graph and loss-report output paths refuse symlink targets
- generated files keep database passwords as variables rather than literal values
- explicit plan validation
- tests for malformed input, unsafe plan paths, unsafe template paths, unsupported imported resources and graph validation failures
- tests for local web Host rejection, security headers, malformed browser input and oversized browser input
- Terraform review reports avoid printing full before/after resource values

Future archive parsing/writing, generated scripts, Helm/Kustomize support, AWS SDK support, custom policy packs, OPA/Rego integration and live scans must update this threat model before implementation.
