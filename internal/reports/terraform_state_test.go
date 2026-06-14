// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/review/terraformstate"
)

func TestRenderTerraformStateInventory(t *testing.T) {
	t.Parallel()

	report := RenderTerraformStateInventory(terraformstate.Result{
		Source:           "state.json",
		FormatVersion:    "1.0",
		TerraformVersion: "1.15.5",
		ResourceCount:    1,
		Resources: []terraformstate.Resource{{
			Address:             "aws_db_instance.app",
			Type:                "aws_db_instance",
			ProviderName:        "registry.terraform.io/hashicorp/aws",
			Supported:           true,
			SensitiveAttributes: []string{"password"},
		}},
	})

	for _, want := range []string{
		"# Terraform State Inventory",
		"- Source: `state.json`",
		"- Resources: `1`",
		"`aws_db_instance.app`",
		"`password`",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("report = %s, want %q", report, want)
		}
	}
	if strings.Contains(report, "super-secret-password") {
		t.Fatalf("report leaked sensitive state value: %s", report)
	}
}
