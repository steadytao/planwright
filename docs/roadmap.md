# Roadmap

Planwright advances through version gates. Each gate requires a full codebase pass for security, correctness, documentation consistency, compatibility claims and accidental artefacts before the next gate starts.

Planwright's product thesis is both simple and stable; infrastructure conversion is a migration and evidence problem, not a syntax conversion problem.

The project should remain engine-first:
```text
typed plan / imports / CLI / local web / future TUI
  -> Planwright Engine
  -> graph, provenance, validation, analysis and reports
  -> generated IaC, diagrams, SARIF, deployment packs and review evidence
```

## Non-Negotiable Principles

Every roadmap gate must preserve these rules:
- the engine is the product; interfaces must not grow private infrastructure logic
- compatibility is loss-aware; unsupported, ambiguous and unsafe constructs must be reported
- local-first is the default; hosted pages are demos only unless governance explicitly changes that boundary
- review comes before deploy; Planwright must not become click-to-mutate cloud tooling
- compatibility is a matrix; no universal lossless-conversion claims
- accessibility and usability are correctness concerns; diagnostics must explain what happened, why it matters and how to fix it
- no cloud credentials are read unless a future read-only scan gate explicitly implements that with account and region confirmation
- no telemetry, remote scripts, remote fonts or hosted persistence by default

## Compatibility Levels

Planwright uses compatibility levels rather than broad slogans:
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

## Current Supported Matrix

This table describes the checked-in implementation. Future rows below are targets, not current claims.
Format or surface | Import | Analyse | Generate | Round-trip | Deploy tested
:--- | :--- | :--- | :--- | :--- | :---
Planwright YAML | Level 4 for one AWS web application pattern | Basic validation, security and cost notes | Level 5 review artefacts for Terraform/OpenTofu, Mermaid and Markdown reports | No | No
Planwright graph JSON | Level 1 structural schema validation | JSON Schema plus semantic graph validation | Schema export and graph diff input | No | No
Planwright graph diff | Local graph JSON comparison | Added, removed and changed graph elements plus selected risk-increasing findings | Markdown diff review | No | No
Planwright policy profiles | Local graph JSON only | Built-in static profile checks | Markdown and SARIF | No | No
Terraform plan JSON | Level 2 review-only input | Destructive change, replacement, public RDS and unknown-value findings | Markdown and SARIF review | No | No
CloudFormation | Level 4 for a small supported subset | Inventory and relationship notes | No generator yet | No | No
SAM | Level 4 for a small supported subset | Inventory and relationship notes | No generator yet | No | No
Kubernetes YAML | Level 4 for a rendered-manifest subset | Workload, Service, Ingress, Gateway API and Cilium inventory notes | No generator yet | No | No
AWS scan bundle | Level 4 from local JSON bundle files | Selected inventory and relationship inference | No live scan yet | No | No

## Graph Model Targets

`planwright.graph.v1` remains the centre of the product. It must grow deliberately enough to model:
- resources, relationships and deployment dependencies
- network flows, routes, exposure and trust boundaries
- IAM permissions, assume-role paths, pass-role risks and secret access
- accounts, regions, VPCs, subnets, namespaces and ownership metadata
- data reads, writes, publishing, subscriptions and encryption edges
- cost surfaces, lifecycle behaviour, backup evidence and observability evidence
- source preservation, provenance and unsupported-source annotations
- validation diagnostics with severity, resource, source location, explanation, fix, confidence and auto-fix status where possible

Core node families should grow in this order:
- AWS scopes, VPCs, subnets, security groups and route surfaces
- compute resources: EC2, ECS, Lambda, Kubernetes workloads and Docker Compose services
- edge and routing resources: ALB, NLB, API Gateway, CloudFront, Kubernetes Service, Ingress, Gateway API, Caddy reverse proxies and target groups
- data resources: RDS, DynamoDB, S3, ElastiCache, EFS, SQS, SNS and Kinesis
- identity resources: IAM roles, policies, users, groups, KMS keys, Secrets Manager secrets, SSM parameters and Kubernetes RBAC objects
- observability resources: CloudWatch log groups, alarms, CloudTrail trails, X-Ray, Prometheus targets, Grafana dashboards and OpenTelemetry collectors
- security resources: WAF, GuardDuty, Security Hub, AWS Config, Kubernetes NetworkPolicy, Cilium policies and future policy-pack artefacts

