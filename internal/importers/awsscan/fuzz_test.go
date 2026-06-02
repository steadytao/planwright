// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package awsscan

import "testing"

func FuzzRejectDuplicateJSONKeys(f *testing.F) {
	f.Add([]byte(`{"schema":"planwright.awsscan.v1","region":"ap-southeast-2","profile":"lab"}`))
	f.Add([]byte(`{"region":"ap-southeast-2","region":"us-east-1"}`))
	f.Add([]byte(`{"Vpcs":[{"VpcId":"vpc-123"}]}`))
	f.Add([]byte(`{`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		if err := rejectDuplicateJSONKeys(data, "fuzz.json"); err != nil {
			return
		}
	})
}
