// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
)

func TestRenderGraphDiff(t *testing.T) {
	t.Parallel()

	report := RenderGraphDiff(graph.Diff{
		AddedNodes: []graph.NodeChange{
			{ID: "db", Kind: "aws.rds.postgres"},
		},
		RemovedNodes: []graph.NodeChange{
			{ID: "old-cache", Kind: "aws.elasticache.cluster"},
		},
		AddedEdges: []graph.EdgeChange{
			{Edge: graph.Edge{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 22}},
		},
		Findings: []graph.DiffFinding{
			{
				Severity: "high",
				RuleID:   "PW-DIFF-NET-001",
				Resource: "internet -> app",
				Message:  "An internet-facing administrative network path was added.",
				Why:      "SSH or RDP exposure materially changes review risk.",
				Fix:      "Restrict the edge or remove it from the graph.",
			},
		},
	}, "old.graph.json", "new.graph.json")

	for _, want := range []string{
		"# Graph Diff Review",
		"old.graph.json",
		"new.graph.json",
		"Added nodes: 1",
		"Removed nodes: 1",
		"PW-DIFF-NET-001",
		"`internet` -> `app`",
		"Planwright compares graph JSON as local static artefacts",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("report = %s, want %q", report, want)
		}
	}
}

func TestRenderGraphDiffEscapesMarkdownFields(t *testing.T) {
	t.Parallel()

	report := RenderGraphDiff(graph.Diff{
		AddedNodes: []graph.NodeChange{
			{ID: "db`\n- forged", Kind: "aws.rds.postgres"},
		},
		AddedEdges: []graph.EdgeChange{
			{Edge: graph.Edge{From: "world`\n- forged", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 22}},
		},
		Findings: []graph.DiffFinding{
			{
				Severity: "high",
				RuleID:   "PW-DIFF-NET-001",
				Resource: "world`\n- forged -> app",
				Message:  "An internet-facing administrative network path was added.",
			},
		},
	}, "old.graph.json\n- forged", "new.graph.json")

	if strings.Contains(report, "\n- forged") {
		t.Fatalf("graph diff report contains forged Markdown line: %s", report)
	}
	if !strings.Contains(report, "`` ") {
		t.Fatalf("graph diff report = %s, want expanded code span for backtick content", report)
	}
}
