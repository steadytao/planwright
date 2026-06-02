// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package awsscan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
	"github.com/steadytao/planwright/internal/limits"
	"github.com/steadytao/planwright/internal/localfs"
)

const (
	maxBundleFileBytes      = 5 * 1024 * 1024
	maxBundleFiles          = 128
	maxBundleTotalFileBytes = 32 * 1024 * 1024
)

type Result struct {
	Graph       graph.Graph
	Loss        loss.Report
	Diagnostics []graph.Diagnostic
	Sources     []SourceFile
}

type SourceFile struct {
	Path string
	Size int
}

type bundleFile struct {
	name string
	path string
	data []byte
}

type manifestFile struct {
	Schema    string `json:"schema"`
	AccountID string `json:"account_id"`
	Region    string `json:"region"`
	Profile   string `json:"profile"`
}

type identityFile struct {
	Account string `json:"Account"`
	Arn     string `json:"Arn"`
	UserID  string `json:"UserId"`
}

type tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type vpcsFile struct {
	Vpcs []vpc `json:"Vpcs"`
}

type vpc struct {
	VpcID     string `json:"VpcId"`
	CidrBlock string `json:"CidrBlock"`
	IsDefault bool   `json:"IsDefault"`
	State     string `json:"State"`
	Tags      []tag  `json:"Tags"`
}

type subnetsFile struct {
	Subnets []subnet `json:"Subnets"`
}

type subnet struct {
	SubnetID            string `json:"SubnetId"`
	VpcID               string `json:"VpcId"`
	CidrBlock           string `json:"CidrBlock"`
	AvailabilityZone    string `json:"AvailabilityZone"`
	MapPublicIPOnLaunch bool   `json:"MapPublicIpOnLaunch"`
}

type securityGroupsFile struct {
	SecurityGroups []securityGroup `json:"SecurityGroups"`
}

type securityGroup struct {
	GroupID       string         `json:"GroupId"`
	GroupName     string         `json:"GroupName"`
	Description   string         `json:"Description"`
	VpcID         string         `json:"VpcId"`
	IPPermissions []ipPermission `json:"IpPermissions"`
}

type ipPermission struct {
	IPProtocol string    `json:"IpProtocol"`
	FromPort   int       `json:"FromPort"`
	ToPort     int       `json:"ToPort"`
	IPRanges   []ipRange `json:"IpRanges"`
	IPv6Ranges []ipRange `json:"Ipv6Ranges"`
}

type ipRange struct {
	CidrIP   string `json:"CidrIp"`
	CidrIPv6 string `json:"CidrIpv6"`
}

type instancesFile struct {
	Reservations []reservation `json:"Reservations"`
}

type reservation struct {
	Instances []instance `json:"Instances"`
}

type instance struct {
	InstanceID       string             `json:"InstanceId"`
	InstanceType     string             `json:"InstanceType"`
	VpcID            string             `json:"VpcId"`
	SubnetID         string             `json:"SubnetId"`
	PrivateIPAddress string             `json:"PrivateIpAddress"`
	PublicIPAddress  string             `json:"PublicIpAddress"`
	State            namedValue         `json:"State"`
	SecurityGroups   []securityGroupRef `json:"SecurityGroups"`
}

type namedValue struct {
	Name string `json:"Name"`
}

type securityGroupRef struct {
	GroupID   string `json:"GroupId"`
	GroupName string `json:"GroupName"`
}

type rdsFile struct {
	DBInstances []dbInstance `json:"DBInstances"`
}

type dbInstance struct {
	DBInstanceIdentifier string                `json:"DBInstanceIdentifier"`
	Engine               string                `json:"Engine"`
	PubliclyAccessible   bool                  `json:"PubliclyAccessible"`
	DBInstanceStatus     string                `json:"DBInstanceStatus"`
	Endpoint             dbEndpoint            `json:"Endpoint"`
	VpcSecurityGroups    []vpcSecurityGroupRef `json:"VpcSecurityGroups"`
}

type dbEndpoint struct {
	Port int `json:"Port"`
}

type vpcSecurityGroupRef struct {
	VpcSecurityGroupID string `json:"VpcSecurityGroupId"`
	Status             string `json:"Status"`
}

type s3File struct {
	Buckets []bucket `json:"Buckets"`
}

type bucket struct {
	Name         string `json:"Name"`
	CreationDate string `json:"CreationDate"`
	BucketRegion string `json:"BucketRegion"`
}

