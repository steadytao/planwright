// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"fmt"
	"strings"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
)

func RenderSecurity(g graph.Graph) string {
	var builder strings.Builder
	builder.WriteString("# Security Report\n\n")
	builder.WriteString("Planwright reports review hints from the typed graph. Planwright does not prove deployability or compliance.\n\n")

	publicDB := false
	openAdmin := false
	for _, node := range g.Nodes {
		if graph.IsDatabaseNode(node) && graph.HasBoolProperty(node.Properties, "publicly_accessible") {
			publicDB = true
		}
	}
	for _, edge := range g.Edges {
		if graph.IsInternetFacingNetworkAllow(g, edge) && graph.IsAdministrativePort(edge.Port) {
			openAdmin = true
		}
	}

	builder.WriteString("## Findings\n\n")
	if publicDB {
		builder.WriteString("- HIGH PW-AWS-RDS-001: A database node is marked publicly accessible. Set `db_public` to false unless this is deliberate and separately controlled.\n")
	} else {
		builder.WriteString("- PASS: No public database exposure was detected in the Planwright graph.\n")
	}
	if openAdmin {
		builder.WriteString("- HIGH PW-AWS-NET-002: SSH or RDP is reachable from the internet in the Planwright graph.\n")
	} else {
		builder.WriteString("- PASS: No internet-facing SSH or RDP edge was detected.\n")
	}
	builder.WriteString("- INFO: Public HTTPS entrypoints still require certificate, WAF, logging and operational review outside Planwright.\n")
	return builder.String()
}

func RenderCostNotes(g graph.Graph) string {
	var builder strings.Builder
	builder.WriteString("# Cost Notes\n\n")
	builder.WriteString("These are review notes, not a bill estimate.\n\n")
	if hasKind(g, "aws.alb") {
		builder.WriteString("- Application Load Balancer creates fixed hourly and usage cost even when idle.\n")
	}
	if hasKind(g, "aws.rds.postgres") {
		builder.WriteString("- RDS creates instance and storage cost while running. Review backups, retention and idle lab teardown.\n")
	}
	builder.WriteString("- NAT Gateway is not generated in the lab profile, avoiding a common fixed-cost trap. Private workloads may need a deliberate egress design later.\n")
	builder.WriteString("- CloudWatch logs, data transfer and database storage growth are not estimated by the current cost notes.\n")
	return builder.String()
}

func RenderDeployability(g graph.Graph) string {
	return fmt.Sprintf("# Deployability Report\n\n"+
		"Planwright writes reviewable Terraform/OpenTofu-oriented files but does not run Terraform, OpenTofu, provider installation, validation, plan or apply.\n\n"+
		"Before use:\n\n"+
		"- review the generated Terraform before running it\n"+
		"- supply `db_password` outside generated files\n"+
		"- attach or model a valid ACM certificate before using the HTTPS listener\n"+
		"- confirm region %s and account manually\n"+
		"- decide whether lab teardown settings are acceptable\n", markdownCode(g.Region))
}

func RenderCleanup(g graph.Graph) string {
	_ = g
	return "# Cleanup Guide\n\n" +
		"Planwright does not run destroy commands.\n\n" +
		"Before deleting resources:\n\n" +
		"- inspect the generated Terraform state and plan output\n" +
		"- confirm the AWS account and region manually\n" +
		"- check whether the database contains data that must be retained\n" +
		"- review final snapshot and deletion protection choices\n\n" +
		"The lab output sets `skip_final_snapshot = true` and `deletion_protection = false` for review convenience. Change those settings before production-like use.\n"
}

func RenderAssumptions(g graph.Graph) string {
	return fmt.Sprintf("# Assumptions\n\n"+
		"- The graph provider is %s.\n"+
		"- The graph uses region %s.\n"+
		"- The graph profile is %s.\n"+
		"- The AWS web application pattern maps to one VPC, two public subnets, two private subnets, one ALB, one ECS service and one PostgreSQL RDS instance.\n"+
		"- Planwright does not infer availability-zone redundancy, IAM least privilege, TLS certificate ownership or runtime health checks.\n", markdownCode(g.Provider), markdownCode(g.Region), markdownCode(g.Profile))
}

func RenderLossReport(loss loss.Report) string {
	var builder strings.Builder
	builder.WriteString("# Loss Report\n\n")
	builder.WriteString("Planwright import reports are compatibility evidence. They do not claim lossless conversion or deployability.\n\n")
	if loss.SourceFormat != "" || loss.Source != "" {
		builder.WriteString("## Source\n\n")
		if loss.SourceFormat != "" {
			fmt.Fprintf(&builder, "- Format: %s\n", markdownCode(loss.SourceFormat))
		}
		if loss.Source != "" {
			fmt.Fprintf(&builder, "- Source: %s\n", markdownCode(loss.Source))
		}
		builder.WriteString("\n")
	}
	renderLossSection(&builder, "Lowered", loss.Lowered, "No resources were lowered into the Planwright graph.")
	renderLossSection(&builder, "Unsupported", loss.Unsupported, "No unsupported resources were reported.")
	renderLossSection(&builder, "Ambiguous", loss.Ambiguous, "No ambiguous constructs were reported.")
	renderLossSection(&builder, "Preserved", loss.Preserved, "No extra template sections were reported as preserved-only.")
	builder.WriteString("## Review\n\n")
	builder.WriteString("Manual review required before using generated graph output for migration, deployment planning or security decisions.\n")
	return builder.String()
}

func renderLossSection(builder *strings.Builder, title string, items []loss.Item, empty string) {
	fmt.Fprintf(builder, "## %s\n\n", title)
	if len(items) == 0 {
		builder.WriteString("- " + empty + "\n\n")
		return
	}
	for _, item := range items {
		resource := item.Resource
		if resource == "" {
			resource = "template"
		}
		kind := item.Kind
		if kind == "" {
			kind = "unknown"
		}
		fmt.Fprintf(builder, "- %s (%s): %s\n", markdownCode(resource), markdownCode(kind), markdownText(item.Message))
	}
	builder.WriteString("\n")
}

func hasKind(g graph.Graph, kind string) bool {
	for _, node := range g.Nodes {
		if node.Kind == kind {
			return true
		}
	}
	return false
}
