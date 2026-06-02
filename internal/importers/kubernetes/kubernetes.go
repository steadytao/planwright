// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
	"github.com/steadytao/planwright/internal/limits"
	"github.com/steadytao/planwright/internal/localfs"
	"github.com/steadytao/planwright/internal/yamlutil"
)

const (
	maxManifestBytes      = 5 * 1024 * 1024
	maxManifestFiles      = 128
	maxManifestTotalBytes = 32 * 1024 * 1024
	maxManifestResources  = 4096
	maxInferredEdges      = 8192
)

type Result struct {
	Graph       graph.Graph
	Loss        loss.Report
	Diagnostics []graph.Diagnostic
	Sources     []SourceFile
}

type SourceFile struct {
	Path string
	Data []byte
	Size int
}

type resource struct {
	apiVersion      string
	kind            string
	id              string
	name            string
	namespace       string
	source          string
	graphKind       string
	properties      map[string]any
	labels          map[string]string
	templateLabels  map[string]string
	serviceSelector map[string]string
	ingressBackends []backendRef
	routeBackends   []backendRef
	routeParents    []parentRef
}

type backendRef struct {
	namespace string
	name      string
	kind      string
	group     string
	port      int
}

type parentRef struct {
	namespace string
	name      string
	kind      string
	group     string
}

func ImportFile(path string) (Result, error) {
	sources, sourceLabel, err := readSources(path)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Graph: graph.Graph{
			Version:  graph.Version,
			Provider: "kubernetes",
			Region:   "local",
			Nodes:    []graph.Node{},
			Edges:    []graph.Edge{},
		},
		Loss: loss.Report{
			SourceFormat: "kubernetes",
			Source:       sourceLabel,
		},
		Sources: sourceMetadata(sources),
	}

	var resources []resource
	for _, source := range sources {
		parsed, err := parseSource(source)
		if err != nil {
			return Result{}, err
		}
		resources = append(resources, parsed...)
		if len(resources) > maxManifestResources {
			return Result{}, fmt.Errorf("read %s: manifest resources exceed %d", sourceLabel, maxManifestResources)
		}
	}

	resourceByID := map[string]resource{}
	for _, item := range resources {
		if item.graphKind == "" {
			result.Loss.Unsupported = append(result.Loss.Unsupported, loss.Item{
				Resource: resourceName(item),
				Kind:     sourceKind(item),
				Message:  "Resource is preserved in the source manifest but is not lowered into the Planwright graph by the current Kubernetes supported subset; manual review required.",
			})
			continue
		}
		resourceByID[item.id] = item
		result.Graph.Nodes = append(result.Graph.Nodes, graph.Node{
			ID:         item.id,
			Kind:       item.graphKind,
			Name:       item.name,
			Properties: item.properties,
		})
		result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  "Resource inventory was lowered into the Planwright graph.",
		})
		appendSemanticLoss(item, &result.Loss)
	}

	result.Graph.Edges = inferEdges(resources, resourceByID, &result.Loss)
	sortGraph(&result.Graph)
	sortLossReport(&result.Loss)
	result.Diagnostics = graph.Validate(result.Graph)
	return result, nil
}

func sourceMetadata(sources []SourceFile) []SourceFile {
	out := make([]SourceFile, 0, len(sources))
	for _, source := range sources {
		out = append(out, SourceFile{
			Path: source.Path,
			Size: len(source.Data),
		})
	}
	return out
}

