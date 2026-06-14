// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraformstate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/steadytao/planwright/internal/localfs"
)

const maxStateJSONBytes = 25 * 1024 * 1024

type Result struct {
	Source           string
	FormatVersion    string
	TerraformVersion string
	ResourceCount    int
	Resources        []Resource
}

type Resource struct {
	Address             string
	Mode                string
	Type                string
	Name                string
	ProviderName        string
	Supported           bool
	SensitiveAttributes []string
}

type stateFile struct {
	FormatVersion    string      `json:"format_version"`
	TerraformVersion string      `json:"terraform_version"`
	Values           stateValues `json:"values"`
}

type stateValues struct {
	RootModule stateModule `json:"root_module"`
}

type stateModule struct {
	Resources    []stateResource `json:"resources"`
	ChildModules []stateModule   `json:"child_modules"`
}

type stateResource struct {
	Address         string         `json:"address"`
	Mode            string         `json:"mode"`
	Type            string         `json:"type"`
	Name            string         `json:"name"`
	ProviderName    string         `json:"provider_name"`
	SensitiveValues map[string]any `json:"sensitive_values"`
}

func ReviewFile(path string) (Result, error) {
	data, err := readState(path)
	if err != nil {
		return Result{}, err
	}
	return ReviewBytes(data, path)
}

func ReviewBytes(data []byte, sourceName string) (Result, error) {
	var state stateFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&state); err != nil {
		return Result{}, fmt.Errorf("parse %s: %w", sourceName, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Result{}, fmt.Errorf("parse %s: Terraform state JSON contains trailing content", sourceName)
	}
	if err := validateFormatVersion(state.FormatVersion); err != nil {
		return Result{}, err
	}

	result := Result{
		Source:           sourceName,
		FormatVersion:    state.FormatVersion,
		TerraformVersion: state.TerraformVersion,
	}
	appendModuleResources(&result.Resources, state.Values.RootModule)
	result.ResourceCount = len(result.Resources)
	sortResources(result.Resources)
	return result, nil
}

func readState(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("read Terraform state JSON: path must not be empty")
	}
	return localfs.ReadNamedRegularFile(path, maxStateJSONBytes, "Terraform state JSON")
}

func validateFormatVersion(version string) error {
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("terraform JSON format_version is required")
	}
	major, _, _ := strings.Cut(version, ".")
	if major != "1" {
		return fmt.Errorf("unsupported Terraform JSON format major version %q", version)
	}
	return nil
}

func appendModuleResources(resources *[]Resource, module stateModule) {
	for _, stateResource := range module.Resources {
		resource := Resource{
			Address:             stateResource.Address,
			Mode:                stateResource.Mode,
			Type:                stateResource.Type,
			Name:                stateResource.Name,
			ProviderName:        stateResource.ProviderName,
			Supported:           supportedProviderResource(stateResource.ProviderName, stateResource.Type),
			SensitiveAttributes: sensitiveAttributePaths(stateResource.SensitiveValues),
		}
		if strings.TrimSpace(resource.Address) == "" {
			resource.Address = resource.Type + "." + resource.Name
		}
		*resources = append(*resources, resource)
	}
	for _, child := range module.ChildModules {
		appendModuleResources(resources, child)
	}
}

func supportedProviderResource(providerName string, resourceType string) bool {
	if providerName != "registry.terraform.io/hashicorp/aws" {
		return false
	}
	switch resourceType {
	case "aws_db_instance", "aws_instance", "aws_lb", "aws_s3_bucket", "aws_security_group", "aws_subnet", "aws_vpc":
		return true
	default:
		return false
	}
}

func sensitiveAttributePaths(values map[string]any) []string {
	var paths []string
	collectSensitivePaths(&paths, "", values)
	sort.Strings(paths)
	return paths
}

func collectSensitivePaths(paths *[]string, prefix string, value any) {
	switch typed := value.(type) {
	case bool:
		if typed && prefix != "" {
			*paths = append(*paths, prefix)
		}
	case []any:
		for index, child := range typed {
			childPrefix := fmt.Sprintf("%s[%d]", prefix, index)
			collectSensitivePaths(paths, childPrefix, child)
		}
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			childPrefix := key
			if prefix != "" {
				childPrefix = prefix + "." + key
			}
			collectSensitivePaths(paths, childPrefix, typed[key])
		}
	}
}

func sortResources(resources []Resource) {
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Address < resources[j].Address
	})
}
