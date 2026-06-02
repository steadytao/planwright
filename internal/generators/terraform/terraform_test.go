// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/plan"
)

func TestRenderAWSWebAppTerraformFiles(t *testing.T) {
	t.Parallel()

	files, err := Render(loadExampleGraph(t))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, path := range []string{
		"README.md",
		"versions.tf",
		"providers.tf",
		"variables.tf",
		"network.tf",
		"security-groups.tf",
		"iam.tf",
		"observability.tf",
		"app.tf",
		"database.tf",
		"outputs.tf",
	} {
		if _, ok := artifact.ByPath(files, path); !ok {
			t.Fatalf("Render() missing %s from files %#v", path, artifact.Paths(files))
		}
	}
}

func TestRenderAWSWebAppTerraformReadmeAvoidsStaleVersionText(t *testing.T) {
	t.Parallel()

	files, err := Render(loadExampleGraph(t))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	readme, ok := artifact.ByPath(files, "README.md")
	if !ok {
		t.Fatal("README.md missing")
	}
	if strings.Contains(string(readme.Data), "v0.2") || strings.Contains(string(readme.Data), "v0.3") {
		t.Fatalf("README.md contains stale stage-specific wording: %s", string(readme.Data))
	}
	if !strings.Contains(string(readme.Data), "assign_public_ip = true") || !strings.Contains(string(readme.Data), "review/lab trade-off") {
		t.Fatalf("README.md = %s, want explicit public-subnet lab networking warning", string(readme.Data))
	}
}