Core edge families should grow in this order:
- `network.allow`, `network.route` and `network.deny`
- `iam.allow`, `iam.assume_role` and `iam.pass_role`
- `data.reads_from`, `data.writes_to`, `publishes_to` and `subscribes_to`
- `logs_to`, `emits_metric_to`, `backs_up_to` and `encrypts_with`
- `depends_on`, `runs_as`, `exposes`, `protects`, `managed_by` and `generated_from`

## v0.0.1 Engine Skeleton

Goal: prove the engine shape.

Scope:
- Go CLI entrypoint
- `planwright.v1` typed plan parser
- `planwright.graph.v1` graph model
- basic validation
- basic diagnostics
- one AWS web application example

Explicitly excluded:
- Terraform generation
- imports
- deployment packs
- SARIF
- local web UI
- TUI
- live cloud scans

Status: implemented and audit-gated.

## v0.1 First Useful AWS Pack

Goal: generate reviewable artefacts for one AWS web application pattern.

Scope:
- Terraform/OpenTofu-oriented review files
- Mermaid diagram output
- security report
- cost notes
- deployability report
- cleanup guide
- assumptions report
- directory-based pack output

Resource subset:
- VPC-oriented generated shape
- public and private subnet intent
- security groups
- ALB, listener and target group intent
- ECS service intent
- RDS Postgres intent
- CloudWatch log group intent
- limited IAM role intent

Checks:
- no public RDS
- no open SSH/RDP
- security group and network-edge consistency where represented
- region required
- tags and ownership notes where represented
- NAT Gateway, ALB, RDS and fixed-cost warnings
- hardcoded secret-looking value warnings
- cleanup and destroy guidance

Explicitly excluded:
- Terraform/OpenTofu execution
- deployability proof
- zip archive output
- imports
- SARIF
- live cloud scans

Status: implemented and audit-gated.

## v0.2 Import Begins

Goal: import a CloudFormation/SAM subset with loss reports.

Scope:
- CloudFormation YAML/JSON subset import
- SAM subset import
- intrinsic-function tracking for supported relationships
- resource inventory extraction
- graph JSON output
- Markdown loss report output
- examples for CloudFormation and SAM import

Resource subset:
- `AWS::EC2::VPC`
- `AWS::EC2::Subnet`
- `AWS::EC2::SecurityGroup`
- `AWS::ElasticLoadBalancingV2::LoadBalancer`
- `AWS::RDS::DBInstance`
- `AWS::S3::Bucket`
- `AWS::IAM::Role`
- `AWS::Serverless::Function`
- `AWS::Serverless::HttpApi`
- `AWS::Serverless::SimpleTable`

Loss report categories:
- preserved
- normalised
- inferred
- ambiguous
- unsupported
- unsafe
- manual review required

Explicitly excluded:
- complete CloudFormation/SAM coverage
- CloudFormation/SAM generation
- Terraform/OpenTofu plan import
- Kubernetes import
- SARIF
- live cloud scans

Status: implemented and audit-gated.

## v0.3 Terraform Review Mode

Goal: review Terraform plan JSON and emit human and machine-readable findings.

Scope:
- Terraform plan JSON review from local files
- destructive-change findings
- replacement findings
- public RDS findings
- unknown security-sensitive value findings
- Markdown review output
- SARIF 2.1.0 output
- SARIF schema validation in CI or release verification

Explicitly excluded:
- Terraform state import
- Terraform HCL/module evaluation
- Terraform provider schema ingestion
- Terraform/OpenTofu execution
- graph lowering from Terraform plan JSON
- cloud account calls
- GitHub Action packaging

