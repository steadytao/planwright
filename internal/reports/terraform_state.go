// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"fmt"
	"strings"

	"github.com/steadytao/planwright/internal/importers/loss"
	"github.com/steadytao/planwright/internal/review/terraformstate"
)

func RenderTerraformStateInventory(result terraformstate.Result) string {
	var builder strings.Builder
	builder.WriteString("# Terraform State Inventory\n\n")
	builder.WriteString("Planwright state inventory is local evidence. It does not execute Terraform/OpenTofu, evaluate HCL, load provider plugins or prove deployability.\n\n")
	builder.WriteString("## Source\n\n")
	fmt.Fprintf(&builder, "- Source: %s\n", markdownCode(result.Source))
	fmt.Fprintf(&builder, "- Format version: %s\n", markdownCode(result.FormatVersion))
	if result.TerraformVersion != "" {
		fmt.Fprintf(&builder, "- Terraform version: %s\n", markdownCode(result.TerraformVersion))
	}
	fmt.Fprintf(&builder, "- Resources: %s\n\n", markdownCode(fmt.Sprintf("%d", result.ResourceCount)))
	builder.WriteString("## Resources\n\n")
	if len(result.Resources) == 0 {
		builder.WriteString("- No resources were found in the supported Terraform state JSON fields.\n\n")
	} else {
		for _, resource := range result.Resources {
			status := "unsupported"
			if resource.Supported {
				status = "supported inventory"
			}
			fmt.Fprintf(&builder, "- %s (%s, %s): %s\n", markdownCode(resource.Address), markdownCode(resource.Type), markdownCode(resource.ProviderName), status)
			if len(resource.SensitiveAttributes) > 0 {
				fmt.Fprintf(&builder, "  - Sensitive attributes: %s\n", markdownCode(strings.Join(resource.SensitiveAttributes, ", ")))
			}
		}
		builder.WriteString("\n")
	}
	builder.WriteString("## Review\n\n")
	builder.WriteString("State inventory does not lower resources into `planwright.graph.v1` in this release. Use it as review evidence only.\n")
	return builder.String()
}

func TerraformStateLossReport(result terraformstate.Result) loss.Report {
	report := loss.Report{
		SourceFormat: "terraform-state-json",
		Source:       result.Source,
	}
	for _, resource := range result.Resources {
		item := loss.Item{
			Resource: resource.Address,
			Kind:     resource.Type,
		}
		if resource.Supported {
			item.Message = "Resource inventory extracted from Terraform/OpenTofu state JSON. Graph lowering is not implemented for state JSON in this release."
			report.Preserved = append(report.Preserved, item)
			continue
		}
		item.Message = "Resource is present in Terraform/OpenTofu state JSON but is not in the current supported state inventory subset; manual review required."
		report.Unsupported = append(report.Unsupported, item)
	}
	return report
}
