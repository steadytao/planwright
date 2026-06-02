// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/review/terraformplan"
)

func TestRenderTerraformReviewMarkdown(t *testing.T) {
	t.Parallel()

	result := terraformplan.Result{
		Source:        "tfplan.json",
		FormatVersion: "1.0",
		ChangeCount:   1,
		Findings: []terraformplan.Finding{{
			RuleID:       "PW-TF-CHANGE-001",
			Severity:     "high",
			Address:      "aws_s3_bucket.logs",
			ResourceType: "aws_s3_bucket",
			Actions:      []string{"delete"},
			Message:      "Resource is planned for deletion.",
			Fix:          "Review whether deletion is intended.",
		}},
	}
	report := RenderTerraformReview(result)
	for _, want := range []string{"# Terraform Plan Review", "tfplan.json", "PW-TF-CHANGE-001", "aws_s3_bucket.logs"} {
		if !strings.Contains(report, want) {
			t.Fatalf("report = %s, want %q", report, want)
		}
	}
}

func TestRenderTerraformReviewEscapesMarkdownFields(t *testing.T) {
	t.Parallel()

	result := terraformplan.Result{
		Source:        "tfplan.json\n- forged",
		FormatVersion: "1.0",
		Findings: []terraformplan.Finding{{
			RuleID:       "PW-TF-CHANGE-001",
			Severity:     "high",
			Address:      "aws_s3_bucket.bad`\n- forged",
			ResourceType: "aws_s3_bucket",
			Actions:      []string{"delete"},
			Message:      "Resource is planned for deletion.",
		}},
	}

	report := RenderTerraformReview(result)
	if strings.Contains(report, "\n- forged") {
		t.Fatalf("terraform review contains forged Markdown line: %s", report)
	}
	if !strings.Contains(report, "`` ") {
		t.Fatalf("terraform review = %s, want expanded code span for backtick content", report)
	}
}

func TestRenderTerraformReviewSARIF(t *testing.T) {
	t.Parallel()

	result := terraformplan.Result{
		Source: "tfplan.json",
		Findings: []terraformplan.Finding{{
			RuleID:       "PW-TF-CHANGE-001",
			Severity:     "high",
			Address:      "aws_s3_bucket.logs",
			ResourceType: "aws_s3_bucket",
			Message:      "Resource is planned for deletion.",
		}},
	}
	data, err := RenderTerraformReviewSARIF(result)
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("SARIF JSON invalid: %v", err)
	}
	if decoded["version"] != "2.1.0" {
		t.Fatalf("version = %v, want 2.1.0", decoded["version"])
	}
	if !strings.Contains(string(data), "PW-TF-CHANGE-001") {
		t.Fatalf("SARIF = %s, want rule ID", data)
	}
}