Status: implemented and audit-gated.

## v0.4 Local Web UI

Goal: provide a local-only browser workbench over the existing engine.

Scope:
- `planwright serve [project-dir] [--addr 127.0.0.1:5786]`
- loopback-only bind validation
- Host allowlist enforcement
- security headers for static and API responses
- static embedded workbench UI
- in-memory browser plan validation
- graph, diagnostics, report, Terraform and Mermaid previews
- keyboard-accessible editor and tables for the current surface

Current boundary:
- this is a text-and-table workbench, not the future drag-and-drop visual planner
- no canvas editor
- no resource-card graph editor

Accessibility target:
- keyboard-first flows
- no drag-only workflow
- focus-visible states
- no colour-only severity indication
- reduced motion compatibility
- text reports and copyable commands

Explicitly excluded:
- hosted demo
- drag-and-drop visual canvas editing
- resource-card graph planning
- browser deployment
- browser file writes
- credential loading
- cloud API calls
- TUI
- GitHub Action packaging

Status: implemented and audit-gated.

## v0.5 Kubernetes/Gateway/Cilium Import

Goal: import a rendered Kubernetes manifest subset with Gateway API and Cilium policy inventory.

Scope:
- `planwright import k8s <manifest-path-or-dir> --out <graph.json> --loss-report <loss.md>`
- rendered YAML/JSON manifest parsing from explicit local files or direct manifest directories
- Kubernetes workload, Service, Ingress, NetworkPolicy, ConfigMap and Secret inventory
- Secret metadata and key-name recording without Secret value lowering
- Gateway API Gateway and route inventory
- Service selector, Ingress backend and Gateway API backend relationship inference
- CiliumNetworkPolicy and CiliumClusterwideNetworkPolicy inventory with semantic loss notes
- Markdown loss reports for unsupported, ambiguous and deliberately redacted constructs

Explicitly excluded:
- live Kubernetes cluster scans
- `kubectl` execution
- Helm template evaluation
- Kustomize execution
- Kubernetes manifest generation
- full Kubernetes NetworkPolicy semantics
- full Cilium policy semantics
- Secret value decoding or display
- deployment or mutation

Status: implemented and audit-gated.

## v0.6 AWS Scan Bundle Import

Goal: import a local AWS CLI JSON scan bundle without adding live credentialed AWS calls.

Scope:
- `planwright import awsscan <bundle-dir> --out <graph.json> --loss-report <loss.md>`
- local directory reading for selected AWS CLI JSON outputs
- VPC, subnet, security group, EC2, RDS, S3, Lambda and ELBv2 inventory extraction
- public security-group ingress inference for tcp/udp rules with one concrete port
- selected dependency inference from VPC, subnet and security-group IDs
- conservative STS identity handling without lowering ARN or UserId values
- preserved-only reporting for unknown JSON files in the bundle

Explicitly excluded:
- live AWS account scans
- AWS SDK integration
- AWS CLI execution
- credential loading
- IAM policy analysis
- drift proof
- deployability proof
- infrastructure mutation

Status: implemented and audit-gated.

## v0.7 Graph Diff Review

Goal: compare two local Planwright graph JSON files and make risk-increasing architecture changes visible.

Scope:
- `planwright diff <old.graph.json> <new.graph.json> --out <review.md>`
- safe local graph JSON file loading
- graph validation before diff output
- deterministic added, removed and changed node reporting
- deterministic added and removed edge reporting
- Markdown graph diff review output
- findings for new public database exposure
- findings for new internet-facing network paths including SSH and RDP

Explicitly excluded:
- live cloud drift proof
- AWS SDK integration
- AWS CLI execution
- Kubernetes API calls
- Terraform state import
- Terraform plan comparison to graph
- automatic safety decisions for deployment
- infrastructure mutation

Status: implemented and audit-gated.

## v0.8 Graph Schema Hardening

Goal: make `planwright.graph.v1` structurally reviewable outside Planwright while preserving semantic graph validation.

