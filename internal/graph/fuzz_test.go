// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import "testing"

func FuzzValidateJSON(f *testing.F) {
	f.Add([]byte(`{
  "version": "planwright.graph.v1",
  "provider": "aws",
  "region": "ap-southeast-2",
  "nodes": [
    {"id": "internet", "kind": "external.internet"},
    {"id": "app", "kind": "aws.ecs.service"}
  ],
  "edges": [
    {"from": "internet", "to": "app", "kind": "network.allow", "protocol": "tcp", "port": 443}
  ]
}`))
	f.Add([]byte(`{"version":"planwright.graph.v1","provider":"aws","region":"ap-southeast-2","nodes":[],"edges":[]}`))
	f.Add([]byte(`{`))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		ValidateJSON(data, "fuzz.graph.json")
	})
}
