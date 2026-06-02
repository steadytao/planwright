// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package localfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReadRegularFile(path string, maxBytes int64) ([]byte, error) {
	return ReadNamedRegularFile(path, maxBytes, "file")
}

func ReadNamedRegularFile(path string, maxBytes int64, sizeLabel string) ([]byte, error) {
	cleanPath, err := cleanPath(path)
	if err != nil {
		return nil, err
	}
	return readRegularFile(filepath.Dir(cleanPath), filepath.Base(cleanPath), cleanPath, maxBytes, sizeLabel)
}

func ReadRegularFileInRoot(rootPath string, relativePath string, maxBytes int64) ([]byte, error) {
	if strings.TrimSpace(rootPath) == "" {
		return nil, fmt.Errorf("read file: root path must not be empty")
	}
	relative := strings.TrimSpace(filepath.Clean(relativePath))
	if relative == "" || relative == "." || filepath.IsAbs(relative) || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("read %s: relative path escapes root", relativePath)
	}
	return readRegularFile(filepath.Clean(rootPath), relative, filepath.Join(rootPath, relative), maxBytes, "file")
}

func readRegularFile(rootPath string, name string, displayPath string, maxBytes int64, sizeLabel string) (_ []byte, err error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open root for %s: %w", displayPath, err)
	}
	defer func() {
		if closeErr := root.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close root for %s: %w", displayPath, closeErr)
		}
	}()

	info, err := root.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", displayPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("read %s: symlink files are not accepted", displayPath)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("read %s: not a regular file", displayPath)
	}
	if maxBytes > 0 && info.Size() > maxBytes {
		return nil, fmt.Errorf("read %s: %s exceeds %d bytes", displayPath, sizeLabel, maxBytes)
	}
	data, err := root.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", displayPath, err)
	}
	return data, nil
}

func cleanPath(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", fmt.Errorf("read file: path must not be empty")
	}
	return filepath.Clean(trimmed), nil
}
