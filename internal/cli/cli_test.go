// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/steadytao/planwright/internal/fixtures"
	"github.com/steadytao/planwright/internal/localfs"
	"github.com/steadytao/planwright/internal/reports"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(version) exitCode = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if got, want := strings.TrimSpace(stdout.String()), "planwright 0.11.0"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"future"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(unknown) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), `unknown command "future"`) {
		t.Fatalf("stderr = %q, want unknown command", stderr.String())
	}
}

func TestRunDocsCheckAcceptsCleanMarkdown(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "README.md"), []byte("Text:\n- item\n\nBehaviour matters.\n"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"docs", "check", dir}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(docs check) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	if stdout.String() != "" {
		t.Fatalf("stdout = %q, want empty output", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty output", stderr.String())
	}
}

func TestRunDocsCheckRejectsStyleFinding(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "README.md"), []byte("Text:\n\n- item\n"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"docs", "check", dir}, &stdout, &stderr)

	if exitCode != ExitValidation {
		t.Fatalf("Run(docs check invalid) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "PW-DOC-LIST-001") {
		t.Fatalf("stderr = %q, want docs style diagnostic", stderr.String())
	}
}

func TestRunValidateExamplePlan(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"validate", filepath.Join("..", "..", "examples", "aws-webapp-basic", "planwright.yaml")}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(validate) exitCode = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "validation passed") {
		t.Fatalf("stdout = %q, want validation passed", stdout.String())
	}
}

func TestRunExplainExamplePlan(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"explain", filepath.Join("..", "..", "examples", "aws-webapp-basic", "planwright.yaml")}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(explain) exitCode = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"# Planwright Graph", "- Nodes: 4", "- Edges: 3"} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout = %q, want %q", output, want)
		}
	}
}

func TestRunValidateRejectsInvalidPlan(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "invalid.yaml")
	writeTestFile(t, path, []byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
components:
  worker:
    pattern: aws.worker.future
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"validate", path}, &stdout, &stderr)

	if exitCode != ExitValidation {
		t.Fatalf("Run(validate invalid) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "PW-PLAN-PATTERN-001") {
		t.Fatalf("stdout = %q, want diagnostic code", stdout.String())
	}
}

func TestRunValidateGraphAcceptsValidGraphJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "graph.json")
	writeTestFile(t, path, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"validate-graph", path}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(validate-graph) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "graph validation passed") {
		t.Fatalf("stdout = %q, want graph validation passed", stdout.String())
	}
}

func TestRunValidateGraphRejectsSchemaViolation(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "graph.json")
	writeTestFile(t, path, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [],
	  "edges": [],
	  "unexpected": true
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"validate-graph", path}, &stdout, &stderr)

	if exitCode != ExitValidation {
		t.Fatalf("Run(validate-graph invalid) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "PW-GRAPH-SCHEMA-001") {
		t.Fatalf("stdout = %q, want schema diagnostic", stdout.String())
	}
}

func TestRunSchemaGraphWritesJSONSchema(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "planwright.graph.v1.schema.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"schema", "graph", "--out", out}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(schema graph) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	assertFileContains(t, out, `"https://json-schema.org/draft/2020-12/schema"`)
	assertFileContains(t, out, `"planwright.graph.v1"`)
}

func TestRunSchemaRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"schema"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(schema invalid) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "planwright schema graph") {
		t.Fatalf("stderr = %q, want schema usage", stderr.String())
	}
}

func TestRunPolicyProfilesListsBuiltInProfiles(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"policy", "profiles"}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(policy profiles) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	for _, want := range []string{"lab", "small-business", "production"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestRunPolicyGraphWritesMarkdownAndSARIF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	graphPath := filepath.Join(dir, "graph.json")
	reportPath := filepath.Join(dir, "policy.md")
	sarifPath := filepath.Join(dir, "policy.sarif")
	writeTestFile(t, graphPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "profile": "lab",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"},
	    {"id": "db", "kind": "aws.rds.postgres", "properties": {"publicly_accessible": false}}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"policy", "graph", graphPath, "--profile", "lab", "--out", reportPath, "--sarif", sarifPath}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(policy graph) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	assertFileContains(t, reportPath, "# Policy Profile Review")
	assertFileContains(t, reportPath, "Profile: `lab`")
	assertFileContains(t, sarifPath, `"version": "2.1.0"`)
}

