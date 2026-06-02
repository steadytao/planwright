// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraformplan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"slices"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/localfs"
)

const maxPlanJSONBytes = 25 * 1024 * 1024

var nonPublicCIDRPrefixes = mustParsePrefixes(
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"::ffff:0:0/96",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"100::/64",
	"2001::/23",
	"2001:2::/48",
	"2001:db8::/32",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
)

type Result struct {
	Source           string
	FormatVersion    string
	TerraformVersion string
	Applyable        bool
	Complete         bool
	Errored          bool
	ChangeCount      int
	Findings         []Finding
}

type Finding struct {
	RuleID       string
	Severity     string
	Address      string
	ResourceType string
	Actions      []string
	Message      string
	Why          string
	Fix          string
}

type planFile struct {
	FormatVersion    string           `json:"format_version"`
	TerraformVersion string           `json:"terraform_version"`
	Applyable        bool             `json:"applyable"`
	Complete         bool             `json:"complete"`
	Errored          bool             `json:"errored"`
	ResourceChanges  []resourceChange `json:"resource_changes"`
}

type resourceChange struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Change  change `json:"change"`
}

type change struct {
	Actions      []string       `json:"actions"`
	Before       map[string]any `json:"before"`
	After        map[string]any `json:"after"`
	AfterUnknown map[string]any `json:"after_unknown"`
}

func ReviewFile(path string) (Result, error) {
	data, err := readPlan(path)
	if err != nil {
		return Result{}, err
	}
	return ReviewBytes(data, path)
}

func ReviewBytes(data []byte, sourceName string) (Result, error) {
	var plan planFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&plan); err != nil {
		return Result{}, fmt.Errorf("parse %s: %w", sourceName, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Result{}, fmt.Errorf("parse %s: Terraform plan JSON contains trailing content", sourceName)
	}
	if err := validateFormatVersion(plan.FormatVersion); err != nil {
		return Result{}, err
	}

	result := Result{
		Source:           sourceName,
		FormatVersion:    plan.FormatVersion,
		TerraformVersion: plan.TerraformVersion,
		Applyable:        plan.Applyable,
		Complete:         plan.Complete,
		Errored:          plan.Errored,
		ChangeCount:      len(plan.ResourceChanges),
	}
	if plan.Errored {
		result.Findings = append(result.Findings, Finding{
			RuleID:   "PW-TF-PLAN-001",
			Severity: "medium",
			Address:  sourceName,
			Message:  "Terraform reported that planning errored.",
			Why:      "An errored plan cannot be treated as an applyable review artefact.",
			Fix:      "Inspect the Terraform plan error output and regenerate plan JSON after fixing the configuration.",
		})
	}
	if !plan.Complete {
		result.Findings = append(result.Findings, Finding{
			RuleID:   "PW-TF-PLAN-002",
			Severity: "info",
			Address:  sourceName,
			Message:  "Terraform reported that the plan is incomplete.",
			Why:      "The planned state may require another plan/apply round before it converges.",
			Fix:      "Review whether an incomplete plan is expected before using it for approval.",
		})
	}

	for _, resource := range plan.ResourceChanges {
		result.Findings = append(result.Findings, reviewChange(resource)...)
	}
	sortFindings(result.Findings)
	return result, nil
}

func readPlan(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("read Terraform plan JSON: path must not be empty")
	}
	data, err := localfs.ReadNamedRegularFile(path, maxPlanJSONBytes, "Terraform plan JSON")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func validateFormatVersion(version string) error {
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("terraform JSON format_version is required")
	}
	major, _, _ := strings.Cut(version, ".")
	if major != "1" {
		return fmt.Errorf("unsupported Terraform JSON format major version %q", version)
	}
	return nil
}

func reviewChange(resource resourceChange) []Finding {
	var findings []Finding
	actions := append([]string(nil), resource.Change.Actions...)
	address := strings.TrimSpace(resource.Address)
	if address == "" {
		address = resource.Type + "." + resource.Name
	}

	if hasOnlyAction(actions, "delete") {
		findings = append(findings, Finding{
			RuleID:       "PW-TF-CHANGE-001",
			Severity:     "high",
			Address:      address,
			ResourceType: resource.Type,
			Actions:      actions,
			Message:      "Resource is planned for deletion.",
			Why:          "Deletion can remove data, networking controls or service capacity.",
			Fix:          "Review whether deletion is intended and confirm backups, retention and rollback before applying.",
		})
	}
	if hasAction(actions, "delete") && hasAction(actions, "create") {
		findings = append(findings, Finding{
			RuleID:       "PW-TF-CHANGE-002",
			Severity:     "high",
			Address:      address,
			ResourceType: resource.Type,
			Actions:      actions,
			Message:      "Resource is planned for replacement.",
			Why:          "Replacement includes deletion and creation, which can cause downtime or data loss.",
			Fix:          "Review replacement paths and provider notes before applying.",
		})
	}
	if resource.Type == "aws_db_instance" && boolValue(resource.Change.After, "publicly_accessible") {
		findings = append(findings, Finding{
			RuleID:       "PW-TF-AWS-RDS-001",
			Severity:     "high",
			Address:      address,
			ResourceType: resource.Type,
			Actions:      actions,
			Message:      "RDS instance is planned to be publicly accessible.",
			Why:          "A public database increases exposure and usually requires separate network, authentication and monitoring controls.",
			Fix:          "Set publicly_accessible to false unless this exposure is deliberate and separately controlled.",
		})
	}
	if hasKnownPublicIngress(resource.Type, resource.Change.After) {
		findings = append(findings, Finding{
			RuleID:       "PW-TF-AWS-SG-001",
			Severity:     "high",
			Address:      address,
			ResourceType: resource.Type,
			Actions:      actions,
			Message:      "Security group ingress is planned from a public CIDR range.",
			Why:          "Public ingress changes the external attack surface and can expose administrative or service ports.",
			Fix:          "Restrict the CIDR range or document why the public ingress is deliberate and separately controlled.",
		})
	}
	if hasUnknownSecurityAttribute(resource.Type, resource.Change.AfterUnknown) {
		findings = append(findings, Finding{
			RuleID:       "PW-TF-UNKNOWN-001",
			Severity:     "medium",
			Address:      address,
			ResourceType: resource.Type,
			Actions:      actions,
			Message:      "Security-sensitive planned values are unknown until apply.",
			Why:          "Unknown network or exposure values can hide a security-relevant change during review.",
			Fix:          "Review Terraform output and provider behaviour before applying or make the value explicit where practical.",
		})
	}
	return findings
}

