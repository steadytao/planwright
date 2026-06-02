// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/review/terraformplan"
)

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri,omitempty"`
	Rules          []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	ShortDescription sarifMessage      `json:"shortDescription"`
	FullDescription  sarifMessage      `json:"fullDescription"`
	Help             sarifMarkdownHelp `json:"help"`
	Properties       sarifRuleProperty `json:"properties"`
}

type sarifRuleProperty struct {
	Tags []string `json:"tags,omitempty"`
}

type sarifMarkdownHelp struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine,omitempty"`
}

func RenderTerraformReview(result terraformplan.Result) string {
	var builder strings.Builder
	builder.WriteString("# Terraform Plan Review\n\n")
	builder.WriteString("Planwright reviews Terraform plan JSON as a local static artefact. It does not run Terraform, apply a plan, inspect provider schemas or contact cloud APIs.\n\n")
	builder.WriteString("## Summary\n\n")
	fmt.Fprintf(&builder, "- Source: %s\n", markdownCode(result.Source))
	fmt.Fprintf(&builder, "- Terraform JSON format: %s\n", markdownCode(emptyAsUnknown(result.FormatVersion)))
	if result.TerraformVersion != "" {
		fmt.Fprintf(&builder, "- Terraform version: %s\n", markdownCode(result.TerraformVersion))
	}
	fmt.Fprintf(&builder, "- Resource changes: %d\n", result.ChangeCount)
	fmt.Fprintf(&builder, "- Findings: %d\n", len(result.Findings))
	fmt.Fprintf(&builder, "- Applyable flag: `%t`\n", result.Applyable)
	fmt.Fprintf(&builder, "- Complete flag: `%t`\n", result.Complete)
	fmt.Fprintf(&builder, "- Errored flag: `%t`\n\n", result.Errored)

	builder.WriteString("## Findings\n\n")
	if len(result.Findings) == 0 {
		builder.WriteString("- PASS: No v0.3 Terraform plan review findings were detected.\n")
		return builder.String()
	}
	for _, finding := range result.Findings {
		fmt.Fprintf(&builder, "### %s %s\n\n", strings.ToUpper(finding.Severity), finding.RuleID)
		fmt.Fprintf(&builder, "- Resource: %s\n", markdownCode(finding.Address))
		if finding.ResourceType != "" {
			fmt.Fprintf(&builder, "- Type: %s\n", markdownCode(finding.ResourceType))
		}
		if len(finding.Actions) > 0 {
			fmt.Fprintf(&builder, "- Actions: %s\n", markdownCode(strings.Join(finding.Actions, ", ")))
		}
		fmt.Fprintf(&builder, "- Problem: %s\n", markdownText(finding.Message))
		if finding.Why != "" {
			fmt.Fprintf(&builder, "- Why it matters: %s\n", markdownText(finding.Why))
		}
		if finding.Fix != "" {
			fmt.Fprintf(&builder, "- Fix: %s\n", markdownText(finding.Fix))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func RenderTerraformReviewSARIF(result terraformplan.Result) ([]byte, error) {
	rules := sarifRules(result.Findings)
	sarif := sarifLog{
		Schema:  "https://docs.oasis-open.org/sarif/sarif/v2.1.0/os/schemas/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "Planwright",
						InformationURI: "https://github.com/steadytao/planwright",
						Rules:          rules,
					},
				},
				Results: sarifResults(result),
			},
		},
	}
	data, err := json.MarshalIndent(sarif, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func sarifRules(findings []terraformplan.Finding) []sarifRule {
	byID := map[string]terraformplan.Finding{}
	for _, finding := range findings {
		if _, exists := byID[finding.RuleID]; !exists {
			byID[finding.RuleID] = finding
		}
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rules := make([]sarifRule, 0, len(ids))
	for _, id := range ids {
		finding := byID[id]
		rules = append(rules, sarifRule{
			ID:               id,
			Name:             id,
			ShortDescription: sarifMessage{Text: finding.Message},
			FullDescription:  sarifMessage{Text: finding.Why},
			Help: sarifMarkdownHelp{
				Text:     finding.Fix,
				Markdown: finding.Fix,
			},
			Properties: sarifRuleProperty{Tags: []string{"terraform", "infrastructure"}},
		})
	}
	return rules
}

func sarifResults(result terraformplan.Result) []sarifResult {
	results := make([]sarifResult, 0, len(result.Findings))
	for _, finding := range result.Findings {
		results = append(results, sarifResult{
			RuleID:  finding.RuleID,
			Level:   sarifLevel(finding.Severity),
			Message: sarifMessage{Text: fmt.Sprintf("%s %s: %s", finding.Address, strings.Join(finding.Actions, ","), finding.Message)},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: sarifArtifactURI(result.Source)},
						Region:           sarifRegion{StartLine: 1},
					},
				},
			},
		})
	}
	return results
}

func sarifLevel(severity string) string {
	switch severity {
	case "high", "medium":
		return "error"
	case "low":
		return "warning"
	default:
		return "note"
	}
}

func emptyAsUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