func readSources(path string) ([]SourceFile, string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, "", fmt.Errorf("read manifests: path must not be empty")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, "", fmt.Errorf("read %s: symlink manifest paths are not accepted", path)
	}
	if info.Mode().IsRegular() {
		data, err := readRegularFile(path, info)
		if err != nil {
			return nil, "", err
		}
		return []SourceFile{{Path: path, Data: data}}, path, nil
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("read %s: not a regular file or directory", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, "", fmt.Errorf("read directory %s: %w", path, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var sources []SourceFile
	totalBytes := 0
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := os.Lstat(fullPath)
		if err != nil {
			return nil, "", fmt.Errorf("inspect %s: %w", fullPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, "", fmt.Errorf("read %s: symlink manifest paths are not accepted", fullPath)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if !isManifestPath(entry.Name()) {
			continue
		}
		data, err := readRegularFile(fullPath, info)
		if err != nil {
			return nil, "", err
		}
		if len(sources)+1 > maxManifestFiles {
			return nil, "", fmt.Errorf("read %s: too many manifest files; maximum is %d", path, maxManifestFiles)
		}
		totalBytes += len(data)
		if totalBytes > maxManifestTotalBytes {
			return nil, "", fmt.Errorf("read %s: manifest files exceed %d total bytes", path, maxManifestTotalBytes)
		}
		sources = append(sources, SourceFile{Path: fullPath, Data: data})
	}
	if len(sources) == 0 {
		return nil, "", fmt.Errorf("read %s: no direct .yaml, .yml or .json manifest files found", path)
	}
	return sources, path, nil
}

func readRegularFile(path string, info os.FileInfo) ([]byte, error) {
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("read %s: not a regular file", path)
	}
	if info.Size() > maxManifestBytes {
		return nil, fmt.Errorf("read %s: manifest exceeds %d bytes", path, maxManifestBytes)
	}
	data, err := localfs.ReadNamedRegularFile(path, maxManifestBytes, "manifest")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func isManifestPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}

func parseSource(source SourceFile) ([]resource, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(source.Data))
	var resources []resource
	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", source.Path, err)
		}
		document := documentNode(&root)
		if document == nil || document.Kind == 0 {
			continue
		}
		if document.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("parse %s: expected Kubernetes mapping document", source.Path)
		}
		if err := yamlutil.RejectDuplicateMappingKeys(document, source.Path); err != nil {
			return nil, err
		}
		if isKubernetesList(document) {
			items := mappingValue(document, "items")
			if items == nil || items.Kind != yaml.SequenceNode {
				return nil, fmt.Errorf("parse %s: List document must contain an items sequence", source.Path)
			}
			for index, item := range items.Content {
				if item.Kind != yaml.MappingNode {
					return nil, fmt.Errorf("parse %s: List item %d must be a mapping document", source.Path, index)
				}
				resources = append(resources, lowerResource(item, source.Path))
			}
			continue
		}
		resources = append(resources, lowerResource(document, source.Path))
	}
	return resources, nil
}

func isKubernetesList(document *yaml.Node) bool {
	kind, _ := scalarString(mappingValue(document, "kind"))
	return strings.TrimSpace(kind) == "List"
}

func lowerResource(node *yaml.Node, source string) resource {
	apiVersion, _ := scalarString(mappingValue(node, "apiVersion"))
	kind, _ := scalarString(mappingValue(node, "kind"))
	metadata := mappingValue(node, "metadata")
	spec := mappingValue(node, "spec")
	name, _ := scalarString(mappingValue(metadata, "name"))
	namespace, _ := scalarString(mappingValue(metadata, "namespace"))

	item := resource{
		apiVersion: strings.TrimSpace(apiVersion),
		kind:       strings.TrimSpace(kind),
		name:       strings.TrimSpace(name),
		namespace:  normaliseNamespace(kind, namespace),
		source:     source,
		labels:     stringMap(mappingValue(metadata, "labels")),
	}
	item.graphKind = graphKindFor(item.apiVersion, item.kind)
	item.id = resourceID(item)
	item.properties = baseProperties(item)

	switch item.graphKind {
	case "k8s.namespace":
		item.properties["labels"] = item.labels
	case "k8s.deployment", "k8s.statefulset", "k8s.daemonset", "k8s.job", "k8s.cronjob":
		lowerWorkload(spec, &item)
	case "k8s.service":
		lowerService(spec, &item)
	case "k8s.ingress":
		lowerIngress(spec, &item)
	case "k8s.network_policy":
		lowerNetworkPolicy(spec, &item)
	case "k8s.config_map":
		lowerConfigMap(node, &item)
	case "k8s.secret":
		lowerSecret(node, &item)
	case "gatewayapi.gateway":
		lowerGateway(spec, &item)
	case "gatewayapi.httproute", "gatewayapi.tcproute", "gatewayapi.tlsroute":
		lowerGatewayRoute(spec, &item)
	case "cilium.network_policy", "cilium.clusterwide_network_policy":
		lowerCiliumPolicy(spec, mappingValue(node, "specs"), &item)
	}

	return item
}