func TestRunPolicyGraphReturnsPolicyExitForBlockingFindings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	graphPath := filepath.Join(dir, "graph.json")
	reportPath := filepath.Join(dir, "policy.md")
	sarifPath := filepath.Join(dir, "policy.sarif")
	writeTestFile(t, graphPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "profile": "lab",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"},
	    {"id": "db", "kind": "aws.rds.postgres", "properties": {"publicly_accessible": false}}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"policy", "graph", graphPath, "--profile", "production", "--out", reportPath, "--sarif", sarifPath}, &stdout, &stderr)

	if exitCode != ExitPolicy {
		t.Fatalf("Run(policy graph production) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitPolicy, stdout.String(), stderr.String())
	}
	assertFileContains(t, reportPath, "PW-POL-BACKUP-001")
	assertFileContains(t, sarifPath, "PW-POL-BACKUP-001")
}

func TestRunPolicyGraphRejectsWhitespacePaddedGraphIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	graphPath := filepath.Join(dir, "graph.json")
	reportPath := filepath.Join(dir, "policy.md")
	sarifPath := filepath.Join(dir, "policy.sarif")
	writeTestFile(t, graphPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "profile": "lab",
	  "nodes": [
	    {"id": " internet ", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 22}
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"policy", "graph", graphPath, "--profile", "lab", "--out", reportPath, "--sarif", sarifPath}, &stdout, &stderr)

	if exitCode != ExitValidation {
		t.Fatalf("Run(policy graph padded identity) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "PW-GRAPH-NODE-003") {
		t.Fatalf("stdout = %q, want whitespace node diagnostic", stdout.String())
	}
	if _, err := os.Stat(reportPath); !os.IsNotExist(err) {
		t.Fatalf("policy report should not be written for invalid graph, stat error = %v", err)
	}
	if _, err := os.Stat(sarifPath); !os.IsNotExist(err) {
		t.Fatalf("policy SARIF should not be written for invalid graph, stat error = %v", err)
	}
}

func TestRunPolicyRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"policy", "graph", "graph.json"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(policy invalid usage) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "planwright policy graph") {
		t.Fatalf("stderr = %q, want policy graph usage", stderr.String())
	}
}

func TestRunGenerateTerraformWritesFiles(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "terraform")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"generate", "terraform", examplePlanPath(), "--out", out}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(generate terraform) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(out, "versions.tf")); err != nil {
		t.Fatalf("versions.tf missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "security-groups.tf")); err != nil {
		t.Fatalf("security-groups.tf missing: %v", err)
	}
}

func TestRunGenerateMermaidWritesDiagram(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "diagrams")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"generate", "mermaid", examplePlanPath(), "--out", out}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(generate mermaid) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(filepath.Join(out, "architecture.mmd")); err != nil {
		t.Fatalf("architecture.mmd missing: %v", err)
	}
}

func TestRunRisksPrintsSecurityReport(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"risks", examplePlanPath()}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(risks) exitCode = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Security Report") {
		t.Fatalf("stdout = %q, want security report", stdout.String())
	}
}

func TestRunCostNotesPrintsCostReport(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"cost-notes", examplePlanPath()}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(cost-notes) exitCode = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Cost Notes") {
		t.Fatalf("stdout = %q, want cost notes", stdout.String())
	}
}

func TestRunAWSWebappProofPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	terraformOut := filepath.Join(root, "terraform")
	diagramsOut := filepath.Join(root, "diagrams")
	packOut := filepath.Join(root, "pack")

	for _, step := range []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "validate",
			args: []string{"validate", examplePlanPath()},
			want: []string{"validation passed"},
		},
		{
			name: "risks",
			args: []string{"risks", examplePlanPath()},
			want: []string{"# Security Report", "No public database exposure was detected"},
		},
		{
			name: "cost-notes",
			args: []string{"cost-notes", examplePlanPath()},
			want: []string{"# Cost Notes", "NAT Gateway is not generated in the lab profile"},
		},
		{
			name: "generate terraform",
			args: []string{"generate", "terraform", examplePlanPath(), "--out", terraformOut},
			want: []string{"wrote Terraform/OpenTofu files to"},
		},
		{
			name: "generate mermaid",
			args: []string{"generate", "mermaid", examplePlanPath(), "--out", diagramsOut},
			want: []string{"wrote Mermaid files to"},
		},
		{
			name: "pack",
			args: []string{"pack", examplePlanPath(), "--out", packOut},
			want: []string{"wrote Planwright pack to"},
		},
	} {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		exitCode := Run(context.Background(), step.args, &stdout, &stderr)
		if exitCode != ExitOK {
			t.Fatalf("Run(%s) exitCode = %d, want %d; stdout=%q stderr=%q", step.name, exitCode, ExitOK, stdout.String(), stderr.String())
		}
		for _, want := range step.want {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("Run(%s) stdout = %q, want %q", step.name, stdout.String(), want)
			}
		}
	}

	for _, path := range []string{
		filepath.Join(terraformOut, "README.md"),
		filepath.Join(terraformOut, "network.tf"),
		filepath.Join(terraformOut, "security-groups.tf"),
		filepath.Join(terraformOut, "database.tf"),
		filepath.Join(diagramsOut, "architecture.mmd"),
		filepath.Join(packOut, "manifest.json"),
		filepath.Join(packOut, "planwright.graph.json"),
		filepath.Join(packOut, "reports", "security-report.md"),
		filepath.Join(packOut, "reports", "cost-notes.md"),
		filepath.Join(packOut, "reports", "deployability-report.md"),
		filepath.Join(packOut, "reports", "cleanup.md"),
		filepath.Join(packOut, "generated", "terraform", "README.md"),
		filepath.Join(packOut, "diagrams", "architecture.mmd"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("proof-path file %s missing: %v", path, err)
		}
	}

	assertFileContains(t, filepath.Join(packOut, "manifest.json"), `"schema": "planwright.pack.v1"`)
	assertFileContains(t, filepath.Join(packOut, "manifest.json"), `"requires_review": true`)
}

