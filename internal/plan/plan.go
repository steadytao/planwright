// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/limits"
	"github.com/steadytao/planwright/internal/localfs"
	"github.com/steadytao/planwright/internal/yamlutil"
)

const Version = "planwright.v1"

const (
	awsWebAppPattern = "aws.webapp.alb_ecs_rds"
	maxPlanBytes     = 5 * 1024 * 1024
)

var awsWebAppPropertyKeys = map[string]struct{}{
	"app_port":  {},
	"db_engine": {},
	"db_public": {},
}

type Document struct {
	Version    string               `yaml:"version"`
	Provider   string               `yaml:"provider"`
	Region     string               `yaml:"region"`
	Profile    string               `yaml:"profile"`
	Components map[string]Component `yaml:"components"`
	Flows      []Flow               `yaml:"flows"`
}

type Component struct {
	Pattern    string         `yaml:"pattern"`
	Properties map[string]any `yaml:"properties"`
}

type Flow struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Kind     string `yaml:"kind"`
	Protocol string `yaml:"protocol"`
	Port     int    `yaml:"port"`
	Intent   string `yaml:"intent"`
}

func Load(path string) (Document, error) {
	data, err := readUserSelectedPlan(path)
	if err != nil {
		return Document{}, err
	}
	return Parse(data, path)
}

func LoadWithSource(path string) (Document, []byte, error) {
	data, err := readUserSelectedPlan(path)
	if err != nil {
		return Document{}, nil, err
	}
	document, err := Parse(data, path)
	if err != nil {
		return Document{}, nil, err
	}
	return document, append([]byte(nil), data...), nil
}

func readUserSelectedPlan(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)
	if strings.TrimSpace(cleanPath) == "." {
		return nil, fmt.Errorf("read %s: path must not be empty", path)
	}

	data, err := localfs.ReadNamedRegularFile(cleanPath, maxPlanBytes, "plan")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Parse(data []byte, sourceName string) (Document, error) {
	var root yaml.Node
	nodeDecoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := nodeDecoder.Decode(&root); err != nil {
		return Document{}, fmt.Errorf("parse %s: %w", sourceName, err)
	}
	var extra yaml.Node
	if err := nodeDecoder.Decode(&extra); err != io.EOF {
		return Document{}, fmt.Errorf("parse %s: plan must contain exactly one YAML document", sourceName)
	}
	if err := yamlutil.RejectDuplicateMappingKeys(&root, sourceName); err != nil {
		return Document{}, err
	}

	var document Document
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("parse %s: %w", sourceName, err)
	}
	if strings.TrimSpace(document.Version) != Version {
		return Document{}, fmt.Errorf("unsupported Planwright plan version %q", document.Version)
	}
	if document.Components == nil {
		document.Components = map[string]Component{}
	}
	return document, nil
}

func (document Document) ToGraph() (graph.Graph, []graph.Diagnostic) {
	lowered := graph.Graph{
		Version:  graph.Version,
		Provider: strings.TrimSpace(document.Provider),
		Region:   strings.TrimSpace(document.Region),
		Profile:  strings.TrimSpace(document.Profile),
	}

	var diagnostics []graph.Diagnostic
	if usesInternet(document.Flows) {
		lowered.Nodes = append(lowered.Nodes, graph.Node{
			ID:   "internet",
			Kind: "external.internet",
			Name: "Internet",
		})
	}

	componentIDs := make([]string, 0, len(document.Components))
	for id := range document.Components {
		componentIDs = append(componentIDs, id)
	}
	sort.Strings(componentIDs)

	for _, id := range componentIDs {
		component := document.Components[id]
		if !validComponentID(id) {
			diagnostics = append(diagnostics, graph.Diagnostic{
				Severity: graph.SeverityError,
				Code:     "PW-PLAN-COMPONENT-001",
				Resource: id,
				Message:  "Component ID must start with a lowercase letter and contain only lowercase letters, digits, underscores or hyphens.",
				Fix:      "Use a stable component ID such as webapp or worker-1.",
			})
		}
		switch strings.TrimSpace(component.Pattern) {
		case awsWebAppPattern:
			nodes, componentDiagnostics := lowerAWSWebApp(id, component)
			lowered.Nodes = append(lowered.Nodes, nodes...)
			diagnostics = append(diagnostics, componentDiagnostics...)
		default:
			diagnostics = append(diagnostics, graph.Diagnostic{
				Severity: graph.SeverityError,
				Code:     "PW-PLAN-PATTERN-001",
				Resource: id,
				Message:  fmt.Sprintf("Component pattern %q is not supported.", component.Pattern),
				Fix:      "Use aws.webapp.alb_ecs_rds or remove the component until the pattern is supported.",
			})
		}
	}

	for _, flow := range document.Flows {
		lowered.Edges = append(lowered.Edges, graph.Edge{
			From:     flow.From,
			To:       flow.To,
			Kind:     strings.TrimSpace(flow.Kind),
			Protocol: strings.TrimSpace(flow.Protocol),
			Port:     flow.Port,
			Intent:   strings.TrimSpace(flow.Intent),
		})
	}

	diagnostics = append(diagnostics, graph.Validate(lowered)...)
	return lowered, diagnostics
}

