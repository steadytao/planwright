# AWS Web Application Proof Path

This example is the current canonical Planwright proof path.

It shows one typed plan becoming:
- a validated Planwright graph
- a security review note
- a cost note
- Terraform/OpenTofu-oriented review files
- a Mermaid architecture diagram
- a directory-based Planwright pack

It is intentionally small. The goal is to make the current engine easy to inspect rather than to claim broad AWS coverage.

## What This Models

The plan models a low-cost lab profile for:
- an external internet entrypoint
- an internet-facing Application Load Balancer
- an ECS service listening on port `8080`
- an RDS PostgreSQL database listening on port `5432`
- explicit network flows between those resources

The source file is [planwright.yaml](planwright.yaml).

## What This Does Not Do

This example does not:
- contact AWS
- read AWS credentials
- run Terraform or OpenTofu
- create infrastructure
- prove deployability
- estimate a bill
- certify compliance
- model every resource a production application needs

The generated Terraform/OpenTofu-oriented files are review artefacts. They still need human inspection and normal infrastructure planning before use.

## Try It

From the repository root:
```bash
go run ./cmd/planwright validate examples/aws-webapp-basic/planwright.yaml
```

Expected output:
```text
validation passed
```

Render the security report:
```bash
go run ./cmd/planwright risks examples/aws-webapp-basic/planwright.yaml
```

Expected output starts with:
```text
# Security Report

Planwright reports review hints from the typed graph. Planwright does not prove deployability or compliance.
```

Render the cost notes:
```bash
go run ./cmd/planwright cost-notes examples/aws-webapp-basic/planwright.yaml
```

Expected output starts with:
```text
# Cost Notes

These are review notes, not a bill estimate.
```

Generate Terraform/OpenTofu-oriented review files:
```bash
go run ./cmd/planwright generate terraform examples/aws-webapp-basic/planwright.yaml --out ./generated/terraform
```

Expected output:
```text
wrote Terraform/OpenTofu files to ./generated/terraform
```

Generate a Mermaid diagram:
```bash
go run ./cmd/planwright generate mermaid examples/aws-webapp-basic/planwright.yaml --out ./generated/diagrams
```

Expected output:
```text
wrote Mermaid files to ./generated/diagrams
```

Create a Planwright pack:
```bash
go run ./cmd/planwright pack examples/aws-webapp-basic/planwright.yaml --out ./planwright-pack
```

Expected output:
```text
wrote Planwright pack to ./planwright-pack
```

The output paths above are ignored by Git so they are safe for local inspection. Remove them when finished:
```bash
rm -rf ./generated ./planwright-pack
```

## Pack Layout

The pack is a directory with reviewable evidence:
```text
planwright-pack/
  manifest.json
  planwright.yaml
  planwright.graph.json
  generated/
    terraform/
      README.md
      versions.tf
      providers.tf
      variables.tf
      network.tf
      security-groups.tf
      iam.tf
      observability.tf
      app.tf
      database.tf
      outputs.tf
  diagrams/
    architecture.mmd
  reports/
    security-report.md
    cost-notes.md
    deployability-report.md
    cleanup.md
    assumptions.md
```

Important files to inspect first:
- `manifest.json`, pack metadata, checksums and `requires_review`
- `planwright.graph.json`, the lowered graph Planwright understood
- `reports/security-report.md`, security review notes
- `reports/cost-notes.md`, cost review notes
- `generated/terraform/README.md`, assumptions for the generated Terraform/OpenTofu-oriented files
- `diagrams/architecture.mmd`, the Mermaid architecture diagram

## Review Notes

Current safe-by-default behaviour in this example:
- the database is not marked publicly accessible
- no internet-facing SSH or RDP flow is present
- the lab profile avoids generating a NAT Gateway
- the Terraform/OpenTofu-oriented output does not hardcode database passwords

Current intentional limitations:
- Planwright does not validate the generated files with Terraform or OpenTofu
- Planwright does not know whether the AWS account has required quotas or permissions
- Planwright does not model ACM certificate ownership or WAF configuration
- Planwright does not prove the ECS task image, registry access or runtime health
- Planwright does not model every IAM permission a production service may need

## Companion Risk Check

To see the public database finding, change this property in a temporary copy:
```yaml
db_public: true
```

The risk report then includes:
```text
HIGH PW-AWS-RDS-001: A database node is marked publicly accessible.
```

The Terraform/OpenTofu generator refuses that graph because the current generator does not support publicly accessible database nodes.

## Why This Example Matters

This example is the proof path for the v0.12 usability work. It keeps Planwright boring, reviewable and evidence-first:
- the typed source is small enough to read
- the graph is explicit enough to review
- the reports are plain Markdown
- the generated files stay local
- compatibility claims are tied to the implemented AWS web application pattern
