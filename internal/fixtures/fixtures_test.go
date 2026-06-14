// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFileAcceptsExampleFixture(t *testing.T) {
	t.Parallel()

	fixture, err := LoadFile(filepath.Join("..", "..", "examples", "aws-webapp-basic", "fixture.yaml"))
	if err != nil {
		t.Fatalf("LoadFile(example fixture) error = %v", err)
	}
	if got, want := fixture.ID, "aws-webapp-basic"; got != want {
		t.Fatalf("fixture.ID = %q, want %q", got, want)
	}
	if got, want := fixture.CompatibilityLevel, 5; got != want {
		t.Fatalf("fixture.CompatibilityLevel = %d, want %d", got, want)
	}
	if fixture.Source() == "" {
		t.Fatalf("fixture.Source() is empty")
	}
}

func TestDiscoverFindsFixtureMetadata(t *testing.T) {
	t.Parallel()

	fixtures, err := Discover(filepath.Join("..", "..", "examples"))
	if err != nil {
		t.Fatalf("Discover(examples) error = %v", err)
	}
	ids := map[string]bool{}
	for _, fixture := range fixtures {
		ids[fixture.ID] = true
	}
	for _, want := range []string{"aws-webapp-basic", "aws-webapp-public-db"} {
		if !ids[want] {
			t.Fatalf("Discover(examples) did not find fixture %q; got %v", want, ids)
		}
	}
}

func TestDiscoverRejectsDuplicateFixtureID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, dirName := range []string{"one", "two"} {
		dir := filepath.Join(root, dirName)
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatalf("Mkdir(%q) error = %v", dir, err)
		}
		writeFixtureTestFile(t, filepath.Join(dir, "planwright.yaml"), []byte("version: planwright.v1\n"))
		writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: duplicate
name: Duplicate fixture
source_format: planwright.yaml
source_path: planwright.yaml
compatibility_level: 1
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
`))
	}

	_, err := Discover(root)
	if err == nil {
		t.Fatalf("Discover(duplicate IDs) error = nil, want error")
	}
	if !strings.Contains(err.Error(), `fixture id "duplicate" is duplicated`) {
		t.Fatalf("Discover(duplicate IDs) error = %v, want duplicate ID refusal", err)
	}
}

func TestLoadFileRejectsPathTraversalSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: bad
name: Bad fixture
source_format: planwright.yaml
source_path: ../secret.yaml
compatibility_level: 1
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
`))

	_, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err == nil {
		t.Fatalf("LoadFile(path traversal) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "must stay inside the fixture directory") {
		t.Fatalf("LoadFile(path traversal) error = %v, want traversal refusal", err)
	}
}

func TestLoadFileRejectsAbsoluteSlashSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: bad
name: Bad fixture
source_format: planwright.yaml
source_path: /secret.yaml
compatibility_level: 1
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
`))

	_, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err == nil {
		t.Fatalf("LoadFile(absolute source) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "must be a relative slash-separated path") {
		t.Fatalf("LoadFile(absolute source) error = %v, want absolute path refusal", err)
	}
}

func TestLoadFileAcceptsDirectorySource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	bundle := filepath.Join(dir, "bundle")
	if err := os.Mkdir(bundle, 0o700); err != nil {
		t.Fatalf("Mkdir(%q) error = %v", bundle, err)
	}
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: directory-source
name: Directory source fixture
source_format: awsscan.bundle
source_kind: directory
source_path: bundle
compatibility_level: 4
commands:
  - name: import
    args: ["import", "awsscan", "${source}", "--out", "${temp}/graph.json", "--loss-report", "${temp}/loss.md"]
    want_exit: 0
`))

	fixture, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err != nil {
		t.Fatalf("LoadFile(directory source) error = %v", err)
	}
	if got, want := fixture.Source(), bundle; got != want {
		t.Fatalf("fixture.Source() = %q, want %q", got, want)
	}
}