func TestRenderAWSWebAppTerraformIsFmtClean(t *testing.T) {
	t.Parallel()

	_, err := exec.LookPath("terraform")
	if err != nil {
		t.Skipf("terraform not installed: %v", err)
	}

	files, err := Render(loadExampleGraph(t))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	root := t.TempDir()
	for _, file := range files {
		target := filepath.Join(root, filepath.FromSlash(file.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(target, file.Data, 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	cmd := exec.Command("terraform", "fmt", "-check", "-recursive")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("terraform fmt check failed: %v\n%s", err, output)
	}
}

func TestRenderAWSWebAppTerraformUsesSafeReviewDefaults(t *testing.T) {
	t.Parallel()

	files, err := Render(loadExampleGraph(t))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	combined := string(artifact.JoinContents(files))

	for _, forbidden := range []string{
		`password = "`,
		`access_key = "`,
		`secret_key = "`,
		"tcp/22",
		"tcp/3389",
	} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("rendered Terraform contains forbidden literal %q", forbidden)
		}
	}

	database, ok := artifact.ByPath(files, "database.tf")
	if !ok {
		t.Fatal("database.tf missing")
	}
	databaseText := string(database.Data)
	if strings.Contains(databaseText, `0.0.0.0/0`) {
		t.Fatalf("database.tf contains public database ingress: %s", databaseText)
	}
	if !strings.Contains(databaseText, "publicly_accessible    = false") {
		t.Fatalf("database.tf = %s, want publicly_accessible false", databaseText)
	}
	if !strings.Contains(databaseText, "password               = var.db_password") {
		t.Fatalf("database.tf = %s, want password variable", databaseText)
	}

	variables, ok := artifact.ByPath(files, "variables.tf")
	if !ok {
		t.Fatal("variables.tf missing")
	}
	if !strings.Contains(string(variables.Data), `variable "acm_certificate_arn"`) {
		t.Fatalf("variables.tf = %s, want ACM certificate variable", string(variables.Data))
	}

	app, ok := artifact.ByPath(files, "app.tf")
	if !ok {
		t.Fatal("app.tf missing")
	}
	appText := string(app.Data)
	if !strings.Contains(appText, "certificate_arn   = var.acm_certificate_arn") {
		t.Fatalf("app.tf = %s, want certificate_arn variable", appText)
	}
	if !strings.Contains(appText, "execution_role_arn       = aws_iam_role.ecs_task_execution.arn") {
		t.Fatalf("app.tf = %s, want ECS task execution role", appText)
	}
	if !strings.Contains(appText, "aws_cloudwatch_log_group.app.name") {
		t.Fatalf("app.tf = %s, want CloudWatch log configuration", appText)
	}
	if strings.Contains(appText, "condition     = false") {
		t.Fatalf("app.tf contains intentionally failing precondition: %s", appText)
	}

	iam, ok := artifact.ByPath(files, "iam.tf")
	if !ok {
		t.Fatal("iam.tf missing")
	}
	if !strings.Contains(string(iam.Data), "AmazonECSTaskExecutionRolePolicy") {
		t.Fatalf("iam.tf = %s, want ECS task execution policy attachment", string(iam.Data))
	}

	observability, ok := artifact.ByPath(files, "observability.tf")
	if !ok {
		t.Fatal("observability.tf missing")
	}
	if !strings.Contains(string(observability.Data), `resource "aws_cloudwatch_log_group" "app"`) {
		t.Fatalf("observability.tf = %s, want CloudWatch log group", string(observability.Data))
	}
}

func TestRenderAWSWebAppTerraformUsesTwoSubnetReviewTopology(t *testing.T) {
	t.Parallel()

	files, err := Render(loadExampleGraph(t))
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	network, ok := artifact.ByPath(files, "network.tf")
	if !ok {
		t.Fatal("network.tf missing")
	}
	networkText := string(network.Data)
	for _, want := range []string{
		`resource "aws_subnet" "public_b"`,
		`resource "aws_subnet" "private_b"`,
		"data.aws_availability_zones.available.names[1]",
	} {
		if !strings.Contains(networkText, want) {
			t.Fatalf("network.tf = %s, want %q", networkText, want)
		}
	}

	app, _ := artifact.ByPath(files, "app.tf")
	appText := string(app.Data)
	if !strings.Contains(appText, "subnets            = [aws_subnet.public_a.id, aws_subnet.public_b.id]") {
		t.Fatalf("app.tf = %s, want ALB across two public subnets", string(app.Data))
	}
	if !strings.Contains(appText, "assign_public_ip = true") {
		t.Fatalf("app.tf = %s, want low-cost public-subnet Fargate image-pull path", appText)
	}

	database, _ := artifact.ByPath(files, "database.tf")
	if !strings.Contains(string(database.Data), "subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]") {
		t.Fatalf("database.tf = %s, want database subnet group across two private subnets", string(database.Data))
	}
}

func TestRenderRejectsGraphWithoutRequiredNetworkIntent(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	g.Edges = nil

	_, err := Render(g)
	if err == nil || !strings.Contains(err.Error(), "requires an internet-facing network.allow edge") {
		t.Fatalf("Render() error = %v, want missing internet edge refusal", err)
	}
}

func TestRenderRejectsPublicDatabaseGraph(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	for index := range g.Nodes {
		if g.Nodes[index].Kind == "aws.rds.postgres" {
			g.Nodes[index].Properties["publicly_accessible"] = true
		}
	}

	_, err := Render(g)
	if err == nil || !strings.Contains(err.Error(), "does not support publicly accessible database") {
		t.Fatalf("Render() error = %v, want public database refusal", err)
	}
}

func TestRenderRejectsNonLabProfile(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	g.Profile = "production"

	_, err := Render(g)
	if err == nil || !strings.Contains(err.Error(), "supports lab profile") {
		t.Fatalf("Render() error = %v, want non-lab refusal", err)
	}
}

func TestRenderReadmeEscapesGraphMetadata(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	g.Region = "ap-southeast-2\n- forged"

	files, err := Render(g)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	readme, ok := artifact.ByPath(files, "README.md")
	if !ok {
		t.Fatal("README.md missing")
	}
	if strings.Contains(string(readme.Data), "\n- forged") {
		t.Fatalf("README.md contains forged Markdown line: %s", string(readme.Data))
	}
}

func TestRenderRejectsUnsupportedGraph(t *testing.T) {
	t.Parallel()

	_, err := Render(graph.Graph{
		Version:  graph.Version,
		Provider: "aws",
		Region:   "ap-southeast-2",
		Nodes: []graph.Node{
			{ID: "only", Kind: "aws.s3.bucket"},
		},
	})
	if err == nil {
		t.Fatal("Render() error = nil, want unsupported graph error")
	}
	if !strings.Contains(err.Error(), "requires aws.alb, aws.ecs.service and aws.rds.postgres nodes") {
		t.Fatalf("Render() error = %q, want unsupported graph error", err.Error())
	}
}

func TestRenderRejectsMultipleWebAppComponents(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	g.Nodes = append(g.Nodes, graph.Node{
		ID:   "second-app",
		Kind: "aws.ecs.service",
		Properties: map[string]any{
			"port": 8080,
		},
	})

	_, err := Render(g)
	if err == nil || !strings.Contains(err.Error(), "aws.ecs.service=2") {
		t.Fatalf("Render() error = %v, want duplicate web app component refusal", err)
	}
}

func TestRenderRejectsUnsupportedExtraNode(t *testing.T) {
	t.Parallel()

	g := loadExampleGraph(t)
	g.Nodes = append(g.Nodes, graph.Node{
		ID:   "logs",
		Kind: "aws.cloudwatch.log_group",
	})

	_, err := Render(g)
	if err == nil || !strings.Contains(err.Error(), "unsupported node kinds") {
		t.Fatalf("Render() error = %v, want unsupported extra node refusal", err)
	}
}

func loadExampleGraph(t *testing.T) graph.Graph {
	t.Helper()

	document, err := plan.Load(filepath.Join("..", "..", "..", "examples", "aws-webapp-basic", "planwright.yaml"))
	if err != nil {
		t.Fatalf("plan.Load() error = %v", err)
	}
	lowered, diagnostics := document.ToGraph()
	if graph.HasBlockingDiagnostics(diagnostics) {
		t.Fatalf("ToGraph() diagnostics = %#v", diagnostics)
	}
	return lowered
}
