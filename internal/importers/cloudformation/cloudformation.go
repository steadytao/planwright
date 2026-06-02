// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"bytes"
	"fmt"
	"io"
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

type Format string

const (
	FormatCloudFormation Format = "cloudformation"
	FormatSAM            Format = "sam"
)

const (
	maxTemplateBytes      = 5 * 1024 * 1024
	maxIntrinsicWalkDepth = 128
	maxIntrinsicWalkNodes = 10000
)

type Result struct {
	Graph       graph.Graph
	Loss        loss.Report
	Diagnostics []graph.Diagnostic
	Source      []byte
}

type resourceNode struct {
	id         string
	kind       string
	properties *yaml.Node
}

func ImportFile(path string, format Format) (Result, error) {
	source, err := readTemplate(path)
	if err != nil {
		return Result{}, err
	}
	result, err := Import(source, path, format)
	if err != nil {
		return Result{}, err
	}
	result.Source = append([]byte(nil), source...)
	return result, nil
}

func Import(source []byte, sourceName string, format Format) (Result, error) {
	var root yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(source))
	if err := decoder.Decode(&root); err != nil {
		return Result{}, fmt.Errorf("parse %s: %w", sourceName, err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return Result{}, fmt.Errorf("parse %s: template must contain exactly one YAML document", sourceName)
	}
	document := documentNode(&root)
	if document == nil || document.Kind != yaml.MappingNode {
		return Result{}, fmt.Errorf("parse %s: expected CloudFormation mapping document", sourceName)
	}
	if err := yamlutil.RejectDuplicateMappingKeys(document, sourceName); err != nil {
		return Result{}, err
	}

	resourcesNode := mappingValue(document, "Resources")
	if resourcesNode == nil || resourcesNode.Kind != yaml.MappingNode {
		return Result{}, fmt.Errorf("parse %s: expected Resources mapping", sourceName)
	}

	result := Result{
		Graph: graph.Graph{
			Version:  graph.Version,
			Provider: "aws",
			Region:   "unknown",
			Nodes:    []graph.Node{},
			Edges:    []graph.Edge{},
		},
		Loss: loss.Report{
			SourceFormat: string(format),
			Source:       sourceName,
		},
	}

	preserveTemplateSections(document, &result.Loss)
	if format == FormatSAM && mappingValue(document, "Transform") == nil {
		result.Loss.Ambiguous = append(result.Loss.Ambiguous, loss.Item{
			Resource: "Transform",
			Kind:     "sam.transform",
			Message:  "SAM import was requested but the template has no Transform section. Resources were still parsed by type.",
		})
	}

	resources := collectResources(resourcesNode)
	internetNodeAdded := false
	for _, resource := range resources {
		kind, properties, supportedProperties := lowerResource(resource)
		if kind == "" {
			result.Loss.Unsupported = append(result.Loss.Unsupported, loss.Item{
				Resource: resource.id,
				Kind:     resource.kind,
				Message:  "Resource type is preserved in the source template but is not lowered into the Planwright graph by the current supported subset; manual review required.",
			})
			continue
		}

		result.Graph.Nodes = append(result.Graph.Nodes, graph.Node{
			ID:         resource.id,
			Kind:       kind,
			Name:       resource.id,
			Properties: properties,
		})
		result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
			Resource: resource.id,
			Kind:     resource.kind,
			Message:  "Resource inventory was lowered into the Planwright graph.",
		})
		appendUnsupportedPropertyLoss(resource, supportedProperties, &result.Loss)
		appendIntrinsicAmbiguity(resource, &result.Loss)

		if resource.kind == "AWS::EC2::SecurityGroup" {
			edges, ambiguous := publicIngressEdges(resource)
			result.Loss.Ambiguous = append(result.Loss.Ambiguous, ambiguous...)
			if len(edges) > 0 && !internetNodeAdded {
				result.Graph.Nodes = append(result.Graph.Nodes, graph.Node{
					ID:   "internet",
					Kind: "external.internet",
					Name: "Internet",
				})
				internetNodeAdded = true
			}
			result.Graph.Edges = append(result.Graph.Edges, edges...)
		}
	}

	result.Diagnostics = graph.Validate(result.Graph)
	sortLossReport(&result.Loss)
	return result, nil
}

func readTemplate(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("read template: path must not be empty")
	}
	data, err := localfs.ReadNamedRegularFile(path, maxTemplateBytes, "template")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func documentNode(root *yaml.Node) *yaml.Node {
	if root == nil {
		return nil
	}
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		return root.Content[0]
	}
	return root
}

func collectResources(resourcesNode *yaml.Node) []resourceNode {
	resources := make([]resourceNode, 0, len(resourcesNode.Content)/2)
	for index := 0; index+1 < len(resourcesNode.Content); index += 2 {
		id := resourcesNode.Content[index].Value
		body := resourcesNode.Content[index+1]
		if body.Kind != yaml.MappingNode {
			resources = append(resources, resourceNode{id: id})
			continue
		}
		typeNode := mappingValue(body, "Type")
		propertiesNode := mappingValue(body, "Properties")
		kind, _ := scalarString(typeNode)
		resources = append(resources, resourceNode{
			id:         strings.TrimSpace(id),
			kind:       strings.TrimSpace(kind),
			properties: propertiesNode,
		})
	}
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].id < resources[j].id
	})
	return resources
}

