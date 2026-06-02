// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/steadytao/planwright/internal/artifact"
)

const outputFileMode os.FileMode = 0o600

func WriteFiles(root string, files []artifact.File) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("output root must not be empty")
	}
	if err := prepareRoot(root); err != nil {
		return err
	}
	for _, file := range files {
		target, err := safeArtifactPath(root, file.Path)
		if err != nil {
			return err
		}
		if err := mkdirAllWithinRoot(root, filepath.Dir(target)); err != nil {
			return fmt.Errorf("create parent for %s: %w", file.Path, err)
		}
		if info, err := os.Lstat(target); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("write %s: refusing to overwrite symlink", file.Path)
		} else if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("inspect %s: %w", file.Path, err)
		}
		if err := os.WriteFile(target, file.Data, outputFileMode); err != nil {
			return fmt.Errorf("write %s: %w", file.Path, err)
		}
	}
	return nil
}

func WriteFile(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("output path must not be empty")
	}
	if err := rejectSymlinkAncestors(path); err != nil {
		return err
	}
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("write %s: refusing to overwrite symlink", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create parent for %s: %w", path, err)
	}
	if err := rejectSymlinkAncestors(path); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, outputFileMode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func prepareRoot(root string) error {
	if err := rejectSymlinkAncestors(root); err != nil {
		return err
	}
	info, err := os.Lstat(root)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("output root %s is a symlink", root)
		}
		if !info.IsDir() {
			return fmt.Errorf("output root %s is not a directory", root)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output root %s: %w", root, err)
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return fmt.Errorf("create output root %s: %w", root, err)
	}
	if err := rejectSymlinkAncestors(root); err != nil {
		return err
	}
	return nil
}

func safeArtifactPath(root string, relativePath string) (string, error) {
	clean := filepath.Clean(relativePath)
	if clean == "." || clean == "" || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", fmt.Errorf("unsafe artefact path %q", relativePath)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve output root: %w", err)
	}
	target := filepath.Join(absRoot, clean)
	rel, err := filepath.Rel(absRoot, target)
	if err != nil {
		return "", fmt.Errorf("resolve artefact path %q: %w", relativePath, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("unsafe artefact path %q", relativePath)
	}
	return target, nil
}

func mkdirAllWithinRoot(root string, parent string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve output root: %w", err)
	}
	absParent, err := filepath.Abs(parent)
	if err != nil {
		return fmt.Errorf("resolve parent: %w", err)
	}
	rel, err := filepath.Rel(absRoot, absParent)
	if err != nil {
		return fmt.Errorf("resolve parent %s: %w", parent, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("parent %s escapes output root", parent)
	}
	if rel == "." {
		return nil
	}
	current := absRoot
	for part := range strings.SplitSeq(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			if err := os.Mkdir(current, 0o750); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("parent %s is a symlink", current)
		}
		if !info.IsDir() {
			return fmt.Errorf("parent %s is not a directory", current)
		}
	}
	return nil
}

func rejectSymlinkAncestors(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", path, err)
	}
	dir := filepath.Dir(abs)
	var ancestors []string
	for {
		ancestors = append(ancestors, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	for _, ancestor := range slices.Backward(ancestors) {
		info, err := os.Lstat(ancestor)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect parent %s: %w", ancestor, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if isPlatformPathAlias(ancestor) {
				continue
			}
			return fmt.Errorf("parent %s is a symlink", ancestor)
		}
		if !info.IsDir() {
			return fmt.Errorf("parent %s is not a directory", ancestor)
		}
	}
	return nil
}

func isPlatformPathAlias(path string) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	clean := filepath.Clean(path)
	return clean == "/tmp" || clean == "/var"
}
