// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/loss"
)

func TestImportFileLowersCloudFormationSubset(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `AWSTemplateFormatVersion: '2010-09-09'
Description: Basic CloudFormation import fixture.
Parameters:
  EnvironmentName:
    Type: String
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
  PublicSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.1.0/24
      MapPublicIpOnLaunch: true
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public HTTPS access.
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0
  LoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Type: application
      Scheme: internet-facing
  Database:
    Type: AWS::RDS::DBInstance
    Properties:
      Engine: postgres
      PubliclyAccessible: false
      Port: 5432
  Bucket:
    Type: AWS::S3::Bucket
    Properties:
      BucketName: planwright-example-bucket
  Role:
    Type: AWS::IAM::Role
    Properties:
      RoleName: planwright-example-role
  CDN:
    Type: AWS::CloudFront::Distribution
    Properties: {}
`)

	result, err := ImportFile(path, FormatCloudFormation)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	for _, kind := range []string{
		"aws.vpc",
		"aws.subnet",
		"aws.security_group",
		"aws.alb",
		"aws.rds.postgres",
		"aws.s3.bucket",
		"aws.iam.role",
		"external.internet",
	} {
		if !hasNodeKind(result.Graph.Nodes, kind) {
			t.Fatalf("imported graph missing node kind %s: %#v", kind, result.Graph.Nodes)
		}
	}
	if !hasNetworkEdge(result.Graph.Edges, "internet", "PublicSecurityGroup", 443) {
		t.Fatalf("imported graph missing public HTTPS edge: %#v", result.Graph.Edges)
	}
	if len(result.Loss.Unsupported) != 1 || result.Loss.Unsupported[0].Resource != "CDN" {
		t.Fatalf("unsupported = %#v, want CDN", result.Loss.Unsupported)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want !Ref relationship note")
	}
	assertNoStaleLossVersionText(t, result.Loss)
}

func TestImportFileLowersSAMSubset(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Transform: AWS::Serverless-2016-10-31
Resources:
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Runtime: provided.al2023
      Handler: bootstrap
  Api:
    Type: AWS::Serverless::HttpApi
    Properties:
      StageName: prod
  Table:
    Type: AWS::Serverless::SimpleTable
    Properties:
      TableName: planwright-table
`)

	result, err := ImportFile(path, FormatSAM)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	for _, kind := range []string{
		"aws.lambda.function",
		"aws.apigateway.http_api",
		"aws.dynamodb.table",
	} {
		if !hasNodeKind(result.Graph.Nodes, kind) {
			t.Fatalf("imported SAM graph missing node kind %s: %#v", kind, result.Graph.Nodes)
		}
	}
	if len(result.Loss.Unsupported) != 0 {
		t.Fatalf("unsupported = %#v, want none", result.Loss.Unsupported)
	}
	assertNoStaleLossVersionText(t, result.Loss)
}

func TestImportFileReportsUnsupportedPublicIngressAsAmbiguous(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Resources:
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public all-protocol rule.
      SecurityGroupIngress:
        - IpProtocol: -1
          CidrIp: 0.0.0.0/0
`)

	result, err := ImportFile(path, FormatCloudFormation)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	if len(result.Graph.Edges) != 0 {
		t.Fatalf("edges = %#v, want unsupported public ingress skipped", result.Graph.Edges)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want unsupported public ingress note")
	}
	assertNoStaleLossVersionText(t, result.Loss)
}

func TestImportFileReportsUnsupportedPropertiesAsPreserved(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Resources:
  Database:
    Type: AWS::RDS::DBInstance
    Properties:
      Engine: postgres
      PubliclyAccessible: false
      Port: 5432
      DeletionProtection: true
      BackupRetentionPeriod: 7
`)

	result, err := ImportFile(path, FormatCloudFormation)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	for _, want := range []string{"DeletionProtection", "BackupRetentionPeriod"} {
		if !hasLossMessageContaining(result.Loss.Preserved, "Database", "cloudformation.property", want) {
			t.Fatalf("preserved = %#v, want unsupported property note for %s", result.Loss.Preserved, want)
		}
	}
}

func TestImportFileLowersIPv6PublicIngress(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Resources:
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public IPv6 HTTPS rule.
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIpv6: ::/0
`)

	result, err := ImportFile(path, FormatCloudFormation)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	if !hasNetworkEdge(result.Graph.Edges, "internet", "PublicSecurityGroup", 443) {
		t.Fatalf("imported graph missing public IPv6 HTTPS edge: %#v", result.Graph.Edges)
	}
}

