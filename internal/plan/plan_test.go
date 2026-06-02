// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
)

func TestLoadExamplePlanLowersToGraph(t *testing.T) {
	t.Parallel()

	document, err := Load(filepath.Join("..", "..", "examples", "aws-webapp-basic", "planwright.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	lowered, diagnostics := document.ToGraph()
	if graph.HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("ToGraph() returned blocking diagnostics: %#v", diagnostics)
	}
	if got, want := lowered.Version, "planwright.graph.v1"; got != want {
		t.Fatalf("lowered.Version = %q, want %q", got, want)
	}
	if got, want := lowered.Provider, "aws"; got != want {
		t.Fatalf("lowered.Provider = %q, want %q", got, want)
	}
	if got, want := lowered.Region, "ap-southeast-2"; got != want {
		t.Fatalf("lowered.Region = %q, want %q", got, want)
	}
	if got, want := len(lowered.Nodes), 4; got != want {
		t.Fatalf("len(lowered.Nodes) = %d, want %d", got, want)
	}
	if got, want := len(lowered.Edges), 3; got != want {
		t.Fatalf("len(lowered.Edges) = %d, want %d", got, want)
	}
}

func TestParseRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`
version: planwright.v2
provider: aws
region: ap-southeast-2
`), "bad-version.yaml")
	if err == nil {
		t.Fatal("Parse() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), `unsupported Planwright plan version "planwright.v2"`) {
		t.Fatalf("Parse() error = %q, want unsupported-version error", err.Error())
	}
}

func TestToGraphRejectsUnsupportedPattern(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  worker:
    pattern: aws.worker.future
`), "unsupported-pattern.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	diagnostic := assertPlanDiagnostic(t, diagnostics, "PW-PLAN-PATTERN-001", "worker")
	if strings.Contains(diagnostic.Message, "v0.2") {
		t.Fatalf("unsupported-pattern diagnostic contains stale version text: %#v", diagnostic)
	}
}

func TestToGraphRejectsInvalidComponentID(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  "web app":
    pattern: aws.webapp.alb_ecs_rds
`), "invalid-component-id.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertPlanDiagnostic(t, diagnostics, "PW-PLAN-COMPONENT-001", "web app")
}

func TestToGraphRejectsUnsupportedDatabaseEngine(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      db_engine: mysql
`), "unsupported-db-engine.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertPlanDiagnostic(t, diagnostics, "PW-PLAN-PROPERTY-001", "webapp.db_engine")
}

func TestToGraphRejectsInvalidAppPort(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      app_port: 70000
`), "invalid-app-port.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertPlanDiagnostic(t, diagnostics, "PW-PLAN-PROPERTY-002", "webapp.app_port")
}

func TestToGraphRejectsInvalidDBPublicType(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      db_public: "false"
`), "invalid-db-public.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertPlanDiagnostic(t, diagnostics, "PW-PLAN-PROPERTY-003", "webapp.db_public")
}

func TestToGraphRejectsUnsupportedComponentProperty(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      db_pubic: true
`), "unsupported-component-property.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertPlanDiagnostic(t, diagnostics, "PW-PLAN-PROPERTY-004", "webapp.db_pubic")
}

func TestToGraphRejectsWhitespacePaddedFlowEndpoint(t *testing.T) {
	t.Parallel()

	document, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
flows:
  - from: " internet"
    to: webapp.alb
    kind: network.allow
    protocol: tcp
    port: 443
`), "whitespace-flow.yaml")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	_, diagnostics := document.ToGraph()

	assertDiagnosticCode(t, diagnostics, "PW-GRAPH-EDGE-004")
}

func TestParseRejectsInvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("version: ["), "invalid.yaml")
	if err == nil {
		t.Fatal("Parse() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "parse invalid.yaml") {
		t.Fatalf("Parse() error = %q, want source name", err.Error())
	}
}

func TestParseRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
future: true
`), "unknown-field.yaml")
	if err == nil {
		t.Fatal("Parse() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "field future not found") {
		t.Fatalf("Parse() error = %q, want unknown-field error", err.Error())
	}
}

func TestParseRejectsDuplicateMappingKeys(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte(`
version: planwright.v1
provider: aws
provider: kubernetes
region: ap-southeast-2
`), "duplicate-field.yaml")
	if err == nil || !strings.Contains(err.Error(), `duplicate mapping key "provider"`) {
		t.Fatalf("Parse() error = %v, want duplicate-key error", err)
	}
}

func TestLoadRejectsDirectory(t *testing.T) {
	t.Parallel()

	_, err := Load(t.TempDir())
	if err == nil {
		t.Fatal("Load(directory) error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("Load(directory) error = %q, want regular-file error", err.Error())
	}
}

func TestLoadRejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.yaml")
	if err := os.WriteFile(target, []byte("version: planwright.v1\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	link := filepath.Join(dir, "link.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := Load(link)
	if err == nil {
		t.Fatal("Load(symlink) error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("Load(symlink) error = %q, want symlink error", err.Error())
	}
}

func TestLoadRejectsOversizedPlan(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "large.yaml")
	data := bytes.Repeat([]byte("x"), maxPlanBytes+1)
	if err := os.WriteFile(target, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := Load(target)
	if err == nil || !strings.Contains(err.Error(), "plan exceeds") {
		t.Fatalf("Load() error = %v, want size refusal", err)
	}
}

func TestParseRejectsMultipleDocuments(t *testing.T) {
	t.Parallel()

	_, err := Parse([]byte("version: planwright.v1\n---\nversion: planwright.v1\n"), "multi.yaml")
	if err == nil || !strings.Contains(err.Error(), "exactly one YAML document") {
		t.Fatalf("Parse() error = %v, want multi-document refusal", err)
	}
}

func assertPlanDiagnostic(t *testing.T, diagnostics []graph.Diagnostic, code string, resource string) graph.Diagnostic {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Resource == resource {
			return diagnostic
		}
	}
	t.Fatalf("diagnostics = %#v, want code %q for resource %q", diagnostics, code, resource)
	return graph.Diagnostic{}
}

func assertDiagnosticCode(t *testing.T, diagnostics []graph.Diagnostic, code string) graph.Diagnostic {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return diagnostic
		}
	}
	t.Fatalf("diagnostics = %#v, want code %q", diagnostics, code)
	return graph.Diagnostic{}
}