func graphKindFor(apiVersion string, kind string) string {
	group := apiGroup(apiVersion)
	switch kind {
	case "Namespace":
		if apiVersion == "v1" {
			return "k8s.namespace"
		}
	case "ConfigMap":
		if apiVersion == "v1" {
			return "k8s.config_map"
		}
	case "Secret":
		if apiVersion == "v1" {
			return "k8s.secret"
		}
	case "Service":
		if apiVersion == "v1" {
			return "k8s.service"
		}
	case "Deployment":
		if group == "apps" {
			return "k8s.deployment"
		}
	case "StatefulSet":
		if group == "apps" {
			return "k8s.statefulset"
		}
	case "DaemonSet":
		if group == "apps" {
			return "k8s.daemonset"
		}
	case "Job":
		if group == "batch" {
			return "k8s.job"
		}
	case "CronJob":
		if group == "batch" {
			return "k8s.cronjob"
		}
	case "Ingress":
		if group == "networking.k8s.io" {
			return "k8s.ingress"
		}
	case "NetworkPolicy":
		if group == "networking.k8s.io" {
			return "k8s.network_policy"
		}
	case "Gateway":
		if group == "gateway.networking.k8s.io" {
			return "gatewayapi.gateway"
		}
	case "HTTPRoute":
		if group == "gateway.networking.k8s.io" {
			return "gatewayapi.httproute"
		}
	case "TCPRoute":
		if group == "gateway.networking.k8s.io" {
			return "gatewayapi.tcproute"
		}
	case "TLSRoute":
		if group == "gateway.networking.k8s.io" {
			return "gatewayapi.tlsroute"
		}
	case "CiliumNetworkPolicy":
		if group == "cilium.io" {
			return "cilium.network_policy"
		}
	case "CiliumClusterwideNetworkPolicy":
		if group == "cilium.io" {
			return "cilium.clusterwide_network_policy"
		}
	}
	return ""
}

func apiGroup(apiVersion string) string {
	if apiVersion == "" || apiVersion == "v1" {
		return ""
	}
	before, _, ok := strings.Cut(apiVersion, "/")
	if !ok {
		return apiVersion
	}
	return before
}

func normaliseNamespace(kind string, namespace string) string {
	if kind == "Namespace" || kind == "CiliumClusterwideNetworkPolicy" {
		return ""
	}
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return "default"
	}
	return namespace
}

func resourceID(item resource) string {
	if item.name == "" {
		return ""
	}
	switch item.graphKind {
	case "k8s.namespace":
		return "namespace/" + item.name
	case "cilium.clusterwide_network_policy":
		return "cluster/ciliumclusterwidenetworkpolicy/" + item.name
	}
	segment := idKindSegment(item.kind)
	if item.namespace == "" || segment == "" {
		return ""
	}
	return item.namespace + "/" + segment + "/" + item.name
}

func idKindSegment(kind string) string {
	switch kind {
	case "NetworkPolicy":
		return "networkpolicy"
	case "HTTPRoute":
		return "httproute"
	case "TCPRoute":
		return "tcproute"
	case "TLSRoute":
		return "tlsroute"
	case "ConfigMap":
		return "configmap"
	case "CiliumNetworkPolicy":
		return "ciliumnetworkpolicy"
	default:
		return strings.ToLower(kind)
	}
}

func baseProperties(item resource) map[string]any {
	properties := map[string]any{
		"api_version": item.apiVersion,
		"source_kind": sourceKind(item),
	}
	if item.namespace != "" {
		properties["namespace"] = item.namespace
	}
	if len(item.labels) > 0 {
		properties["labels"] = item.labels
	}
	return properties
}

