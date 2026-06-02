// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
	"github.com/steadytao/planwright/internal/plan"
	"github.com/steadytao/planwright/internal/review/terraformplan"
)

func TestRenderSecurityReport(t *testing.T) {
	t.Parallel()

	report := RenderSecurity(loadExampleGraph(t))
	for _, want := range []string{
		"# Security Report",
		"No public database exposure was detected",
		"Planwright does not prove deployability or compliance",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("security report = %s, want %q", report, want)
		}
	}
	if strings.Contains(report, "v0.2") {
		t.Fatalf("security report contains stale version text: %s", report)
	}
}

func TestRenderCostNotes(t *testing.T) {
	t.Parallel()

	report := RenderCostNotes(loadExampleGraph(t))
	for _, want := range []string{
		"# Cost Notes",
		"Application Load Balancer",
		"RDS",
		"NAT Gateway is not generated in the lab profile",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("cost notes = %s, want %q", report, want)
		}
	}
	if strings.Contains(report, "v0.2") {
		t.Fatalf("cost notes contains stale version text: %s", report)
	}
}

func TestRenderOperationalReports(t *testing.T) {
	t.Parallel()

	lowered := loadExampleGraph(t)
	reports := map[string]string{
		"deployability": RenderDeployability(lowered),
		"cleanup":       RenderCleanup(lowered),
		"assumptions":   RenderAssumptions(lowered),
	}

	expected := map[string][]string{
		"deployability": {"# Deployability Report", "review the generated Terraform before running it"},
		"cleanup":       {"# Cleanup Guide", "Planwright does not run destroy commands"},
		"assumptions":   {"# Assumptions", "region `ap-southeast-2`"},
	}

	for name, report := range reports {
		for _, want := range expected[name] {
			if !strings.Contains(report, want) {
				t.Fatalf("%s report = %s, want %q", name, report, want)
			}
		}
		if strings.Contains(report, "v0.2") {
			t.Fatalf("%s report contains stale version text: %s", name, report)
		}
	}
}

func TestRenderLossReport(t *testing.T) {
	t.Parallel()

	report := RenderLossReport(loss.Report{
		SourceFormat: "kubernetes",
		Source:       "manifests.yaml",
		Lowered: []loss.Item{
			{Resource: "default/app", Kind: "apps/v1/Deployment", Message: "Resource inventory was lowered into the Planwright graph."},
		},
		Unsupported: []loss.Item{
			{Resource: "default/future", Kind: "example.io/v1/Future", Message: "manual review required"},
		},
	})
	for _, want := range []string{
		"# Loss Report",
		"## Lowered",
		"## Unsupported",
		"manual review required",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("loss report = %s, want %q", report, want)
		}
	}
}

func TestRenderLossReportEscapesMarkdownFields(t *testing.T) {
	t.Parallel()

	report := RenderLossReport(loss.Report{
		SourceFormat: "kubernetes`",
		Source:       "source.md\n- forged",
		Unsupported: []loss.Item{
			{Resource: "bad`\n- forged", Kind: "kind`", Message: "manual review required"},
		},
	})

	if strings.Contains(report, "\n- forged") {
		t.Fatalf("loss report contains forged Markdown line: %s", report)
	}
	if !strings.Contains(report, "`` ") {
		t.Fatalf("loss report = %s, want expanded code span for backtick content", report)
	}
}

func TestTerraformReviewSARIFUsesRepositoryRelativeURI(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	source := filepath.Join(cwd, "fixtures", "tfplan.json")
	data, err := RenderTerraformReviewSARIF(terraformplan.Result{
		Source: source,
		Findings: []terraformplan.Finding{
			{
				RuleID:   "PW-TF-CHANGE-001",
				Severity: "high",
				Address:  "aws_s3_bucket.logs",
				Actions:  []string{"delete"},
				Message:  "Resource is planned for deletion.",
				Why:      "Deletion can remove data.",
				Fix:      "Review the deletion.",
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	uri := firstSARIFURI(t, data)
	if strings.Contains(uri, string(filepath.Separator)) || strings.Contains(uri, ":") {
		t.Fatalf("SARIF URI = %q, want repository-relative slash path without drive or scheme", uri)
	}
	if uri != "fixtures/tfplan.json" {
		t.Fatalf("SARIF URI = %q, want fixtures/tfplan.json", uri)
	}
}

func TestTerraformReviewSARIFUsesBasenameForExternalAbsolutePath(t *testing.T) {
	t.Parallel()

	source := filepath.Join(t.TempDir(), "tfplan.json")
	data, err := RenderTerraformReviewSARIF(terraformplan.Result{
		Source: source,
		Findings: []terraformplan.Finding{
			{
				RuleID:   "PW-TF-CHANGE-001",
				Severity: "high",
				Address:  "aws_s3_bucket.logs",
				Actions:  []string{"delete"},
				Message:  "Resource is planned for deletion.",
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	uri := firstSARIFURI(t, data)
	if uri != "tfplan.json" {
		t.Fatalf("SARIF URI = %q, want basename without local absolute path", uri)
	}
}

func TestTerraformReviewSARIFEscapesArtifactURI(t *testing.T) {
	t.Parallel()

	source := filepath.Join("fixtures", "plan #1?.json")
	data, err := RenderTerraformReviewSARIF(terraformplan.Result{
		Source: source,
		Findings: []terraformplan.Finding{
			{
				RuleID:   "PW-TF-CHANGE-001",
				Severity: "high",
				Address:  "aws_s3_bucket.logs",
				Actions:  []string{"delete"},
				Message:  "Resource is planned for deletion.",
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	uri := firstSARIFURI(t, data)
	if uri != "fixtures/plan%20%231%3F.json" {
		t.Fatalf("SARIF URI = %q, want escaped repository-relative URI", uri)
	}
}

func TestTerraformReviewSARIFUsesBasenameForRelativeTraversal(t *testing.T) {
	t.Parallel()

	data, err := RenderTerraformReviewSARIF(terraformplan.Result{
		Source: filepath.Join("..", "outside plan.json"),
		Findings: []terraformplan.Finding{
			{
				RuleID:   "PW-TF-CHANGE-001",
				Severity: "high",
				Address:  "aws_s3_bucket.logs",
				Actions:  []string{"delete"},
				Message:  "Resource is planned for deletion.",
			},
		},
	})
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	uri := firstSARIFURI(t, data)
	if uri != "outside%20plan.json" {
		t.Fatalf("SARIF URI = %q, want escaped basename without traversal", uri)
	}
}

func firstSARIFURI(t *testing.T, data []byte) string {
	t.Helper()

	var decoded struct {
		Runs []struct {
			Results []struct {
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							URI string `json:"uri"`
						} `json:"artifactLocation"`
					} `json:"physicalLocation"`
				} `json:"locations"`
			} `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("SARIF JSON invalid: %v\n%s", err, data)
	}
	if len(decoded.Runs) == 0 || len(decoded.Runs[0].Results) == 0 || len(decoded.Runs[0].Results[0].Locations) == 0 {
		t.Fatalf("SARIF has no first location: %s", data)
	}
	return decoded.Runs[0].Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI
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
