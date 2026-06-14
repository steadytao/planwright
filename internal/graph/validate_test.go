// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import "testing"

func TestValidateAcceptsBasicAWSWebAppGraph(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Profile:  "lab",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "webapp.alb", Kind: "aws.alb"},
			{ID: "webapp.app", Kind: "aws.ecs.service"},
			{ID: "webapp.db", Kind: "aws.rds.postgres"},
		},
		Edges: []Edge{
			{From: "internet", To: "webapp.alb", Kind: "network.allow", Protocol: "tcp", Port: 443},
			{From: "webapp.alb", To: "webapp.app", Kind: "network.allow", Protocol: "tcp", Port: 8080},
			{From: "webapp.app", To: "webapp.db", Kind: "network.allow", Protocol: "tcp", Port: 5432},
		},
	})
	if HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("Validate() returned blocking diagnostics: %#v", diagnostics)
	}
}

func TestNormalizeSlicesMakesNilSlicesEmpty(t *testing.T) {
	t.Parallel()

	g := NormalizeSlices(Graph{})
	if g.Nodes == nil {
		t.Fatalf("NormalizeSlices(Graph{}).Nodes is nil, want empty slice")
	}
	if g.Edges == nil {
		t.Fatalf("NormalizeSlices(Graph{}).Edges is nil, want empty slice")
	}
}

func TestValidateRejectsMissingRegion(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-REGION-001", "graph")
}

func TestValidateRejectsUnknownEdgeEndpoint(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
		},
		Edges: []Edge{
			{From: "internet", To: "missing", Kind: "network.allow", Protocol: "tcp", Port: 443},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-001", "missing")
}

func TestValidateRejectsWhitespacePaddedNodeID(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: " internet ", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service"},
		},
		Edges: []Edge{
			{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 443},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-NODE-003", "internet")
	assertDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-001", "internet")
}

func TestValidateRejectsWhitespacePaddedEdgeEndpoint(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service"},
		},
		Edges: []Edge{
			{From: " internet ", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 443},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-004", " internet  -> app")
}

func TestValidateRejectsInvalidNetworkPort(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "webapp.alb", Kind: "aws.alb"},
		},
		Edges: []Edge{
			{From: "internet", To: "webapp.alb", Kind: "network.allow", Protocol: "tcp", Port: 0},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-NET-001", "internet -> webapp.alb")
}

func TestValidateWarnsAboutPublicDatabase(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{
				ID:   "webapp.db",
				Kind: "aws.rds.postgres",
				Properties: map[string]any{
					"publicly_accessible": true,
				},
			},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-AWS-RDS-001", "webapp.db")
}

func TestValidateWarnsAboutGenericPublicRDSInstance(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{
				ID:   "imported.db",
				Kind: "aws.rds.instance",
				Properties: map[string]any{
					"publicly_accessible": true,
				},
			},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-AWS-RDS-001", "imported.db")
}

func TestValidateRejectsMalformedPublicDatabaseProperty(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{
				ID:   "db",
				Kind: "aws.rds.postgres",
				Properties: map[string]any{
					"publicly_accessible": "true",
				},
			},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-PROPERTY-001", "db.publicly_accessible")
}

func TestValidateRejectsUnknownEdgeKind(t *testing.T) {
	t.Parallel()

	diagnostics := Validate(Graph{
		Version:  "planwright.graph.v1",
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service"},
		},
		Edges: []Edge{
			{From: "internet", To: "app", Kind: "network.alow", Protocol: "tcp", Port: 22},
		},
	})

	assertDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-002", "internet -> app")
}

func assertDiagnostic(t *testing.T, diagnostics []Diagnostic, code string, resource string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Resource == resource {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want code %q for resource %q", diagnostics, code, resource)
}