func lowerWorkload(spec *yaml.Node, item *resource) {
	item.templateLabels = stringMap(mappingValue(mappingValue(mappingValue(spec, "template"), "metadata"), "labels"))
	if len(item.templateLabels) > 0 {
		item.properties["template_labels"] = item.templateLabels
	}
	selectorLabels := stringMap(mappingValue(mappingValue(spec, "selector"), "matchLabels"))
	if len(selectorLabels) > 0 {
		item.properties["selector_match_labels"] = selectorLabels
	}
	if replicas, ok := scalarInt(mappingValue(spec, "replicas")); ok {
		item.properties["replicas"] = replicas
	}
	containers := containerSummaries(mappingValue(mappingValue(mappingValue(spec, "template"), "spec"), "containers"))
	if len(containers) > 0 {
		item.properties["containers"] = containers
	}
}

func lowerService(spec *yaml.Node, item *resource) {
	item.serviceSelector = stringMap(mappingValue(spec, "selector"))
	if len(item.serviceSelector) > 0 {
		item.properties["selector"] = item.serviceSelector
	}
	serviceType, _ := scalarString(mappingValue(spec, "type"))
	if strings.TrimSpace(serviceType) == "" {
		serviceType = "ClusterIP"
	}
	item.properties["type"] = strings.TrimSpace(serviceType)
	ports := servicePorts(mappingValue(spec, "ports"))
	if len(ports) > 0 {
		item.properties["ports"] = ports
	}
}

func lowerIngress(spec *yaml.Node, item *resource) {
	if spec == nil || spec.Kind != yaml.MappingNode {
		return
	}
	if defaultBackend := mappingValue(spec, "defaultBackend"); defaultBackend != nil {
		item.ingressBackends = append(item.ingressBackends, ingressBackendRef(defaultBackend, item.namespace))
	}
	rules := mappingValue(spec, "rules")
	if rules == nil || rules.Kind != yaml.SequenceNode {
		return
	}
	for _, rule := range rules.Content {
		http := mappingValue(rule, "http")
		paths := mappingValue(http, "paths")
		if paths == nil || paths.Kind != yaml.SequenceNode {
			continue
		}
		for _, path := range paths.Content {
			item.ingressBackends = append(item.ingressBackends, ingressBackendRef(mappingValue(path, "backend"), item.namespace))
		}
	}
}

func lowerNetworkPolicy(spec *yaml.Node, item *resource) {
	item.properties["policy_semantics"] = "preserved_not_fully_modelled"
	policyTypes := stringSlice(mappingValue(spec, "policyTypes"))
	if len(policyTypes) > 0 {
		item.properties["policy_types"] = policyTypes
	}
	podSelector := stringMap(mappingValue(mappingValue(spec, "podSelector"), "matchLabels"))
	if len(podSelector) > 0 {
		item.properties["pod_selector_match_labels"] = podSelector
	}
}

func lowerConfigMap(node *yaml.Node, item *resource) {
	dataKeys := sortedMappingKeys(mappingValue(node, "data"))
	binaryDataKeys := sortedMappingKeys(mappingValue(node, "binaryData"))
	if len(dataKeys) > 0 {
		item.properties["data_keys"] = dataKeys
	}
	if len(binaryDataKeys) > 0 {
		item.properties["binary_data_keys"] = binaryDataKeys
	}
}

func lowerSecret(node *yaml.Node, item *resource) {
	secretType, _ := scalarString(mappingValue(node, "type"))
	if strings.TrimSpace(secretType) != "" {
		item.properties["type"] = strings.TrimSpace(secretType)
	}
	dataKeys := sortedMappingKeys(mappingValue(node, "data"))
	stringDataKeys := sortedMappingKeys(mappingValue(node, "stringData"))
	if len(dataKeys) > 0 {
		item.properties["data_keys"] = dataKeys
		item.properties["has_data"] = true
	}
	if len(stringDataKeys) > 0 {
		item.properties["string_data_keys"] = stringDataKeys
		item.properties["has_string_data"] = true
	}
}

