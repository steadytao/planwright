// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/graph"
)

const (
	SeverityError = "error"
	SeverityWarn  = "warn"
	SeverityInfo  = "info"
)

type Profile struct {
	ID          string
	Name        string
	Description string
	Rules       []Rule
}

type Rule struct {
	ID          string
	Severity    string
	Description string
}

type Result struct {
	Source   string
	Profile  Profile
	Graph    graph.Graph
	Findings []Finding
}

type Finding struct {
	RuleID   string
	Severity string
	Resource string
	Message  string
	Why      string
	Fix      string
}

var builtInProfiles = []Profile{
	{
		ID:          "lab",
		Name:        "Lab",
		Description: "Local review checks for low-cost learning and test environments.",
		Rules: []Rule{
			{ID: "PW-POL-NET-001", Severity: SeverityError, Description: "Databases must not be public."},
			{ID: "PW-POL-NET-002", Severity: SeverityError, Description: "SSH and RDP must not be internet-facing."},
		},
	},
	{
		ID:          "small-business",
		Name:        "Small Business",
		Description: "Baseline review checks for small business infrastructure planning.",
		Rules: []Rule{
			{ID: "PW-POL-NET-001", Severity: SeverityError, Description: "Databases must not be public."},
			{ID: "PW-POL-NET-002", Severity: SeverityError, Description: "SSH and RDP must not be internet-facing."},
			{ID: "PW-POL-BACKUP-001", Severity: SeverityWarn, Description: "Databases should include backup evidence."},
		},
	},
	{
		ID:          "production",
		Name:        "Production",
		Description: "Stricter local review checks for production-profile planning.",
		Rules: []Rule{
			{ID: "PW-POL-NET-001", Severity: SeverityError, Description: "Databases must not be public."},
			{ID: "PW-POL-NET-002", Severity: SeverityError, Description: "SSH and RDP must not be internet-facing."},
			{ID: "PW-POL-BACKUP-001", Severity: SeverityError, Description: "Databases must include backup evidence."},
			{ID: "PW-POL-OBS-001", Severity: SeverityWarn, Description: "Production graphs should include observability evidence."},
			{ID: "PW-POL-META-001", Severity: SeverityWarn, Description: "Production policy should not be applied silently to lab-profile graphs."},
		},
	},
}

func Profiles() []Profile {
	profiles := make([]Profile, 0, len(builtInProfiles))
	for _, profile := range builtInProfiles {
		profiles = append(profiles, cloneProfile(profile))
	}
	return profiles
}

func Evaluate(g graph.Graph, profileID string) (Result, error) {
	profile, ok := profileByID(profileID)
	if !ok {
		return Result{}, fmt.Errorf("unknown policy profile %q", profileID)
	}

	result := Result{
		Profile: profile,
		Graph:   g,
	}
	rules := ruleIndex(profile.Rules)

	for _, node := range g.Nodes {
		if graph.IsDatabaseNode(node) && graph.HasBoolProperty(node.Properties, "publicly_accessible") {
			result.Findings = append(result.Findings, findingForRule(rules, "PW-POL-NET-001", node.ID,
				"Database is marked publicly accessible.",
				"Public database exposure materially changes the infrastructure risk profile and usually requires explicit review.",
				"Set publicly_accessible to false or document the exposure with compensating controls before deployment."))
		}
		if graph.IsDatabaseNode(node) && hasRule(rules, "PW-POL-BACKUP-001") && !hasBackupEvidence(node.Properties) {
			result.Findings = append(result.Findings, findingForRule(rules, "PW-POL-BACKUP-001", node.ID,
				"Database backup evidence is not present in the graph.",
				"Database recovery expectations should be reviewed before production-like use.",
				"Add explicit backup evidence to the graph or document why backup handling is outside this plan."))
		}
	}

	for _, edge := range g.Edges {
		if hasRule(rules, "PW-POL-NET-002") && graph.IsInternetFacingNetworkAllow(g, edge) && graph.IsAdministrativePort(edge.Port) {
			result.Findings = append(result.Findings, findingForRule(rules, "PW-POL-NET-002", edge.From+" -> "+edge.To,
				"Administrative network access is internet-facing.",
				"SSH and RDP exposure is commonly abused and should not be treated as a default access path.",
				"Remove the public administrative edge or restrict access to a deliberate private access path."))
		}
	}

	if hasRule(rules, "PW-POL-OBS-001") && !hasObservabilityEvidence(g) {
		result.Findings = append(result.Findings, findingForRule(rules, "PW-POL-OBS-001", "graph",
			"Observability evidence is not present in the graph.",
			"Production review usually needs logging, metrics or tracing ownership evidence.",
			"Add observability nodes or logging edges or document why observability is handled outside this graph."))
	}
	if hasRule(rules, "PW-POL-META-001") && strings.EqualFold(strings.TrimSpace(g.Profile), "lab") {
		result.Findings = append(result.Findings, findingForRule(rules, "PW-POL-META-001", "graph.profile",
			"Production policy is being applied to a lab-profile graph.",
			"Lab defaults may deliberately omit production controls such as backup and observability detail.",
			"Change the graph profile or treat the production policy output as a gap report."))
	}

	sortFindings(result.Findings)
	return result, nil
}