func hasOnlyAction(actions []string, action string) bool {
	return len(actions) == 1 && actions[0] == action
}

func hasAction(actions []string, action string) bool {
	return slices.Contains(actions, action)
}

func boolValue(values map[string]any, key string) bool {
	value, ok := values[key]
	if !ok {
		return false
	}
	boolValue, ok := value.(bool)
	return ok && boolValue
}

func hasUnknownSecurityAttribute(resourceType string, unknown map[string]any) bool {
	if !securitySensitiveResource(resourceType) {
		return false
	}
	for _, key := range []string{"cidr_blocks", "ipv6_cidr_blocks", "cidr_ipv4", "cidr_ipv6", "from_port", "to_port", "publicly_accessible", "ingress", "egress"} {
		if unknownValue(unknown, key) {
			return true
		}
	}
	return false
}

func securitySensitiveResource(resourceType string) bool {
	switch resourceType {
	case "aws_db_instance", "aws_security_group", "aws_security_group_rule", "aws_vpc_security_group_ingress_rule", "aws_vpc_security_group_egress_rule":
		return true
	default:
		return false
	}
}

func unknownValue(values map[string]any, key string) bool {
	value, ok := values[key]
	if !ok {
		return false
	}
	return anyUnknown(value)
}

func anyUnknown(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case []any:
		return slices.ContainsFunc(typed, anyUnknown)
	case map[string]any:
		for _, child := range typed {
			if anyUnknown(child) {
				return true
			}
		}
	}
	return false
}

func hasKnownPublicIngress(resourceType string, values map[string]any) bool {
	switch resourceType {
	case "aws_security_group_rule":
		return stringValue(values, "type") == "ingress" && publicCIDRValues(values)
	case "aws_vpc_security_group_ingress_rule":
		return publicCIDRValues(values)
	case "aws_security_group":
		return listHasPublicIngress(values["ingress"])
	default:
		return false
	}
}

func listHasPublicIngress(value any) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		fields, ok := item.(map[string]any)
		if ok && publicCIDRValues(fields) {
			return true
		}
	}
	return false
}

func publicCIDRValues(values map[string]any) bool {
	if publicCIDR(stringValue(values, "cidr_ipv4")) || publicCIDR(stringValue(values, "cidr_ipv6")) {
		return true
	}
	return stringListHasPublicCIDR(values["cidr_blocks"]) || stringListHasPublicCIDR(values["ipv6_cidr_blocks"])
}

func stringValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok {
		return ""
	}
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(stringValue)
}

func stringListHasPublicCIDR(value any) bool {
	items, ok := value.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		stringItem, ok := item.(string)
		if ok && publicCIDR(stringItem) {
			return true
		}
	}
	return false
}

func publicCIDR(value string) bool {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	prefix = prefix.Masked()
	if !prefix.Addr().IsValid() {
		return false
	}
	for _, nonPublic := range nonPublicCIDRPrefixes {
		if prefixWithin(prefix, nonPublic) {
			return false
		}
	}
	return prefix.Addr().Is4() || prefix.Addr().Is6()
}

func prefixWithin(child netip.Prefix, parent netip.Prefix) bool {
	return parent.Overlaps(child) && parent.Contains(child.Addr()) && child.Bits() >= parent.Bits()
}

func mustParsePrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(fmt.Sprintf("parse non-public CIDR prefix %q: %v", value, err))
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity == findings[j].Severity {
			if findings[i].RuleID == findings[j].RuleID {
				return findings[i].Address < findings[j].Address
			}
			return findings[i].RuleID < findings[j].RuleID
		}
		return severityRank(findings[i].Severity) < severityRank(findings[j].Severity)
	})
}

func severityRank(severity string) int {
	switch severity {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	case "info":
		return 3
	default:
		return 4
	}
}
