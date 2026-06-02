// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraformplan

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewDetectsDestructiveAndReplacementChanges(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "terraform_version": "1.15.5",
	  "applyable": true,
	  "complete": true,
	  "resource_changes": [
	    {
	      "address": "aws_db_instance.app",
	      "mode": "managed",
	      "type": "aws_db_instance",
	      "name": "app",
	      "change": {
	        "actions": ["delete", "create"],
	        "before": {"publicly_accessible": false},
	        "after": {"publicly_accessible": false}
	      }
	    },
	    {
	      "address": "aws_s3_bucket.logs",
	      "mode": "managed",
	      "type": "aws_s3_bucket",
	      "name": "logs",
	      "change": {"actions": ["delete"]}
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-CHANGE-002", "aws_db_instance.app")
	assertFinding(t, result.Findings, "PW-TF-CHANGE-001", "aws_s3_bucket.logs")
}

func TestReviewDetectsPublicDatabaseAndUnknownSecurityChange(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "resource_changes": [
	    {
	      "address": "aws_db_instance.app",
	      "mode": "managed",
	      "type": "aws_db_instance",
	      "name": "app",
	      "change": {
	        "actions": ["update"],
	        "before": {"publicly_accessible": false},
	        "after": {"publicly_accessible": true}
	      }
	    },
	    {
	      "address": "aws_security_group_rule.admin",
	      "mode": "managed",
	      "type": "aws_security_group_rule",
	      "name": "admin",
	      "change": {
	        "actions": ["update"],
	        "after_unknown": {"cidr_blocks": true}
	      }
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-AWS-RDS-001", "aws_db_instance.app")
	assertFinding(t, result.Findings, "PW-TF-UNKNOWN-001", "aws_security_group_rule.admin")
}

func TestReviewDetectsNestedUnknownSecurityChange(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "resource_changes": [
	    {
	      "address": "aws_security_group.web",
	      "mode": "managed",
	      "type": "aws_security_group",
	      "name": "web",
	      "change": {
	        "actions": ["update"],
	        "after_unknown": {
	          "ingress": [
	            {
	              "cidr_blocks": true
	            }
	          ]
	        }
	      }
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-UNKNOWN-001", "aws_security_group.web")
}

func TestReviewDetectsKnownPublicSecurityGroupIngress(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "resource_changes": [
	    {
	      "address": "aws_vpc_security_group_ingress_rule.admin",
	      "mode": "managed",
	      "type": "aws_vpc_security_group_ingress_rule",
	      "name": "admin",
	      "change": {
	        "actions": ["create"],
	        "after": {
	          "cidr_ipv4": "0.0.0.0/0",
	          "from_port": 22,
	          "to_port": 22,
	          "ip_protocol": "tcp"
	        }
	      }
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-AWS-SG-001", "aws_vpc_security_group_ingress_rule.admin")
}

func TestReviewDetectsPublicHostCIDRIngress(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "resource_changes": [
	    {
	      "address": "aws_vpc_security_group_ingress_rule.dns",
	      "mode": "managed",
	      "type": "aws_vpc_security_group_ingress_rule",
	      "name": "dns",
	      "change": {
	        "actions": ["create"],
	        "after": {
	          "cidr_ipv4": "8.8.8.8/32",
	          "from_port": 53,
	          "to_port": 53,
	          "ip_protocol": "udp"
	        }
	      }
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-AWS-SG-001", "aws_vpc_security_group_ingress_rule.dns")
}

func TestReviewDoesNotFlagPrivateCIDRIngressAsPublic(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "resource_changes": [
	    {
	      "address": "aws_security_group.private",
	      "mode": "managed",
	      "type": "aws_security_group",
	      "name": "private",
	      "change": {
	        "actions": ["update"],
	        "after": {
	          "ingress": [
	            {
	              "cidr_blocks": ["10.0.0.0/8"],
	              "from_port": 5432,
	              "to_port": 5432,
	              "protocol": "tcp"
	            }
	          ]
	        }
	      }
	    }
	  ]
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertNoFinding(t, result.Findings, "PW-TF-AWS-SG-001", "aws_security_group.private")
}

func TestReviewDetectsPlanStatusFindings(t *testing.T) {
	t.Parallel()

	result, err := ReviewBytes([]byte(`{
	  "format_version": "1.0",
	  "errored": true,
	  "complete": false,
	  "resource_changes": []
	}`), "tfplan.json")
	if err != nil {
		t.Fatalf("ReviewBytes() error = %v", err)
	}
	assertFinding(t, result.Findings, "PW-TF-PLAN-001", "tfplan.json")
	assertFinding(t, result.Findings, "PW-TF-PLAN-002", "tfplan.json")
}

func TestReviewRejectsUnsupportedMajorFormatVersion(t *testing.T) {
	t.Parallel()

	_, err := ReviewBytes([]byte(`{"format_version":"2.0","resource_changes":[]}`), "tfplan.json")
	if err == nil || !strings.Contains(err.Error(), "unsupported Terraform JSON format major version") {
		t.Fatalf("ReviewBytes() error = %v, want unsupported major version", err)
	}
}

func TestReviewFileRejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "plan.json")
	if err := os.WriteFile(target, []byte(`{"format_version":"1.0","resource_changes":[]}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := ReviewFile(link)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("ReviewFile() error = %v, want symlink refusal", err)
	}
}

func TestReviewFileRejectsOversizedPlanJSON(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "tfplan.json")
	data := bytes.Repeat([]byte("x"), maxPlanJSONBytes+1)
	if err := os.WriteFile(target, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReviewFile(target)
	if err == nil || !strings.Contains(err.Error(), "Terraform plan JSON exceeds") {
		t.Fatalf("ReviewFile() error = %v, want size refusal", err)
	}
}

func TestReviewBytesRejectsTrailingContent(t *testing.T) {
	t.Parallel()

	_, err := ReviewBytes([]byte(`{"format_version":"1.0"} {"format_version":"1.0"}`), "tfplan.json")
	if err == nil || !strings.Contains(err.Error(), "trailing content") {
		t.Fatalf("ReviewBytes() error = %v, want trailing content refusal", err)
	}
}

func assertFinding(t *testing.T, findings []Finding, ruleID string, address string) {
	t.Helper()

	for _, finding := range findings {
		if finding.RuleID == ruleID && finding.Address == address {
			return
		}
	}
	t.Fatalf("findings = %#v, want %s for %s", findings, ruleID, address)
}

func assertNoFinding(t *testing.T, findings []Finding, ruleID string, address string) {
	t.Helper()

	for _, finding := range findings {
		if finding.RuleID == ruleID && finding.Address == address {
			t.Fatalf("findings = %#v, did not want %s for %s", findings, ruleID, address)
		}
	}
}
