// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/policy"
)

func RenderPolicy(result policy.Result) string {
	var builder strings.Builder
	builder.WriteString("# Policy Profile Review\n\n")
	builder.WriteString("Planwright policy profiles are local static review checks. They do not certify compliance, prove deployability or contact cloud APIs.\n\n")
	builder.WriteString("## Summary\n\n")
	if result.Source != "" {
		fmt.Fprintf(&builder, "- Source: %s\n", markdownCode(result.Source))
	}
	fmt.Fprintf(&builder, "- Profile: %s\n", markdownCode(result.Profile.ID))
	fmt.Fprintf(&builder, "- Profile name: %s\n", markdownCode(result.Profile.Name))
	fmt.Fprintf(&builder, "- Rules: %d\n", len(result.Profile.Rules))
	fmt.Fprintf(&builder, "- Findings: %d\n", len(result.Findings))
	fmt.Fprintf(&builder, "- Blocking findings: `%t`\n\n", policy.HasBlockingFindings(result.Findings))

	builder.WriteString("## Findings\n\n")
	if len(result.Findings) == 0 {
		builder.WriteString("- PASS: No built-in policy profile findings were detected.\n\n")
	} else {
		for _, finding := range result.Findings {
			fmt.Fprintf(&builder, "### %s %s\n\n", strings.ToUpper(finding.Severity), finding.RuleID)
			fmt.Fprintf(&builder, "- Resource: %s\n", markdownCode(finding.Resource))
			fmt.Fprintf(&builder, "- Problem: %s\n", markdownText(finding.Message))
			if finding.Why != "" {
				fmt.Fprintf(&builder, "- Why it matters: %s\n", markdownText(finding.Why))
			}
			if finding.Fix != "" {
				fmt.Fprintf(&builder, "- Fix: %s\n", markdownText(finding.Fix))
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("## Rules\n\n")
	for _, rule := range result.Profile.Rules {
		fmt.Fprintf(&builder, "- %s (%s): %s\n", markdownCode(rule.ID), markdownText(rule.Severity), markdownText(rule.Description))
	}
	builder.WriteString("\n")
	builder.WriteString("## Review Notes\n\n")
	builder.WriteString("- Policy profiles are built into Planwright; custom policy packs are not implemented.\n")
	builder.WriteString("- A quiet policy report does not prove the graph is complete, deployable or compliant.\n")
	builder.WriteString("- Findings should be reviewed with the source plan, generated artefacts and operational context.\n")
	return builder.String()
}

func RenderPolicySARIF(result policy.Result) ([]byte, error) {
	sarif := sarifLog{
		Schema:  "https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/schemas/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "Planwright",
						InformationURI: "https://github.com/steadytao/planwright",
						Rules:          policySARIFRules(result),
					},
				},
				Results: policySARIFResults(result),
			},
		},
	}
	data, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func policySARIFRules(result policy.Result) []sarifRule {
	rules := make([]policy.Rule, 0, len(result.Profile.Rules))
	rules = append(rules, result.Profile.Rules...)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})
	output := make([]sarifRule, 0, len(rules))
	for _, rule := range rules {
		output = append(output, sarifRule{
			ID:               rule.ID,
			Name:             rule.ID,
			ShortDescription: sarifMessage{Text: rule.Description},
			FullDescription:  sarifMessage{Text: rule.Description},
			Help: sarifMarkdownHelp{
				Text:     rule.Description,
				Markdown: rule.Description,
			},
			Properties: sarifRuleProperty{Tags: []string{"policy", "infrastructure", result.Profile.ID}},
		})
	}
	return output
}

func policySARIFResults(result policy.Result) []sarifResult {
	source := result.Source
	if source == "" {
		source = "planwright.graph.json"
	}
	output := make([]sarifResult, 0, len(result.Findings))
	for _, finding := range result.Findings {
		output = append(output, sarifResult{
			RuleID:  finding.RuleID,
			Level:   policySARIFLevel(finding.Severity),
			Message: sarifMessage{Text: fmt.Sprintf("%s: %s", finding.Resource, finding.Message)},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: sarifArtifactURI(source)},
						Region:           sarifRegion{StartLine: 1},
					},
				},
			},
		})
	}
	return output
}

func policySARIFLevel(severity string) string {
	switch severity {
	case policy.SeverityError:
		return "error"
	case policy.SeverityWarn:
		return "warning"
	default:
		return "note"
	}
}