type lambdaFile struct {
	Functions []lambdaFunction `json:"Functions"`
}

type lambdaFunction struct {
	FunctionName string `json:"FunctionName"`
	Runtime      string `json:"Runtime"`
	Handler      string `json:"Handler"`
	Role         string `json:"Role"`
	FunctionArn  string `json:"FunctionArn"`
}

type loadBalancersFile struct {
	LoadBalancers []loadBalancer `json:"LoadBalancers"`
}

type loadBalancer struct {
	LoadBalancerArn  string `json:"LoadBalancerArn"`
	LoadBalancerName string `json:"LoadBalancerName"`
	DNSName          string `json:"DNSName"`
	Scheme           string `json:"Scheme"`
	Type             string `json:"Type"`
	VpcID            string `json:"VpcId"`
}

func ImportDirectory(path string) (Result, error) {
	files, sources, err := readBundle(path)
	if err != nil {
		return Result{}, err
	}
	result := Result{
		Graph: graph.Graph{
			Version:  graph.Version,
			Provider: "aws",
			Region:   "unknown",
			Profile:  "awsscan-bundle",
			Nodes:    []graph.Node{},
			Edges:    []graph.Edge{},
		},
		Loss: loss.Report{
			SourceFormat: "awsscan",
			Source:       path,
		},
		Sources: sources,
	}

	knownFiles := 0
	nodesByID := map[string]graph.Node{}
	edges := []graph.Edge{}
	edgeSeen := map[string]struct{}{}
	internetNeeded := false

	for _, file := range files {
		switch file.name {
		case "manifest.json":
			knownFiles++
			manifest, err := parseJSON[manifestFile](file)
			if err != nil {
				return Result{}, err
			}
			applyManifest(manifest, &result)
		case "sts-get-caller-identity.json":
			knownFiles++
			identity, err := parseJSON[identityFile](file)
			if err != nil {
				return Result{}, err
			}
			applyIdentity(identity, &result)
		case "ec2-describe-vpcs.json":
			knownFiles++
			document, err := parseJSON[vpcsFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.Vpcs {
				if addNode(nodesByID, vpcNode(item, result.Graph.Profile), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.VpcID,
						Kind:     "ec2.vpc",
						Message:  "VPC inventory was lowered into the Planwright graph.",
					})
				}
			}
		case "ec2-describe-subnets.json":
			knownFiles++
			document, err := parseJSON[subnetsFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.Subnets {
				if addNode(nodesByID, subnetNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.SubnetID,
						Kind:     "ec2.subnet",
						Message:  "Subnet inventory was lowered into the Planwright graph.",
					})
					appendDependency(&edges, edgeSeen, subnetID(item.SubnetID), vpcID(item.VpcID), item.SubnetID, "subnet_vpc", &result.Loss, nodesByID)
				}
			}
		case "ec2-describe-security-groups.json":
			knownFiles++
			document, err := parseJSON[securityGroupsFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.SecurityGroups {
				if addNode(nodesByID, securityGroupNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.GroupID,
						Kind:     "ec2.security_group",
						Message:  "Security group inventory was lowered into the Planwright graph.",
					})
					appendDependency(&edges, edgeSeen, securityGroupID(item.GroupID), vpcID(item.VpcID), item.GroupID, "security_group_vpc", &result.Loss, nodesByID)
					if appendPublicIngressEdges(&edges, edgeSeen, item, &result.Loss) {
						internetNeeded = true
					}
				}
			}
		case "ec2-describe-instances.json":
			knownFiles++
			document, err := parseJSON[instancesFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, reservation := range document.Reservations {
				for _, item := range reservation.Instances {
					if addNode(nodesByID, instanceNode(item), &result.Diagnostics) {
						result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
							Resource: item.InstanceID,
							Kind:     "ec2.instance",
							Message:  "EC2 instance inventory was lowered into the Planwright graph.",
						})
						appendDependency(&edges, edgeSeen, instanceID(item.InstanceID), subnetID(item.SubnetID), item.InstanceID, "instance_subnet", &result.Loss, nodesByID)
						appendDependency(&edges, edgeSeen, instanceID(item.InstanceID), vpcID(item.VpcID), item.InstanceID, "instance_vpc", &result.Loss, nodesByID)
						for _, group := range item.SecurityGroups {
							appendDependency(&edges, edgeSeen, instanceID(item.InstanceID), securityGroupID(group.GroupID), item.InstanceID, "instance_security_group", &result.Loss, nodesByID)
						}
					}
				}
			}
		case "rds-describe-db-instances.json":
			knownFiles++
			document, err := parseJSON[rdsFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.DBInstances {
				if addNode(nodesByID, dbInstanceNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.DBInstanceIdentifier,
						Kind:     "rds.db_instance",
						Message:  "RDS instance inventory was lowered into the Planwright graph.",
					})
					for _, group := range item.VpcSecurityGroups {
						appendDependency(&edges, edgeSeen, dbInstanceID(item.DBInstanceIdentifier), securityGroupID(group.VpcSecurityGroupID), item.DBInstanceIdentifier, "rds_security_group", &result.Loss, nodesByID)
					}
				}
			}
		case "s3-list-buckets.json":
			knownFiles++
			document, err := parseJSON[s3File](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.Buckets {
				if addNode(nodesByID, bucketNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.Name,
						Kind:     "s3.bucket",
						Message:  "S3 bucket inventory was lowered into the Planwright graph.",
					})
				}
			}
		case "lambda-list-functions.json":
			knownFiles++
			document, err := parseJSON[lambdaFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.Functions {
				if addNode(nodesByID, lambdaNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.FunctionName,
						Kind:     "lambda.function",
						Message:  "Lambda function inventory was lowered into the Planwright graph.",
					})
				}
			}
		case "elbv2-describe-load-balancers.json":
			knownFiles++
			document, err := parseJSON[loadBalancersFile](file)
			if err != nil {
				return Result{}, err
			}
			for _, item := range document.LoadBalancers {
				if addNode(nodesByID, loadBalancerNode(item), &result.Diagnostics) {
					result.Loss.Lowered = append(result.Loss.Lowered, loss.Item{
						Resource: item.LoadBalancerName,
						Kind:     "elbv2.load_balancer",
						Message:  "ELBv2 load balancer inventory was lowered into the Planwright graph.",
					})
					appendDependency(&edges, edgeSeen, loadBalancerID(item.LoadBalancerName), vpcID(item.VpcID), item.LoadBalancerName, "load_balancer_vpc", &result.Loss, nodesByID)
				}
			}
		default:
			if strings.EqualFold(filepath.Ext(file.name), ".json") {
				result.Loss.Preserved = append(result.Loss.Preserved, loss.Item{
					Resource: file.name,
					Kind:     "awsscan.unknown_json",
					Message:  "JSON file is preserved as a local bundle artefact but is not parsed by the current AWS scan supported subset.",
				})
			}
		}
	}

	if knownFiles == 0 {
		return Result{}, fmt.Errorf("read %s: no known AWS scan JSON files found", path)
	}
	if internetNeeded {
		addNode(nodesByID, graph.Node{ID: "internet", Kind: "external.internet", Name: "Internet"}, &result.Diagnostics)
	}
	result.Graph.Nodes = sortedNodes(nodesByID)
	sortEdges(edges)
	result.Graph.Edges = edges
	sortLossReport(&result.Loss)
	result.Diagnostics = append(result.Diagnostics, graph.Validate(result.Graph)...)
	return result, nil
}

