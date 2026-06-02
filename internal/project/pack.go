// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"

	"github.com/steadytao/planwright/internal/artifact"
	"github.com/steadytao/planwright/internal/generators/mermaid"
	terraformgen "github.com/steadytao/planwright/internal/generators/terraform"
	"github.com/steadytao/planwright/internal/graph"
	"github.com/steadytao/planwright/internal/reports"
	"github.com/steadytao/planwright/internal/version"
)

type Manifest struct {
	Schema         string        `json:"schema"`
	CreatedBy      string        `json:"created_by"`
	Source         string        `json:"source"`
	Files          []ManifestRef `json:"files"`
	RequiresReview bool          `json:"requires_review"`
}

type ManifestRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int    `json:"size"`
}

func BuildPack(sourceName string, sourceData []byte, g graph.Graph) ([]artifact.File, error) {
	terraformFiles, err := terraformgen.Render(g)
	if err != nil {
		return nil, err
	}

	graphData, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return nil, err
	}
	graphData = append(graphData, '\n')

	files := []artifact.File{
		{Path: "planwright.yaml", Data: append([]byte(nil), sourceData...)},
		{Path: "planwright.graph.json", Data: graphData},
		{Path: "reports/security-report.md", Data: []byte(reports.RenderSecurity(g))},
		{Path: "reports/cost-notes.md", Data: []byte(reports.RenderCostNotes(g))},
		{Path: "reports/deployability-report.md", Data: []byte(reports.RenderDeployability(g))},
		{Path: "reports/cleanup.md", Data: []byte(reports.RenderCleanup(g))},
		{Path: "reports/assumptions.md", Data: []byte(reports.RenderAssumptions(g))},
	}
	files = append(files, artifact.Prefix("generated/terraform", terraformFiles)...)
	files = append(files, artifact.Prefix("diagrams", mermaid.Render(g))...)
	artifact.Sort(files)

	manifestData, err := renderManifest(sourceName, files)
	if err != nil {
		return nil, err
	}
	files = append(files, artifact.File{Path: "manifest.json", Data: manifestData})
	artifact.Sort(files)
	return files, nil
}

func renderManifest(sourceName string, files []artifact.File) ([]byte, error) {
	refs := make([]ManifestRef, 0, len(files))
	for _, file := range files {
		// Pack manifests use SHA-256 as a file integrity checksum, not as a password hash.
		// codeql[go/weak-sensitive-data-hashing]
		sum := sha256.Sum256(file.Data)
		refs = append(refs, ManifestRef{
			Path:   filepath.ToSlash(file.Path),
			SHA256: hex.EncodeToString(sum[:]),
			Size:   len(file.Data),
		})
	}
	manifest := Manifest{
		Schema:         "planwright.pack.v1",
		CreatedBy:      version.Created,
		Source:         filepath.Base(sourceName),
		Files:          refs,
		RequiresReview: true,
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
