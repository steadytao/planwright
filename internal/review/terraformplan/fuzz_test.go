// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package terraformplan

import "testing"

func FuzzReviewBytes(f *testing.F) {
	f.Add([]byte(`{
  "format_version": "1.0",
  "terraform_version": "1.15.5",
  "resource_changes": [
    {
      "address": "aws_db_instance.app",
      "mode": "managed",
      "type": "aws_db_instance",
      "name": "app",
      "change": {
        "actions": ["update"],
        "after": {"publicly_accessible": true}
      }
    }
  ]
}`))
	f.Add([]byte(`{"format_version":"1.0","resource_changes":[]}`))
	f.Add([]byte(`{"format_version":"2.0"}`))
	f.Add([]byte(`{`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		if _, err := ReviewBytes(data, "fuzz.tfplan.json"); err != nil {
			return
		}
	})
}