func lowerGateway(spec *yaml.Node, item *resource) {
	gatewayClass, _ := scalarString(mappingValue(spec, "gatewayClassName"))
	if strings.TrimSpace(gatewayClass) != "" {
		item.properties["gateway_class_name"] = strings.TrimSpace(gatewayClass)
	}
	listeners := listenerSummaries(mappingValue(spec, "listeners"))
	if len(listeners) > 0 {
		item.properties["listeners"] = listeners
	}
}

func lowerGatewayRoute(spec *yaml.Node, item *resource) {
	item.routeParents = parentRefs(mappingValue(spec, "parentRefs"), item.namespace)
	item.routeBackends = gatewayBackendRefs(mappingValue(spec, "rules"), item.namespace)
	hostnames := stringSlice(mappingValue(spec, "hostnames"))
	if len(hostnames) > 0 {
		item.properties["hostnames"] = hostnames
	}
}

func lowerCiliumPolicy(spec *yaml.Node, specs *yaml.Node, item *resource) {
	item.properties["policy_semantics"] = "preserved_not_fully_modelled"
	if spec != nil {
		item.properties["has_spec"] = true
	}
	if specs != nil && specs.Kind == yaml.SequenceNode {
		item.properties["specs_count"] = len(specs.Content)
	}
}

func containerSummaries(containers *yaml.Node) []map[string]any {
	if containers == nil || containers.Kind != yaml.SequenceNode {
		return nil
	}
	var out []map[string]any
	for _, container := range containers.Content {
		name, _ := scalarString(mappingValue(container, "name"))
		image, _ := scalarString(mappingValue(container, "image"))
		summary := map[string]any{}
		if name != "" {
			summary["name"] = name
		}
		if image != "" {
			summary["image"] = image
		}
		ports := containerPorts(mappingValue(container, "ports"))
		if len(ports) > 0 {
			summary["ports"] = ports
		}
		if len(summary) > 0 {
			out = append(out, summary)
		}
	}
	return out
}

func containerPorts(ports *yaml.Node) []int {
	if ports == nil || ports.Kind != yaml.SequenceNode {
		return nil
	}
	var out []int
	for _, portNode := range ports.Content {
		port, ok := scalarInt(mappingValue(portNode, "containerPort"))
		if ok {
			out = append(out, port)
		}
	}
	sort.Ints(out)
	return out
}

func servicePorts(ports *yaml.Node) []map[string]any {
	if ports == nil || ports.Kind != yaml.SequenceNode {
		return nil
	}
	var out []map[string]any
	for _, portNode := range ports.Content {
		summary := map[string]any{}
		if name, ok := scalarString(mappingValue(portNode, "name")); ok && name != "" {
			summary["name"] = name
		}
		if port, ok := scalarInt(mappingValue(portNode, "port")); ok {
			summary["port"] = port
		}
		if target, ok := scalarInt(mappingValue(portNode, "targetPort")); ok {
			summary["target_port"] = target
		} else if targetName, ok := scalarString(mappingValue(portNode, "targetPort")); ok && targetName != "" {
			summary["target_port_name"] = targetName
		}
		if protocol, ok := scalarString(mappingValue(portNode, "protocol")); ok && protocol != "" {
			summary["protocol"] = protocol
		}
		if len(summary) > 0 {
			out = append(out, summary)
		}
	}
	return out
}

func listenerSummaries(listeners *yaml.Node) []map[string]any {
	if listeners == nil || listeners.Kind != yaml.SequenceNode {
		return nil
	}
	var out []map[string]any
	for _, listener := range listeners.Content {
		summary := map[string]any{}
		if name, ok := scalarString(mappingValue(listener, "name")); ok && name != "" {
			summary["name"] = name
		}
		if protocol, ok := scalarString(mappingValue(listener, "protocol")); ok && protocol != "" {
			summary["protocol"] = protocol
		}
		if port, ok := scalarInt(mappingValue(listener, "port")); ok {
			summary["port"] = port
		}
		if len(summary) > 0 {
			out = append(out, summary)
		}
	}
	return out
}

