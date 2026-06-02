// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"fmt"
	"strings"

	"github.com/steadytao/planwright/internal/limits"
)

func Validate(g Graph) []Diagnostic {
	var diagnostics []Diagnostic

	if strings.TrimSpace(g.Version) != Version {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "PW-GRAPH-VERSION-001",
			Resource: "graph",
			Message:  fmt.Sprintf("Graph version must be %s.", Version),
			Fix:      "Set version to planwright.graph.v1.",
		})
	}
	if strings.TrimSpace(g.Provider) == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "PW-GRAPH-PROVIDER-001",
			Resource: "graph",
			Message:  "Graph provider is required.",
			Fix:      "Set provider to the infrastructure provider being modelled.",
		})
	}
	if strings.TrimSpace(g.Region) == "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "PW-GRAPH-REGION-001",
			Resource: "graph",
			Message:  "Graph region is required.",
			Fix:      "Set region to the intended deployment region.",
		})
	}

	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		id := node.ID
		trimmedID := strings.TrimSpace(node.ID)
		if trimmedID == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityError,
				Code:     "PW-GRAPH-NODE-001",
				Resource: "node",
				Message:  "Node ID is required.",
				Fix:      "Give each node a stable ID.",
			})
			continue
		}
		if id != trimmedID {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityError,
				Code:     "PW-GRAPH-NODE-003",
				Resource: trimmedID,
				Message:  "Node ID must not contain leading or trailing whitespace.",
				Fix:      "Use the exact node ID without surrounding whitespace.",
			})
			continue
		}
		if _, exists := nodeIDs[id]; exists {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityError,
				Code:     "PW-GRAPH-NODE-002",
				Resource: id,
				Message:  "Node ID is duplicated.",
				Fix:      "Use one stable ID per node.",
			})
			continue
		}
		nodeIDs[id] = struct{}{}

		if IsDatabaseNode(node) {
			if _, ok := node.Properties["publicly_accessible"]; ok {
				public, valid := BoolProperty(node.Properties, "publicly_accessible")
				if !valid {
					diagnostics = append(diagnostics, Diagnostic{
						Severity: SeverityError,
						Code:     "PW-GRAPH-PROPERTY-001",
						Resource: id + ".publicly_accessible",
						Message:  "Database node property publicly_accessible must be a boolean.",
						Fix:      "Use true or false without quotes.",
					})
				} else if public {
					diagnostics = append(diagnostics, Diagnostic{
						Severity: SeverityWarn,
						Code:     "PW-AWS-RDS-001",
						Resource: id,
						Message:  "RDS node is marked publicly accessible.",
						Fix:      "Set publicly_accessible to false unless the public exposure is deliberate and separately controlled.",
					})
				}
			}
		}
	}

	for _, edge := range g.Edges {
		resource := edge.From + " -> " + edge.To
		fromValid := edgeEndpointHasExactIdentity(edge.From, "from", resource, &diagnostics)
		toValid := edgeEndpointHasExactIdentity(edge.To, "to", resource, &diagnostics)
		if fromValid {
			if _, exists := nodeIDs[edge.From]; !exists {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityError,
					Code:     "PW-GRAPH-EDGE-001",
					Resource: edge.From,
					Message:  fmt.Sprintf("Edge endpoint %q does not match any node.", edge.From),
					Fix:      "Add the missing node or correct the edge endpoint.",
				})
			}
		}
		if toValid {
			if _, exists := nodeIDs[edge.To]; !exists {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityError,
					Code:     "PW-GRAPH-EDGE-001",
					Resource: edge.To,
					Message:  fmt.Sprintf("Edge endpoint %q does not match any node.", edge.To),
					Fix:      "Add the missing node or correct the edge endpoint.",
				})
			}
		}
		if !IsAllowedEdgeKind(edge.Kind) {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityError,
				Code:     "PW-GRAPH-EDGE-002",
				Resource: resource,
				Message:  fmt.Sprintf("Edge kind %q is not supported by the current graph model.", edge.Kind),
				Fix:      "Use a supported edge kind such as network.allow, network.route or depends_on.",
			})
			continue
		}
		if edge.Kind == "network.allow" {
			if edge.Port < limits.MinNetworkPort || edge.Port > limits.MaxNetworkPort {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityError,
					Code:     "PW-GRAPH-NET-001",
					Resource: resource,
					Message:  fmt.Sprintf("Network edge port must be between %d and %d.", limits.MinNetworkPort, limits.MaxNetworkPort),
					Fix:      "Use a valid TCP or UDP port.",
				})
			}
			if edge.Protocol != "tcp" && edge.Protocol != "udp" {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: SeverityError,
					Code:     "PW-GRAPH-NET-002",
					Resource: resource,
					Message:  "Network edge protocol must be tcp or udp.",
					Fix:      "Set protocol to tcp or udp.",
				})
			}
		}
	}

	return diagnostics
}

func edgeEndpointHasExactIdentity(endpoint string, field string, resource string, diagnostics *[]Diagnostic) bool {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		*diagnostics = append(*diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "PW-GRAPH-EDGE-003",
			Resource: resource,
			Message:  fmt.Sprintf("Edge %s endpoint is required.", field),
			Fix:      "Set the endpoint to an exact node ID.",
		})
		return false
	}
	if endpoint != trimmed {
		*diagnostics = append(*diagnostics, Diagnostic{
			Severity: SeverityError,
			Code:     "PW-GRAPH-EDGE-004",
			Resource: resource,
			Message:  fmt.Sprintf("Edge %s endpoint must not contain leading or trailing whitespace.", field),
			Fix:      "Use the exact node ID without surrounding whitespace.",
		})
		return false
	}
	return true
}
