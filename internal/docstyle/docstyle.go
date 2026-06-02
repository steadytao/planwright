// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package docstyle

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Finding struct {
	Path    string
	Line    int
	Code    string
	Message string
	Fix     string
}

var (
	bulletPattern    = regexp.MustCompile(`^\s*[-*+]\s+`)
	commaAndPattern  = regexp.MustCompile(`(?i),\s+and\b`)
	commaWordPattern = regexp.MustCompile(`(?i),\s+(but|including|however|or|then)\b`)
	wordPattern      = regexp.MustCompile(`[A-Za-z][A-Za-z'-]*`)
)

var britishSpellings = map[string]string{
	"behavior":      "behaviour",
	"behaviors":     "behaviours",
	"color":         "colour",
	"colors":        "colours",
	"colored":       "coloured",
	"organize":      "organise",
	"organized":     "organised",
	"organizing":    "organising",
	"organization":  "organisation",
	"organizations": "organisations",
	"normalize":     "normalise",
	"normalized":    "normalised",
	"normalizing":   "normalising",
	"serialize":     "serialise",
	"serialized":    "serialised",
	"serializing":   "serialising",
	"artifact":      "artefact",
	"artifacts":     "artefacts",
	"license":       "licence",
	"licenses":      "licences",
}

func CheckMarkdown(path string, data []byte) []Finding {
	lines := bytes.Split(data, []byte{'\n'})
	inFence := false
	var findings []Finding

	for index, rawLine := range lines {
		lineNumber := index + 1
		line := string(bytes.TrimRight(rawLine, "\r"))
		if isFenceLine(line) {
			if !inFence && index > 0 && isBlankLine(lines[index-1]) {
				findings = append(findings, Finding{
					Path:    path,
					Line:    lineNumber,
					Code:    "PW-DOC-CODE-001",
					Message: "Code blocks must hug the text above them.",
					Fix:     "Remove the blank line immediately before the code block.",
				})
			}
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		if bulletPattern.MatchString(line) && index > 0 && isBlankLine(lines[index-1]) {
			findings = append(findings, Finding{
				Path:    path,
				Line:    lineNumber,
				Code:    "PW-DOC-LIST-001",
				Message: "Bullet lists must hug the text above them.",
				Fix:     "Remove the blank line immediately before the bullet list.",
			})
		}
		if commaWordPattern.MatchString(line) {
			findings = append(findings, Finding{
				Path:    path,
				Line:    lineNumber,
				Code:    "PW-DOC-COMMA-001",
				Message: "Do not put a comma before these words in documentation prose: but; including; however; or; then.",
				Fix:     "Remove the comma before the matched word unless the text is quoted or code.",
			})
		}
		if hasSimpleCommaBeforeAnd(line) {
			findings = append(findings, Finding{
				Path:    path,
				Line:    lineNumber,
				Code:    "PW-DOC-COMMA-002",
				Message: "Avoid comma-before-and unless it separates larger topics.",
				Fix:     "Remove the comma before and or rewrite the sentence with a semicolon.",
			})
		}
		for _, word := range wordPattern.FindAllString(line, -1) {
			normalised := strings.ToLower(strings.Trim(word, "'-"))
			replacement, ok := britishSpellings[normalised]
			if !ok || allowedAmericanContext(line, normalised) {
				continue
			}
			findings = append(findings, Finding{
				Path:    path,
				Line:    lineNumber,
				Code:    "PW-DOC-EN-GB-001",
				Message: fmt.Sprintf("Use British English spelling: %s -> %s.", word, replacement),
				Fix:     fmt.Sprintf("Use %q unless this is a protocol, filename or official title.", replacement),
			})
		}
	}

	return findings
}

func isFenceLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}

func isBlankLine(line []byte) bool {
	return strings.TrimSpace(string(bytes.TrimRight(line, "\r"))) == ""
}

func CheckPaths(paths []string) ([]Finding, error) {
	var findings []Finding
	for _, path := range paths {
		pathFindings, err := checkPath(path)
		if err != nil {
			return nil, err
		}
		findings = append(findings, pathFindings...)
	}
	sortFindings(findings)
	return findings, nil
}

func checkPath(path string) (findings []Finding, err error) {
	cleanPath := filepath.Clean(path)
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s: symlink paths are not accepted", path)
	}
	if !info.IsDir() {
		if strings.ToLower(filepath.Ext(cleanPath)) != ".md" {
			return nil, nil
		}
		root, err := os.OpenRoot(filepath.Dir(cleanPath))
		if err != nil {
			return nil, err
		}
		defer func() {
			if closeErr := root.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}()
		data, err := root.ReadFile(filepath.Base(cleanPath))
		if err != nil {
			return nil, err
		}
		return CheckMarkdown(filepath.ToSlash(cleanPath), data), nil
	}

	root, err := os.OpenRoot(cleanPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := root.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	err = fs.WalkDir(root.FS(), ".", func(current string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		displayPath := filepath.Join(cleanPath, current)
		if shouldSkipPath(displayPath, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(current)) != ".md" {
			return nil
		}
		info, err := root.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		data, err := root.ReadFile(current)
		if err != nil {
			return err
		}
		findings = append(findings, CheckMarkdown(filepath.ToSlash(displayPath), data)...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return findings, nil
}

func FormatFindings(findings []Finding) string {
	var builder strings.Builder
	for _, finding := range findings {
		fmt.Fprintf(&builder, "%s:%d: %s: %s\n", finding.Path, finding.Line, finding.Code, finding.Message)
		if finding.Fix != "" {
			fmt.Fprintf(&builder, "  fix: %s\n", finding.Fix)
		}
	}
	return builder.String()
}

func hasSimpleCommaBeforeAnd(line string) bool {
	matches := commaAndPattern.FindAllStringIndex(line, -1)
	for _, match := range matches {
		before := line[:match[0]]
		start := strings.LastIndexAny(before, ".;:!?")
		clause := before
		if start >= 0 {
			clause = before[start+1:]
		}
		if !strings.Contains(strings.ToLower(clause), " and ") {
			return true
		}
	}
	return false
}

func allowedAmericanContext(line string, word string) bool {
	if strings.Contains(line, "`") {
		return true
	}
	if strings.Contains(line, "https://img.shields.io/") {
		return true
	}
	if strings.Contains(line, "SPDX-License-Identifier") || strings.Contains(line, "Apache License") || strings.Contains(line, "LICENSE") {
		return true
	}
	switch word {
	case "artifact", "artifacts":
		return strings.Contains(line, "GitHub Actions") || strings.Contains(line, "SARIF")
	default:
		return false
	}
}

func shouldSkipPath(path string, entry fs.DirEntry) bool {
	name := entry.Name()
	if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "bin" || name == "generated" {
		return true
	}
	return false
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		return findings[i].Code < findings[j].Code
	})
}
