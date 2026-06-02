// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import "strings"

var allowedEdgeKinds = map[string]struct{}{
	"network.allow":   {},
	"network.route":   {},
	"network.deny":    {},
	"iam.allow":       {},
	"iam.assume_role": {},
	"iam.pass_role":   {},
	"data.reads_from": {},
	"data.writes_to":  {},
	"publishes_to":    {},
	"subscribes_to":   {},
	"logs_to":         {},
	"emits_metric_to": {},
	"depends_on":      {},
	"runs_as":         {},
	"exposes":         {},
	"protects":        {},
	"encrypts_with":   {},
	"backs_up_to":     {},
	"managed_by":      {},
	"generated_from":  {},
}

func IsAllowedEdgeKind(kind string) bool {
	_, ok := allowedEdgeKinds[strings.TrimSpace(kind)]
	return ok
}

func IsDatabaseNode(node Node) bool {
	return strings.HasPrefix(node.Kind, "aws.rds.")
}

func BoolProperty(properties map[string]any, key string) (bool, bool) {
	if properties == nil {
		return false, false
	}
	value, ok := properties[key]
	if !ok {
		return false, false
	}
	boolValue, ok := value.(bool)
	return boolValue, ok
}

func HasBoolProperty(properties map[string]any, key string) bool {
	value, ok := BoolProperty(properties, key)
	return ok && value
}

func InternetNodeIDs(g Graph) map[string]struct{} {
	ids := map[string]struct{}{}
	for _, node := range g.Nodes {
		if node.Kind == "external.internet" {
			ids[node.ID] = struct{}{}
		}
	}
	return ids
}

func IsInternetFacingNetworkAllow(g Graph, edge Edge) bool {
	if edge.Kind != "network.allow" {
		return false
	}
	_, ok := InternetNodeIDs(g)[edge.From]
	return ok
}

func IsAdministrativePort(port int) bool {
	return port == 22 || port == 3389
}
