// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/steadytao/planwright/internal/graph"
)

func TestProfilesListBuiltInProfiles(t *testing.T) {
	t.Parallel()

	profiles := Profiles()
	for _, want := range []string{"lab", "small-business", "production"} {
		if !hasProfile(profiles, want) {
			t.Fatalf("Profiles() = %#v, want %q", profiles, want)
		}
	}
}

func TestEvaluateLabProfileAcceptsBaselineGraph(t *testing.T) {
	t.Parallel()

	result, err := Evaluate(baselineGraph(), "lab")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if result.Profile.ID != "lab" {
		t.Fatalf("Result.Profile.ID = %q, want lab", result.Profile.ID)
	}
	if HasBlockingFindings(result.Findings) {
		t.Fatalf("lab profile findings = %#v, want no blocking findings", result.Findings)
	}
}

func TestEvaluateFlagsPublicDatabase(t *testing.T) {
	t.Parallel()

	g := baselineGraph()
	g.Nodes = append(g.Nodes, graph.Node{
		ID:   "public-db",
		Kind: "aws.rds.postgres",
		Properties: map[string]any{
			"publicly_accessible": true,
		},
	})

	result, err := Evaluate(g, "lab")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-POL-NET-001", "public-db")
}

func TestEvaluateFlagsInternetFacingAdminIngress(t *testing.T) {
	t.Parallel()

	g := baselineGraph()
	g.Edges = append(g.Edges, graph.Edge{
		From:     "internet",
		To:       "app",
		Kind:     "network.allow",
		Protocol: "tcp",
		Port:     22,
	})

	result, err := Evaluate(g, "lab")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-POL-NET-002", "internet -> app")
}

func TestEvaluateFlagsInternetFacingAdminIngressFromExternalInternetNode(t *testing.T) {
	t.Parallel()

	g := baselineGraph()
	g.Nodes[0] = graph.Node{ID: "world", Kind: "external.internet"}
	g.Edges = []graph.Edge{
		{
			From:     "world",
			To:       "app",
			Kind:     "network.allow",
			Protocol: "tcp",
			Port:     3389,
		},
	}

	result, err := Evaluate(g, "lab")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-POL-NET-002", "world -> app")
}

func TestEvaluateProductionFlagsMissingDatabaseBackupEvidence(t *testing.T) {
	t.Parallel()

	result, err := Evaluate(baselineGraph(), "production")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	finding := assertFinding(t, result.Findings, "PW-POL-BACKUP-001", "db")
	if finding.Severity != SeverityError {
		t.Fatalf("PW-POL-BACKUP-001 severity = %q, want %q", finding.Severity, SeverityError)
	}
	if !HasBlockingFindings(result.Findings) {
		t.Fatalf("production findings = %#v, want blocking backup finding", result.Findings)
	}
}

func TestEvaluateRejectsUnknownProfile(t *testing.T) {
	t.Parallel()

	_, err := Evaluate(baselineGraph(), "future")
	if err == nil {
		t.Fatal("Evaluate() error = nil, want unknown profile error")
	}
}

func baselineGraph() graph.Graph {
	return graph.Graph{
		Version:  graph.Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Profile:  "lab",
		Nodes: []graph.Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service"},
			{
				ID:   "db",
				Kind: "aws.rds.postgres",
				Properties: map[string]any{
					"publicly_accessible": false,
				},
			},
		},
		Edges: []graph.Edge{
			{From: "internet", To: "app", Kind: "network.allow", Protocol: "tcp", Port: 443},
		},
	}
}

func hasProfile(profiles []Profile, id string) bool {
	for _, profile := range profiles {
		if profile.ID == id {
			return true
		}
	}
	return false
}

func assertFinding(t *testing.T, findings []Finding, ruleID string, resource string) Finding {
	t.Helper()

	for _, finding := range findings {
		if finding.RuleID == ruleID && finding.Resource == resource {
			return finding
		}
	}
	t.Fatalf("findings = %#v, want %s for %s", findings, ruleID, resource)
	return Finding{}
}
