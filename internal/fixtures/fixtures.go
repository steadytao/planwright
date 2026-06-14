// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package fixtures

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/steadytao/planwright/internal/localfs"
)

const (
	Schema                  = "planwright.fixture.v1"
	maxFixtureMetadataBytes = 128 * 1024
)

var allowedLossCategories = []string{
	"ambiguous",
	"inferred",
	"manual-review-required",
	"normalised",
	"preserved",
	"unsafe",
	"unsupported",
}

type Fixture struct {
	Schema                 string               `yaml:"schema"`
	ID                     string               `yaml:"id"`
	Name                   string               `yaml:"name"`
	SourceFormat           string               `yaml:"source_format"`
	SourceKind             string               `yaml:"source_kind"`
	SourcePath             string               `yaml:"source_path"`
	CompatibilityLevel     int                  `yaml:"compatibility_level"`
	ExpectedLossCategories []string             `yaml:"expected_loss_categories"`
	Commands               []CommandExpectation `yaml:"commands"`

	dir string
}

type CommandExpectation struct {
	Name               string                   `yaml:"name"`
	Args               []string                 `yaml:"args"`
	WantExit           int                      `yaml:"want_exit"`
	WantStdoutContains []string                 `yaml:"want_stdout_contains"`
	WantStderrContains []string                 `yaml:"want_stderr_contains"`
	WantFiles          []string                 `yaml:"want_files"`
	WantFileContains   []FileContentExpectation `yaml:"want_file_contains"`
}

type FileContentExpectation struct {
	Path     string   `yaml:"path"`
	Contains []string `yaml:"contains"`
}

