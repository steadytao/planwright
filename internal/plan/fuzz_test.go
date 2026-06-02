// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package plan

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte(`version: planwright.v1
provider: aws
region: ap-southeast-2
profile: lab
components:
  webapp:
    pattern: aws.webapp.alb_ecs_rds
    properties:
      app_port: 8080
      db_engine: postgres
      db_public: false
flows:
  - from: internet
    to: webapp.alb
    kind: network.allow
    protocol: tcp
    port: 443
`))
	f.Add([]byte("version: planwright.v2\n"))
	f.Add([]byte("not: [valid\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		document, err := Parse(data, "fuzz.yaml")
		if err != nil {
			return
		}
		document.ToGraph()
	})
}