func ingressBackendRef(backend *yaml.Node, namespace string) backendRef {
	if backend == nil || backend.Kind != yaml.MappingNode {
		return backendRef{}
	}
	service := mappingValue(backend, "service")
	if service == nil {
		return backendRef{namespace: namespace, kind: "resource"}
	}
	name, _ := scalarString(mappingValue(service, "name"))
	port := firstInt(mappingValue(mappingValue(service, "port"), "number"))
	return backendRef{namespace: namespace, name: strings.TrimSpace(name), kind: "Service", port: port}
}

func parentRefs(refs *yaml.Node, defaultNamespace string) []parentRef {
	if refs == nil || refs.Kind != yaml.SequenceNode {
		return nil
	}
	var out []parentRef
	for _, ref := range refs.Content {
		name, _ := scalarString(mappingValue(ref, "name"))
		namespace, _ := scalarString(mappingValue(ref, "namespace"))
		kind, _ := scalarString(mappingValue(ref, "kind"))
		group, _ := scalarString(mappingValue(ref, "group"))
		if namespace == "" {
			namespace = defaultNamespace
		}
		if kind == "" {
			kind = "Gateway"
		}
		if group == "" {
			group = "gateway.networking.k8s.io"
		}
		out = append(out, parentRef{
			namespace: namespace,
			name:      strings.TrimSpace(name),
			kind:      strings.TrimSpace(kind),
			group:     strings.TrimSpace(group),
		})
	}
	return out
}

func gatewayBackendRefs(rules *yaml.Node, defaultNamespace string) []backendRef {
	if rules == nil || rules.Kind != yaml.SequenceNode {
		return nil
	}
	var out []backendRef
	for _, rule := range rules.Content {
		backendRefs := mappingValue(rule, "backendRefs")
		if backendRefs == nil || backendRefs.Kind != yaml.SequenceNode {
			continue
		}
		for _, ref := range backendRefs.Content {
			name, _ := scalarString(mappingValue(ref, "name"))
			namespace, _ := scalarString(mappingValue(ref, "namespace"))
			kind, _ := scalarString(mappingValue(ref, "kind"))
			group, _ := scalarString(mappingValue(ref, "group"))
			if namespace == "" {
				namespace = defaultNamespace
			}
			if kind == "" {
				kind = "Service"
			}
			port := firstInt(mappingValue(ref, "port"))
			out = append(out, backendRef{
				namespace: namespace,
				name:      strings.TrimSpace(name),
				kind:      strings.TrimSpace(kind),
				group:     strings.TrimSpace(group),
				port:      port,
			})
		}
	}
	return out
}

