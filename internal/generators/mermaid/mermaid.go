// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package mermaid

import (
	"fmt"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/graph"
)

func Render(g graph.Graph) []artifact.File {
	var builder strings.Builder
	builder.WriteString("flowchart LR\n")

	nodes := append([]graph.Node(nil), g.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	nodeIDs := stableMermaidIDs(nodes)
	for _, node := range nodes {
		fmt.Fprintf(&builder, "  %s[\"%s<br/>%s\"]\n", nodeIDs[node.ID], mermaidLabel(node.ID), mermaidLabel(node.Kind))
	}

	edges := append([]graph.Edge(nil), g.Edges...)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
	for _, edge := range edges {
		label := edge.Kind
		if edge.Protocol != "" && edge.Port > 0 {
			label = fmt.Sprintf("%s %s/%d", edge.Kind, edge.Protocol, edge.Port)
		}
		fmt.Fprintf(&builder, "  %s -->|\"%s\"| %s\n", mermaidReferenceID(edge.From, nodeIDs), mermaidLabel(label), mermaidReferenceID(edge.To, nodeIDs))
	}

	return []artifact.File{
		{Path: "architecture.mmd", Data: []byte(builder.String())},
	}
}

func stableMermaidIDs(nodes []graph.Node) map[string]string {
	ids := map[string]string{}
	used := map[string]struct{}{}
	for _, node := range nodes {
		base := mermaidID(node.ID)
		if base == "" {
			base = "node"
		}
		candidate := base
		for suffix := 2; ; suffix++ {
			if _, exists := used[candidate]; !exists {
				break
			}
			candidate = fmt.Sprintf("%s_%d", base, suffix)
		}
		used[candidate] = struct{}{}
		ids[node.ID] = candidate
	}
	return ids
}

func mermaidReferenceID(id string, nodeIDs map[string]string) string {
	if mapped, ok := nodeIDs[id]; ok {
		return mapped
	}
	fallback := mermaidID(id)
	if fallback == "" {
		return "node"
	}
	return fallback
}

func mermaidLabel(label string) string {
	label = strings.ReplaceAll(label, "\r\n", " ")
	label = strings.ReplaceAll(label, "\n", " ")
	label = strings.ReplaceAll(label, "\r", " ")
	label = strings.ReplaceAll(label, "\t", " ")
	label = strings.ReplaceAll(label, "-->", "-&gt;")
	replacer := strings.NewReplacer(
		`"`, "#quot;",
		"|", "&#124;",
		"<", "&lt;",
		">", "&gt;",
		"[", "&#91;",
		"]", "&#93;",
	)
	return replacer.Replace(strings.Join(strings.Fields(label), " "))
}

func mermaidID(id string) string {
	var builder strings.Builder
	for _, value := range id {
		switch {
		case value >= 'a' && value <= 'z':
			builder.WriteRune(value)
		case value >= 'A' && value <= 'Z':
			builder.WriteRune(value)
		case value >= '0' && value <= '9':
			builder.WriteRune(value)
		default:
			builder.WriteRune('_')
		}
	}
	return builder.String()
}