func readBundle(path string) ([]bundleFile, []SourceFile, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil, fmt.Errorf("read AWS scan bundle: path must not be empty")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, fmt.Errorf("read %s: symlink bundle directories are not accepted", path)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("read %s: AWS scan bundle must be a directory", path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read directory %s: %w", path, err)
	}
	var files []bundleFile
	var sources []SourceFile
	totalBytes := 0
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := os.Lstat(fullPath)
		if err != nil {
			return nil, nil, fmt.Errorf("inspect %s: %w", fullPath, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, nil, fmt.Errorf("read %s: symlink bundle files are not accepted", fullPath)
		}
		if !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		if info.Size() > maxBundleFileBytes {
			return nil, nil, fmt.Errorf("read %s: bundle file exceeds %d bytes", fullPath, maxBundleFileBytes)
		}
		if len(files)+1 > maxBundleFiles {
			return nil, nil, fmt.Errorf("read %s: too many bundle JSON files; maximum is %d", path, maxBundleFiles)
		}
		data, err := localfs.ReadRegularFile(fullPath, maxBundleFileBytes)
		if err != nil {
			return nil, nil, err
		}
		totalBytes += len(data)
		if totalBytes > maxBundleTotalFileBytes {
			return nil, nil, fmt.Errorf("read %s: bundle JSON files exceed %d total bytes", path, maxBundleTotalFileBytes)
		}
		files = append(files, bundleFile{name: entry.Name(), path: fullPath, data: data})
		sources = append(sources, SourceFile{Path: fullPath, Size: len(data)})
	}
	sort.Slice(files, func(i, j int) bool {
		left := knownFileOrder(files[i].name)
		right := knownFileOrder(files[j].name)
		if left == right {
			return files[i].name < files[j].name
		}
		return left < right
	})
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Path < sources[j].Path
	})
	return files, sources, nil
}