Scope:
- embedded JSON Schema 2020-12 for `planwright.graph.v1`
- `planwright schema graph --out <schema.json>`
- `planwright validate-graph <planwright.graph.json>`
- structural validation before semantic graph validation
- schema checks for required graph, node and edge fields
- schema checks for network protocol values and port ranges
- documentation of the schema boundary and non-goals

Explicitly excluded:
- v1.0 schema compatibility policy
- automatic migration between graph schema versions
- graph schema generation from Go types
- live drift proof
- deployment validation
- broader graph taxonomy completion

Status: implemented and audit-gated.

## v0.9 Built-In Policy Profile Review

Goal: review local Planwright graph JSON against small built-in policy profiles.

Scope:
- `planwright policy profiles`
- `planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>`
- built-in `lab`, `small-business` and `production` profiles
- local graph JSON validation before policy evaluation
- policy findings for public databases and internet-facing SSH/RDP
- database backup evidence checks for stricter profiles
- production observability and lab-profile metadata notes
- Markdown policy profile review output
- SARIF 2.1.0 policy finding output

Explicitly excluded:
- custom policy packs
- OPA/Rego integration
- Conftest-compatible bundles
- organisation-specific policy inheritance
- compliance certification
- live cloud or Kubernetes checks
- automatic deployment safety decisions

Status: implemented and audit-gated.

## v0.10 CI and Documentation Style Hardening

Goal: make repository checks broad enough for release, security and documentation hygiene while enforcing Planwright's documentation style.

Scope:
- reusable GitHub Actions workflows
- pinned checkout and Go setup actions
- Linux runner hardening on CI jobs
- build, cross-platform build, test, race test, lint, module hygiene, static analysis and vulnerability jobs
- action pin drift checks
- file-header checks
- Go Report Card checks for public repositories
- script-quality checks for GitHub helper scripts
- CodeQL and Scorecards workflows
- workflow validation with actionlint
- release verification, prepare-release and release workflow metadata
- Apache-2.0 licence metadata, AUTHORS and CONTRIBUTORS
- local documentation style checker
- British-English and spelling checks with cspell
- documentation style rules for bullet-list spacing, code-block spacing, comma-before-`but` and simple comma-before-`and`

Explicitly excluded:
- proof that a release workflow has published artefacts
- package publishing
- hosted deployment
- proof that CI covers every possible security or correctness issue

Status: implemented and audit-gated.

## v0.11 Rename, Roadmap, UI and Hardening Pass

Goal: make the project consistently Planwright, reduce root clutter, sharpen the roadmap and harden the current local engine.

Scope:
- make `repo/planwright` the canonical repository path
- rename CLI command, module path, schema names, examples, docs, diagnostics and generated artefact names
- remove AI planning artefacts while retaining `AGENTS.md`
- minimise root files where GitHub and project conventions still work
- add a changelog
- fill the roadmap with all planned gates from the product plan
- make the local web UI square, dark and pure-black by default
- harden `.gitignore` and project metadata
- centralise runtime version metadata
- run a full codebase pass for security, correctness, readability and maturity before continuing

Explicitly excluded:
- changing the Apache-2.0 licence
- hiding repository maturity gaps
- adding live cloud scans
- adding deployment execution
- publishing a release

Status: implemented and audit-gated.

## v0.12 Usability and Proof Release

Goal: make Planwright understandable, runnable and reviewable in under five minutes.

Strategic rule:
- do not expand the feature surface unless the work directly improves the proof path
- do not add another importer, generator or interface to look broader
- make the existing engine easier to inspect, verify and explain
- keep compatibility claims brutally tied to checked fixtures and loss evidence

Scope:
- better README demo flow
- one polished `examples/aws-webapp-basic` walkthrough
- generated deployment pack walkthrough
- real CLI transcript snippets from checked commands
- Mermaid diagram preview or maintained screenshot only if the verification path is clear
- one companion bad-example path such as public database exposure if it improves the proof story
- clearer “what Planwright is not” section
- generated evidence examples for validation, security notes, cost notes, Mermaid and pack layout
- golden tests or fixture checks for the reports and generated artefacts used in public docs
- documentation that explains why the local web workbench is not the future drag-and-drop canvas

