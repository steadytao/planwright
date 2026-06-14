// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCompatibilityMatrixReferencesKnownFixtures(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	fixtures, err := Discover(filepath.Join(root, "examples"))
	if err != nil {
		t.Fatalf("Discover(examples) error = %v", err)
	}

	known := make(map[string]bool, len(fixtures))
	for _, fixture := range fixtures {
		known[fixture.ID] = true
	}

	matrix, err := readCompatibilityMatrix()
	if err != nil {
		t.Fatalf("read compatibility matrix error = %v", err)
	}

	referenced := map[string]bool{}
	for _, row := range matrix {
		evidence := row[len(row)-1]
		if strings.Contains(evidence, "Documentation-only") {
			continue
		}
		ids := fixtureIDs(evidence)
		if len(ids) == 0 {
			t.Fatalf("compatibility row %q has no fixture IDs or documentation-only marker", row[0])
		}
		for _, id := range ids {
			if !known[id] {
				t.Fatalf("compatibility row %q references unknown fixture %q", row[0], id)
			}
			referenced[id] = true
		}
	}

	for id := range known {
		if !referenced[id] {
			t.Fatalf("fixture %q is not referenced by the compatibility matrix", id)
		}
	}
}

func readCompatibilityMatrix() ([][]string, error) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "compatibility.md"))
	if err != nil {
		return nil, err
	}

	var rows [][]string
	inMatrix := false
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "## Current Matrix" {
			inMatrix = true
			continue
		}
		if inMatrix && strings.HasPrefix(line, "#") {
			break
		}
		if !inMatrix || line == "" || !strings.Contains(line, "|") {
			continue
		}
		cells := splitMarkdownRow(line)
		if len(cells) == 0 || cells[0] == "Format" || strings.HasPrefix(cells[0], ":---") {
			continue
		}
		if len(cells) != 7 {
			return nil, fmt.Errorf("compatibility matrix row has %d cells: %s", len(cells), line)
		}
		rows = append(rows, cells)
	}
	return rows, nil
}

func splitMarkdownRow(line string) []string {
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func fixtureIDs(evidence string) []string {
	matches := regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(evidence, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}
	return ids
}