func inferEdges(resources []resource, resourceByID map[string]resource, report *loss.Report) []graph.Edge {
	var edges []graph.Edge
	seen := map[string]struct{}{}
	for _, item := range resources {
		if item.id == "" || item.graphKind == "" {
			continue
		}
		switch item.graphKind {
		case "k8s.service":
			matched := false
			for _, candidate := range resources {
				if !isWorkload(candidate) || candidate.namespace != item.namespace {
					continue
				}
				if selectorMatches(item.serviceSelector, candidate.templateLabels) {
					edges = appendUniqueEdge(edges, seen, graph.Edge{
						From:   item.id,
						To:     candidate.id,
						Kind:   "network.route",
						Intent: "k8s_service_selector",
					})
					if len(edges) > maxInferredEdges {
						report.Ambiguous = append(report.Ambiguous, loss.Item{
							Resource: resourceName(item),
							Kind:     sourceKind(item),
							Message:  "Relationship inference reached the current edge budget; remaining inferred edges require manual review.",
						})
						sortEdges(edges)
						return edges
					}
					matched = true
				}
			}
			if len(item.serviceSelector) > 0 && !matched {
				report.Ambiguous = append(report.Ambiguous, loss.Item{
					Resource: resourceName(item),
					Kind:     sourceKind(item),
					Message:  "Service selector did not match any imported workload template labels; route relationship was not lowered.",
				})
			}
		case "k8s.ingress":
			for _, backend := range item.ingressBackends {
				appendBackendEdge(item, backend, "k8s_ingress_backend", resourceByID, report, &edges, seen)
			}
		case "gatewayapi.httproute", "gatewayapi.tcproute", "gatewayapi.tlsroute":
			for _, parent := range item.routeParents {
				if parent.kind != "Gateway" || parent.group != "gateway.networking.k8s.io" {
					report.Ambiguous = append(report.Ambiguous, loss.Item{
						Resource: resourceName(item),
						Kind:     sourceKind(item),
						Message:  "Route parentRef does not target a Gateway in the supported Gateway API subset; parent relationship was not lowered.",
					})
					continue
				}
				targetID := namespacedID(parent.namespace, "Gateway", parent.name)
				if _, ok := resourceByID[targetID]; !ok {
					report.Ambiguous = append(report.Ambiguous, loss.Item{
						Resource: resourceName(item),
						Kind:     sourceKind(item),
						Message:  fmt.Sprintf("Route parentRef %s/%s was not found in imported manifests; parent relationship was not lowered.", parent.namespace, parent.name),
					})
					continue
				}
				edges = appendUniqueEdge(edges, seen, graph.Edge{
					From:   item.id,
					To:     targetID,
					Kind:   "depends_on",
					Intent: "gatewayapi_parent_ref",
				})
			}
			for _, backend := range item.routeBackends {
				appendBackendEdge(item, backend, "gatewayapi_backend_ref", resourceByID, report, &edges, seen)
			}
		}
	}
	sortEdges(edges)
	return edges
}

func appendBackendEdge(item resource, backend backendRef, intent string, resourceByID map[string]resource, report *loss.Report, edges *[]graph.Edge, seen map[string]struct{}) {
	if backend.kind != "Service" || backend.group != "" {
		report.Ambiguous = append(report.Ambiguous, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  "Backend reference does not target a core Service in the supported Kubernetes subset; route relationship was not lowered.",
		})
		return
	}
	targetID := namespacedID(backend.namespace, "Service", backend.name)
	if _, ok := resourceByID[targetID]; !ok {
		report.Ambiguous = append(report.Ambiguous, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  fmt.Sprintf("Backend service %s/%s was not found in imported manifests; route relationship was not lowered.", backend.namespace, backend.name),
		})
		return
	}
	edge := graph.Edge{
		From:   item.id,
		To:     targetID,
		Kind:   "network.route",
		Intent: intent,
	}
	if backend.port >= limits.MinNetworkPort && backend.port <= limits.MaxNetworkPort {
		edge.Port = backend.port
	}
	*edges = appendUniqueEdge(*edges, seen, edge)
}

func appendUniqueEdge(edges []graph.Edge, seen map[string]struct{}, edge graph.Edge) []graph.Edge {
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s", edge.From, edge.To, edge.Kind, edge.Port, edge.Intent)
	if _, ok := seen[key]; ok {
		return edges
	}
	seen[key] = struct{}{}
	return append(edges, edge)
}

func namespacedID(namespace string, kind string, name string) string {
	if namespace == "" {
		namespace = "default"
	}
	return namespace + "/" + idKindSegment(kind) + "/" + name
}

