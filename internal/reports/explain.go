// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"fmt"
	"strings"

	"github.com/steadytao/planwright/internal/graph"
)

func Explain(g graph.Graph, diagnostics []graph.Diagnostic) string {
	var builder strings.Builder

	builder.WriteString("# Planwright Graph\n\n")
	fmt.Fprintf(&builder, "- Version: %s\n", markdownText(g.Version))
	fmt.Fprintf(&builder, "- Provider: %s\n", markdownText(g.Provider))
	fmt.Fprintf(&builder, "- Region: %s\n", markdownText(g.Region))
	fmt.Fprintf(&builder, "- Profile: %s\n", markdownText(g.Profile))
	fmt.Fprintf(&builder, "- Nodes: %d\n", len(g.Nodes))
	fmt.Fprintf(&builder, "- Edges: %d\n", len(g.Edges))

	if len(g.Nodes) > 0 {
		builder.WriteString("\n## Nodes\n\n")
		for _, node := range g.Nodes {
			fmt.Fprintf(&builder, "- %s (%s)\n", markdownCode(node.ID), markdownText(node.Kind))
		}
	}

	if len(g.Edges) > 0 {
		builder.WriteString("\n## Edges\n\n")
		for _, edge := range g.Edges {
			fmt.Fprintf(&builder, "- %s -> %s (%s", markdownCode(edge.From), markdownCode(edge.To), markdownText(edge.Kind))
			if edge.Protocol != "" && edge.Port > 0 {
				fmt.Fprintf(&builder, " %s/%d", markdownText(edge.Protocol), edge.Port)
			}
			builder.WriteString(")\n")
		}
	}

	if len(diagnostics) > 0 {
		builder.WriteString("\n## Diagnostics\n\n")
		builder.WriteString(RenderDiagnostics(diagnostics))
	}

	return builder.String()
}

func RenderDiagnostics(diagnostics []graph.Diagnostic) string {
	var builder strings.Builder
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(&builder, "- %s %s", strings.ToUpper(diagnostic.Severity), diagnostic.Code)
		if diagnostic.Resource != "" {
			fmt.Fprintf(&builder, " %s", markdownCode(diagnostic.Resource))
		}
		if diagnostic.Message != "" {
			fmt.Fprintf(&builder, ": %s", markdownText(diagnostic.Message))
		}
		if diagnostic.Fix != "" {
			fmt.Fprintf(&builder, " Fix: %s", markdownText(diagnostic.Fix))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