Exit criteria:
- a new user can install or build Planwright, run the canonical example and inspect generated evidence without guessing the command order
- the canonical example has documented input, commands, expected outputs and limits
- generated examples used in documentation are either checked by tests or clearly marked as illustrative
- the README explains the current proof path before listing the full command surface
- the docs state non-goals plainly enough that Planwright is not mistaken for a one-click deployer, universal converter, compliance tool or live cloud scanner
- release notes describe usability and evidence improvements rather than pretending this is a broad compatibility expansion

Explicitly excluded:
- new import families
- new cloud providers
- live AWS calls
- credential loading
- drag-and-drop GUI expansion
- TUI work
- custom policy packs
- OPA/Rego integration
- plugin SDK work
- cost-estimation precision claims
- compliance language
- deployment execution

## v0.13 Golden Compatibility Fixture Suite

Goal: make compatibility claims testable before adding more broad surface area.

Scope:
- fixture runner for importers, generators and reports
- fixture metadata for supported source format, expected capability level and expected loss categories
- golden tests for Terraform/OpenTofu, CloudFormation, SAM, Mermaid, Markdown and SARIF outputs where those paths already exist
- malformed input fixtures
- unsupported construct fixtures
- lossy conversion fixtures
- round-trip tests only where the compatibility level claims round-trip support
- public compatibility matrix generated or checked from fixture metadata

Exit criteria:
- every current compatibility matrix row is backed by at least one fixture or explicitly marked as documentation-only
- unsupported and ambiguous fixture cases produce visible loss evidence
- no new compatibility level can be raised without a fixture update

Explicitly excluded:
- broad service coverage without fixtures
- deployability testing
- invisible compatibility claims in docs

## v0.14 Terraform/OpenTofu State and Provider Schema Import

Goal: make Terraform/OpenTofu compatibility more than plan-review evidence.

Scope:
- `terraform show -json` state import
- OpenTofu-compatible state import where the JSON format matches supported fields
- resource inventory extraction from state JSON
- prior-state and planned-value comparison where both are present
- sensitive value tracking and redaction
- provider schema ingestion from `terraform providers schema -json`
- compatibility reports tied to provider/resource schemas

Exit criteria:
- state JSON fixtures cover empty, malformed, sensitive and unsupported resources
- provider schema fixtures prove that sensitivity metadata is preserved in reports
- Terraform and OpenTofu paths share code where the supported JSON shape is equivalent

Explicitly excluded:
- full HCL module interpreter
- Terraform expression evaluation
- provider plugin execution
- Terraform/OpenTofu apply
- automatic import-block generation unless separately designed

## v0.15 Terraform/OpenTofu Graph Lowering

Goal: lower supported Terraform/OpenTofu plan and state resources into `planwright.graph.v1`.

Scope:
- supported AWS resource mapping to graph nodes
- selected relationship inference from IDs, references and security group rules
- destructive-change graph annotations
- drift-ish comparison between prior and planned state
- Markdown and SARIF review improvements
- loss and confidence reporting for unknown, computed and sensitive values

Exit criteria:
- supported AWS resources lower into graph fixtures with deterministic node and edge IDs
- sensitive, unknown and computed values are never silently treated as known safe values
- graph lowering has loss reports for unsupported provider resources and attributes

Explicitly excluded:
- lossless Terraform-to-Planwright conversion
- direct Terraform HCL module execution
- provider-specific coverage beyond declared support

## v0.16 Deployment Pack v1

Goal: make the pack a business-credible evidence artefact.