func lowerAWSWebApp(id string, component Component) ([]graph.Node, []graph.Diagnostic) {
	appPort, appPortDiagnostics := appPortProperty(id, component.Properties)
	dbEngine, dbEngineDiagnostics := dbEngineProperty(id, component.Properties)
	dbPublic, dbPublicDiagnostics := dbPublicProperty(id, component.Properties)
	unknownPropertyDiagnostics := unsupportedPropertyDiagnostics(id, component.Properties, awsWebAppPropertyKeys)
	diagnostics := make([]graph.Diagnostic, 0, len(appPortDiagnostics)+len(dbEngineDiagnostics)+len(dbPublicDiagnostics)+len(unknownPropertyDiagnostics))
	diagnostics = append(diagnostics, appPortDiagnostics...)
	diagnostics = append(diagnostics, dbEngineDiagnostics...)
	diagnostics = append(diagnostics, dbPublicDiagnostics...)
	diagnostics = append(diagnostics, unknownPropertyDiagnostics...)

	return []graph.Node{
		{
			ID:   id + ".alb",
			Kind: "aws.alb",
			Name: id + "-alb",
			Properties: map[string]any{
				"scheme": "internet-facing",
			},
		},
		{
			ID:   id + ".app",
			Kind: "aws.ecs.service",
			Name: id + "-app",
			Properties: map[string]any{
				"port": appPort,
			},
		},
		{
			ID:   id + ".db",
			Kind: "aws.rds." + dbEngine,
			Name: id + "-db",
			Properties: map[string]any{
				"publicly_accessible": dbPublic,
				"port":                5432,
			},
		},
	}, diagnostics
}

func unsupportedPropertyDiagnostics(componentID string, properties map[string]any, supported map[string]struct{}) []graph.Diagnostic {
	if len(properties) == 0 {
		return nil
	}
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var diagnostics []graph.Diagnostic
	for _, key := range keys {
		if _, ok := supported[key]; ok {
			continue
		}
		diagnostics = append(diagnostics, graph.Diagnostic{
			Severity: graph.SeverityError,
			Code:     "PW-PLAN-PROPERTY-004",
			Resource: componentID + "." + key,
			Message:  fmt.Sprintf("aws.webapp.alb_ecs_rds property %q is not supported.", key),
			Fix:      "Remove the property or use one of app_port, db_engine or db_public.",
		})
	}
	return diagnostics
}

func usesInternet(flows []Flow) bool {
	for _, flow := range flows {
		if strings.TrimSpace(flow.From) == "internet" || strings.TrimSpace(flow.To) == "internet" {
			return true
		}
	}
	return false
}

func validComponentID(id string) bool {
	if id == "" {
		return false
	}
	for index, value := range id {
		switch {
		case value >= 'a' && value <= 'z':
		case index > 0 && value >= '0' && value <= '9':
		case index > 0 && (value == '_' || value == '-'):
		default:
			return false
		}
	}
	return true
}

func appPortProperty(componentID string, properties map[string]any) (int, []graph.Diagnostic) {
	const fallback = 8080
	value, exists := properties["app_port"]
	if !exists {
		return fallback, nil
	}
	intValue, ok := value.(int)
	if !ok || intValue < limits.MinNetworkPort || intValue > limits.MaxNetworkPort {
		return fallback, []graph.Diagnostic{
			{
				Severity: graph.SeverityError,
				Code:     "PW-PLAN-PROPERTY-002",
				Resource: componentID + ".app_port",
				Message:  fmt.Sprintf("aws.webapp.alb_ecs_rds app_port must be an integer between %d and %d.", limits.MinNetworkPort, limits.MaxNetworkPort),
				Fix:      "Set app_port to the application listener port, for example 8080.",
			},
		}
	}
	return intValue, nil
}

func dbEngineProperty(componentID string, properties map[string]any) (string, []graph.Diagnostic) {
	const fallback = "postgres"
	value, exists := properties["db_engine"]
	if !exists {
		return fallback, nil
	}
	stringValue, ok := value.(string)
	if !ok || strings.TrimSpace(stringValue) != fallback {
		return fallback, []graph.Diagnostic{
			{
				Severity: graph.SeverityError,
				Code:     "PW-PLAN-PROPERTY-001",
				Resource: componentID + ".db_engine",
				Message:  "aws.webapp.alb_ecs_rds currently supports only postgres for db_engine.",
				Fix:      "Set db_engine to postgres or wait for another database pattern to be implemented.",
			},
		}
	}
	return fallback, nil
}

func dbPublicProperty(componentID string, properties map[string]any) (bool, []graph.Diagnostic) {
	value, exists := properties["db_public"]
	if !exists {
		return false, nil
	}
	boolValue, ok := value.(bool)
	if !ok {
		return false, []graph.Diagnostic{
			{
				Severity: graph.SeverityError,
				Code:     "PW-PLAN-PROPERTY-003",
				Resource: componentID + ".db_public",
				Message:  "aws.webapp.alb_ecs_rds db_public must be a boolean.",
				Fix:      "Set db_public to false or true without quotes.",
			},
		}
	}
	return boolValue, nil
}
