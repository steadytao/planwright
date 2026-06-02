// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"fmt"
	"strings"

	"github.com/steadytao/planwright/internal/graph"
)

func RenderGraphDiff(diff graph.Diff, oldSource string, newSource string) string {
	var builder strings.Builder
	builder.WriteString("# Graph Diff Review\n\n")
	builder.WriteString("Planwright compares graph JSON as local static artefacts. It does not contact cloud APIs, prove drift or deploy changes.\n\n")
	builder.WriteString("## Summary\n\n")
	fmt.Fprintf(&builder, "- Old graph: %s\n", markdownCode(oldSource))
	fmt.Fprintf(&builder, "- New graph: %s\n", markdownCode(newSource))
	fmt.Fprintf(&builder, "- Added nodes: %d\n", len(diff.AddedNodes))
	fmt.Fprintf(&builder, "- Removed nodes: %d\n", len(diff.RemovedNodes))
	fmt.Fprintf(&builder, "- Changed nodes: %d\n", len(diff.ChangedNodes))
	fmt.Fprintf(&builder, "- Added edges: %d\n", len(diff.AddedEdges))
	fmt.Fprintf(&builder, "- Removed edges: %d\n", len(diff.RemovedEdges))
	fmt.Fprintf(&builder, "- Findings: %d\n\n", len(diff.Findings))

	renderDiffFindings(&builder, diff.Findings)
	renderNodeChanges(&builder, "Added Nodes", diff.AddedNodes, "No nodes were added.")
	renderNodeChanges(&builder, "Removed Nodes", diff.RemovedNodes, "No nodes were removed.")
	renderNodeChanges(&builder, "Changed Nodes", diff.ChangedNodes, "No nodes changed kind or properties.")
	renderEdgeChanges(&builder, "Added Edges", diff.AddedEdges, "No edges were added.")
	renderEdgeChanges(&builder, "Removed Edges", diff.RemovedEdges, "No edges were removed.")

	builder.WriteString("## Review Notes\n\n")
	builder.WriteString("- This diff compares Planwright graph shape, not live infrastructure state.\n")
	builder.WriteString("- Added public ingress, removed controls and database exposure changes still require manual review.\n")
	builder.WriteString("- A quiet diff does not prove that generated infrastructure is deployable or compliant.\n")
	return builder.String()
}

func renderDiffFindings(builder *strings.Builder, findings []graph.DiffFinding) {
	builder.WriteString("## Findings\n\n")
	if len(findings) == 0 {
		builder.WriteString("- PASS: No v0.7 graph diff findings were detected.\n\n")
		return
	}
	for _, finding := range findings {
		fmt.Fprintf(builder, "### %s %s\n\n", strings.ToUpper(finding.Severity), finding.RuleID)
		fmt.Fprintf(builder, "- Resource: %s\n", markdownCode(finding.Resource))
		fmt.Fprintf(builder, "- Problem: %s\n", markdownText(finding.Message))
		if finding.Why != "" {
			fmt.Fprintf(builder, "- Why it matters: %s\n", markdownText(finding.Why))
		}
		if finding.Fix != "" {
			fmt.Fprintf(builder, "- Fix: %s\n", markdownText(finding.Fix))
		}
		builder.WriteString("\n")
	}
}

func renderNodeChanges(builder *strings.Builder, title string, changes []graph.NodeChange, empty string) {
	fmt.Fprintf(builder, "## %s\n\n", title)
	if len(changes) == 0 {
		builder.WriteString("- " + empty + "\n\n")
		return
	}
	for _, change := range changes {
		kind := change.Kind
		if kind == "" {
			kind = change.NewKind
		}
		if change.OldKind != "" && change.NewKind != "" && change.OldKind != change.NewKind {
			fmt.Fprintf(builder, "- %s: %s -> %s\n", markdownCode(change.ID), markdownCode(change.OldKind), markdownCode(change.NewKind))
			continue
		}
		fmt.Fprintf(builder, "- %s (%s)\n", markdownCode(change.ID), markdownCode(kind))
	}
	builder.WriteString("\n")
}

func renderEdgeChanges(builder *strings.Builder, title string, changes []graph.EdgeChange, empty string) {
	fmt.Fprintf(builder, "## %s\n\n", title)
	if len(changes) == 0 {
		builder.WriteString("- " + empty + "\n\n")
		return
	}
	for _, change := range changes {
		edge := change.Edge
		fmt.Fprintf(builder, "- %s -> %s (%s", markdownCode(edge.From), markdownCode(edge.To), markdownCode(edge.Kind))
		if edge.Protocol != "" && edge.Port > 0 {
			fmt.Fprintf(builder, " %s/%d", markdownText(edge.Protocol), edge.Port)
		}
		builder.WriteString(")\n")
	}
	builder.WriteString("\n")
}
