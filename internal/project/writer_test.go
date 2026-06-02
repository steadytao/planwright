// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/localfs"
)

func TestWriteFilesWritesNamedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	err := WriteFiles(root, []artifact.File{
		{Path: "generated/terraform/main.tf", Data: []byte("terraform {}\n")},
		{Path: "reports/security-report.md", Data: []byte("# Security Report\n")},
	})
	if err != nil {
		t.Fatalf("WriteFiles() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "generated", "terraform", "main.tf")); err != nil {
		t.Fatalf("generated file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "reports", "security-report.md")); err != nil {
		t.Fatalf("report file missing: %v", err)
	}
}

func TestWriteFilesRejectsTraversal(t *testing.T) {
	t.Parallel()

	err := WriteFiles(t.TempDir(), []artifact.File{
		{Path: "../escape.txt", Data: []byte("escape")},
	})
	if err == nil {
		t.Fatal("WriteFiles() error = nil, want traversal error")
	}
	if !strings.Contains(err.Error(), "unsafe artefact path") {
		t.Fatalf("WriteFiles() error = %q, want unsafe path error", err.Error())
	}
}

func TestWriteFilesRejectsSymlinkRoot(t *testing.T) {
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

	err := WriteFiles(link, []artifact.File{{Path: "file.txt", Data: []byte("data")}})
	if err == nil {
		t.Fatal("WriteFiles() error = nil, want symlink root error")
	}
	if !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("WriteFiles() error = %q, want symlink error", err.Error())
	}
}

func TestWriteFilesRejectsSymlinkAncestorOfRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatalf("Mkdir(target) error = %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	err := WriteFiles(filepath.Join(link, "out"), []artifact.File{{Path: "report.md", Data: []byte("unsafe")}})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("WriteFiles() error = %v, want symlink ancestor refusal", err)
	}
	if _, err := os.Stat(filepath.Join(target, "out", "report.md")); !os.IsNotExist(err) {
		t.Fatalf("outside file stat error = %v, want file not created", err)
	}
}

func TestWriteFilesRejectsSymlinkParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "out")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatalf("Mkdir(root) error = %v", err)
	}
	outside := filepath.Join(dir, "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatalf("Mkdir(outside) error = %v", err)
	}
	link := filepath.Join(root, "reports")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	err := WriteFiles(root, []artifact.File{{Path: "reports/security-report.md", Data: []byte("unsafe")}})
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("WriteFiles() error = %v, want symlink parent refusal", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "security-report.md")); !os.IsNotExist(err) {
		t.Fatalf("outside file stat error = %v, want file not created", err)
	}
}

func TestWriteFileWritesExplicitOutputFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "graph.json")
	if err := WriteFile(path, []byte("{}\n")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	data, err := localfs.ReadRegularFile(path, 1024)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "{}\n" {
		t.Fatalf("output = %q, want graph JSON", data)
	}
}

func TestWriteFilesUsesPrivatePermissions(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX write permissions through os.FileMode")
	}

	root := t.TempDir()
	path := "reports/security-report.md"
	if err := WriteFiles(root, []artifact.File{{Path: path, Data: []byte("# Security Report\n")}}); err != nil {
		t.Fatalf("WriteFiles() error = %v", err)
	}
	info, err := os.Stat(filepath.Join(root, path))
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != outputFileMode {
		t.Fatalf("permissions = %#o, want %#o", got, outputFileMode)
	}
}

func TestWriteFileUsesPrivatePermissions(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX write permissions through os.FileMode")
	}

	path := filepath.Join(t.TempDir(), "graph.json")
	if err := WriteFile(path, []byte("{}\n")); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != outputFileMode {
		t.Fatalf("permissions = %#o, want %#o", got, outputFileMode)
	}
}

func TestWriteFileRejectsSymlinkTarget(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("old"), 0o600); err != nil {
		t.Fatalf("WriteFile() setup error = %v", err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	err := WriteFile(link, []byte("new"))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("WriteFile() error = %v, want symlink refusal", err)
	}
}

func TestWriteFileRejectsSymlinkParent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	outside := filepath.Join(dir, "outside")
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatalf("Mkdir(outside) error = %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	err := WriteFile(filepath.Join(link, "report.md"), []byte("unsafe"))
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("WriteFile() error = %v, want symlink parent refusal", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "report.md")); !os.IsNotExist(err) {
		t.Fatalf("outside file stat error = %v, want file not created", err)
	}
}