func TestLoadFileAcceptsCurrentDirectorySource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: current-directory-source
name: Current directory source fixture
source_format: awsscan.bundle
source_kind: directory
source_path: .
compatibility_level: 4
commands:
  - name: import
    args: ["import", "awsscan", "${source}", "--out", "${temp}/graph.json", "--loss-report", "${temp}/loss.md"]
    want_exit: 0
`))

	fixture, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err != nil {
		t.Fatalf("LoadFile(current directory source) error = %v", err)
	}
	if got, want := fixture.Source(), filepath.Clean(dir); got != want {
		t.Fatalf("fixture.Source() = %q, want %q", got, want)
	}
}

func TestLoadFileRejectsUnknownSourceKind(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFixtureTestFile(t, filepath.Join(dir, "planwright.yaml"), []byte("version: planwright.v1\n"))
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: bad-kind
name: Bad kind fixture
source_format: planwright.yaml
source_kind: stream
source_path: planwright.yaml
compatibility_level: 1
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
`))

	_, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err == nil {
		t.Fatalf("LoadFile(unknown source kind) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "source_kind must be file or directory") {
		t.Fatalf("LoadFile(unknown source kind) error = %v, want source kind refusal", err)
	}
}

func TestCommandExpectationExpandsPaths(t *testing.T) {
	t.Parallel()

	fixture := Fixture{
		SourcePath: "planwright.yaml",
		dir:        filepath.Join("root", "fixture"),
	}
	command := CommandExpectation{
		Args:      []string{"generate", "terraform", "${source}", "--out", "${temp}/terraform"},
		WantFiles: []string{"${temp}/terraform/versions.tf"},
		WantSARIFFiles: []string{
			"${temp}/terraform/planwright.sarif",
		},
		WantFileContains: []FileContentExpectation{
			{
				Path:     "${temp}/terraform/README.md",
				Contains: []string{"Terraform"},
			},
		},
	}

	args := command.ExpandArgs(fixture, filepath.Join("tmp", "out"))
	if got, want := args[2], filepath.Join("root", "fixture", "planwright.yaml"); got != want {
		t.Fatalf("expanded source arg = %q, want %q", got, want)
	}
	if got, want := args[4], filepath.Join("tmp", "out", "terraform"); got != want {
		t.Fatalf("expanded output arg = %q, want %q", got, want)
	}
	files := command.ExpectedFiles(filepath.Join("tmp", "out"))
	if got, want := files[0], filepath.Join("tmp", "out", "terraform", "versions.tf"); got != want {
		t.Fatalf("expanded expected file = %q, want %q", got, want)
	}
	contents := command.ExpectedFileContents(filepath.Join("tmp", "out"))
	if got, want := contents[0].Path, filepath.Join("tmp", "out", "terraform", "README.md"); got != want {
		t.Fatalf("expanded expected file content path = %q, want %q", got, want)
	}
	sarifFiles := command.ExpectedSARIFFiles(filepath.Join("tmp", "out"))
	if got, want := sarifFiles[0], filepath.Join("tmp", "out", "terraform", "planwright.sarif"); got != want {
		t.Fatalf("expanded expected SARIF file = %q, want %q", got, want)
	}
}

func TestLoadFileRejectsEmptyFileContentExpectation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFixtureTestFile(t, filepath.Join(dir, "planwright.yaml"), []byte("version: planwright.v1\n"))
	writeFixtureTestFile(t, filepath.Join(dir, "fixture.yaml"), []byte(`
schema: planwright.fixture.v1
id: bad-file-content
name: Bad file content fixture
source_format: planwright.yaml
source_path: planwright.yaml
compatibility_level: 1
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
    want_file_contains:
      - path: "${temp}/report.md"
        contains: []
`))

	_, err := LoadFile(filepath.Join(dir, "fixture.yaml"))
	if err == nil {
		t.Fatalf("LoadFile(empty file content expectation) error = nil, want error")
	}
	if !strings.Contains(err.Error(), "contains must not be empty") {
		t.Fatalf("LoadFile(empty file content expectation) error = %v, want contains refusal", err)
	}
}

func writeFixtureTestFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
