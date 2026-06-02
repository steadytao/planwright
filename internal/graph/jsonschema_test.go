// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONSchemaUsesDraft202012AndGraphVersion(t *testing.T) {
	t.Parallel()

	var decoded map[string]any
	if err := json.Unmarshal(JSONSchema(), &decoded); err != nil {
		t.Fatalf("JSONSchema() is not JSON: %v", err)
	}
	if got, want := decoded["$schema"], "https://json-schema.org/draft/2020-12/schema"; got != want {
		t.Fatalf("$schema = %#v, want %q", got, want)
	}
	if !strings.Contains(string(JSONSchema()), Version) {
		t.Fatalf("JSONSchema() = %s, want graph version %s", JSONSchema(), Version)
	}
}

func TestValidateJSONAcceptsValidGraph(t *testing.T) {
	t.Parallel()

	loaded, diagnostics := ValidateJSON([]byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}
	  ]
	}`), "graph.json")

	if HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("ValidateJSON() diagnostics = %#v, want no blocking diagnostics", diagnostics)
	}
	if got, want := loaded.Version, Version; got != want {
		t.Fatalf("loaded.Version = %q, want %q", got, want)
	}
}

func TestValidateJSONRejectsUnknownGraphField(t *testing.T) {
	t.Parallel()

	_, diagnostics := ValidateJSON([]byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [],
	  "edges": [],
	  "unexpected": true
	}`), "graph.json")

	assertSchemaDiagnostic(t, diagnostics, "PW-GRAPH-SCHEMA-001")
}

func TestValidateJSONIncludesSemanticDiagnostics(t *testing.T) {
	t.Parallel()

	_, diagnostics := ValidateJSON([]byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [{"id": "app", "kind": "aws.ecs.service"}],
	  "edges": [{"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}]
	}`), "graph.json")

	assertSchemaDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-001")
}

func TestValidateJSONRejectsWhitespacePaddedIdentity(t *testing.T) {
	t.Parallel()

	_, diagnostics := ValidateJSON([]byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": " internet ", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 22}
	  ]
	}`), "graph.json")

	assertSchemaDiagnostic(t, diagnostics, "PW-GRAPH-NODE-003")
	assertSchemaDiagnostic(t, diagnostics, "PW-GRAPH-EDGE-001")
}

func TestValidateJSONRejectsInvalidNetworkPortStructurally(t *testing.T) {
	t.Parallel()

	_, diagnostics := ValidateJSON([]byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 70000}
	  ]
	}`), "graph.json")

	assertSchemaDiagnostic(t, diagnostics, "PW-GRAPH-SCHEMA-001")
}

func assertSchemaDiagnostic(t *testing.T, diagnostics []Diagnostic, code string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want %s", diagnostics, code)
}
