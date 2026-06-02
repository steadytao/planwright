// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/steadytao/planwright/internal/docstyle"
	"github.com/steadytao/planwright/internal/generators/mermaid"
	terraformgen "github.com/steadytao/planwright/internal/generators/terraform"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/importers/awsscan"
	"github.com/steadytao/planwright/internal/importers/cloudformation"
	"github.com/steadytao/planwright/internal/importers/kubernetes"
	"github.com/steadytao/planwright/internal/importers/loss"
	"github.com/steadytao/planwright/internal/localfs"
	"github.com/steadytao/planwright/internal/plan"
	"github.com/steadytao/planwright/internal/policy"
	"github.com/steadytao/planwright/internal/project"
	"github.com/steadytao/planwright/internal/reports"
	"github.com/steadytao/planwright/internal/review/terraformplan"
	cloudserver "github.com/steadytao/planwright/internal/server"
	"github.com/steadytao/planwright/internal/version"
)

const Version = version.Number

const maxGraphJSONSize = 10 * 1024 * 1024

const (
	ExitOK         = 0
	ExitInternal   = 1
	ExitValidation = 2
	ExitPolicy     = 3
	ExitUsage      = 4
	ExitUnsafe     = 5
)

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	checkedStdout := &checkedWriter{target: stdout}
	checkedStderr := &checkedWriter{target: stderr}
	exitCode := run(ctx, args, checkedStdout, checkedStderr)
	if checkedStdout.err != nil || checkedStderr.err != nil {
		return ExitInternal
	}
	return exitCode
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return ExitUsage
	}

	command, rest := splitCommandArgs(args)
	switch command {
	case "validate":
		return runValidate(rest, stdout, stderr)
	case "validate-graph":
		return runValidateGraph(rest, stdout, stderr)
	case "explain":
		return runExplain(rest, stdout, stderr)
	case "generate":
		return runGenerate(rest, stdout, stderr)
	case "risks":
		return runRisks(rest, stdout, stderr)
	case "cost-notes":
		return runCostNotes(rest, stdout, stderr)
	case "docs":
		return runDocs(rest, stdout, stderr)
	case "import":
		return runImport(rest, stdout, stderr)
	case "diff":
		return runDiff(rest, stdout, stderr)
	case "schema":
		return runSchema(rest, stdout, stderr)
	case "policy":
		return runPolicy(rest, stdout, stderr)
	case "pack":
		return runPack(rest, stdout, stderr)
	case "review":
		return runReview(rest, stdout, stderr)
	case "serve":
		return runServe(ctx, rest, stdout, stderr)
	case "version":
		return runVersion(rest, stdout, stderr)
	case "-h", "--help", "help":
		printUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown command %q\n\n", command)
		printUsage(stderr)
		return ExitUsage
	}
}

type checkedWriter struct {
	target io.Writer
	err    error
}

func (writer *checkedWriter) Write(data []byte) (int, error) {
	if writer.err != nil {
		return 0, writer.err
	}
	n, err := writer.target.Write(data)
	if err != nil {
		writer.err = err
	}
	return n, err
}

func write(writer io.Writer, values ...any) {
	_, err := fmt.Fprint(writer, values...)
	recordWriteError(writer, err)
}

func writef(writer io.Writer, format string, values ...any) {
	_, err := fmt.Fprintf(writer, format, values...)
	recordWriteError(writer, err)
}

func writeln(writer io.Writer, values ...any) {
	_, err := fmt.Fprintln(writer, values...)
	recordWriteError(writer, err)
}

func recordWriteError(writer io.Writer, err error) {
	if err == nil {
		return
	}
	checked, ok := writer.(*checkedWriter)
	if ok {
		checked.err = err
	}
}

func splitCommandArgs(args []string) (string, []string) {
	var command string
	rest := make([]string, 0, max(len(args)-1, 0))
	for index, arg := range args {
		if index == 0 {
			command = arg
			continue
		}
		rest = append(rest, arg)
	}
	return command, rest
}

