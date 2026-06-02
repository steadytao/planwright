// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/graph"
)

func TestImportFileLowersKubernetesGatewaySubset(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "manifests.yaml", kubernetesGatewayFixture())

	result, err := ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if graph.HasBlockingDiagnostics(result.Diagnostics) {
		t.Fatalf("ImportFile() diagnostics = %#v, want no blocking diagnostics", result.Diagnostics)
	}
	if got, want := result.Graph.Provider, "kubernetes"; got != want {
		t.Fatalf("Graph.Provider = %q, want %q", got, want)
	}
	if got, want := result.Graph.Region, "local"; got != want {
		t.Fatalf("Graph.Region = %q, want %q", got, want)
	}
	for _, kind := range []string{
		"k8s.namespace",
		"k8s.deployment",
		"k8s.service",
		"gatewayapi.gateway",
		"gatewayapi.httproute",
		"k8s.network_policy",
		"k8s.secret",
	} {
		if !hasNodeKind(result.Graph.Nodes, kind) {
			t.Fatalf("imported graph missing node kind %s: %#v", kind, result.Graph.Nodes)
		}
	}
	if !hasEdge(result.Graph.Edges, "demo/service/app", "demo/deployment/app", "network.route") {
		t.Fatalf("imported graph missing service-to-workload route edge: %#v", result.Graph.Edges)
	}
	if !hasEdge(result.Graph.Edges, "demo/httproute/app", "demo/service/app", "network.route") {
		t.Fatalf("imported graph missing HTTPRoute-to-service route edge: %#v", result.Graph.Edges)
	}
	if !hasEdge(result.Graph.Edges, "demo/httproute/app", "demo/gateway/public", "depends_on") {
		t.Fatalf("imported graph missing HTTPRoute-to-Gateway dependency edge: %#v", result.Graph.Edges)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want NetworkPolicy semantics note")
	}
}

func TestImportFileRedactsSecretValues(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "secret.yaml", `apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: demo
type: Opaque
data:
  password: c2VjcmV0
stringData:
  token: plaintext-token
`)

	result, err := ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	data, err := json.Marshal(result.Graph)
	if err != nil {
		t.Fatalf("Marshal(graph) error = %v", err)
	}
	output := string(data)
	for _, leaked := range []string{"c2VjcmV0", "plaintext-token"} {
		if strings.Contains(output, leaked) {
			t.Fatalf("graph output leaked secret value %q: %s", leaked, output)
		}
		for _, source := range result.Sources {
			if strings.Contains(string(source.Data), leaked) {
				t.Fatalf("result source metadata leaked secret value %q", leaked)
			}
		}
	}
	secret := findNode(result.Graph.Nodes, "demo/secret/app-secret")
	if secret == nil {
		t.Fatalf("secret node missing: %#v", result.Graph.Nodes)
	}
	if got, want := secret.Properties["type"], "Opaque"; got != want {
		t.Fatalf("secret type = %#v, want %q", got, want)
	}
	if !containsStringSlice(secret.Properties["data_keys"], "password") {
		t.Fatalf("secret data_keys = %#v, want password", secret.Properties["data_keys"])
	}
	if !containsStringSlice(secret.Properties["string_data_keys"], "token") {
		t.Fatalf("secret string_data_keys = %#v, want token", secret.Properties["string_data_keys"])
	}
}

func TestImportFileImportsDirectoryDeterministically(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeManifestAt(t, filepath.Join(dir, "b-service.yaml"), `apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: demo
spec:
  selector:
    app: web
  ports:
    - port: 80
`)
	writeManifestAt(t, filepath.Join(dir, "a-deployment.yaml"), `apiVersion: apps/v1
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
`)

	first, err := ImportFile(dir)
	if err != nil {
		t.Fatalf("ImportFile(first) error = %v", err)
	}
	second, err := ImportFile(dir)
	if err != nil {
		t.Fatalf("ImportFile(second) error = %v", err)
	}
	firstData, err := json.Marshal(first.Graph)
	if err != nil {
		t.Fatalf("Marshal(first) error = %v", err)
	}
	secondData, err := json.Marshal(second.Graph)
	if err != nil {
		t.Fatalf("Marshal(second) error = %v", err)
	}
	if string(firstData) != string(secondData) {
		t.Fatalf("directory import is not deterministic:\nfirst=%s\nsecond=%s", firstData, secondData)
	}
	if !hasEdge(first.Graph.Edges, "demo/service/app", "demo/deployment/app", "network.route") {
		t.Fatalf("directory import missing inferred service route: %#v", first.Graph.Edges)
	}
}

func TestImportFileRejectsDirectoryWithTooManyManifestFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for i := range 129 {
		writeManifestAt(t, filepath.Join(dir, "manifest-"+strconv.Itoa(i)+".yaml"), `apiVersion: v1
kind: ConfigMap
metadata:
  name: example
`)
	}

	_, err := ImportFile(dir)
	if err == nil || !strings.Contains(err.Error(), "too many manifest files") {
		t.Fatalf("ImportFile() error = %v, want aggregate file-count refusal", err)
	}
}