func TestImportFileReportsPublicPortRangeAsAmbiguous(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Resources:
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public range rule.
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 1
          ToPort: 65535
          CidrIp: 0.0.0.0/0
`)

	result, err := ImportFile(path, FormatCloudFormation)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	if len(result.Graph.Edges) != 0 {
		t.Fatalf("edges = %#v, want public range preserved as ambiguous", result.Graph.Edges)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want public range note")
	}
}

func TestImportFileRejectsDuplicateMappingKeys(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, `Resources:
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public HTTPS access.
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 10.0.0.0/8
          CidrIp: 0.0.0.0/0
`)

	_, err := ImportFile(path, FormatCloudFormation)
	if err == nil || !strings.Contains(err.Error(), `duplicate mapping key "CidrIp"`) {
		t.Fatalf("ImportFile() error = %v, want duplicate CidrIp refusal", err)
	}
}

func TestImportFileReportsJSONFormIntrinsicsAsAmbiguous(t *testing.T) {
	t.Parallel()

	result, err := Import([]byte(`{"Resources":{"Bucket":{"Type":"AWS::S3::Bucket","Properties":{"BucketName":{"Fn::Sub":"${Name}-bucket"}}}}}`), "template.json", FormatCloudFormation)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want JSON-form intrinsic note")
	}
	if !hasLossItem(result.Loss.Ambiguous, "Bucket", "cloudformation.intrinsic") {
		t.Fatalf("ambiguous = %#v, want Bucket intrinsic ambiguity", result.Loss.Ambiguous)
	}
}

func TestImportFileRejectsMalformedTemplate(t *testing.T) {
	t.Parallel()

	path := writeTemplate(t, "Resources:\n  Broken: [\n")
	if _, err := ImportFile(path, FormatCloudFormation); err == nil {
		t.Fatal("ImportFile() error = nil, want parse error")
	}
}

func TestImportFileRejectsOversizedTemplate(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "template.yaml")
	data := bytes.Repeat([]byte("x"), maxTemplateBytes+1)
	if err := os.WriteFile(target, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ImportFile(target, FormatCloudFormation)
	if err == nil || !strings.Contains(err.Error(), "template exceeds") {
		t.Fatalf("ImportFile() error = %v, want size refusal", err)
	}
}

func TestImportRejectsMultipleDocuments(t *testing.T) {
	t.Parallel()

	_, err := Import([]byte("Resources: {}\n---\nResources: {}\n"), "multi.yaml", FormatCloudFormation)
	if err == nil || !strings.Contains(err.Error(), "exactly one YAML document") {
		t.Fatalf("Import() error = %v, want multi-document refusal", err)
	}
}

func writeTemplate(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "template.yaml")
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func hasNodeKind(nodes []graph.Node, kind string) bool {
	for _, node := range nodes {
		if node.Kind == kind {
			return true
		}
	}
	return false
}

func hasNetworkEdge(edges []graph.Edge, from string, to string, port int) bool {
	for _, edge := range edges {
		if edge.Kind == "network.allow" && edge.From == from && edge.To == to && edge.Port == port {
			return true
		}
	}
	return false
}

func hasLossItem(items []loss.Item, resource string, kind string) bool {
	for _, item := range items {
		if item.Resource == resource && item.Kind == kind {
			return true
		}
	}
	return false
}

func hasLossMessageContaining(items []loss.Item, resource string, kind string, text string) bool {
	for _, item := range items {
		if item.Resource == resource && item.Kind == kind && strings.Contains(item.Message, text) {
			return true
		}
	}
	return false
}

func assertNoStaleLossVersionText(t *testing.T, report loss.Report) {
	t.Helper()

	for _, items := range [][]loss.Item{report.Lowered, report.Unsupported, report.Ambiguous, report.Preserved} {
		for _, item := range items {
			if strings.Contains(item.Message, "v0.2") {
				t.Fatalf("loss item contains stale version text: %#v", item)
			}
		}
	}
}
