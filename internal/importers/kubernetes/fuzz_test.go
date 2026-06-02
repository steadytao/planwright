// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import "testing"

func FuzzParseSource(f *testing.F) {
	f.Add([]byte(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: demo
spec:
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: app
          image: example/app:1.0
`))
	f.Add([]byte(`apiVersion: v1
kind: List
items:
  - apiVersion: v1
    kind: Service
    metadata:
      name: app
    spec:
      ports:
        - port: 80
`))
	f.Add([]byte("apiVersion: v1\nkind: List\nitems:\n  - plain-string\n"))
	f.Add([]byte("kind: [\n"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			return
		}
		if _, err := parseSource(SourceFile{Path: "fuzz.yaml", Data: data}); err != nil {
			return
		}
	})
}