func TestImportFileReportsUnsupportedKind(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "future.yaml", `apiVersion: example.io/v1
kind: FutureDatabase
metadata:
  name: database
  namespace: demo
spec: {}
`)

	result, err := ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.Loss.Unsupported) != 1 {
		t.Fatalf("unsupported = %#v, want one unsupported resource", result.Loss.Unsupported)
	}
	if got, want := result.Loss.Unsupported[0].Kind, "example.io/v1/FutureDatabase"; got != want {
		t.Fatalf("unsupported kind = %q, want %q", got, want)
	}
}

func TestImportFileDoesNotTrimSelectorKeysForInference(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "selector-whitespace.yaml", `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: demo
spec:
  selector:
    matchLabels:
      " app ": web
  template:
    metadata:
      labels:
        " app ": web
    spec:
      containers:
        - name: app
          image: example/app:1.0
---
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: demo
spec:
  selector:
    app: web
  ports:
    - port: 80
`)

	result, err := ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if hasEdge(result.Graph.Edges, "demo/service/app", "demo/deployment/app", "network.route") {
		t.Fatalf("edges = %#v, did not want selector route inferred from trimmed key", result.Graph.Edges)
	}
	if len(result.Loss.Ambiguous) == 0 {
		t.Fatal("ambiguous loss items empty, want unmatched selector note")
	}
}

func TestImportFileRejectsDuplicateMappingKeys(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "duplicate.yaml", `apiVersion: v1
kind: ConfigMap
kind: Secret
metadata:
  name: ambiguous
`)

	_, err := ImportFile(path)
	if err == nil || !strings.Contains(err.Error(), `duplicate mapping key "kind"`) {
		t.Fatalf("ImportFile() error = %v, want duplicate kind refusal", err)
	}
}

func TestImportFileRejectsNonMappingListItem(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "bad-list.yaml", `apiVersion: v1
kind: List
items:
  - plain-string
`)

	_, err := ImportFile(path)
	if err == nil || !strings.Contains(err.Error(), "List item 0 must be a mapping document") {
		t.Fatalf("ImportFile() error = %v, want non-mapping List item refusal", err)
	}
}

func TestImportFileLowersCiliumPoliciesWithAmbiguity(t *testing.T) {
	t.Parallel()

	path := writeManifest(t, "cilium.yaml", `apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: app-policy
  namespace: demo
spec:
  endpointSelector:
    matchLabels:
      app: web
---
apiVersion: cilium.io/v2
kind: CiliumClusterwideNetworkPolicy
metadata:
  name: cluster-policy
specs:
  - endpointSelector:
      matchLabels:
        role: gateway
  - endpointSelector:
      matchLabels:
        role: worker
`)

	result, err := ImportFile(path)
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	for _, kind := range []string{"cilium.network_policy", "cilium.clusterwide_network_policy"} {
		if !hasNodeKind(result.Graph.Nodes, kind) {
			t.Fatalf("imported graph missing node kind %s: %#v", kind, result.Graph.Nodes)
		}
	}
	clusterPolicy := findNode(result.Graph.Nodes, "cluster/ciliumclusterwidenetworkpolicy/cluster-policy")
	if clusterPolicy == nil {
		t.Fatalf("cluster-wide Cilium policy node missing: %#v", result.Graph.Nodes)
	}
	if got, want := clusterPolicy.Properties["specs_count"], 2; got != want {
		t.Fatalf("specs_count = %#v, want %d", got, want)
	}
	if len(result.Loss.Ambiguous) < 2 {
		t.Fatalf("ambiguous = %#v, want Cilium policy semantics notes", result.Loss.Ambiguous)
	}
}

func TestImportFileRejectsSymlinkManifest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "manifest.yaml")
	writeManifestAt(t, target, kubernetesGatewayFixture())
	link := filepath.Join(dir, "link.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := ImportFile(link)
	if err == nil {
		t.Fatal("ImportFile() error = nil, want symlink rejection")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("ImportFile() error = %q, want symlink refusal", err.Error())
	}
}

func kubernetesGatewayFixture() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  name: demo
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: demo
spec:
  replicas: 2
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
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: demo
spec:
  type: ClusterIP
  selector:
    app: web
  ports:
    - port: 80
      targetPort: 8080
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: public
  namespace: demo
spec:
  gatewayClassName: example
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app
  namespace: demo
spec:
  parentRefs:
    - name: public
  rules:
    - backendRefs:
        - name: app
          port: 80
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: app-ingress
  namespace: demo
spec:
  podSelector:
    matchLabels:
      app: web
  policyTypes:
    - Ingress
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secret
  namespace: demo
type: Opaque
data:
  password: c2VjcmV0
stringData:
  token: plaintext-token
`
}

func writeManifest(t *testing.T, name string, data string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	writeManifestAt(t, path, data)
	return path
}

func writeManifestAt(t *testing.T, path string, data string) {
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

func findNode(nodes []graph.Node, id string) *graph.Node {
	for index := range nodes {
		if nodes[index].ID == id {
			return &nodes[index]
		}
	}
	return nil
}

func hasEdge(edges []graph.Edge, from string, to string, kind string) bool {
	for _, edge := range edges {
		if edge.From == from && edge.To == to && edge.Kind == kind {
			return true
		}
	}
	return false
}

func containsStringSlice(value any, want string) bool {
	items, ok := value.([]string)
	return ok && slices.Contains(items, want)
}