func preserveTemplateSections(document *yaml.Node, report *loss.Report) {
	for _, section := range []string{"Parameters", "Mappings", "Conditions", "Outputs", "Transform"} {
		if mappingValue(document, section) != nil {
			report.Preserved = append(report.Preserved, loss.Item{
				Resource: section,
				Kind:     "cloudformation.section",
				Message:  "Section is preserved in the source template but not fully lowered into the Planwright graph by the current supported subset.",
			})
		}
	}
}

func lowerResource(resource resourceNode) (string, map[string]any, map[string]string) {
	switch resource.kind {
	case "AWS::EC2::VPC":
		keys := map[string]string{
			"CidrBlock": "cidr_block",
		}
		return "aws.vpc", selectedProperties(resource.properties, keys), keys
	case "AWS::EC2::Subnet":
		keys := map[string]string{
			"CidrBlock":                   "cidr_block",
			"AvailabilityZone":            "availability_zone",
			"MapPublicIpOnLaunch":         "map_public_ip_on_launch",
			"AssignIpv6AddressOnCreation": "assign_ipv6_address_on_creation",
		}
		return "aws.subnet", selectedProperties(resource.properties, keys), keys
	case "AWS::EC2::SecurityGroup":
		keys := map[string]string{
			"GroupDescription": "description",
			"GroupName":        "name",
		}
		return "aws.security_group", selectedProperties(resource.properties, keys), keys
	case "AWS::ElasticLoadBalancingV2::LoadBalancer":
		keys := map[string]string{
			"Scheme": "scheme",
			"Type":   "type",
		}
		properties := selectedProperties(resource.properties, keys)
		if strings.EqualFold(stringProperty(properties, "type"), "network") {
			return "aws.nlb", properties, keys
		}
		return "aws.alb", properties, keys
	case "AWS::RDS::DBInstance":
		keys := map[string]string{
			"Engine":             "engine",
			"PubliclyAccessible": "publicly_accessible",
			"Port":               "port",
		}
		properties := selectedProperties(resource.properties, keys)
		if strings.EqualFold(stringProperty(properties, "engine"), "postgres") {
			return "aws.rds.postgres", properties, keys
		}
		return "aws.rds.instance", properties, keys
	case "AWS::S3::Bucket":
		keys := map[string]string{
			"BucketName": "bucket_name",
		}
		return "aws.s3.bucket", selectedProperties(resource.properties, keys), keys
	case "AWS::IAM::Role":
		keys := map[string]string{
			"RoleName": "role_name",
		}
		return "aws.iam.role", selectedProperties(resource.properties, keys), keys
	case "AWS::Serverless::Function":
		keys := map[string]string{
			"Runtime": "runtime",
			"Handler": "handler",
		}
		return "aws.lambda.function", selectedProperties(resource.properties, keys), keys
	case "AWS::Serverless::HttpApi":
		keys := map[string]string{
			"StageName": "stage_name",
		}
		return "aws.apigateway.http_api", selectedProperties(resource.properties, keys), keys
	case "AWS::Serverless::SimpleTable":
		keys := map[string]string{
			"TableName": "table_name",
		}
		return "aws.dynamodb.table", selectedProperties(resource.properties, keys), keys
	default:
		return "", nil, nil
	}
}

func selectedProperties(properties *yaml.Node, keys map[string]string) map[string]any {
	out := map[string]any{}
	if properties == nil || properties.Kind != yaml.MappingNode {
		return out
	}
	sourceKeys := make([]string, 0, len(keys))
	for key := range keys {
		sourceKeys = append(sourceKeys, key)
	}
	sort.Strings(sourceKeys)
	for _, sourceKey := range sourceKeys {
		valueNode := mappingValue(properties, sourceKey)
		if valueNode == nil {
			continue
		}
		if value, ok := scalarValue(valueNode); ok {
			out[keys[sourceKey]] = value
		}
	}
	return out
}

func appendUnsupportedPropertyLoss(resource resourceNode, supported map[string]string, report *loss.Report) {
	if resource.properties == nil || resource.properties.Kind != yaml.MappingNode {
		return
	}
	var unsupported []string
	for index := 0; index+1 < len(resource.properties.Content); index += 2 {
		key := resource.properties.Content[index].Value
		if _, ok := supported[key]; !ok {
			unsupported = append(unsupported, key)
		}
	}
	sort.Strings(unsupported)
	for _, property := range unsupported {
		report.Preserved = append(report.Preserved, loss.Item{
			Resource: resource.id,
			Kind:     "cloudformation.property",
			Message:  fmt.Sprintf("Property %s is preserved in the source template but is not lowered into the Planwright graph by the current supported subset; manual review required.", property),
		})
	}
}