Scope:
- `planwright.pack.v1` manifest stability
- optional zip output
- source preservation folder
- generated Terraform/OpenTofu, CloudFormation/SAM, Kubernetes and diagram folders as available
- reports folder with README, compatibility, loss, security, cost, deployability, cleanup, assumptions, threat model and runbook notes
- diagrams folder for Mermaid, future D2 and future Graphviz
- scripts folder for identity checks, plan commands, verification and destroy guidance
- manifest file hashes and generated hash list

Exit criteria:
- pack manifest has a tested `planwright.pack.v1` shape
- pack output is deterministic for the same input
- source preservation never stores credentials or secret values that Planwright has redacted
- generated scripts are review aids only and are never run by the pack command

Explicitly excluded:
- storing secrets
- executing generated scripts
- claiming deployability without sandbox evidence

## v0.17 Examples Gallery and Documentation Site

Goal: make examples teachable and reviewable using real engine output.

Scope:
- `examples/aws-webapp-basic`
- `examples/aws-webapp-bad-public-db`
- `examples/aws-serverless-sam-import`
- `examples/terraform-plan-risk-review`
- `examples/cloudformation-to-terraform-loss-report`
- `examples/kubernetes-gateway-basic`
- `examples/docker-compose-import`
- `examples/small-business-web-stack`
- `examples/student-lab-low-cost`
- expected findings for each example
- generated outputs and reports for examples where checked-in output is justified
- static documentation site content sourced from checked-in docs
- screenshots only where maintained by an explicit verification process

Each example must document:
- what it creates or represents
- what it does not create
- cost notes
- public and private surfaces
- permissions required
- verification steps
- destroy guidance
- what Planwright could not infer

Exit criteria:
- every example has commands, expected findings and scope boundaries
- documentation site content does not overclaim beyond the checked-in implementation
- examples are exercised by tests where practical

Explicitly excluded:
- hosted demo deployment
- screenshots without a maintained verification process
- generated output snapshots that cannot be kept deterministic

## v0.18 CI Review Action and SARIF Hardening

Goal: make pull request review practical while keeping machine-readable findings stable.

Scope:
- GitHub Action wrapper
- Planwright graph, Terraform plan and policy profile review in CI
- SARIF upload path
- Markdown PR comment summary
- stable exit codes
- documented GitHub token permissions
- SARIF helper package if duplication starts affecting rule stability
- schema validation tests for emitted SARIF
- stable rule IDs
- source location support where importers can provide it
- baseline and suppression strategy discussion

Exit criteria:
- Action examples follow least-privilege GitHub token permissions
- SARIF output validates against the supported schema in tests
- rule IDs and exit codes are documented as compatibility-sensitive

Explicitly excluded:
- cloud credentials by default
- cloud mutation
- secret display
- unstable rule IDs after release without documented compatibility impact

## v0.19 Security and Accessibility Hardening

Goal: document and test every current trust boundary before declaring a stable core.

Scope:
- threat model refresh
- Host allowlist tests
- CORS refusal tests
- CSRF design for future state-changing browser actions
- request body size tests
- path traversal and symlink tests
- archive extraction safety before zip pack input is added
- secret redaction tests
- generated artefact secret scanning checks
- keyboard-first local web flows
- screen-reader graph summary for the current workbench
- high-contrast and focus-visible checks
- severity icons and text rather than colour alone
- reduced-motion compatibility
- ARIA live validation region where it improves the current workbench
- Playwright accessibility smoke checks if a browser test harness is added

Exit criteria:
- threat model matches the implemented CLI, importer, server and release surfaces
- local web security tests cover loopback, Host, CORS and request-size boundaries
- accessibility docs distinguish target practices from formal audit claims

Explicitly excluded:
- weakening local-only defaults for convenience
- claiming formal accessibility certification without an audit

## v1.0 Stable Core

Goal: make Planwright business-proposable as a local-first compatibility and evidence engine.

