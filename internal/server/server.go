// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/generators/mermaid"
	terraformgen "github.com/steadytao/planwright/internal/generators/terraform"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/localfs"
	"github.com/steadytao/planwright/internal/plan"
	"github.com/steadytao/planwright/internal/reports"
)

const defaultMaxBodyBytes int64 = 1 << 20

type Options struct {
	ProjectDir   string
	AllowedHosts []string
	MaxBodyBytes int64
}

type validateRequest struct {
	Plan string `json:"plan"`
}

type validateResponse struct {
	Diagnostics    []graph.Diagnostic `json:"diagnostics"`
	Blocking       bool               `json:"blocking"`
	Graph          graph.Graph        `json:"graph"`
	GraphJSON      string             `json:"graph_json"`
	Reports        map[string]string  `json:"reports"`
	TerraformFiles []filePreview      `json:"terraform_files,omitempty"`
	MermaidFiles   []filePreview      `json:"mermaid_files,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type exampleResponse struct {
	Plan string `json:"plan"`
}

type filePreview struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type handler struct {
	projectDir   string
	allowedHosts map[string]struct{}
	maxBodyBytes int64
	static       http.Handler
}

func New(opts Options) http.Handler {
	projectDir := strings.TrimSpace(opts.ProjectDir)
	if projectDir == "" {
		projectDir = "."
	}
	allowedHosts := opts.AllowedHosts
	if len(allowedHosts) == 0 {
		allowedHosts = DefaultAllowedHosts("")
	}
	maxBodyBytes := opts.MaxBodyBytes
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultMaxBodyBytes
	}
	static := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server static assets unavailable", http.StatusInternalServerError)
	}))
	if staticRoot, err := fs.Sub(staticFiles, "static"); err == nil {
		static = http.FileServer(http.FS(staticRoot))
	}
	return &handler{
		projectDir:   projectDir,
		allowedHosts: allowedHostSet(allowedHosts),
		maxBodyBytes: maxBodyBytes,
		static:       static,
	}
}

func DefaultAllowedHosts(addr string) []string {
	hosts := []string{"127.0.0.1", "localhost", "::1"}
	if host := hostFromAddr(addr); host != "" {
		hosts = append(hosts, host)
	}
	return hosts
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.setSecurityHeaders(w)
	if !h.hostAllowed(r.Host) {
		http.Error(w, "forbidden host", http.StatusForbidden)
		return
	}

	switch r.URL.Path {
	case "/api/example":
		h.handleExample(w, r)
	case "/api/validate":
		h.handleValidate(w, r)
	default:
		h.static.ServeHTTP(w, r)
	}
}

func (h *handler) handleExample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	path := filepath.Join("examples", "aws-webapp-basic", "planwright.yaml")
	data, err := localfs.ReadRegularFileInRoot(h.projectDir, path, defaultMaxBodyBytes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: fmt.Sprintf("read example plan: %v", err)})
		return
	}
	writeJSON(w, http.StatusOK, exampleResponse{Plan: string(data)})
}

func (h *handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, http.MethodPost)
		return
	}
	if !hasJSONContentType(r.Header.Get("Content-Type")) {
		writeJSON(w, http.StatusUnsupportedMediaType, errorResponse{Error: "Content-Type must be application/json"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	defer func() {
		if err := r.Body.Close(); err != nil {
			return
		}
	}()

	var request validateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeJSON(w, http.StatusRequestEntityTooLarge, errorResponse{Error: "request body is too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: fmt.Sprintf("decode request: %v", err)})
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "decode request: trailing JSON content is not accepted"})
		return
	}
	response, err := validatePlanText(request.Plan)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
		return
	}
	status := http.StatusOK
	if response.Blocking {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, response)
}

func validatePlanText(planText string) (validateResponse, error) {
	document, err := plan.Parse([]byte(planText), "browser-input")
	if err != nil {
		return validateResponse{}, err
	}
	lowered, diagnostics := document.ToGraph()
	graphData, err := json.MarshalIndent(lowered, "", "  ")
	if err != nil {
		return validateResponse{}, fmt.Errorf("render graph JSON: %w", err)
	}
	response := validateResponse{
		Diagnostics: diagnostics,
		Blocking:    graph.HasBlockingDiagnostics(diagnostics),
		Graph:       lowered,
		GraphJSON:   string(append(graphData, '\n')),
		Reports: map[string]string{
			"security":      reports.RenderSecurity(lowered),
			"cost":          reports.RenderCostNotes(lowered),
			"deployability": reports.RenderDeployability(lowered),
			"cleanup":       reports.RenderCleanup(lowered),
			"assumptions":   reports.RenderAssumptions(lowered),
		},
	}
	if response.Blocking {
		return response, nil
	}
	terraformFiles, err := terraformgen.Render(lowered)
	if err != nil {
		response.Diagnostics = append(response.Diagnostics, graph.Diagnostic{
			Severity: graph.SeverityWarn,
			Code:     "PW-WEB-GEN-001",
			Resource: "terraform",
			Message:  fmt.Sprintf("Terraform preview could not be generated: %v", err),
			Fix:      "Review graph diagnostics and use the CLI generator for detailed errors.",
		})
	} else {
		response.TerraformFiles = previews(terraformFiles)
	}
	response.MermaidFiles = previews(mermaid.Render(lowered))
	return response, nil
}

func previews(files []artifact.File) []filePreview {
	artifact.Sort(files)
	result := make([]filePreview, 0, len(files))
	for _, file := range files {
		result = append(result, filePreview{
			Path:    filepath.ToSlash(file.Path),
			Content: string(file.Data),
		})
	}
	return result
}

func (h *handler) setSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("X-Frame-Options", "DENY")
}

func (h *handler) hostAllowed(hostport string) bool {
	host := normaliseHost(hostport)
	if host == "" {
		return false
	}
	_, ok := h.allowedHosts[host]
	return ok
}

func allowedHostSet(hosts []string) map[string]struct{} {
	result := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		normalised := normaliseHost(host)
		if normalised != "" {
			result[normalised] = struct{}{}
		}
	}
	return result
}

func normaliseHost(hostport string) string {
	host := strings.TrimSpace(hostport)
	if host == "" {
		return ""
	}
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		host = splitHost
	}
	host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	return strings.ToLower(host)
}

func hostFromAddr(addr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return ""
	}
	return normaliseHost(host)
}

func hasJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil {
		return false
	}
	return strings.EqualFold(mediaType, "application/json")
}

func methodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		http.Error(w, "internal JSON encoding error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		return
	}
}