func runDocs(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "docs requires a target")
		printDocsUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "check":
		return runDocsCheck(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printDocsUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown docs target %q\n\n", args[0])
		printDocsUsage(stderr)
		return ExitUsage
	}
}

func runDocsCheck(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 1 {
		for _, arg := range args {
			switch arg {
			case "-h", "--help", "help":
				printDocsCheckUsage(stdout)
				return ExitOK
			}
		}
	}
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}
	findings, err := docstyle.CheckPaths(paths)
	if err != nil {
		writef(stderr, "docs check: %v\n", err)
		return ExitValidation
	}
	if len(findings) > 0 {
		write(stderr, docstyle.FormatFindings(findings))
		return ExitValidation
	}
	return ExitOK
}

func runDiff(args []string, stdout io.Writer, stderr io.Writer) int {
	oldPath, newPath, outPath, ok := parseDiffArgs(args, stderr)
	if !ok {
		printDiffUsage(stderr)
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"old graph", oldPath}, namedPath{"new graph", newPath}, namedPath{"output", outPath}) {
		return ExitUsage
	}
	oldGraph, oldDiagnostics, err := loadGraphJSON(oldPath)
	if err != nil {
		writef(stderr, "load old graph: %v\n", err)
		return ExitValidation
	}
	newGraph, newDiagnostics, err := loadGraphJSON(newPath)
	if err != nil {
		writef(stderr, "load new graph: %v\n", err)
		return ExitValidation
	}

	if len(oldDiagnostics) > 0 {
		writeln(stdout, "Old graph diagnostics:")
		write(stdout, reports.RenderDiagnostics(oldDiagnostics))
	}
	if len(newDiagnostics) > 0 {
		writeln(stdout, "New graph diagnostics:")
		write(stdout, reports.RenderDiagnostics(newDiagnostics))
	}
	if graph.HasBlockingDiagnostics(oldDiagnostics) || graph.HasBlockingDiagnostics(newDiagnostics) {
		return ExitValidation
	}

	diff := graph.Compare(oldGraph, newGraph)
	if err := project.WriteFile(outPath, []byte(reports.RenderGraphDiff(diff, oldPath, newPath))); err != nil {
		writef(stderr, "write graph diff review: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote graph diff review to %s\n", outPath)
	return ExitOK
}

func runValidateGraph(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		writeln(stderr, "validate-graph requires exactly one graph JSON path")
		printValidateGraphUsage(stderr)
		return ExitUsage
	}
	data, err := readGraphJSONFile(args[0])
	if err != nil {
		writef(stderr, "load graph: %v\n", err)
		return ExitValidation
	}
	_, diagnostics := graph.ValidateJSON(data, args[0])
	if len(diagnostics) > 0 {
		write(stdout, reports.RenderDiagnostics(diagnostics))
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		writeln(stdout, "graph validation failed")
		return ExitValidation
	}
	writeln(stdout, "graph validation passed")
	return ExitOK
}

func runSchema(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "schema requires a target")
		printSchemaUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "graph":
		return runSchemaGraph(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printSchemaUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown schema target %q\n\n", args[0])
		printSchemaUsage(stderr)
		return ExitUsage
	}
}

func runSchemaGraph(args []string, stdout io.Writer, stderr io.Writer) int {
	outPath, ok := parseSchemaGraphArgs(args, stderr)
	if !ok {
		printSchemaGraphUsage(stderr)
		return ExitUsage
	}
	if err := project.WriteFile(outPath, graph.JSONSchema()); err != nil {
		writef(stderr, "write graph schema: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote Planwright graph schema to %s\n", outPath)
	return ExitOK
}

func runPolicy(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "policy requires a target")
		printPolicyUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "profiles":
		return runPolicyProfiles(args[1:], stdout, stderr)
	case "graph":
		return runPolicyGraph(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printPolicyUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown policy target %q\n\n", args[0])
		printPolicyUsage(stderr)
		return ExitUsage
	}
}

