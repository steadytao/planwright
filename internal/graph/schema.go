// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

const Version = "planwright.graph.v1"

const (
	SeverityInfo  = "info"
	SeverityWarn  = "warn"
	SeverityError = "error"
)

type Graph struct {
	Version  string `json:"version"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Profile  string `json:"profile,omitempty"`
	Nodes    []Node `json:"nodes"`
	Edges    []Edge `json:"edges"`
}

type Node struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Name       string         `json:"name,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Kind     string `json:"kind"`
	Protocol string `json:"protocol,omitempty"`
	Port     int    `json:"port,omitempty"`
	Intent   string `json:"intent,omitempty"`
}

type Diagnostic struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Resource string `json:"resource,omitempty"`
	Message  string `json:"message"`
	Fix      string `json:"fix,omitempty"`
}

func HasBlockingDiagnostics(diagnostics []Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			return true
		}
	}
	return false
}

func NormalizeSlices(g Graph) Graph {
	if g.Nodes == nil {
		g.Nodes = []Node{}
	}
	if g.Edges == nil {
		g.Edges = []Edge{}
	}
	return g
}
