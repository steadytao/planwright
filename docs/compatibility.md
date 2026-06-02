# Compatibility

Planwright describes compatibility as a level, not a slogan.
- [Compatibility Levels](#compatibility-levels)
- [Current Matrix](#current-matrix)
- [Import Boundaries](#import-boundaries)
- [Compatibility Rule](#compatibility-rule)

# Compatibility Levels

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

## Current Matrix

Format | Import | Analyse | Generate | Round-trip | Deploy tested
:--- | :--- | :--- | :--- | :--- | :---
Planwright YAML | Level 4 for one AWS web application pattern | Basic validation, security and cost notes | Level 5 review artefacts for Terraform/OpenTofu, Mermaid and Markdown reports | Not yet | No
Terraform/OpenTofu plan JSON | Level 2 for resource change inventory | v0.3 review findings for destructive changes, replacement, public RDS and unknown security values | SARIF and Markdown review output | No | No
Terraform/OpenTofu state JSON | Not yet | Not yet | Not yet | No | No
CloudFormation | Level 4 for a small AWS resource subset with loss reports | Basic graph validation after import | Not yet | No | No
SAM | Level 4 for `Function`, `HttpApi` and `SimpleTable` with loss reports | Basic graph validation after import | Not yet | No | No
Kubernetes YAML | Level 4 for a rendered-manifest subset | Basic graph validation plus route relationship inference for Services, Ingress and Gateway API routes | Not yet | No | No
Gateway API manifests | Level 4 for Gateway and HTTPRoute/TCPRoute/TLSRoute inventory | Route parent/backend relationship inference for the supported subset | Not yet | No | No
Cilium policies | Level 4 inventory with semantic loss notes | Policy presence and scope notes only | Not yet | No | No
AWS scan bundle JSON | Level 4 for a local AWS CLI JSON subset | Basic graph validation plus security-group public ingress and reference inference | Not yet | No | No
Planwright graph JSON | Level 1 structural schema validation | JSON Schema 2020-12 plus semantic graph validation | Schema export and graph diff input | No | No
Planwright graph JSON diff | Local graph JSON comparison only | Added, removed and changed nodes or edges plus selected risk-increasing changes | Markdown diff review | No | No
Planwright policy profiles | Local graph JSON input only | Built-in `lab`, `small-business` and `production` static checks | Markdown and SARIF review output | No | No

# Import Boundaries

## v0.2 CloudFormation/SAM Subset

Planwright v0.2 lowers only these resource types:
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

Intrinsic functions, parameters, mappings, conditions and outputs may be preserved or reported as ambiguous; they are not fully evaluated in v0.2.

## v0.3 Terraform Plan JSON Review

Planwright v0.3 reads local JSON produced by `terraform show -json <PLAN FILE>`.

It reviews:
- resource changes whose action list is `["delete"]`
- replacement changes whose action list contains both `delete` and `create`
- `aws_db_instance` changes whose planned `publicly_accessible` value is `true`
- selected security-sensitive unknown values on AWS database and security-group resources
- plan-level `errored` and incomplete flags

It does not evaluate HCL, inspect provider schemas, import Terraform state, run Terraform, apply plans or prove deployability.

## v0.4 Local Web UI

Planwright v0.4 adds a local browser workbench for the existing typed-plan path.

The workbench:
- validates browser-posted plan text in memory
- previews graph data, diagnostics, Markdown reports, Terraform/OpenTofu-oriented generated files and Mermaid output
- uses a text editor and tables for the current surface
- runs only when explicitly started with `planwright serve`
- defaults to loopback binding

It does not add a new compatibility level by itself, because it is an interface over the existing typed-plan and generator paths.

It is not the future drag-and-drop visual planner, canvas editor or resource-card graph editor.

## v0.5 Kubernetes Manifest Import

Planwright v0.5 reads rendered local Kubernetes YAML or JSON manifests from an explicit file path or a direct manifest directory.

It lowers:
- `Namespace`
- `Deployment`
- `StatefulSet`
- `DaemonSet`
- `Job`
- `CronJob`
- `Service`
- `Ingress`
- `NetworkPolicy`
- `ConfigMap`
- `Secret` metadata and key names only
- Gateway API `Gateway`, `HTTPRoute`, `TCPRoute` and `TLSRoute`
- Cilium `CiliumNetworkPolicy` and `CiliumClusterwideNetworkPolicy`

It infers:
- Service selector routes to imported workloads with matching template labels
- Ingress backend routes to imported Services
- Gateway API route backend references to imported Services
- Gateway API route parent references to imported Gateways

It does not run `kubectl`, contact a cluster, evaluate Helm templates, run Kustomize, decode Secret values or fully model Kubernetes NetworkPolicy or Cilium policy semantics. Those unsupported or partial semantics are reported in the loss report rather than silently converted.

## v0.6 AWS Scan Bundle Import

Planwright v0.6 reads a local directory of selected AWS CLI JSON artefacts.

It recognises:
- `manifest.json` for optional Planwright bundle metadata
- `sts-get-caller-identity.json`
- `ec2-describe-vpcs.json`
- `ec2-describe-subnets.json`
- `ec2-describe-security-groups.json`
- `ec2-describe-instances.json`
- `rds-describe-db-instances.json`
- `s3-list-buckets.json`
- `lambda-list-functions.json`
- `elbv2-describe-load-balancers.json`

It lowers:
- VPCs
- subnets
- security groups
- EC2 instances
- RDS instances
- S3 buckets
- Lambda functions
- Application and Network Load Balancers
- public security-group ingress edges when the rule has a concrete tcp/udp port

It infers selected `depends_on` relationships from IDs present in the bundle.

It does not call AWS, run the AWS CLI, use the AWS SDK, load credentials, verify live account identity, inspect IAM policy documents or prove drift. STS ARN and UserId values are deliberately not lowered into graph properties or loss messages.

## v0.7 Graph Diff Review

Planwright v0.7 compares two local `planwright.graph.v1` JSON files.

It reports:
- added graph nodes
- removed graph nodes
- graph nodes whose kind or properties changed
- added graph edges
- removed graph edges
- new public database exposure findings
- new internet-facing network path findings

It validates both graph files before writing a diff report. Invalid graph input is reported through graph diagnostics and no diff report is written.

It does not contact cloud APIs, inspect Terraform state, compare against Kubernetes clusters, prove live drift, execute generated infrastructure or decide whether a change is safe to deploy.

## v0.8 Graph Schema Hardening

Planwright v0.8 embeds and exports a JSON Schema 2020-12 document for `planwright.graph.v1`.

It supports:
- `planwright schema graph --out <schema.json>`
- `planwright validate-graph <planwright.graph.json>`
- structural validation of graph JSON
- semantic validation through the existing graph validator after schema validation
- schema checks for required top-level graph fields, node fields, edge fields, network protocol values and valid port ranges

The JSON Schema is not a deployment format and not a v1.0 compatibility guarantee. It is a structural contract for the current `planwright.graph.v1` graph representation. Planwright's semantic validator remains responsible for checks that JSON Schema cannot express cleanly in this version including duplicate node IDs and edge endpoint existence.

## v0.9 Built-In Policy Profiles

Planwright v0.9 reviews local `planwright.graph.v1` JSON against built-in policy profiles.

It supports:
- `planwright policy profiles`
- `planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>`
- `lab`, `small-business` and `production` profiles
- local graph JSON schema and semantic validation before policy evaluation
- Markdown policy profile reviews
- SARIF 2.1.0 output for policy findings

The v0.9 policy checks cover selected graph risks only: public databases, internet-facing SSH/RDP, missing database backup evidence in stricter profiles, production observability evidence and applying production checks to a lab-profile graph.

They do not run custom policy files, execute OPA/Rego, contact cloud APIs, inspect live infrastructure, prove deployability or certify compliance.

## Compatibility Rule

If Planwright cannot preserve or understand a construct, it must report that honestly rather than silently dropping it.
