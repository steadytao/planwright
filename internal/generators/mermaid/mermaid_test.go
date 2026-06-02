// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package mermaid

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/plan"
)

func TestRenderArchitectureDiagram(t *testing.T) {
	t.Parallel()

	files := Render(loadExampleGraph(t))
	file, ok := artifact.ByPath(files, "architecture.mmd")
	if !ok {
		t.Fatalf("Render() missing architecture.mmd from files %#v", artifact.Paths(files))
	}

	text := string(file.Data)
	for _, want := range []string{
		"flowchart LR",
		"internet",
		"webapp_alb",
		"webapp_app",
		"webapp_db",
		"tcp/443",
		"tcp/8080",
		"tcp/5432",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("architecture.mmd = %s, want %q", text, want)
		}
	}
}

func TestRenderEscapesMermaidLabels(t *testing.T) {
	t.Parallel()

	files := Render(graph.Graph{
		Version:  graph.Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []graph.Node{
			{ID: "internet", Kind: "external.internet"},
			{ID: "app", Kind: "aws.ecs.service\" ]\n  injected[\"bad"},
		},
		Edges: []graph.Edge{
			{From: "internet", To: "app", Kind: "network.allow\"| injected --> app |\"", Protocol: "tcp", Port: 443},
		},
	})
	file, ok := artifact.ByPath(files, "architecture.mmd")
	if !ok {
		t.Fatalf("Render() missing architecture.mmd from files %#v", artifact.Paths(files))
	}

	text := string(file.Data)
	if strings.Contains(text, "injected[") || strings.Contains(text, "injected -->") {
		t.Fatalf("architecture.mmd contains unescaped Mermaid injection: %s", text)
	}
	if !strings.Contains(text, "#quot;") {
		t.Fatalf("architecture.mmd = %s, want escaped quote entity", text)
	}
}

func TestRenderKeepsCollidingMermaidIDsDistinct(t *testing.T) {
	t.Parallel()

	files := Render(graph.Graph{
		Version:  graph.Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []graph.Node{
			{ID: "web-app.alb", Kind: "aws.alb"},
			{ID: "web_app.alb", Kind: "aws.alb"},
		},
		Edges: []graph.Edge{
			{From: "web-app.alb", To: "web_app.alb", Kind: "depends_on"},
		},
	})
	file, ok := artifact.ByPath(files, "architecture.mmd")
	if !ok {
		t.Fatalf("Render() missing architecture.mmd from files %#v", artifact.Paths(files))
	}

	text := string(file.Data)
	for _, want := range []string{
		`web_app_alb["web-app.alb<br/>aws.alb"]`,
		`web_app_alb_2["web_app.alb<br/>aws.alb"]`,
		`web_app_alb -->|"depends_on"| web_app_alb_2`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("architecture.mmd = %s, want %q", text, want)
		}
	}
}

func loadExampleGraph(t *testing.T) graph.Graph {
	t.Helper()

	document, err := plan.Load(filepath.Join("..", "..", "..", "examples", "aws-webapp-basic", "planwright.yaml"))
	if err != nil {
		t.Fatalf("plan.Load() error = %v", err)
	}
	lowered, diagnostics := document.ToGraph()
	if graph.HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("ToGraph() diagnostics = %#v", diagnostics)
	}
	return lowered
}
