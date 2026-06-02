// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func sarifArtifactURI(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "unknown"
	}

	absolute, err := filepath.Abs(filepath.Clean(source))
	if err != nil {
		return escapeSARIFPath(filepath.ToSlash(filepath.Base(source)))
	}

	cwd, err := os.Getwd()
	if err == nil {
		if rel, err := filepath.Rel(cwd, absolute); err == nil && isLocalRelativePath(rel) {
			return escapeSARIFPath(filepath.ToSlash(rel))
		}
	}
	return escapeSARIFPath(filepath.ToSlash(filepath.Base(absolute)))
}

func isLocalRelativePath(path string) bool {
	if path == "." || filepath.IsAbs(path) {
		return false
	}
	return path != ".." && !strings.HasPrefix(path, ".."+string(filepath.Separator))
}

func escapeSARIFPath(path string) string {
	path = strings.ReplaceAll(filepath.ToSlash(path), "\\", "/")
	parts := strings.Split(path, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