func publicIngressEdges(resource resourceNode) ([]graph.Edge, []loss.Item) {
	if resource.properties == nil || resource.properties.Kind != yaml.MappingNode {
		return nil, nil
	}
	ingress := mappingValue(resource.properties, "SecurityGroupIngress")
	if ingress == nil || ingress.Kind != yaml.SequenceNode {
		return nil, nil
	}
	var edges []graph.Edge
	var ambiguous []loss.Item
	for _, rule := range ingress.Content {
		if rule.Kind != yaml.MappingNode {
			continue
		}
		if !isPublicIngressRule(rule) {
			continue
		}
		protocol, _ := scalarString(mappingValue(rule, "IpProtocol"))
		fromPort, fromOK := scalarInt(mappingValue(rule, "FromPort"))
		toPort, toOK := scalarInt(mappingValue(rule, "ToPort"))
		if (protocol != "tcp" && protocol != "udp") || !fromOK || !toOK || fromPort < limits.MinNetworkPort || fromPort > limits.MaxNetworkPort || fromPort != toPort {
			ambiguous = append(ambiguous, loss.Item{
				Resource: resource.id,
				Kind:     "cloudformation.security_group_ingress",
				Message:  "A public security group ingress rule is preserved but not lowered because the current supported subset only models tcp/udp rules with one concrete port; manual review required.",
			})
			continue
		}
		edges = append(edges, graph.Edge{
			From:     "internet",
			To:       resource.id,
			Kind:     "network.allow",
			Protocol: strings.TrimSpace(protocol),
			Port:     fromPort,
			Intent:   "cloudformation_security_group_ingress",
		})
	}
	return edges, ambiguous
}

func isPublicIngressRule(rule *yaml.Node) bool {
	cidr, _ := scalarString(mappingValue(rule, "CidrIp"))
	if strings.TrimSpace(cidr) == "0.0.0.0/0" {
		return true
	}
	cidr, _ = scalarString(mappingValue(rule, "CidrIpv6"))
	return strings.TrimSpace(cidr) == "::/0"
}

func appendIntrinsicAmbiguity(resource resourceNode, report *loss.Report) {
	if resource.properties == nil {
		return
	}
	seen := map[string]bool{}
	walked := 0
	var walk func(*yaml.Node, int) bool
	walk = func(node *yaml.Node, depth int) bool {
		if node == nil {
			return true
		}
		walked++
		if depth > maxIntrinsicWalkDepth || walked > maxIntrinsicWalkNodes {
			report.Ambiguous = append(report.Ambiguous, loss.Item{
				Resource: resource.id,
				Kind:     "cloudformation.intrinsic",
				Message:  "Intrinsic scan reached the current traversal budget; manual review required for remaining nested properties.",
			})
			return false
		}
		if strings.HasPrefix(node.Tag, "!") && !strings.HasPrefix(node.Tag, "!!") {
			key := node.Tag + ":" + node.Value
			if !seen[key] {
				seen[key] = true
				report.Ambiguous = append(report.Ambiguous, loss.Item{
					Resource: resource.id,
					Kind:     "cloudformation.intrinsic",
					Message:  fmt.Sprintf("Intrinsic %s is preserved but not evaluated by the current supported subset; manual review required.", node.Tag),
				})
			}
		}
		if node.Kind == yaml.MappingNode && isJSONFormIntrinsic(node) {
			keyNode := node.Content[0]
			key := keyNode.Value
			if !seen[key] {
				seen[key] = true
				report.Ambiguous = append(report.Ambiguous, loss.Item{
					Resource: resource.id,
					Kind:     "cloudformation.intrinsic",
					Message:  fmt.Sprintf("Intrinsic %s is preserved but not evaluated by the current supported subset; manual review required.", key),
				})
			}
		}
		for _, child := range node.Content {
			if !walk(child, depth+1) {
				return false
			}
		}
		return true
	}
	walk(resource.properties, 0)
}

func isJSONFormIntrinsic(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.MappingNode || len(node.Content) != 2 {
		return false
	}
	key := node.Content[0].Value
	switch key {
	case "Ref",
		"Fn::And",
		"Fn::Base64",
		"Fn::Cidr",
		"Fn::Equals",
		"Fn::FindInMap",
		"Fn::ForEach",
		"Fn::GetAtt",
		"Fn::GetAZs",
		"Fn::If",
		"Fn::ImportValue",
		"Fn::Join",
		"Fn::Length",
		"Fn::Not",
		"Fn::Or",
		"Fn::Select",
		"Fn::Split",
		"Fn::Sub",
		"Fn::ToJsonString",
		"Fn::Transform":
		return true
	default:
		return false
	}
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
	if strings.HasPrefix(node.Tag, "!") && !strings.HasPrefix(node.Tag, "!!") {
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
	return stringValue, ok
}

func scalarInt(node *yaml.Node) (int, bool) {
	value, ok := scalarValue(node)
	if !ok {
		return 0, false
	}
	intValue, ok := value.(int)
	return intValue, ok
}

func stringProperty(properties map[string]any, key string) string {
	value, ok := properties[key]
	if !ok {
		return ""
	}
	stringValue, ok := value.(string)
	if !ok {
		return ""
	}
	return stringValue
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