func HasBlockingFindings(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Severity == SeverityError {
			return true
		}
	}
	return false
}

func cloneProfile(profile Profile) Profile {
	profile.Rules = append([]Rule(nil), profile.Rules...)
	return profile
}

func profileByID(id string) (Profile, bool) {
	normalised := strings.TrimSpace(id)
	for _, profile := range builtInProfiles {
		if profile.ID == normalised {
			return cloneProfile(profile), true
		}
	}
	return Profile{}, false
}

func ruleIndex(rules []Rule) map[string]Rule {
	index := make(map[string]Rule, len(rules))
	for _, rule := range rules {
		index[rule.ID] = rule
	}
	return index
}

func hasRule(rules map[string]Rule, id string) bool {
	_, ok := rules[id]
	return ok
}

func findingForRule(rules map[string]Rule, ruleID string, resource string, message string, why string, fix string) Finding {
	rule := rules[ruleID]
	return Finding{
		RuleID:   ruleID,
		Severity: rule.Severity,
		Resource: resource,
		Message:  message,
		Why:      why,
		Fix:      fix,
	}
}

func hasBackupEvidence(properties map[string]any) bool {
	if graph.HasBoolProperty(properties, "backup_enabled") {
		return true
	}
	if intProperty(properties, "backup_retention_days") > 0 {
		return true
	}
	return false
}

func hasObservabilityEvidence(g graph.Graph) bool {
	for _, node := range g.Nodes {
		if strings.HasPrefix(node.Kind, "aws.cloudwatch.") || strings.HasPrefix(node.Kind, "prometheus.") || strings.HasPrefix(node.Kind, "grafana.") || strings.HasPrefix(node.Kind, "otel.") {
			return true
		}
	}
	for _, edge := range g.Edges {
		if edge.Kind == "logs_to" || edge.Kind == "emits_metric_to" {
			return true
		}
	}
	return false
}

func intProperty(properties map[string]any, key string) int {
	if properties == nil {
		return 0
	}
	value, ok := properties[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		if typed == float64(int(typed)) {
			return int(typed)
		}
	}
	return 0
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		left := severityRank(findings[i].Severity)
		right := severityRank(findings[j].Severity)
		if left != right {
			return left < right
		}
		if findings[i].RuleID != findings[j].RuleID {
			return findings[i].RuleID < findings[j].RuleID
		}
		return findings[i].Resource < findings[j].Resource
	})
}

func severityRank(severity string) int {
	switch severity {
	case SeverityError:
		return 0
	case SeverityWarn:
		return 1
	case SeverityInfo:
		return 2
	default:
		return 3
	}
}