func Discover(root string) ([]Fixture, error) {
	var found []Fixture
	ids := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "fixture.yaml" {
			return nil
		}
		fixture, err := LoadFile(path)
		if err != nil {
			return err
		}
		if previous, ok := ids[fixture.ID]; ok {
			return fmt.Errorf("fixture id %q is duplicated by %s and %s", fixture.ID, previous, path)
		}
		ids[fixture.ID] = path
		found = append(found, fixture)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

func LoadFile(path string) (Fixture, error) {
	data, err := localfs.ReadNamedRegularFile(path, maxFixtureMetadataBytes, "fixture metadata")
	if err != nil {
		return Fixture{}, err
	}
	var fixture Fixture
	if err := yaml.Unmarshal(data, &fixture); err != nil {
		return Fixture{}, fmt.Errorf("parse fixture metadata %s: %w", path, err)
	}
	fixture.dir = filepath.Dir(path)
	if err := fixture.Validate(); err != nil {
		return Fixture{}, fmt.Errorf("validate fixture metadata %s: %w", path, err)
	}
	return fixture, nil
}

func (fixture Fixture) Dir() string {
	return fixture.dir
}

func (fixture Fixture) Source() string {
	return filepath.Join(fixture.dir, filepath.FromSlash(fixture.SourcePath))
}

func (fixture Fixture) Validate() error {
	var errs []error
	if fixture.Schema != Schema {
		errs = append(errs, fmt.Errorf("schema must be %q", Schema))
	}
	if strings.TrimSpace(fixture.ID) == "" {
		errs = append(errs, errors.New("id must not be empty"))
	}
	if strings.TrimSpace(fixture.Name) == "" {
		errs = append(errs, errors.New("name must not be empty"))
	}
	if strings.TrimSpace(fixture.SourceFormat) == "" {
		errs = append(errs, errors.New("source_format must not be empty"))
	}
	sourceKind := fixture.sourceKind()
	if sourceKind != "file" && sourceKind != "directory" {
		errs = append(errs, fmt.Errorf("source_kind must be file or directory"))
	}
	if fixture.CompatibilityLevel < 0 || fixture.CompatibilityLevel > 8 {
		errs = append(errs, fmt.Errorf("compatibility_level must be between 0 and 8"))
	}
	if err := validateSourcePath("source_path", fixture.SourcePath, sourceKind == "directory"); err != nil {
		errs = append(errs, err)
	} else if fixture.dir != "" {
		source := fixture.Source()
		info, err := os.Lstat(source)
		if err != nil {
			errs = append(errs, fmt.Errorf("source_path %q is not readable: %w", fixture.SourcePath, err))
		} else if sourceKind == "file" && !info.Mode().IsRegular() {
			errs = append(errs, fmt.Errorf("source_path %q must be a regular file", fixture.SourcePath))
		} else if sourceKind == "directory" && !info.IsDir() {
			errs = append(errs, fmt.Errorf("source_path %q must be a directory", fixture.SourcePath))
		}
	}
	for _, category := range fixture.ExpectedLossCategories {
		if !slices.Contains(allowedLossCategories, category) {
			errs = append(errs, fmt.Errorf("expected_loss_categories contains unsupported category %q", category))
		}
	}
	if len(fixture.Commands) == 0 {
		errs = append(errs, errors.New("commands must include at least one command expectation"))
	}
	for index, command := range fixture.Commands {
		errs = append(errs, command.validate(index)...)
	}
	return errors.Join(errs...)
}

func (fixture Fixture) sourceKind() string {
	if strings.TrimSpace(fixture.SourceKind) == "" {
		return "file"
	}
	return fixture.SourceKind
}

func (command CommandExpectation) ExpandArgs(fixture Fixture, tempDir string) []string {
	args := make([]string, 0, len(command.Args))
	for _, arg := range command.Args {
		expanded := strings.ReplaceAll(arg, "${source}", fixture.Source())
		expanded = strings.ReplaceAll(expanded, "${fixture}", fixture.Dir())
		expanded = strings.ReplaceAll(expanded, "${temp}", tempDir)
		if strings.Contains(arg, "${") {
			expanded = filepath.Clean(filepath.FromSlash(expanded))
		}
		args = append(args, expanded)
	}
	return args
}

func (command CommandExpectation) ExpectedFiles(tempDir string) []string {
	files := make([]string, 0, len(command.WantFiles))
	for _, path := range command.WantFiles {
		expanded := strings.ReplaceAll(path, "${temp}", tempDir)
		files = append(files, filepath.Clean(filepath.FromSlash(expanded)))
	}
	return files
}

func (command CommandExpectation) ExpectedFileContents(tempDir string) []FileContentExpectation {
	expectations := make([]FileContentExpectation, 0, len(command.WantFileContains))
	for _, expectation := range command.WantFileContains {
		expanded := strings.ReplaceAll(expectation.Path, "${temp}", tempDir)
		expectation.Path = filepath.Clean(filepath.FromSlash(expanded))
		expectations = append(expectations, expectation)
	}
	return expectations
}

func (command CommandExpectation) validate(index int) []error {
	var errs []error
	prefix := fmt.Sprintf("commands[%d]", index)
	if strings.TrimSpace(command.Name) == "" {
		errs = append(errs, fmt.Errorf("%s.name must not be empty", prefix))
	}
	if len(command.Args) == 0 {
		errs = append(errs, fmt.Errorf("%s.args must not be empty", prefix))
	}
	for _, path := range command.WantFiles {
		if err := validateTemplatePath(prefix+".want_files", path); err != nil {
			errs = append(errs, err)
		}
	}
	for itemIndex, expectation := range command.WantFileContains {
		itemPrefix := fmt.Sprintf("%s.want_file_contains[%d]", prefix, itemIndex)
		if err := validateTemplatePath(itemPrefix+".path", expectation.Path); err != nil {
			errs = append(errs, err)
		}
		if len(expectation.Contains) == 0 {
			errs = append(errs, fmt.Errorf("%s.contains must not be empty", itemPrefix))
		}
		for containsIndex, contains := range expectation.Contains {
			if strings.TrimSpace(contains) == "" {
				errs = append(errs, fmt.Errorf("%s.contains[%d] must not be empty", itemPrefix, containsIndex))
			}
		}
	}
	return errs
}

func validateSourcePath(field string, value string, allowCurrentDirectory bool) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	if path.IsAbs(value) || strings.Contains(value, `\`) || strings.Contains(value, ":") {
		return fmt.Errorf("%s %q must be a relative slash-separated path", field, value)
	}
	clean := path.Clean(value)
	if clean == "." && allowCurrentDirectory {
		return nil
	}
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("%s %q must stay inside the fixture directory", field, value)
	}
	return nil
}

func validateTemplatePath(field string, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not contain an empty path", field)
	}
	withoutTemp := strings.ReplaceAll(value, "${temp}", "temp")
	return validateSourcePath(field, withoutTemp, false)
}
