// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package loss

type Report struct {
	SourceFormat string
	Source       string
	Lowered      []Item
	Unsupported  []Item
	Ambiguous    []Item
	Preserved    []Item
}

type Item struct {
	Resource string
	Kind     string
	Message  string
}