func knownFileOrder(name string) int {
	switch name {
	case "manifest.json":
		return 0
	case "sts-get-caller-identity.json":
		return 1
	case "ec2-describe-vpcs.json":
		return 2
	case "ec2-describe-subnets.json":
		return 3
	case "ec2-describe-security-groups.json":
		return 4
	case "ec2-describe-instances.json":
		return 5
	case "rds-describe-db-instances.json":
		return 6
	case "elbv2-describe-load-balancers.json":
		return 7
	case "s3-list-buckets.json":
		return 8
	case "lambda-list-functions.json":
		return 9
	default:
		return 100
	}
}

func parseJSON[T any](file bundleFile) (T, error) {
	var out T
	if err := rejectDuplicateJSONKeys(file.data, file.path); err != nil {
		return out, err
	}
	if err := json.Unmarshal(file.data, &out); err != nil {
		return out, fmt.Errorf("parse %s: %w", file.path, err)
	}
	return out, nil
}

func rejectDuplicateJSONKeys(data []byte, sourceName string) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder, sourceName, "$"); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("parse %s: JSON contains trailing content", sourceName)
		}
		return fmt.Errorf("parse %s: %w", sourceName, err)
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder, sourceName string, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("parse %s: %w", sourceName, err)
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]struct{}{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return fmt.Errorf("parse %s: %w", sourceName, err)
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("parse %s: expected JSON object key at %s", sourceName, path)
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("parse %s: duplicate JSON object key %q at %s", sourceName, key, path)
			}
			seen[key] = struct{}{}
			if err := scanJSONValue(decoder, sourceName, path+"."+key); err != nil {
				return err
			}
		}
		end, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("parse %s: %w", sourceName, err)
		}
		if end != json.Delim('}') {
			return fmt.Errorf("parse %s: expected end of JSON object at %s", sourceName, path)
		}
	case '[':
		index := 0
		for decoder.More() {
			if err := scanJSONValue(decoder, sourceName, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
			index++
		}
		end, err := decoder.Token()
		if err != nil {
			return fmt.Errorf("parse %s: %w", sourceName, err)
		}
		if end != json.Delim(']') {
			return fmt.Errorf("parse %s: expected end of JSON array at %s", sourceName, path)
		}
	}
	return nil
}

func applyManifest(manifest manifestFile, result *Result) {
	if strings.TrimSpace(manifest.Region) != "" {
		result.Graph.Region = strings.TrimSpace(manifest.Region)
	}
	if strings.TrimSpace(manifest.Profile) != "" {
		result.Graph.Profile = strings.TrimSpace(manifest.Profile)
	}
	if strings.TrimSpace(manifest.AccountID) != "" {
		result.Loss.Preserved = append(result.Loss.Preserved, loss.Item{
			Resource: "manifest.json",
			Kind:     "awsscan.identity",
			Message:  "Bundle account ID metadata was parsed for context. Planwright does not verify account identity from local files.",
		})
	}
}

func applyIdentity(identity identityFile, result *Result) {
	if strings.TrimSpace(identity.Account) == "" {
		return
	}
	if result.Graph.Profile == "awsscan-bundle" {
		result.Graph.Profile = "account/" + strings.TrimSpace(identity.Account)
	}
	result.Loss.Preserved = append(result.Loss.Preserved, loss.Item{
		Resource: "sts-get-caller-identity.json",
		Kind:     "sts.identity",
		Message:  "STS account ID was parsed for local context. ARN and UserId are intentionally not lowered into graph properties or loss messages.",
	})
}

func vpcNode(item vpc, profile string) graph.Node {
	properties := map[string]any{
		"cidr_block": item.CidrBlock,
		"is_default": item.IsDefault,
		"state":      item.State,
	}
	if name := tagValue(item.Tags, "Name"); name != "" {
		properties["tag_name"] = name
	}
	if accountID, ok := strings.CutPrefix(profile, "account/"); ok {
		properties["account_id"] = accountID
	}
	return graph.Node{ID: vpcID(item.VpcID), Kind: "aws.vpc", Name: firstNonEmpty(tagValue(item.Tags, "Name"), item.VpcID), Properties: properties}
}