func runPolicyProfiles(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 0 {
		writeln(stderr, "policy profiles does not accept arguments")
		printPolicyProfilesUsage(stderr)
		return ExitUsage
	}
	writeln(stdout, "Built-in policy profiles:")
	for _, profile := range policy.Profiles() {
		writef(stdout, "- %s: %s\n", profile.ID, profile.Description)
	}
	return ExitOK
}

func runPolicyGraph(args []string, stdout io.Writer, stderr io.Writer) int {
	graphPath, profileID, reportPath, sarifPath, ok := parsePolicyGraphArgs(args, stderr)
	if !ok {
		printPolicyGraphUsage(stderr)
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"graph", graphPath}, namedPath{"report", reportPath}, namedPath{"SARIF", sarifPath}) {
		return ExitUsage
	}
	loaded, diagnostics, err := loadGraphJSON(graphPath)
	if err != nil {
		writef(stderr, "load graph: %v\n", err)
		return ExitValidation
	}
	if len(diagnostics) > 0 {
		write(stdout, reports.RenderDiagnostics(diagnostics))
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		return ExitValidation
	}
	result, err := policy.Evaluate(loaded, profileID)
	if err != nil {
		writef(stderr, "policy graph: %v\n", err)
		return ExitUsage
	}
	result.Source = graphPath
	if err := project.WriteFile(reportPath, []byte(reports.RenderPolicy(result))); err != nil {
		writef(stderr, "write policy report: %v\n", err)
		return ExitUnsafe
	}
	sarifData, err := reports.RenderPolicySARIF(result)
	if err != nil {
		writef(stderr, "render policy SARIF: %v\n", err)
		return ExitInternal
	}
	if err := project.WriteFile(sarifPath, sarifData); err != nil {
		writef(stderr, "write policy SARIF: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote policy profile review to %s\n", reportPath)
	writef(stdout, "wrote policy SARIF to %s\n", sarifPath)
	if policy.HasBlockingFindings(result.Findings) {
		writef(stdout, "policy profile %s failed\n", profileID)
		return ExitPolicy
	}
	writef(stdout, "policy profile %s passed\n", profileID)
	return ExitOK
}

func runServe(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		printServeUsage(stdout)
		return ExitOK
	}
	projectDir, addr, ok := parseServeArgs(args, stderr)
	if !ok {
		printServeUsage(stderr)
		return ExitUsage
	}
	if !isLoopbackAddr(addr) {
		writef(stderr, "serve address must use a loopback host such as 127.0.0.1 or localhost: %s\n", addr)
		return ExitUnsafe
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		writef(stderr, "start server: %v\n", err)
		return ExitInternal
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			writef(stderr, "close server listener: %v\n", err)
		}
	}()

	runCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	httpServer := &http.Server{
		Handler:           cloudserver.New(cloudserver.Options{ProjectDir: projectDir, AllowedHosts: cloudserver.DefaultAllowedHosts(listener.Addr().String())}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- httpServer.Serve(listener)
	}()

	writef(stdout, "Serving Planwright on http://%s\n", listener.Addr().String())

	select {
	case <-runCtx.Done():
	case err := <-errc:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			writef(stderr, "serve: %v\n", err)
			return ExitInternal
		}
		return ExitOK
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		writef(stderr, "stop server: %v\n", err)
		return ExitInternal
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		writef(stderr, "serve: %v\n", err)
		return ExitInternal
	}
	return ExitOK
}

func runReview(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "review requires a target")
		printReviewUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "terraform-plan":
		return runReviewTerraformPlan(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printReviewUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown review target %q\n\n", args[0])
		printReviewUsage(stderr)
		return ExitUsage
	}
}

