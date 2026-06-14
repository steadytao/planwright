// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/review/terraformplan"
)

func TestValidateSARIFAcceptsRenderedTerraformReview(t *testing.T) {
	t.Parallel()

	data, err := RenderTerraformReviewSARIF(terraformplan.Result{
		Source: "tfplan.json",
		Findings: []terraformplan.Finding{{
			RuleID:   "PW-TF-RDS-001",
			Severity: "high",
			Address:  "aws_db_instance.app",
			Message:  "Database is publicly accessible.",
		}},
	})
	if err != nil {
		t.Fatalf("RenderTerraformReviewSARIF() error = %v", err)
	}
	if err := ValidateSARIF(data, "terraform-review.sarif"); err != nil {
		t.Fatalf("ValidateSARIF(rendered terraform SARIF) error = %v", err)
	}
}

func TestValidateSARIFRejectsMissingRun(t *testing.T) {
	t.Parallel()

	err := ValidateSARIF([]byte(`{"version":"2.1.0","runs":[]}`), "empty.sarif")
	if err == nil {
		t.Fatalf("ValidateSARIF(missing run) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "runs must contain at least one run") {
		t.Fatalf("ValidateSARIF(missing run) error = %v, want missing run reason", err)
	}
}

func TestValidateSARIFRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	err := ValidateSARIF([]byte(`{`), "bad.sarif")
	if err == nil {
		t.Fatalf("ValidateSARIF(invalid JSON) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "parse bad.sarif") {
		t.Fatalf("ValidateSARIF(invalid JSON) error = %v, want source name", err)
	}
}
