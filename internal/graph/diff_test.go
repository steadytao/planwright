// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import "testing"

func TestDiffReportsAddedRemovedAndChangedGraphElements(t *testing.T) {
	t.Parallel()

	oldGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service", Properties: map[string]any{"port": 8080}},
			{ID: "old-cache", Kind: "aws.elasticache.cluster"},
		},
		Edges: []Edge{
			{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 443},
			{From: "app", To: "old-cache", Kind: "depends_on"},
		},
	}
	newGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service", Properties: map[string]any{"port": 8080, "desired_count": 2}},
			{ID: "db", Kind: "aws.rds.postgres", Properties: map[string]any{"publicly_accessible": true}},
		},
		Edges: []Edge{
			{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 443},
			{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 22},
			{From: "app", To: "db", Kind: "network.allow", Protocol: "tcp", Port: 5432},
		},
	}

	diff := Compare(oldGraph, newGraph)

	if got, want := len(diff.AddedNodes), 1; got != want {
		t.Fatalf("AddedNodes length = %d, want %d: %#v", got, want, diff.AddedNodes)
	}
	if diff.AddedNodes[0].ID != "db" {
		t.Fatalf("AddedNodes[0].ID = %q, want db", diff.AddedNodes[0].ID)
	}
	if got, want := len(diff.RemovedNodes), 1; got != want {
		t.Fatalf("RemovedNodes length = %d, want %d: %#v", got, want, diff.RemovedNodes)
	}
	if diff.RemovedNodes[0].ID != "old-cache" {
		t.Fatalf("RemovedNodes[0].ID = %q, want old-cache", diff.RemovedNodes[0].ID)
	}
	if got, want := len(diff.ChangedNodes), 1; got != want {
		t.Fatalf("ChangedNodes length = %d, want %d: %#v", got, want, diff.ChangedNodes)
	}
	if diff.ChangedNodes[0].ID != "app" {
		t.Fatalf("ChangedNodes[0].ID = %q, want app", diff.ChangedNodes[0].ID)
	}
	if got, want := len(diff.AddedEdges), 2; got != want {
		t.Fatalf("AddedEdges length = %d, want %d: %#v", got, want, diff.AddedEdges)
	}
	if got, want := len(diff.RemovedEdges), 1; got != want {
		t.Fatalf("RemovedEdges length = %d, want %d: %#v", got, want, diff.RemovedEdges)
	}

	assertFinding(t, diff.Findings, "PW-DIFF-RDS-001", "db")
	assertFinding(t, diff.Findings, "PW-DIFF-NET-001", "internet -> app")
}

func TestDiffReportsDatabaseBecomingPublic(t *testing.T) {
	t.Parallel()

	oldGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "db", Kind: "aws.rds.postgres", Properties: map[string]any{"publicly_accessible": false}},
		},
	}
	newGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "db", Kind: "aws.rds.postgres", Properties: map[string]any{"publicly_accessible": true}},
		},
	}

	diff := Compare(oldGraph, newGraph)

	assertFinding(t, diff.Findings, "PW-DIFF-RDS-001", "db")
}

func TestDiffReportsInternetIngressFromExternalInternetNode(t *testing.T) {
	t.Parallel()

	oldGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "world", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service"},
		},
	}
	newGraph := oldGraph
	newGraph.Edges = []Edge{
		{From: "world", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 22},
	}

	diff := Compare(oldGraph, newGraph)

	assertFinding(t, diff.Findings, "PW-DIFF-NET-001", "world -> app")
}

func TestDiffIsDeterministic(t *testing.T) {
	t.Parallel()

	oldGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "b", Kind: "aws.s3.bucket"},
			{ID: "a", Kind: "aws.vpc"},
		},
	}
	newGraph := Graph{
		Version:  Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "d", Kind: "aws.lambda.function"},
			{ID: "c", Kind: "aws.rds.postgres"},
		},
	}

	diff := Compare(oldGraph, newGraph)

	if got, want := diff.RemovedNodes[0].ID, "a"; got != want {
		t.Fatalf("first removed node = %q, want %q", got, want)
	}
	if got, want := diff.AddedNodes[0].ID, "c"; got != want {
		t.Fatalf("first added node = %q, want %q", got, want)
	}
}

func assertFinding(t *testing.T, findings []DiffFinding, ruleID string, resource string) {
	t.Helper()

	for _, finding := range findings {
		if finding.RuleID == ruleID && finding.Resource == resource {
			return
		}
	}
	t.Fatalf("findings = %#v, want %s for %s", findings, ruleID, resource)
}
