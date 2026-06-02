// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"encoding/json"
	"fmt"
	"sort"
)

type Diff struct {
	AddedNodes   []NodeChange
	RemovedNodes []NodeChange
	ChangedNodes []NodeChange
	AddedEdges   []EdgeChange
	RemovedEdges []EdgeChange
	Findings     []DiffFinding
}

type NodeChange struct {
	ID      string
	Kind    string
	OldKind string
	NewKind string
}

type EdgeChange struct {
	Edge Edge
}

type DiffFinding struct {
	Severity string
	RuleID   string
	Resource string
	Message  string
	Why      string
	Fix      string
}

func Compare(oldGraph Graph, newGraph Graph) Diff {
	oldNodes := indexNodes(oldGraph.Nodes)
	newNodes := indexNodes(newGraph.Nodes)
	oldEdges := indexEdges(oldGraph.Edges)
	newEdges := indexEdges(newGraph.Edges)

	var diff Diff
	for id, node := range newNodes {
		oldNode, exists := oldNodes[id]
		if !exists {
			diff.AddedNodes = append(diff.AddedNodes, NodeChange{ID: id, Kind: node.Kind, NewKind: node.Kind})
			if IsDatabaseNode(node) && HasBoolProperty(node.Properties, "publicly_accessible") {
				diff.Findings = append(diff.Findings, publicDatabaseFinding(id))
			}
			continue
		}
		if node.Kind != oldNode.Kind || canonicalProperties(node.Properties) != canonicalProperties(oldNode.Properties) {
			diff.ChangedNodes = append(diff.ChangedNodes, NodeChange{
				ID:      id,
				Kind:    node.Kind,
				OldKind: oldNode.Kind,
				NewKind: node.Kind,
			})
		}
		if IsDatabaseNode(node) && !HasBoolProperty(oldNode.Properties, "publicly_accessible") && HasBoolProperty(node.Properties, "publicly_accessible") {
			diff.Findings = append(diff.Findings, publicDatabaseFinding(id))
		}
	}
	for id, node := range oldNodes {
		if _, exists := newNodes[id]; !exists {
			diff.RemovedNodes = append(diff.RemovedNodes, NodeChange{ID: id, Kind: node.Kind, OldKind: node.Kind})
		}
	}

	for key, edge := range newEdges {
		if _, exists := oldEdges[key]; !exists {
			diff.AddedEdges = append(diff.AddedEdges, EdgeChange{Edge: edge})
			if finding, ok := addedEdgeFinding(newGraph, edge); ok {
				diff.Findings = append(diff.Findings, finding)
			}
		}
	}
	for key, edge := range oldEdges {
		if _, exists := newEdges[key]; !exists {
			diff.RemovedEdges = append(diff.RemovedEdges, EdgeChange{Edge: edge})
		}
	}

	sortNodeChanges(diff.AddedNodes)
	sortNodeChanges(diff.RemovedNodes)
	sortNodeChanges(diff.ChangedNodes)
	sortEdgeChanges(diff.AddedEdges)
	sortEdgeChanges(diff.RemovedEdges)
	sortFindings(diff.Findings)
	return diff
}

func indexNodes(nodes []Node) map[string]Node {
	indexed := make(map[string]Node, len(nodes))
	for _, node := range nodes {
		indexed[node.ID] = node
	}
	return indexed
}

func indexEdges(edges []Edge) map[string]Edge {
	indexed := make(map[string]Edge, len(edges))
	for _, edge := range edges {
		indexed[edgeKey(edge)] = edge
	}
	return indexed
}

func edgeKey(edge Edge) string {
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d\x00%s", edge.From, edge.To, edge.Kind, edge.Protocol, edge.Port, edge.Intent)
}

func canonicalProperties(properties map[string]any) string {
	if len(properties) == 0 {
		return ""
	}
	data, err := json.Marshal(properties)
	if err != nil {
		return fmt.Sprintf("%#v", properties)
	}
	return string(data)
}

func publicDatabaseFinding(resource string) DiffFinding {
	return DiffFinding{
		Severity: "high",
		RuleID:   "PW-DIFF-RDS-001",
		Resource: resource,
		Message:  "A database is newly public or became public.",
		Why:      "Public database exposure materially changes the infrastructure risk profile and usually requires explicit review.",
		Fix:      "Set publicly_accessible to false or document the exposure with compensating controls before deployment.",
	}
}

func addedEdgeFinding(g Graph, edge Edge) (DiffFinding, bool) {
	if !IsInternetFacingNetworkAllow(g, edge) {
		return DiffFinding{}, false
	}
	resource := edge.From + " -> " + edge.To
	if IsAdministrativePort(edge.Port) {
		return DiffFinding{
			Severity: "high",
			RuleID:   "PW-DIFF-NET-001",
			Resource: resource,
			Message:  "An internet-facing administrative network path was added.",
			Why:      "SSH or RDP exposure materially changes review risk and is commonly abused when left open.",
			Fix:      "Remove the public administrative edge or restrict access to a deliberate private access path.",
		}, true
	}
	return DiffFinding{
		Severity: "medium",
		RuleID:   "PW-DIFF-NET-002",
		Resource: resource,
		Message:  "A new internet-facing network path was added.",
		Why:      "New public ingress changes the external attack surface and should be reviewed deliberately.",
		Fix:      "Confirm the public entrypoint is intended and has matching TLS, logging and ownership controls.",
	}, true
}

func sortNodeChanges(changes []NodeChange) {
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].ID < changes[j].ID
	})
}

func sortEdgeChanges(changes []EdgeChange) {
	sort.Slice(changes, func(i, j int) bool {
		return edgeKey(changes[i].Edge) < edgeKey(changes[j].Edge)
	})
}

func sortFindings(findings []DiffFinding) {
	sort.Slice(findings, func(i, j int) bool {
		left := findings[i].Severity + "\x00" + findings[i].RuleID + "\x00" + findings[i].Resource
		right := findings[j].Severity + "\x00" + findings[j].RuleID + "\x00" + findings[j].Resource
		return left < right
	})
}