func runReviewTerraformPlan(args []string, stdout io.Writer, stderr io.Writer) int {
	planPath, reviewPath, sarifPath, ok := parseReviewArgs(args, stderr, printReviewTerraformPlanUsage)
	if !ok {
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"Terraform plan JSON", planPath}, namedPath{"review", reviewPath}, namedPath{"SARIF", sarifPath}) {
		return ExitUsage
	}
	result, err := terraformplan.ReviewFile(planPath)
	if err != nil {
		writef(stderr, "review terraform-plan: %v\n", err)
		return ExitValidation
	}
	if err := project.WriteFile(reviewPath, []byte(reports.RenderTerraformReview(result))); err != nil {
		writef(stderr, "write terraform review: %v\n", err)
		return ExitUnsafe
	}
	sarifData, err := reports.RenderTerraformReviewSARIF(result)
	if err != nil {
		writef(stderr, "render terraform SARIF: %v\n", err)
		return ExitInternal
	}
	if err := project.WriteFile(sarifPath, sarifData); err != nil {
		writef(stderr, "write terraform SARIF: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote Terraform plan review to %s\n", reviewPath)
	writef(stdout, "wrote Terraform plan SARIF to %s\n", sarifPath)
	return ExitOK
}

func runImport(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "import requires a source format")
		printImportUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "cloudformation":
		return runImportCloudFormation(args[1:], stdout, stderr)
	case "sam":
		return runImportSAM(args[1:], stdout, stderr)
	case "k8s":
		return runImportKubernetes(args[1:], stdout, stderr)
	case "awsscan":
		return runImportAWSScan(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printImportUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown import source %q\n\n", args[0])
		printImportUsage(stderr)
		return ExitUsage
	}
}

func runImportCloudFormation(args []string, stdout io.Writer, stderr io.Writer) int {
	return runImportTemplate(args, stdout, stderr, cloudformation.FormatCloudFormation, "cloudformation")
}

func runImportSAM(args []string, stdout io.Writer, stderr io.Writer) int {
	return runImportTemplate(args, stdout, stderr, cloudformation.FormatSAM, "sam")
}

func runImportKubernetes(args []string, stdout io.Writer, stderr io.Writer) int {
	manifestPath, graphPath, lossPath, ok := parseImportArgs(args, stderr, printImportKubernetesUsage)
	if !ok {
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"input", manifestPath}, namedPath{"graph", graphPath}, namedPath{"loss report", lossPath}) {
		return ExitUsage
	}
	result, err := kubernetes.ImportFile(manifestPath)
	if err != nil {
		writef(stderr, "import k8s: %v\n", err)
		return ExitValidation
	}
	if exitCode := writeImportOutputs(stdout, stderr, graphPath, lossPath, result.Graph, result.Diagnostics, result.Loss); exitCode != ExitOK {
		return exitCode
	}
	writef(stdout, "imported k8s manifests to %s\n", graphPath)
	writef(stdout, "wrote loss report to %s\n", lossPath)
	return ExitOK
}

func runImportAWSScan(args []string, stdout io.Writer, stderr io.Writer) int {
	bundlePath, graphPath, lossPath, ok := parseImportArgs(args, stderr, printImportAWSScanUsage)
	if !ok {
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"input", bundlePath}, namedPath{"graph", graphPath}, namedPath{"loss report", lossPath}) {
		return ExitUsage
	}
	result, err := awsscan.ImportDirectory(bundlePath)
	if err != nil {
		writef(stderr, "import awsscan: %v\n", err)
		return ExitValidation
	}
	if exitCode := writeImportOutputs(stdout, stderr, graphPath, lossPath, result.Graph, result.Diagnostics, result.Loss); exitCode != ExitOK {
		return exitCode
	}
	writef(stdout, "imported AWS scan bundle to %s\n", graphPath)
	writef(stdout, "wrote loss report to %s\n", lossPath)
	return ExitOK
}