func subnetNode(item subnet) graph.Node {
	return graph.Node{
		ID:   subnetID(item.SubnetID),
		Kind: "aws.subnet",
		Name: item.SubnetID,
		Properties: map[string]any{
			"availability_zone":        item.AvailabilityZone,
			"cidr_block":               item.CidrBlock,
			"map_public_ip_on_launch":  item.MapPublicIPOnLaunch,
			"vpc_id":                   item.VpcID,
			"source_subnet_id":         item.SubnetID,
			"relationship_confidence":  "scan_reference",
			"source_relationship_vpc":  item.VpcID,
			"source_relationship_kind": "subnet_vpc",
		},
	}
}

func securityGroupNode(item securityGroup) graph.Node {
	return graph.Node{
		ID:   securityGroupID(item.GroupID),
		Kind: "aws.security_group",
		Name: firstNonEmpty(item.GroupName, item.GroupID),
		Properties: map[string]any{
			"description": item.Description,
			"group_name":  item.GroupName,
			"vpc_id":      item.VpcID,
		},
	}
}

func instanceNode(item instance) graph.Node {
	properties := map[string]any{
		"instance_type": item.InstanceType,
		"state":         item.State.Name,
		"vpc_id":        item.VpcID,
		"subnet_id":     item.SubnetID,
	}
	if item.PrivateIPAddress != "" {
		properties["private_ip_address"] = item.PrivateIPAddress
	}
	if item.PublicIPAddress != "" {
		properties["public_ip_address_present"] = true
	}
	return graph.Node{ID: instanceID(item.InstanceID), Kind: "aws.ec2.instance", Name: item.InstanceID, Properties: properties}
}

func dbInstanceNode(item dbInstance) graph.Node {
	kind := "aws.rds.instance"
	if strings.Contains(strings.ToLower(item.Engine), "postgres") {
		kind = "aws.rds.postgres"
	}
	return graph.Node{
		ID:   dbInstanceID(item.DBInstanceIdentifier),
		Kind: kind,
		Name: item.DBInstanceIdentifier,
		Properties: map[string]any{
			"engine":              item.Engine,
			"port":                item.Endpoint.Port,
			"publicly_accessible": item.PubliclyAccessible,
			"status":              item.DBInstanceStatus,
		},
	}
}

func bucketNode(item bucket) graph.Node {
	properties := map[string]any{
		"bucket_name": item.Name,
	}
	if item.CreationDate != "" {
		properties["creation_date"] = item.CreationDate
	}
	if item.BucketRegion != "" {
		properties["bucket_region"] = item.BucketRegion
	}
	return graph.Node{ID: bucketID(item.Name), Kind: "aws.s3.bucket", Name: item.Name, Properties: properties}
}

func lambdaNode(item lambdaFunction) graph.Node {
	properties := map[string]any{
		"runtime": item.Runtime,
		"handler": item.Handler,
	}
	if item.Role != "" {
		properties["role_arn_present"] = true
	}
	if item.FunctionArn != "" {
		properties["function_arn_present"] = true
	}
	return graph.Node{ID: lambdaID(item.FunctionName), Kind: "aws.lambda.function", Name: item.FunctionName, Properties: properties}
}

func loadBalancerNode(item loadBalancer) graph.Node {
	kind := "aws.alb"
	if strings.EqualFold(item.Type, "network") {
		kind = "aws.nlb"
	}
	return graph.Node{
		ID:   loadBalancerID(item.LoadBalancerName),
		Kind: kind,
		Name: item.LoadBalancerName,
		Properties: map[string]any{
			"dns_name": item.DNSName,
			"scheme":   item.Scheme,
			"type":     item.Type,
			"vpc_id":   item.VpcID,
		},
	}
}

