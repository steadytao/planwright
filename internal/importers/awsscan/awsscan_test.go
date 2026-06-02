// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package awsscan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
)

func TestImportDirectoryLowersAWSScanBundle(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)

	result, err := ImportDirectory(dir)
	if err != nil {
		t.Fatalf("ImportDirectory() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportDirectory() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	if got, want := result.Graph.Provider, "aws"; got != want {
		t.Fatalf("Graph.Provider = %q, want %q", got, want)
	}
	if got, want := result.Graph.Region, "ap-southeast-2"; got != want {
		t.Fatalf("Graph.Region = %q, want %q", got, want)
	}
	if got, want := result.Graph.Profile, "lab"; got != want {
		t.Fatalf("Graph.Profile = %q, want %q", got, want)
	}
	for _, kind := range []string{
		"aws.vpc",
		"aws.subnet",
		"aws.security_group",
		"external.internet",
		"aws.ec2.instance",
		"aws.rds.postgres",
		"aws.s3.bucket",
		"aws.lambda.function",
		"aws.alb",
	} {
		if !hasNodeKind(result.Graph.Nodes, kind) {
			t.Fatalf("imported graph missing node kind %s: %#v", kind, result.Graph.Nodes)
		}
	}
	if !hasNetworkAllow(result.Graph.Edges, "internet", "aws/security-group/sg-123", 443) {
		t.Fatalf("graph missing public HTTPS security-group edge: %#v", result.Graph.Edges)
	}
	for _, edge := range []struct {
		from string
		to   string
	}{
		{"aws/subnet/subnet-123", "aws/vpc/vpc-123"},
		{"aws/security-group/sg-123", "aws/vpc/vpc-123"},
		{"aws/ec2-instance/i-123", "aws/subnet/subnet-123"},
		{"aws/ec2-instance/i-123", "aws/security-group/sg-123"},
		{"aws/rds/db-1", "aws/security-group/sg-123"},
		{"aws/load-balancer/app", "aws/vpc/vpc-123"},
	} {
		if !hasEdge(result.Graph.Edges, edge.from, edge.to, "depends_on") {
			t.Fatalf("graph missing dependency edge %s -> %s: %#v", edge.from, edge.to, result.Graph.Edges)
		}
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want all-protocol public ingress note")
	}
}

func TestImportDirectoryRedactsIdentityDetails(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)

	result, err := ImportDirectory(dir)
	if err != nil {
		t.Fatalf("ImportDirectory() error = %v", err)
	}
	graphData, err := json.Marshal(result.Graph)
	if err != nil {
		t.Fatalf("Marshal(graph) error = %v", err)
	}
	lossData, err := json.Marshal(result.Loss)
	if err != nil {
		t.Fatalf("Marshal(loss) error = %v", err)
	}
	for _, leaked := range []string{
		"arn:aws:iam::123456789012:user/example",
		"AIDAEXAMPLEUSERID",
	} {
		if strings.Contains(string(graphData), leaked) {
			t.Fatalf("graph leaked identity detail %q: %s", leaked, graphData)
		}
		if strings.Contains(string(lossData), leaked) {
			t.Fatalf("loss report leaked identity detail %q: %s", leaked, lossData)
		}
	}
}

func TestImportDirectoryRejectsSymlinkBundle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := ImportDirectory(link)
	if err == nil {
		t.Fatal("ImportDirectory() error = nil, want symlink rejection")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("ImportDirectory() error = %q, want symlink refusal", err.Error())
	}
}

func TestImportDirectoryReportsUnknownJSONAsPreserved(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)
	writeJSON(t, filepath.Join(dir, "future-service.json"), `{"FutureResources":[{"ID":"future-1"}]}`)

	result, err := ImportDirectory(dir)
	if err != nil {
		t.Fatalf("ImportDirectory() error = %v", err)
	}
	if !hasLossItem(result.Loss.Preserved, "future-service.json") {
		t.Fatalf("preserved = %#v, want unknown JSON file note", result.Loss.Preserved)
	}
}

