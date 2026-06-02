// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/plan"
)

func TestBuildPackIncludesGeneratedOutputsAndManifest(t *testing.T) {
	t.Parallel()

	files, err := BuildPack("planwright.yaml", []byte("version: planwright.v1\n"), loadExampleGraph(t))
	if err != nil {
		t.Fatalf("BuildPack() error = %v", err)
	}

	for _, path := range []string{
		"planwright.yaml",
		"planwright.graph.json",
		"manifest.json",
		"generated/terraform/versions.tf",
		"generated/terraform/security-groups.tf",
		"diagrams/architecture.mmd",
		"reports/security-report.md",
		"reports/cost-notes.md",
		"reports/deployability-report.md",
		"reports/cleanup.md",
		"reports/assumptions.md",
	} {
		if _, ok := artifact.ByPath(files, path); !ok {
			t.Fatalf("BuildPack() missing %s from files %#v", path, artifact.Paths(files))
		}
	}

	manifest, _ := artifact.ByPath(files, "manifest.json")
	if !strings.Contains(string(manifest.Data), `"schema": "planwright.pack.v1"`) {
		t.Fatalf("manifest.json = %s, want schema", string(manifest.Data))
	}
	if !strings.Contains(string(manifest.Data), `"requires_review": true`) {
		t.Fatalf("manifest.json = %s, want requires_review", string(manifest.Data))
	}
	if !strings.Contains(string(manifest.Data), `"created_by": "planwright 0.11.0"`) {
		t.Fatalf("manifest.json = %s, want current created_by", string(manifest.Data))
	}
	if strings.Contains(string(manifest.Data), "planwright 0.3.0") {
		t.Fatalf("manifest.json contains stale version: %s", string(manifest.Data))
	}
}

func loadExampleGraph(t *testing.T) graph.Graph {
	t.Helper()

	document, err := plan.Load(filepath.Join("..", "..", "examples", "aws-webapp-basic", "planwright.yaml"))
	if err != nil {
		t.Fatalf("plan.Load() error = %v", err)
	}
	lowered, diagnostics := document.ToGraph()
	if graph.HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("ToGraph() diagnostics = %#v", diagnostics)
	}
	return lowered
}