func runImportTemplate(args []string, stdout io.Writer, stderr io.Writer, format cloudformation.Format, label string) int {
	templatePath, graphPath, lossPath, ok := parseImportArgs(args, stderr, func(w io.Writer) {
		printImportFormatUsage(w, label)
	})
	if !ok {
		return ExitUsage
	}
	if !pathsAreDistinct(stderr, namedPath{"input", templatePath}, namedPath{"graph", graphPath}, namedPath{"loss report", lossPath}) {
		return ExitUsage
	}
	result, err := cloudformation.ImportFile(templatePath, format)
	if err != nil {
		writef(stderr, "import %s: %v\n", label, err)
		return ExitValidation
	}
	if exitCode := writeImportOutputs(stdout, stderr, graphPath, lossPath, result.Graph, result.Diagnostics, result.Loss); exitCode != ExitOK {
		return exitCode
	}
	writef(stdout, "imported %s template to %s\n", label, graphPath)
	writef(stdout, "wrote loss report to %s\n", lossPath)
	return ExitOK
}

func writeImportOutputs(stdout io.Writer, stderr io.Writer, graphPath string, lossPath string, importedGraph graph.Graph, diagnostics []graph.Diagnostic, lossReport loss.Report) int {
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	graphData, err := json.MarshalIndent(importedGraph, "", "  ")
	if err != nil {
		writef(stderr, "render graph: %v\n", err)
		return ExitInternal
	}
	graphData = append(graphData, '\n')
	if err := project.WriteFile(graphPath, graphData); err != nil {
		writef(stderr, "write graph output: %v\n", err)
		return ExitUnsafe
	}
	if err := project.WriteFile(lossPath, []byte(reports.RenderLossReport(lossReport))); err != nil {
		writef(stderr, "write loss report: %v\n", err)
		return ExitUnsafe
	}
	return ExitOK
}

func runValidate(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		writeln(stderr, "validate requires exactly one plan path")
		printValidateUsage(stderr)
		return ExitUsage
	}

	_, diagnostics, exitCode := loadGraph(args[0], stderr)
	if exitCode != ExitOK {
		return exitCode
	}

	if len(diagnostics) > 0 {
		write(stdout, reports.RenderDiagnostics(diagnostics))
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		writeln(stdout, "validation failed")
		return ExitValidation
	}

	writeln(stdout, "validation passed")
	return ExitOK
}

func runExplain(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		writeln(stderr, "explain requires exactly one plan path")
		printExplainUsage(stderr)
		return ExitUsage
	}

	lowered, diagnostics, exitCode := loadGraph(args[0], stderr)
	if exitCode != ExitOK {
		return exitCode
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}

	write(stdout, reports.Explain(lowered, diagnostics))
	return ExitOK
}

func runGenerate(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		writeln(stderr, "generate requires a target")
		printGenerateUsage(stderr)
		return ExitUsage
	}
	switch args[0] {
	case "terraform":
		return runGenerateTerraform(args[1:], stdout, stderr)
	case "mermaid":
		return runGenerateMermaid(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		printGenerateUsage(stdout)
		return ExitOK
	default:
		writef(stderr, "unknown generate target %q\n\n", args[0])
		printGenerateUsage(stderr)
		return ExitUsage
	}
}

