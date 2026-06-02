// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/policy"
)

func TestRenderPolicyReport(t *testing.T) {
	t.Parallel()

	result, err := policy.Evaluate(loadExampleGraph(t), "production")
	if err != nil {
		t.Fatalf("policy.Evaluate() error = %v", err)
	}
	report := RenderPolicy(result)
	for _, want := range []string{
		"# Policy Profile Review",
		"Profile: `production`",
		"PW-POL-BACKUP-001",
		"Planwright policy profiles are local static review checks",
	} {
		if !strings.Contains(report, want) {
			t.Fatalf("policy report = %s, want %q", report, want)
		}
	}
}

func TestRenderPolicyReportEscapesMarkdownFields(t *testing.T) {
	t.Parallel()

	result := policy.Result{
		Source: "graph.json\n- forged",
		Profile: policy.Profile{
			ID:    "lab`",
			Name:  "Lab",
			Rules: []policy.Rule{{ID: "PW-POL-NET-002", Severity: policy.SeverityError, Description: "SSH and RDP must not be internet-facing."}},
		},
		Findings: []policy.Finding{{
			RuleID:   "PW-POL-NET-002",
			Severity: policy.SeverityError,
			Resource: "world`\n- forged -> app",
			Message:  "Administrative network access is internet-facing.",
		}},
	}

	report := RenderPolicy(result)
	if strings.Contains(report, "\n- forged") {
		t.Fatalf("policy report contains forged Markdown line: %s", report)
	}
	if !strings.Contains(report, "`` ") {
		t.Fatalf("policy report = %s, want expanded code span for backtick content", report)
	}
}

func TestRenderPolicySARIF(t *testing.T) {
	t.Parallel()

	result, err := policy.Evaluate(loadExampleGraph(t), "production")
	if err != nil {
		t.Fatalf("policy.Evaluate() error = %v", err)
	}
	data, err := RenderPolicySARIF(result)
	if err != nil {
		t.Fatalf("RenderPolicySARIF() error = %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("policy SARIF is not valid JSON: %v", err)
	}
	if got, want := decoded["version"], "2.1.0"; got != want {
		t.Fatalf("SARIF version = %v, want %q", got, want)
	}
	if !strings.Contains(string(data), "PW-POL-BACKUP-001") {
		t.Fatalf("policy SARIF = %s, want backup rule", data)
	}
}