Requirements:
- stable `planwright.graph.v1` schema and compatibility policy
- stable `planwright.pack.v1` pack format
- declared compatibility matrix generated or checked from tested fixtures
- signed releases through a documented human-controlled trust root
- checksums
- SBOMs beside release artefacts when the release tooling supports them
- security policy
- threat model
- accessibility statement
- CI Action
- documentation site
- examples gallery
- Windows, macOS and Linux binary verification
- no telemetry by default
- offline mode
- stable exit codes
- stable diagnostics policy
- full release-candidate security and correctness pass

Explicitly excluded at v1.0 unless separately approved:
- hosted SaaS
- one-click deployment
- AI architect claims
- lossless universal IaC conversion
- broad multi-cloud coverage
- live mutation from browser
- live cloud scans as a stability requirement
- broad generator coverage beyond fixture-backed support

## Post-v1.0 Expansion Gates

The following gates remain part of the full project vision but should not block the first stable core. They expand the product after the core graph, fixture suite, pack format, release evidence and CI review flow are stable.

## v1.1 CloudFormation and SAM Generator Expansion

Goal: generate reviewable CloudFormation and SAM templates from supported graph shapes.

Scope:
- CloudFormation `template.yaml` output
- Parameters, Resources and Outputs for supported AWS web application shapes
- Metadata with Planwright provenance
- SAM output for serverless graph shapes
- CloudFormation/SAM golden tests
- compatibility report for generated constructs

Explicitly excluded:
- source-level CDK generation
- perfect parity with imported CloudFormation
- complete SAM transform expansion

## v1.2 Docker Compose Import and Generate

Goal: support local container topologies as a bridge to cloud and Kubernetes planning.

Scope:
- Docker Compose service, network, port and volume inventory
- `depends_on`, healthcheck, image and environment-key extraction
- secret-looking environment key warnings
- host exposure findings
- limited Docker Compose generation from simple graph shapes
- loss reports for unsupported Compose features

Explicitly excluded:
- Dockerfile execution or build
- container runtime inspection
- Swarm deployment support

## v1.3 Kubernetes, Gateway API and Cilium Generation

Goal: make cloud-native output useful for the supported graph subset.

Scope:
- Namespace, Deployment, StatefulSet, Service and ConfigMap generation
- Secret placeholders without values
- Gateway, HTTPRoute, TLSRoute and TCPRoute generation
- Kubernetes NetworkPolicy generation
- optional CiliumNetworkPolicy generation
- cluster exposure graph reports
- Kubernetes deployment pack folder

Explicitly excluded:
- live cluster apply
- Helm chart generation
- full policy semantic proof

## v1.4 TUI Review Interface

Goal: support terminal-first infrastructure review.

Scope:
- `planwright tui`
- resource tree
- selected resource details
- flow table
- findings panel
- generated output preview
- diff view
- loss report view
- keyboard command palette

Explicitly excluded:
- terminal deployment
- private TUI-only engine logic

## v1.5 Hosted Static Demo

Goal: show the engine safely on `steadytao.com/projects/planwright`.

Scope:
- static project page
- thesis and safety boundaries
- engine diagram
- typed plan example
- CLI playback
- TUI mock or playback
- interactive GUI demo with sample data
- report previews
- sample deployment pack download
- compatibility matrix
- roadmap and source links

Safety boundary:
- no AWS credentials
- no deployment
- no server-side storage
- no secrets
- client-side sample data unless a user explicitly uploads local files in browser memory

## v1.6 VS Code Extension

Goal: improve authoring without moving engine logic into the editor.

Scope:
- schema validation
- autocomplete for Planwright YAML
- graph preview
- findings panel
- generated-output preview
- command invocation through the local CLI

Explicitly excluded:
- extension-only features
- cloud credential handling

## v1.7 Live Read-Only AWS Scan

Goal: help users understand existing AWS environments without mutation.

Scope:
- explicit `planwright scan aws` command
- identity confirmation with account ID and region display
- read-only service inventory for VPC, subnet, security group, EC2, ECS, RDS, S3, Lambda, API Gateway and ELBv2 where supported
- permission-aware missing-resource report
- relationship inference
- local graph output
- loss report
- no secret value collection