func TestRunAWSWebappPublicDatabaseRiskFixture(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "public-db.yaml")
	writeTestFile(t, path, []byte(`
version: planwright.v1
provider: aws
region: ap-southeast-2
profile: lab

components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      app_port: 8080
      db_engine: postgres
      db_public: true

flows:
  - from: internet
    to: webapp.alb
    kind: network.allow
    protocol: tcp
    port: 443
    intent: public_https_entrypoint

  - from: webapp.alb
    to: webapp.app
    kind: network.allow
    protocol: tcp
    port: 8080
    intent: load_balancer_to_app

  - from: webapp.app
    to: webapp.db
    kind: network.allow
    protocol: tcp
    port: 5432
    intent: application_database_access
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"risks", path}, &stdout, &stderr)
	if exitCode != ExitOK {
		t.Fatalf("Run(risks public db) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "HIGH PW-AWS-RDS-001") {
		t.Fatalf("stdout = %q, want public database finding", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Run(context.Background(), []string{"generate", "terraform", path, "--out", filepath.Join(t.TempDir(), "terraform")}, &stdout, &stderr)
	if exitCode != ExitValidation {
		t.Fatalf("Run(generate terraform public db) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "does not support publicly accessible database") {
		t.Fatalf("stderr = %q, want public database generator refusal", stderr.String())
	}
}

func TestRunExampleCompatibilityFixtures(t *testing.T) {
	t.Parallel()

	metadata, err := fixtures.Discover(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("Discover(examples) error = %v", err)
	}
	if len(metadata) == 0 {
		t.Fatalf("Discover(examples) returned no fixtures")
	}

	for _, fixture := range metadata {
		t.Run(fixture.ID, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			for _, command := range fixture.Commands {
				var stdout bytes.Buffer
				var stderr bytes.Buffer
				exitCode := Run(context.Background(), command.ExpandArgs(fixture, tempDir), &stdout, &stderr)
				if exitCode != command.WantExit {
					t.Fatalf("Run(%s/%s) exitCode = %d, want %d; stdout=%q stderr=%q", fixture.ID, command.Name, exitCode, command.WantExit, stdout.String(), stderr.String())
				}
				for _, want := range command.WantStdoutContains {
					if !strings.Contains(stdout.String(), want) {
						t.Fatalf("Run(%s/%s) stdout = %q, want %q", fixture.ID, command.Name, stdout.String(), want)
					}
				}
				for _, want := range command.WantStderrContains {
					if !strings.Contains(stderr.String(), want) {
						t.Fatalf("Run(%s/%s) stderr = %q, want %q", fixture.ID, command.Name, stderr.String(), want)
					}
				}
				for _, path := range command.ExpectedFiles(tempDir) {
					if _, err := os.Stat(path); err != nil {
						t.Fatalf("Run(%s/%s) expected file %s missing: %v", fixture.ID, command.Name, path, err)
					}
				}
				for _, path := range command.ExpectedSARIFFiles(tempDir) {
					data, err := localfs.ReadRegularFile(path, 10*1024*1024)
					if err != nil {
						t.Fatalf("Run(%s/%s) expected SARIF file %s unreadable: %v", fixture.ID, command.Name, path, err)
					}
					if err := reports.ValidateSARIF(data, path); err != nil {
						t.Fatalf("Run(%s/%s) expected SARIF file %s invalid: %v", fixture.ID, command.Name, path, err)
					}
				}
				for _, expectation := range command.ExpectedFileContents(tempDir) {
					data, err := localfs.ReadRegularFile(expectation.Path, 1024*1024)
					if err != nil {
						t.Fatalf("Run(%s/%s) expected file %s unreadable: %v", fixture.ID, command.Name, expectation.Path, err)
					}
					for _, want := range expectation.Contains {
						if !strings.Contains(string(data), want) {
							t.Fatalf("Run(%s/%s) file %s = %q, want %q", fixture.ID, command.Name, expectation.Path, string(data), want)
						}
					}
				}
			}
		})
	}
}

func TestRunPackWritesDeploymentPack(t *testing.T) {
	t.Parallel()

	out := filepath.Join(t.TempDir(), "pack")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"pack", examplePlanPath(), "--out", out}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(pack) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	for _, path := range []string{
		filepath.Join(out, "manifest.json"),
		filepath.Join(out, "generated", "terraform", "versions.tf"),
		filepath.Join(out, "reports", "security-report.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("pack file %s missing: %v", path, err)
		}
	}
}

func TestRunImportCloudFormationWritesGraphAndLossReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	template := filepath.Join(dir, "template.yaml")
	graphPath := filepath.Join(dir, "graph.json")
	lossPath := filepath.Join(dir, "loss.md")
	writeTestFile(t, template, []byte(`Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
  PublicSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Public HTTPS access.
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0
  CDN:
    Type: AWS::CloudFront::Distribution
    Properties: {}
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "cloudformation", template, "--out", graphPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(import cloudformation) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	graphData, err := localfs.ReadRegularFile(graphPath, 1024*1024)
	if err != nil {
		t.Fatalf("ReadFile(graph) error = %v", err)
	}
	if !strings.Contains(string(graphData), `"aws.vpc"`) {
		t.Fatalf("graph output = %s, want aws.vpc", graphData)
	}
	lossData, err := localfs.ReadRegularFile(lossPath, 1024*1024)
	if err != nil {
		t.Fatalf("ReadFile(loss) error = %v", err)
	}
	if !strings.Contains(string(lossData), "# Loss Report") || !strings.Contains(string(lossData), "AWS::CloudFront::Distribution") {
		t.Fatalf("loss output = %s, want loss report with unsupported type", lossData)
	}
}

func TestRunImportRejectsDuplicateOutputPaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	template := filepath.Join(dir, "template.yaml")
	outputPath := filepath.Join(dir, "same.out")
	writeTestFile(t, template, []byte(`Resources:
  Bucket:
    Type: AWS::S3::Bucket
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "cloudformation", template, "--out", outputPath, "--loss-report", outputPath}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(import duplicate outputs) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitUsage, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "must be distinct") {
		t.Fatalf("stderr = %q, want distinct-path error", stderr.String())
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("duplicate output path should not be written, stat error = %v", err)
	}
}

func TestRunImportRejectsHardLinkedInputOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	template := filepath.Join(dir, "template.yaml")
	graphPath := filepath.Join(dir, "graph.json")
	lossPath := filepath.Join(dir, "loss.md")
	source := []byte(`Resources:
  Bucket:
    Type: AWS::S3::Bucket
`)
	writeTestFile(t, template, source)
	if err := os.Link(template, graphPath); err != nil {
		t.Skipf("cannot create hard link: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "cloudformation", template, "--out", graphPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(import hard-linked input/output) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitUsage, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "must be distinct") {
		t.Fatalf("stderr = %q, want distinct-path error", stderr.String())
	}
	data, err := localfs.ReadRegularFile(template, 1024)
	if err != nil {
		t.Fatalf("ReadFile(template) error = %v", err)
	}
	if string(data) != string(source) {
		t.Fatalf("template changed through hard-linked output: %q", data)
	}
	if _, err := os.Stat(lossPath); !os.IsNotExist(err) {
		t.Fatalf("loss report should not be written, stat error = %v", err)
	}
}

func TestRunImportSAMWritesGraphAndLossReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	template := filepath.Join(dir, "template.yaml")
	graphPath := filepath.Join(dir, "sam.graph.json")
	lossPath := filepath.Join(dir, "sam-loss.md")
	writeTestFile(t, template, []byte(`Transform: AWS::Serverless-2016-10-31
Resources:
  Function:
    Type: AWS::Serverless::Function
    Properties:
      Runtime: provided.al2023
      Handler: bootstrap
  Api:
    Type: AWS::Serverless::HttpApi
  Table:
    Type: AWS::Serverless::SimpleTable
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "sam", template, "--out", graphPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(import sam) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	graphData, err := localfs.ReadRegularFile(graphPath, 1024*1024)
	if err != nil {
		t.Fatalf("ReadFile(graph) error = %v", err)
	}
	for _, want := range []string{`"aws.lambda.function"`, `"aws.apigateway.http_api"`, `"aws.dynamodb.table"`} {
		if !strings.Contains(string(graphData), want) {
			t.Fatalf("graph output = %s, want %s", graphData, want)
		}
	}
	if _, err := os.Stat(lossPath); err != nil {
		t.Fatalf("loss report missing: %v", err)
	}
}

func TestRunImportKubernetesWritesGraphAndLossReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	manifest := filepath.Join(dir, "manifests.yaml")
	graphPath := filepath.Join(dir, "kubernetes.graph.json")
	lossPath := filepath.Join(dir, "kubernetes-loss.md")
	writeTestFile(t, manifest, []byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: demo
spec:
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: app
          image: example/app:1.0
`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "k8s", manifest, "--out", graphPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(import k8s) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	assertFileContains(t, graphPath, `"k8s.deployment"`)
	assertFileContains(t, lossPath, "# Loss Report")
}

func TestRunImportAWSScanWritesGraphAndLossReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bundle := filepath.Join(dir, "bundle")
	if err := os.Mkdir(bundle, 0o700); err != nil {
		t.Fatalf("Mkdir(bundle) error = %v", err)
	}
	writeTestFile(t, filepath.Join(bundle, "manifest.json"), []byte(`{"schema":"planwright.awsscan.v1","region":"ap-southeast-2","profile":"lab"}`))
	writeTestFile(t, filepath.Join(bundle, "ec2-describe-vpcs.json"), []byte(`{"Vpcs":[{"VpcId":"vpc-123","CidrBlock":"10.0.0.0/16","IsDefault":false,"State":"available"}]}`))

	graphPath := filepath.Join(dir, "awsscan.graph.json")
	lossPath := filepath.Join(dir, "awsscan-loss.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "awsscan", bundle, "--out", graphPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(import awsscan) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	assertFileContains(t, graphPath, `"aws.vpc"`)
	assertFileContains(t, lossPath, "# Loss Report")
}

func TestRunDiffWritesGraphDiffReview(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.graph.json")
	newPath := filepath.Join(dir, "new.graph.json")
	reviewPath := filepath.Join(dir, "diff.md")
	writeTestFile(t, oldPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": []
	}`))
	writeTestFile(t, newPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [
	    {"id": "internet", "kind": "external.internet"},
	    {"id": "app", "kind": "aws.ecs.service"}
	  ],
	  "edges": [
	    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 22}
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"diff", oldPath, newPath, "--out", reviewPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(diff) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	assertFileContains(t, reviewPath, "# Graph Diff Review")
	assertFileContains(t, reviewPath, "PW-DIFF-NET-001")
}

func TestRunDiffRejectsInvalidGraph(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.graph.json")
	newPath := filepath.Join(dir, "new.graph.json")
	reviewPath := filepath.Join(dir, "diff.md")
	writeTestFile(t, oldPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [{"id": "internet", "kind": "external.internet"}],
	  "edges": []
	}`))
	writeTestFile(t, newPath, []byte(`{
	  "version": "planwright.graph.v1",
	  "provider": "aws",
	  "region": "ap-southeast-2",
	  "nodes": [{"id": "app", "kind": "aws.ecs.service"}],
	  "edges": [{"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 22}]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"diff", oldPath, newPath, "--out", reviewPath}, &stdout, &stderr)

	if exitCode != ExitValidation {
		t.Fatalf("Run(diff invalid) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitValidation, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "PW-GRAPH-EDGE-001") {
		t.Fatalf("stdout = %q, want graph validation diagnostic", stdout.String())
	}
	if _, err := os.Stat(reviewPath); !os.IsNotExist(err) {
		t.Fatalf("diff review should not be written for invalid graph, stat error = %v", err)
	}
}

func TestRunDiffRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"diff", "old.graph.json"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(diff invalid usage) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "planwright diff") {
		t.Fatalf("stderr = %q, want diff usage", stderr.String())
	}
}

func TestRunImportHelpListsCurrentImporters(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "--help"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(import --help) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{
		"planwright import cloudformation",
		"planwright import sam",
		"planwright import k8s",
		"planwright import awsscan",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("import help = %q, want %q", output, want)
		}
	}
}

func TestRunImportRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"import", "cloudformation"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(import invalid) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "import cloudformation") {
		t.Fatalf("stderr = %q, want import usage", stderr.String())
	}
}

