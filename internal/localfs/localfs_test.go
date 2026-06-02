// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package localfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadRegularFileReadsFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := ReadRegularFile(path, 16)
	if err != nil {
		t.Fatalf("ReadRegularFile() error = %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("data = %q, want data", data)
	}
}

func TestReadRegularFileRejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	_, err := ReadRegularFile(link, 16)
	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("ReadRegularFile() error = %v, want symlink refusal", err)
	}
}

func TestReadRegularFileRejectsOversize(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReadRegularFile(path, 3)
	if err == nil || !strings.Contains(err.Error(), "exceeds 3 bytes") {
		t.Fatalf("ReadRegularFile() error = %v, want size refusal", err)
	}
}

func TestReadRegularFileInRootRejectsTraversal(t *testing.T) {
	t.Parallel()

	_, err := ReadRegularFileInRoot(t.TempDir(), "../escape.txt", 16)
	if err == nil || !strings.Contains(err.Error(), "escapes root") {
		t.Fatalf("ReadRegularFileInRoot() error = %v, want traversal refusal", err)
	}
}