func appendPublicIngressEdges(edges *[]graph.Edge, seen map[string]struct{}, item securityGroup, report *loss.Report) bool {
	internetNeeded := false
	for _, permission := range item.IPPermissions {
		if !hasPublicRange(permission) {
			continue
		}
		protocol := strings.ToLower(strings.TrimSpace(permission.IPProtocol))
		if protocol != "tcp" && protocol != "udp" {
			report.Ambiguous = append(report.Ambiguous, loss.Item{
				Resource: item.GroupID,
				Kind:     "ec2.security_group_ingress",
				Message:  "A public security-group ingress rule is preserved but not lowered because the current supported subset only models tcp/udp rules with one concrete port; manual review required.",
			})
			continue
		}
		if permission.FromPort < limits.MinNetworkPort || permission.FromPort > limits.MaxNetworkPort || permission.FromPort != permission.ToPort {
			report.Ambiguous = append(report.Ambiguous, loss.Item{
				Resource: item.GroupID,
				Kind:     "ec2.security_group_ingress",
				Message:  "A public security-group ingress rule is preserved but not lowered because the current supported subset only models one concrete port; manual review required.",
			})
			continue
		}
		*edges = appendUniqueEdge(*edges, seen, graph.Edge{
			From:     "internet",
			To:       securityGroupID(item.GroupID),
			Kind:     "network.allow",
			Protocol: protocol,
			Port:     permission.FromPort,
			Intent:   "awsscan_security_group_ingress",
		})
		internetNeeded = true
	}
	return internetNeeded
}

func hasPublicRange(permission ipPermission) bool {
	for _, cidr := range permission.IPRanges {
		if strings.TrimSpace(cidr.CidrIP) == "0.0.0.0/0" {
			return true
		}
	}
	for _, cidr := range permission.IPv6Ranges {
		if strings.TrimSpace(cidr.CidrIPv6) == "::/0" {
			return true
		}
	}
	return false
}

func appendDependency(edges *[]graph.Edge, seen map[string]struct{}, from string, to string, resource string, intent string, report *loss.Report, nodesByID map[string]graph.Node) {
	if from == "" || to == "" {
		return
	}
	if _, ok := nodesByID[to]; !ok {
		report.Ambiguous = append(report.Ambiguous, loss.Item{
			Resource: resource,
			Kind:     "awsscan.relationship",
			Message:  fmt.Sprintf("Relationship %s -> %s was preserved as a source reference but not lowered because the target was not present in the scan bundle.", from, to),
		})
		return
	}
	*edges = appendUniqueEdge(*edges, seen, graph.Edge{From: from, To: to, Kind: "depends_on", Intent: intent})
}

func appendUniqueEdge(edges []graph.Edge, seen map[string]struct{}, edge graph.Edge) []graph.Edge {
	key := fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%d\x00%s", edge.From, edge.To, edge.Kind, edge.Protocol, edge.Port, edge.Intent)
	if _, ok := seen[key]; ok {
		return edges
	}
	seen[key] = struct{}{}
	return append(edges, edge)
}

func addNode(nodes map[string]graph.Node, node graph.Node, diagnostics *[]graph.Diagnostic) bool {
	id := strings.TrimSpace(node.ID)
	if id == "" {
		return false
	}
	if existing, exists := nodes[id]; exists {
		*diagnostics = append(*diagnostics, graph.Diagnostic{
			Severity: graph.SeverityError,
			Code:     "PW-AWSSCAN-DUPLICATE-001",
			Resource: id,
			Message:  fmt.Sprintf("AWS scan bundle contains duplicate resources that lower to node ID %q (%s and %s).", id, existing.Kind, node.Kind),
			Fix:      "Inspect the source bundle and remove or rename the duplicate resource before trusting the imported graph.",
		})
		return false
	}
	nodes[id] = node
	return true
}

func sortedNodes(nodes map[string]graph.Node) []graph.Node {
	out := make([]graph.Node, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, node)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
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

func tagValue(tags []tag, key string) string {
	for _, item := range tags {
		if item.Key == key {
			return item.Value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func vpcID(id string) string {
	if id == "" {
		return ""
	}
	return "aws/vpc/" + id
}

func subnetID(id string) string {
	if id == "" {
		return ""
	}
	return "aws/subnet/" + id
}

func securityGroupID(id string) string {
	if id == "" {
		return ""
	}
	return "aws/security-group/" + id
}

func instanceID(id string) string {
	if id == "" {
		return ""
	}
	return "aws/ec2-instance/" + id
}

func dbInstanceID(id string) string {
	if id == "" {
		return ""
	}
	return "aws/rds/" + id
}

func bucketID(name string) string {
	if name == "" {
		return ""
	}
	return "aws/s3-bucket/" + name
}

func lambdaID(name string) string {
	if name == "" {
		return ""
	}
	return "aws/lambda-function/" + name
}

func loadBalancerID(name string) string {
	if name == "" {
		return ""
	}
	return "aws/load-balancer/" + name
}
