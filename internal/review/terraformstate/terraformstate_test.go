// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraformstate

import (
	"slices"
	"testing"
)

func TestReviewBytesInventoriesResourcesAndSensitiveAttributes(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "terraform_version": "1.15.5",
	  "values": {
	    "root_module": {
	      "resources": [
	        {
	          "address": "aws_db_instance.app",
	          "mode": "managed",
	          "type": "aws_db_instance",
	          "name": "app",
	          "provider_name": "registry.terraform.io/hashicorp/aws",
	          "values": {
	            "identifier": "app-db",
	            "password": "super-secret-password"
	          },
	          "sensitive_values": {
	            "password": true
	          }
	        }
	      ]
	    }
	  }
	}`), "state.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	if got, want := len(result.Resources), 1; got != want {
		t.Fatalf("resource count = %d, want %d", got, want)
	}
	resource := result.Resources[0]
	if got, want := resource.Address, "aws_db_instance.app"; got != want {
		t.Fatalf("resource address = %q, want %q", got, want)
	}
	if got, want := resource.ProviderName, "registry.terraform.io/hashicorp/aws"; got != want {
		t.Fatalf("provider name = %q, want %q", got, want)
	}
	if got, want := resource.SensitiveAttributes, []string{"password"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("sensitive attributes = %#v, want %#v", got, want)
	}
	if containsSensitiveValue(result, "super-secret-password") {
		t.Fatalf("result leaked sensitive state value")
	}
}

func containsSensitiveValue(result Result, value string) bool {
	for _, resource := range result.Resources {
		if slices.Contains(resource.SensitiveAttributes, value) {
			return true
		}
	}
	return false
}