func TestImportDirectoryRejectsDuplicateJSONKeys(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)
	writeJSON(t, filepath.Join(dir, "manifest.json"), `{"schema":"planwright.awsscan.v1","region":"ap-southeast-2","region":"us-east-1","profile":"lab"}`)

	_, err := ImportDirectory(dir)
	if err == nil || !strings.Contains(err.Error(), `duplicate JSON object key "region"`) {
		t.Fatalf("ImportDirectory() error = %v, want duplicate JSON key refusal", err)
	}
}

func TestImportDirectoryReportsDuplicateLoweredNodeIDs(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)
	writeJSON(t, filepath.Join(dir, "rds-describe-db-instances.json"), `{"DBInstances":[{"DBInstanceIdentifier":"db-1","Engine":"postgres","PubliclyAccessible":true,"DBInstanceStatus":"available","Endpoint":{"Port":5432}},{"DBInstanceIdentifier":"db-1","Engine":"postgres","PubliclyAccessible":false,"DBInstanceStatus":"available","Endpoint":{"Port":5432}}]}`)

	result, err := ImportDirectory(dir)
	if err != nil {
		t.Fatalf("ImportDirectory() error = %v", err)
	}
	if !hasDiagnostic(result.Diagnostics, "PW-AWSSCAN-DUPLICATE-001", "aws/rds/db-1") {
		t.Fatalf("diagnostics = %#v, want duplicate node diagnostic", result.Diagnostics)
	}
	if !hasDiagnostic(result.Diagnostics, "PW-AWS-RDS-001", "aws/rds/db-1") {
		t.Fatalf("diagnostics = %#v, want public database warning from first duplicate", result.Diagnostics)
	}
	node := findNode(result.Graph.Nodes, "aws/rds/db-1")
	if node == nil {
		t.Fatalf("RDS node missing from graph: %#v", result.Graph.Nodes)
	}
	if public, ok := graph.BoolProperty(node.Properties, "publicly_accessible"); !ok || !public {
		t.Fatalf("RDS publicly_accessible = %#v, want first duplicate retained as public", node.Properties["publicly_accessible"])
	}
}

func TestImportDirectoryRejectsTooManyJSONFiles(t *testing.T) {
	t.Parallel()

	dir := writeAWSScanBundle(t)
	for i := range 129 {
		writeJSON(t, filepath.Join(dir, "unknown-"+strconv.Itoa(i)+".json"), `{"items":[]}`)
	}

	_, err := ImportDirectory(dir)
	if err == nil || !strings.Contains(err.Error(), "too many bundle JSON files") {
		t.Fatalf("ImportDirectory() error = %v, want aggregate file-count refusal", err)
	}
}

func TestImportDirectoryRequiresKnownJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "future-service.json"), `{"FutureResources":[]}`)

	_, err := ImportDirectory(dir)
	if err == nil {
		t.Fatal("ImportDirectory() error = nil, want no known JSON error")
	}
	if !strings.Contains(err.Error(), "no known AWS scan JSON files") {
		t.Fatalf("ImportDirectory() error = %q, want no known files error", err.Error())
	}
}