Safety boundary:
- read-only only
- explicit region
- explicit account confirmation
- no mutation
- no browser-triggered credential use

## v1.8 Drift and Diff Expansion

Goal: compare intended, generated and observed states.

Scope:
- graph vs graph diff expansion
- Terraform plan vs graph comparison
- Terraform state vs graph comparison
- CloudFormation vs graph comparison
- AWS scan vs graph comparison
- Kubernetes live state vs manifest comparison if live Kubernetes scan is later approved
- findings for new public paths, permission broadening, destructive changes, logging removal, backup removal and encryption removal

Explicitly excluded:
- absolute proof of drift without explicit source data
- automatic remediation

## v1.9 Policy Packs

Goal: support organisational static checks without forcing OPA immediately.

Scope:
- built-in `lab`, `student`, `small-business`, `public-web-app`, `internal-service`, `production` and `regulated-baseline` profiles
- custom YAML policy rules
- profile inheritance
- policy fixture tests
- machine-readable policy output

Explicitly excluded:
- compliance certification
- arbitrary code execution in policies

## v1.10 OPA/Rego and Conftest Integration

Goal: integrate with established policy tooling after the built-in model is stable.

Scope:
- OPA input JSON export
- optional Rego policy execution behind explicit command flags
- Conftest-compatible bundle output if it proves useful
- SARIF output for policy findings

Explicitly excluded:
- silently running third-party policy code
- network policy downloads by default

## v1.11 Plugin SDK

Goal: let compatibility grow through adapters without destabilising the engine.

Scope:
- importer interface
- generator interface
- analyser interface
- capability metadata
- fixture harness
- plugin compatibility tests
- documentation for loss reports, provenance and diagnostics

Explicitly excluded:
- unreviewed arbitrary plugin execution by default
- marketplace or hosted plugin registry

## v1.12 Operational and Resilience Analysis

Goal: expand reports from security-only review into operations evidence.

Scope:
- single-AZ warnings
- no-backup warnings
- no-healthcheck warnings
- no-log-retention warnings
- no-alarm warnings
- no-deletion-policy decision warnings
- rollback notes
- operator runbook generation
- incident-response notes
- data-flow and threat-model report generation

Explicitly excluded:
- claiming operational readiness without deployed evidence

## v1.13 Cost Analysis Expansion

Goal: turn cost notes into rough cost review while remaining honest.

Scope:
- NAT Gateway warning
- ALB idle cost warning
- RDS idle cost warning
- EBS orphan warning
- Elastic IP unused warning
- CloudWatch log growth warning
- cross-AZ data warning
- managed service fixed-cost warning
- student/lab low-cost mode
- optional Infracost integration if it earns the dependency and boundary

Explicitly excluded:
- exact billing prediction
- cloud pricing calls without explicit user action

## v1.14 Multi-Cloud Research Gate

Goal: decide whether non-AWS targets fit without collapsing scope.

Research topics:
- Azure ARM/Bicep import and generation
- Google Cloud deployment formats
- Crossplane compositions
- Ansible inventory-style explanation
- Packer templates
- Nomad jobs
- cloud-init analysis
- Serverless Framework import

Exit criteria:
- documented fit assessment
- no implementation until AWS, Terraform/OpenTofu, CloudFormation/SAM and Kubernetes foundations are stronger

## v1.15 Other IaC and Catalogue Research Gate

Goal: decide whether adjacent IaC and catalogue targets fit the compatibility model.

Research topics:
- Pulumi stack export import
- Pulumi preview JSON review if a stable local format is available
- Pulumi YAML generation
- CDK synth output import through CloudFormation
- CDK TypeScript or Python skeleton generation
- Backstage `catalog-info.yaml` generation
- OpenAPI-style infrastructure API description
- Serverless Framework import
- Podman systemd-style unit analysis

Exit criteria:
- documented fit assessment for each target
- clear loss-report strategy for code-first or template-driven formats
- no code generation until generated output can be fixture-tested and reviewed
