// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesIndexWithSecurityHeaders(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	req.Host = "127.0.0.1"
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", got)
	}
	assertHeaderContains(t, rr, "Content-Security-Policy", "default-src 'self'")
	assertHeaderContains(t, rr, "X-Content-Type-Options", "nosniff")
	assertHeaderContains(t, rr, "Referrer-Policy", "no-referrer")
	assertHeaderContains(t, rr, "Cross-Origin-Opener-Policy", "same-origin")
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("unexpected CORS header: %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if strings.Contains(rr.Body.String(), "http://") || strings.Contains(rr.Body.String(), "https://") {
		t.Fatalf("index contains remote URL: %s", rr.Body.String())
	}
}

func TestHandlerRejectsUnexpectedHost(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://evil.example/", nil)
	req.Host = "evil.example"
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403; body=%s", rr.Code, rr.Body.String())
	}
}

func TestValidateAcceptsPlanTextAndReturnsPreviews(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"plan":` + quote(examplePlanYAML) + `}`)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/validate", body)
	req.Host = "127.0.0.1"
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var decoded validateResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("response JSON invalid: %v\n%s", err, rr.Body.String())
	}
	if decoded.Blocking {
		t.Fatalf("Blocking = true, want false; diagnostics=%#v", decoded.Diagnostics)
	}
	if decoded.Graph.Version != "planwright.graph.v1" {
		t.Fatalf("graph version = %q, want planwright.graph.v1", decoded.Graph.Version)
	}
	if decoded.Reports["security"] == "" || decoded.Reports["cost"] == "" || decoded.Reports["deployability"] == "" {
		t.Fatalf("reports = %#v, want security, cost and deployability", decoded.Reports)
	}
	if !hasPreview(decoded.TerraformFiles, "versions.tf") {
		t.Fatalf("terraform previews = %#v, want versions.tf", decoded.TerraformFiles)
	}
	if !hasPreview(decoded.MermaidFiles, "architecture.mmd") {
		t.Fatalf("mermaid previews = %#v, want architecture.mmd", decoded.MermaidFiles)
	}
}

func TestValidateRejectsJSONLikeContentType(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"plan":` + quote(examplePlanYAML) + `}`)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/validate", body)
	req.Host = "127.0.0.1"
	req.Header.Set("Content-Type", "application/jsonx")
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415; body=%s", rr.Code, rr.Body.String())
	}
}

func TestValidateRejectsMalformedPlan(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"plan":"version: planwright.v2\n"}`)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/validate", body)
	req.Host = "127.0.0.1"
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "unsupported Planwright plan version") {
		t.Fatalf("body = %s, want parse error", rr.Body.String())
	}
}

func TestValidateRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/validate", strings.NewReader(`{"plan":"too large"}`))
	req.Host = "127.0.0.1"
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: ".", MaxBodyBytes: 8}).ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413; body=%s", rr.Code, rr.Body.String())
	}
}

func TestValidateRejectsTrailingJSONContent(t *testing.T) {
	t.Parallel()

	body := bytes.NewBufferString(`{"plan":` + quote(examplePlanYAML) + `} {"plan":""}`)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/api/validate", body)
	req.Host = "127.0.0.1"
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	New(Options{ProjectDir: "."}).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "trailing JSON content") {
		t.Fatalf("body = %s, want trailing-content error", rr.Body.String())
	}
}

func assertHeaderContains(t *testing.T, rr *httptest.ResponseRecorder, name string, want string) {
	t.Helper()

	if got := rr.Header().Get(name); !strings.Contains(got, want) {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}

func quote(value string) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func hasPreview(files []filePreview, path string) bool {
	for _, file := range files {
		if file.Path == path && file.Content != "" {
			return true
		}
	}
	return false
}

const examplePlanYAML = `version: planwright.v1
provider: aws
region: ap-southeast-2
profile: lab

components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      app_port: 8080
      db_engine: postgres
      db_public: false

flows:
  - from: internet
    to: webapp.alb
    kind: network.allow
    protocol: tcp
    port: 443

  - from: webapp.alb
    to: webapp.app
    kind: network.allow
    protocol: tcp
    port: 8080

  - from: webapp.app
    to: webapp.db
    kind: network.allow
    protocol: tcp
    port: 5432
`