func writeAWSScanBundle(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "manifest.json"), `{"schema":"planwright.awsscan.v1","account_id":"123456789012","region":"ap-southeast-2","profile":"lab"}`)
	writeJSON(t, filepath.Join(dir, "sts-get-caller-identity.json"), `{"Account":"123456789012","Arn":"arn:aws:iam::123456789012:user/example","UserId":"AIDAEXAMPLEUSERID"}`)
	writeJSON(t, filepath.Join(dir, "ec2-describe-vpcs.json"), `{"Vpcs":[{"VpcId":"vpc-123","CidrBlock":"10.0.0.0/16","IsDefault":false,"State":"available","Tags":[{"Key":"Name","Value":"demo-vpc"}]}]}`)
	writeJSON(t, filepath.Join(dir, "ec2-describe-subnets.json"), `{"Subnets":[{"SubnetId":"subnet-123","VpcId":"vpc-123","CidrBlock":"10.0.1.0/24","AvailabilityZone":"ap-southeast-2a","MapPublicIpOnLaunch":true}]}`)
	writeJSON(t, filepath.Join(dir, "ec2-describe-security-groups.json"), `{"SecurityGroups":[{"GroupId":"sg-123","GroupName":"web","Description":"web ingress","VpcId":"vpc-123","IpPermissions":[{"IpProtocol":"tcp","FromPort":443,"ToPort":443,"IpRanges":[{"CidrIp":"0.0.0.0/0"}]},{"IpProtocol":"-1","IpRanges":[{"CidrIp":"0.0.0.0/0"}]}]}]}`)
	writeJSON(t, filepath.Join(dir, "ec2-describe-instances.json"), `{"Reservations":[{"Instances":[{"InstanceId":"i-123","InstanceType":"t3.micro","VpcId":"vpc-123","SubnetId":"subnet-123","PrivateIpAddress":"10.0.1.10","PublicIpAddress":"203.0.113.10","State":{"Name":"running"},"SecurityGroups":[{"GroupId":"sg-123","GroupName":"web"}]}]}]}`)
	writeJSON(t, filepath.Join(dir, "rds-describe-db-instances.json"), `{"DBInstances":[{"DBInstanceIdentifier":"db-1","Engine":"postgres","PubliclyAccessible":false,"DBInstanceStatus":"available","Endpoint":{"Port":5432},"VpcSecurityGroups":[{"VpcSecurityGroupId":"sg-123","Status":"active"}]}]}`)
	writeJSON(t, filepath.Join(dir, "s3-list-buckets.json"), `{"Buckets":[{"Name":"planwright-example-bucket","CreationDate":"2026-05-31T00:00:00Z"}]}`)
	writeJSON(t, filepath.Join(dir, "lambda-list-functions.json"), `{"Functions":[{"FunctionName":"worker","Runtime":"provided.al2023","Handler":"bootstrap","Role":"arn:aws:iam::123456789012:role/lambda-worker","FunctionArn":"arn:aws:lambda:ap-southeast-2:123456789012:function:worker"}]}`)
	writeJSON(t, filepath.Join(dir, "elbv2-describe-load-balancers.json"), `{"LoadBalancers":[{"LoadBalancerArn":"arn:aws:elasticloadbalancing:ap-southeast-2:123456789012:loadbalancer/app/app/123","LoadBalancerName":"app","DNSName":"app.example.elb.amazonaws.com","Scheme":"internet-facing","Type":"application","VpcId":"vpc-123"}]}`)
	return dir
}

func writeJSON(t *testing.T, path string, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func hasNodeKind(nodes []graph.Node, kind string) bool {
	for _, node := range nodes {
		if node.Kind == kind {
			return true
		}
	}
	return false
}

func hasNetworkAllow(edges []graph.Edge, from string, to string, port int) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Kind == "network.allow" && edge.Protocol == "tcp" && edge.Port == port {
			return true
		}
	}
	return false
}

func hasEdge(edges []graph.Edge, from string, to string, kind string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Kind == kind {
			return true
		}
	}
	return false
}

func hasLossItem(items []loss.Item, resource string) bool {
	for _, item := range items {
		if item.Resource == resource {
			return true
		}
	}
	return false
}

func findNode(nodes []graph.Node, id string) *graph.Node {
	for index := range nodes {
		if nodes[index].ID == id {
			return &nodes[index]
		}
	}
	return nil
}

func hasDiagnostic(diagnostics []graph.Diagnostic, code string, resource string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code && diagnostic.Resource == resource {
			return true
		}
	}
	return false
}