func TestRunReviewTerraformPlanWritesMarkdownAndSARIF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	planPath := filepath.Join(dir, "tfplan.json")
	reviewPath := filepath.Join(dir, "review.md")
	sarifPath := filepath.Join(dir, "planwright.sarif")
	writeTestFile(t, planPath, []byte(`{
	  "format_version": "1.0",
	  "terraform_version": "1.15.5",
	  "resource_changes": [
	    {
	      "address": "aws_s3_bucket.logs",
	      "mode": "managed",
	      "type": "aws_s3_bucket",
	      "name": "logs",
	      "change": {"actions": ["delete"]}
	    }
	  ]
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"review", "terraform-plan", planPath, "--out", reviewPath, "--sarif", sarifPath}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Run(review terraform-plan) exitCode = %d, want 0; stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
	}
	assertFileContains(t, reviewPath, "# Terraform Plan Review")
	assertFileContains(t, sarifPath, `"version": "2.1.0"`)
}

func TestRunReviewTerraformPlanRejectsDuplicateOutputPaths(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	planPath := filepath.Join(dir, "tfplan.json")
	outputPath := filepath.Join(dir, "same.out")
	writeTestFile(t, planPath, []byte(`{"format_version":"1.0","resource_changes":[]}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"review", "terraform-plan", planPath, "--out", outputPath, "--sarif", outputPath}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(review duplicate outputs) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitUsage, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "must be distinct") {
		t.Fatalf("stderr = %q, want distinct-path error", stderr.String())
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("duplicate review output path should not be written, stat error = %v", err)
	}
}