func selectorMatches(selector map[string]string, labels map[string]string) bool {
	if len(selector) == 0 || len(labels) == 0 {
		return false
	}
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func isWorkload(item resource) bool {
	switch item.graphKind {
	case "k8s.deployment", "k8s.statefulset", "k8s.daemonset", "k8s.job", "k8s.cronjob":
		return true
	default:
		return false
	}
}

func appendSemanticLoss(item resource, report *loss.Report) {
	switch item.graphKind {
	case "k8s.network_policy":
		report.Ambiguous = append(report.Ambiguous, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  "NetworkPolicy rule semantics are preserved in the source manifest but not fully lowered into Planwright graph edges by the current Kubernetes supported subset; manual review required.",
		})
	case "cilium.network_policy", "cilium.clusterwide_network_policy":
		report.Ambiguous = append(report.Ambiguous, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  "Cilium policy semantics are preserved in the source manifest but not fully lowered into Planwright graph edges by the current Kubernetes supported subset; manual review required.",
		})
	case "k8s.secret":
		report.Preserved = append(report.Preserved, loss.Item{
			Resource: resourceName(item),
			Kind:     sourceKind(item),
			Message:  "Secret data values are intentionally not lowered into the graph; only metadata and key names are recorded.",
		})
	}
}

func resourceName(item resource) string {
	if item.namespace == "" {
		return item.name
	}
	return item.namespace + "/" + item.name
}

func sourceKind(item resource) string {
	if item.apiVersion == "" {
		return item.kind
	}
	if item.kind == "" {
		return item.apiVersion
	}
	return item.apiVersion + "/" + item.kind
}

func documentNode(root *yaml.Node) *yaml.Node {
	if root == nil {
		return nil
	}
	if root.Kind == yaml.DocumentNode {
		if len(root.Content) == 0 {
			return nil
		}
		return root.Content[0]
	}
	return root
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func scalarValue(node *yaml.Node) (any, bool) {
	if node == nil || node.Kind != yaml.ScalarNode {
		return nil, false
	}
	switch node.Tag {
	case "!!bool":
		value, err := strconv.ParseBool(node.Value)
		return value, err == nil
	case "!!int":
		value, err := strconv.Atoi(node.Value)
		return value, err == nil
	default:
		return node.Value, true
	}
}

func scalarString(node *yaml.Node) (string, bool) {
	value, ok := scalarValue(node)
	if !ok {
		return "", false
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", false
	}
	return stringValue, true
}

func scalarInt(node *yaml.Node) (int, bool) {
	value, ok := scalarValue(node)
	if !ok {
		return 0, false
	}
	intValue, ok := value.(int)
	return intValue, ok
}

func firstInt(nodes ...*yaml.Node) int {
	for _, node := range nodes {
		if value, ok := scalarInt(node); ok {
			return value
		}
	}
	return 0
}

func stringMap(node *yaml.Node) map[string]string {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	out := map[string]string{}
	for index := 0; index+1 < len(node.Content); index += 2 {
		value, ok := scalarString(node.Content[index+1])
		if !ok {
			continue
		}
		key := node.Content[index].Value
		if key != "" {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func stringSlice(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.SequenceNode {
		return nil
	}
	var out []string
	for _, item := range node.Content {
		value, ok := scalarString(item)
		if ok && strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	sort.Strings(out)
	return out
}

func sortedMappingKeys(node *yaml.Node) []string {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	keys := make([]string, 0, len(node.Content)/2)
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := strings.TrimSpace(node.Content[index].Value)
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func sortGraph(g *graph.Graph) {
	sort.Slice(g.Nodes, func(i, j int) bool {
		return g.Nodes[i].ID < g.Nodes[j].ID
	})
	sortEdges(g.Edges)
}

func sortEdges(edges []graph.Edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		if edges[i].To != edges[j].To {
			return edges[i].To < edges[j].To
		}
		if edges[i].Kind != edges[j].Kind {
			return edges[i].Kind < edges[j].Kind
		}
		if edges[i].Port != edges[j].Port {
			return edges[i].Port < edges[j].Port
		}
		return edges[i].Intent < edges[j].Intent
	})
}

func sortLossReport(report *loss.Report) {
	sortLossItems(report.Lowered)
	sortLossItems(report.Unsupported)
	sortLossItems(report.Ambiguous)
	sortLossItems(report.Preserved)
}

func sortLossItems(items []loss.Item) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Resource == items[j].Resource {
			return items[i].Kind < items[j].Kind
		}
		return items[i].Resource < items[j].Resource
	})
}
