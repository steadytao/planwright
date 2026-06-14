// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func ValidateSARIF(data []byte, sourceName string) error {
	var log sarifLog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&log); err != nil {
		return fmt.Errorf("parse %s: %w", sourceName, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("parse %s: SARIF JSON contains trailing content", sourceName)
	}
	if log.Version != "2.1.0" {
		return fmt.Errorf("validate %s: version must be 2.1.0", sourceName)
	}
	if len(log.Runs) == 0 {
		return fmt.Errorf("validate %s: runs must contain at least one run", sourceName)
	}
	for runIndex, run := range log.Runs {
		if strings.TrimSpace(run.Tool.Driver.Name) == "" {
			return fmt.Errorf("validate %s: runs[%d].tool.driver.name must not be empty", sourceName, runIndex)
		}
		for resultIndex, result := range run.Results {
			if strings.TrimSpace(result.RuleID) == "" {
				return fmt.Errorf("validate %s: runs[%d].results[%d].ruleId must not be empty", sourceName, runIndex, resultIndex)
			}
			if strings.TrimSpace(result.Message.Text) == "" {
				return fmt.Errorf("validate %s: runs[%d].results[%d].message.text must not be empty", sourceName, runIndex, resultIndex)
			}
		}
	}
	return nil
}