func TestRunReviewTerraformStateWritesInventoryAndLossReport(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	statePath := filepath.Join(dir, "state.json")
	inventoryPath := filepath.Join(dir, "inventory.md")
	lossPath := filepath.Join(dir, "loss.md")
	writeTestFile(t, statePath, []byte(`{
	  "format_version": "1.0",
	  "terraform_version": "1.15.5",
	  "values": {
	    "root_module": {
	      "resources": [
	        {
	          "address": "aws_db_instance.app",
	          "mode": "managed",
	          "type": "aws_db_instance",
	          "name": "app",
	          "provider_name": "registry.terraform.io/hashicorp/aws",
	          "values": {"password": "super-secret-password"},
	          "sensitive_values": {"password": true}
	        }
	      ]
	    }
	  }
	}`))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"review", "terraform-state", statePath, "--out", inventoryPath, "--loss-report", lossPath}, &stdout, &stderr)

	if exitCode != ExitOK {
		t.Fatalf("Run(review terraform-state) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
	}
	assertFileContains(t, inventoryPath, "# Terraform State Inventory")
	assertFileContains(t, inventoryPath, "`aws_db_instance.app`")
	assertFileContains(t, inventoryPath, "`password`")
	assertFileContains(t, lossPath, "# Loss Report")
	for _, path := range []string{inventoryPath, lossPath} {
		data, err := localfs.ReadRegularFile(path, 1024*1024)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(data), "super-secret-password") {
			t.Fatalf("%s leaked sensitive state value: %s", path, data)
		}
	}
}

func TestRunReviewRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"review", "terraform-plan"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(review invalid) exitCode = %d, want %d", exitCode, ExitUsage)
	}
}

func TestRunServeRejectsInvalidUsage(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(context.Background(), []string{"serve", ".", "--addr"}, &stdout, &stderr)

	if exitCode != ExitUsage {
		t.Fatalf("Run(serve invalid) exitCode = %d, want %d", exitCode, ExitUsage)
	}
	if !strings.Contains(stderr.String(), "planwright serve") {
		t.Fatalf("stderr = %q, want serve usage", stderr.String())
	}
}

func TestRunServePrintsCapitalisedAddress(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stdout := newNotifyingWriter()
	var stderr bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- Run(ctx, []string{"serve", t.TempDir(), "--addr", "127.0.0.1:0"}, stdout, &stderr)
	}()

	select {
	case <-stdout.written:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatalf("serve did not write startup output; stderr=%q", stderr.String())
	}

	cancel()
	select {
	case exitCode := <-done:
		if exitCode != ExitOK {
			t.Fatalf("Run(serve) exitCode = %d, want %d; stdout=%q stderr=%q", exitCode, ExitOK, stdout.String(), stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("serve did not stop after context cancellation; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Serving Planwright on http://127.0.0.1:") {
		t.Fatalf("stdout = %q, want capitalised serve message", output)
	}
	if strings.Contains(output, "serving Planwright") {
		t.Fatalf("stdout = %q, did not want lowercase serve message", output)
	}
}

func examplePlanPath() string {
	return filepath.Join("..", "..", "examples", "aws-webapp-basic", "planwright.yaml")
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()

	data, err := localfs.ReadRegularFile(path, 1024*1024)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s = %s, want %q", path, data, want)
	}
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

type notifyingWriter struct {
	mu      sync.Mutex
	buffer  bytes.Buffer
	written chan struct{}
	once    sync.Once
}

func newNotifyingWriter() *notifyingWriter {
	return &notifyingWriter{written: make(chan struct{})}
}

func (w *notifyingWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.buffer.Write(data)
	w.once.Do(func() {
		close(w.written)
	})
	return n, err
}

func (w *notifyingWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buffer.String()
}