func runGenerateTerraform(args []string, stdout io.Writer, stderr io.Writer) int {
	planPath, outPath, ok := parseInputOut(args, stderr, printGenerateTerraformUsage)
	if !ok {
		return ExitUsage
	}
	lowered, diagnostics, exitCode := loadGraph(planPath, stderr)
	if exitCode != ExitOK {
		return exitCode
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	files, err := terraformgen.Render(lowered)
	if err != nil {
		writef(stderr, "generate terraform: %v\n", err)
		return ExitValidation
	}
	if err := project.WriteFiles(outPath, files); err != nil {
		writef(stderr, "write terraform output: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote Terraform/OpenTofu files to %s\n", outPath)
	return ExitOK
}

func runGenerateMermaid(args []string, stdout io.Writer, stderr io.Writer) int {
	planPath, outPath, ok := parseInputOut(args, stderr, printGenerateMermaidUsage)
	if !ok {
		return ExitUsage
	}
	lowered, diagnostics, exitCode := loadGraph(planPath, stderr)
	if exitCode != ExitOK {
		return exitCode
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	if err := project.WriteFiles(outPath, mermaid.Render(lowered)); err != nil {
		writef(stderr, "write mermaid output: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote Mermaid files to %s\n", outPath)
	return ExitOK
}

func runRisks(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		writeln(stderr, "risks requires exactly one plan path")
		printRisksUsage(stderr)
		return ExitUsage
	}
	lowered, diagnostics, exitCode := loadGraph(args[0], stderr)
	if exitCode != ExitOK {
		return exitCode
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	write(stdout, reports.RenderSecurity(lowered))
	return ExitOK
}

func runCostNotes(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		writeln(stderr, "cost-notes requires exactly one plan path")
		printCostNotesUsage(stderr)
		return ExitUsage
	}
	lowered, diagnostics, exitCode := loadGraph(args[0], stderr)
	if exitCode != ExitOK {
		return exitCode
	}
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	write(stdout, reports.RenderCostNotes(lowered))
	return ExitOK
}

func runPack(args []string, stdout io.Writer, stderr io.Writer) int {
	planPath, outPath, ok := parseInputOut(args, stderr, printPackUsage)
	if !ok {
		return ExitUsage
	}
	document, sourceData, err := plan.LoadWithSource(planPath)
	if err != nil {
		writef(stderr, "load plan: %v\n", err)
		return ExitValidation
	}
	lowered, diagnostics := document.ToGraph()
	if graph.HasBlockingDiagnostics(diagnostics) {
		write(stdout, reports.RenderDiagnostics(diagnostics))
		return ExitValidation
	}
	files, err := project.BuildPack(planPath, sourceData, lowered)
	if err != nil {
		writef(stderr, "build pack: %v\n", err)
		return ExitValidation
	}
	if err := project.WriteFiles(outPath, files); err != nil {
		writef(stderr, "write pack: %v\n", err)
		return ExitUnsafe
	}
	writef(stdout, "wrote Planwright pack to %s\n", outPath)
	return ExitOK
}

func runVersion(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 0 {
		writeln(stderr, "version does not accept arguments")
		return ExitUsage
	}
	writef(stdout, "%s %s\n", version.Command, Version)
	return ExitOK
}

func loadGraph(path string, stderr io.Writer) (graph.Graph, []graph.Diagnostic, int) {
	document, err := plan.Load(path)
	if err != nil {
		writef(stderr, "load plan: %v\n", err)
		return graph.Graph{}, nil, ExitValidation
	}

	lowered, diagnostics := document.ToGraph()
	return lowered, diagnostics, ExitOK
}

func parseInputOut(args []string, stderr io.Writer, printUsage func(io.Writer)) (string, string, bool) {
	if len(args) != 3 || args[1] != "--out" {
		writeln(stderr, "command requires <planwright.yaml> --out <dir>")
		printUsage(stderr)
		return "", "", false
	}
	return args[0], args[2], true
}

func parseImportArgs(args []string, stderr io.Writer, printUsage func(io.Writer)) (string, string, string, bool) {
	if len(args) != 5 || args[1] != "--out" || args[3] != "--loss-report" {
		writeln(stderr, "import requires <input> --out <graph.json> --loss-report <loss.md>")
		printUsage(stderr)
		return "", "", "", false
	}
	return args[0], args[2], args[4], true
}

func parseReviewArgs(args []string, stderr io.Writer, printUsage func(io.Writer)) (string, string, string, bool) {
	if len(args) != 5 || args[1] != "--out" || args[3] != "--sarif" {
		writeln(stderr, "review terraform-plan requires <tfplan.json> --out <review.md> --sarif <planwright.sarif>")
		printUsage(stderr)
		return "", "", "", false
	}
	return args[0], args[2], args[4], true
}

func parseDiffArgs(args []string, stderr io.Writer) (string, string, string, bool) {
	if len(args) != 4 || args[2] != "--out" {
		writeln(stderr, "diff requires <old.graph.json> <new.graph.json> --out <review.md>")
		return "", "", "", false
	}
	return args[0], args[1], args[3], true
}

func parseSchemaGraphArgs(args []string, stderr io.Writer) (string, bool) {
	if len(args) != 2 || args[0] != "--out" {
		writeln(stderr, "schema graph requires --out <schema.json>")
		return "", false
	}
	return args[1], true
}

func parsePolicyGraphArgs(args []string, stderr io.Writer) (string, string, string, string, bool) {
	if len(args) != 7 || args[1] != "--profile" || args[3] != "--out" || args[5] != "--sarif" {
		writeln(stderr, "policy graph requires <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>")
		return "", "", "", "", false
	}
	return args[0], args[2], args[4], args[6], true
}

type namedPath struct {
	name string
	path string
}

type resolvedNamedPath struct {
	namedPath
	identity string
	info     os.FileInfo
	exists   bool
}

func pathsAreDistinct(stderr io.Writer, paths ...namedPath) bool {
	var seen []resolvedNamedPath
	for _, item := range paths {
		resolved, err := resolveNamedPath(item)
		if err != nil {
			writef(stderr, "%s path: %v\n", item.name, err)
			return false
		}
		for _, previous := range seen {
			if resolvedPathsMatch(resolved, previous) {
				writef(stderr, "%s path %q must be distinct from %s path %q\n", item.name, item.path, previous.name, previous.path)
				return false
			}
		}
		seen = append(seen, resolved)
	}
	return true
}

func resolveNamedPath(item namedPath) (resolvedNamedPath, error) {
	if strings.TrimSpace(item.path) == "" {
		return resolvedNamedPath{}, fmt.Errorf("path must not be empty")
	}
	absolute, err := filepath.Abs(filepath.Clean(item.path))
	if err != nil {
		return resolvedNamedPath{}, err
	}
	if os.PathSeparator == '\\' {
		absolute = strings.ToLower(absolute)
	}
	info, err := os.Stat(item.path)
	if err == nil {
		return resolvedNamedPath{
			namedPath: item,
			identity:  absolute,
			info:      info,
			exists:    true,
		}, nil
	}
	if !os.IsNotExist(err) {
		return resolvedNamedPath{}, err
	}
	return resolvedNamedPath{
		namedPath: item,
		identity:  absolute,
	}, nil
}

func resolvedPathsMatch(left resolvedNamedPath, right resolvedNamedPath) bool {
	if left.exists && right.exists {
		return os.SameFile(left.info, right.info)
	}
	return left.identity == right.identity
}

func parseServeArgs(args []string, stderr io.Writer) (string, string, bool) {
	projectDir := "."
	addr := "127.0.0.1:5786"

	index := 0
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		projectDir = args[0]
		index = 1
	}
	for index < len(args) {
		switch args[index] {
		case "--addr":
			if index+1 >= len(args) || strings.TrimSpace(args[index+1]) == "" {
				writeln(stderr, "serve --addr requires an address")
				return "", "", false
			}
			addr = args[index+1]
			index += 2
		default:
			writef(stderr, "unknown serve argument %q\n", args[index])
			return "", "", false
		}
	}
	return projectDir, addr, true
}

func isLoopbackAddr(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	host = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(host), "]"), "[")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func loadGraphJSON(path string) (graph.Graph, []graph.Diagnostic, error) {
	data, err := readGraphJSONFile(path)
	if err != nil {
		return graph.Graph{}, nil, err
	}
	loaded, diagnostics := graph.ValidateJSON(data, path)
	return loaded, diagnostics, nil
}

func readGraphJSONFile(path string) ([]byte, error) {
	data, err := localfs.ReadNamedRegularFile(path, maxGraphJSONSize, "graph JSON")
	if err != nil {
		return nil, err
	}
	return data, nil
}

func printUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright validate <planwright.yaml>")
	writeln(w, "  planwright validate-graph <planwright.graph.json>")
	writeln(w, "  planwright explain <planwright.yaml>")
	writeln(w, "  planwright generate terraform <planwright.yaml> --out <dir>")
	writeln(w, "  planwright generate mermaid <planwright.yaml> --out <dir>")
	writeln(w, "  planwright risks <planwright.yaml>")
	writeln(w, "  planwright cost-notes <planwright.yaml>")
	writeln(w, "  planwright docs check [path ...]")
	writeln(w, "  planwright import cloudformation <template.yaml> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import sam <template.yaml> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import k8s <manifest-path-or-dir> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import awsscan <bundle-dir> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright diff <old.graph.json> <new.graph.json> --out <review.md>")
	writeln(w, "  planwright schema graph --out <schema.json>")
	writeln(w, "  planwright policy profiles")
	writeln(w, "  planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>")
	writeln(w, "  planwright pack <planwright.yaml> --out <dir>")
	writeln(w, "  planwright review terraform-plan <tfplan.json> --out <review.md> --sarif <planwright.sarif>")
	writeln(w, "  planwright serve [project-dir] [--addr 127.0.0.1:5786]")
	writeln(w, "  planwright version")
}

func printValidateUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright validate <planwright.yaml>")
}

func printValidateGraphUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright validate-graph <planwright.graph.json>")
}

func printExplainUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright explain <planwright.yaml>")
}

func printGenerateUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright generate terraform <planwright.yaml> --out <dir>")
	writeln(w, "  planwright generate mermaid <planwright.yaml> --out <dir>")
}

func printGenerateTerraformUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright generate terraform <planwright.yaml> --out <dir>")
}

func printGenerateMermaidUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright generate mermaid <planwright.yaml> --out <dir>")
}

func printRisksUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright risks <planwright.yaml>")
}

func printCostNotesUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright cost-notes <planwright.yaml>")
}

func printDocsUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright docs check [path ...]")
}

func printDocsCheckUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright docs check [path ...]")
}

func printImportUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright import cloudformation <template.yaml> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import sam <template.yaml> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import k8s <manifest-path-or-dir> --out <graph.json> --loss-report <loss.md>")
	writeln(w, "  planwright import awsscan <bundle-dir> --out <graph.json> --loss-report <loss.md>")
}

func printImportFormatUsage(w io.Writer, format string) {
	writeln(w, "Usage:")
	writef(w, "  planwright import %s <template.yaml> --out <graph.json> --loss-report <loss.md>\n", format)
}

func printImportKubernetesUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright import k8s <manifest-path-or-dir> --out <graph.json> --loss-report <loss.md>")
}

func printImportAWSScanUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright import awsscan <bundle-dir> --out <graph.json> --loss-report <loss.md>")
}

func printPackUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright pack <planwright.yaml> --out <dir>")
}

func printReviewUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright review terraform-plan <tfplan.json> --out <review.md> --sarif <planwright.sarif>")
}

func printReviewTerraformPlanUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright review terraform-plan <tfplan.json> --out <review.md> --sarif <planwright.sarif>")
}

func printSchemaUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright schema graph --out <schema.json>")
}

func printSchemaGraphUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright schema graph --out <schema.json>")
}

func printPolicyUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright policy profiles")
	writeln(w, "  planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>")
}

func printPolicyProfilesUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright policy profiles")
}

func printPolicyGraphUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright policy graph <planwright.graph.json> --profile <profile> --out <policy.md> --sarif <policy.sarif>")
}

func printDiffUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright diff <old.graph.json> <new.graph.json> --out <review.md>")
}

func printServeUsage(w io.Writer) {
	writeln(w, "Usage:")
	writeln(w, "  planwright serve [project-dir] [--addr 127.0.0.1:5786]")
}
