// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const jsonSchemaID = "urn:planwright:schema:graph:v1"

//go:embed schemas/planwright.graph.v1.schema.json
var graphJSONSchema []byte

var (
	compiledGraphSchemaOnce sync.Once
	compiledGraphSchema     *jsonschema.Schema
	errCompiledGraphSchema  error
)

func JSONSchema() []byte {
	return append([]byte(nil), graphJSONSchema...)
}

func ValidateJSON(data []byte, source string) (Graph, []Diagnostic) {
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return Graph{}, []Diagnostic{{
			Severity: SeverityError,
			Code:     "PW-GRAPH-JSON-001",
			Resource: source,
			Message:  fmt.Sprintf("Graph JSON could not be parsed: %v.", err),
			Fix:      "Provide valid JSON encoded as planwright.graph.v1.",
		}}
	}

	schema, err := compiledSchema()
	if err != nil {
		return Graph{}, []Diagnostic{{
			Severity: SeverityError,
			Code:     "PW-GRAPH-SCHEMA-ENGINE-001",
			Resource: "graph schema",
			Message:  fmt.Sprintf("Planwright graph schema could not be compiled: %v.", err),
			Fix:      "Report this as a Planwright bug.",
		}}
	}
	if err := schema.Validate(instance); err != nil {
		return Graph{}, []Diagnostic{{
			Severity: SeverityError,
			Code:     "PW-GRAPH-SCHEMA-001",
			Resource: source,
			Message:  fmt.Sprintf("Graph JSON does not match planwright.graph.v1 schema: %s.", strings.TrimSpace(err.Error())),
			Fix:      "Fix the graph JSON structure before running semantic validation.",
		}}
	}

	var loaded Graph
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loaded); err != nil {
		return Graph{}, []Diagnostic{{
			Severity: SeverityError,
			Code:     "PW-GRAPH-JSON-002",
			Resource: source,
			Message:  fmt.Sprintf("Graph JSON could not be decoded into the Planwright graph model: %v.", err),
			Fix:      "Check that graph fields match the Planwright graph schema.",
		}}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Graph{}, []Diagnostic{{
			Severity: SeverityError,
			Code:     "PW-GRAPH-JSON-003",
			Resource: source,
			Message:  "Graph JSON contains trailing content.",
			Fix:      "Provide exactly one graph JSON document.",
		}}
	}

	return loaded, Validate(loaded)
}

func compiledSchema() (*jsonschema.Schema, error) {
	compiledGraphSchemaOnce.Do(func() {
		var schemaDoc any
		if err := json.Unmarshal(graphJSONSchema, &schemaDoc); err != nil {
			errCompiledGraphSchema = err
			return
		}
		compiler := jsonschema.NewCompiler()
		compiler.DefaultDraft(jsonschema.Draft2020)
		if err := compiler.AddResource(jsonSchemaID, schemaDoc); err != nil {
			errCompiledGraphSchema = err
			return
		}
		compiledGraphSchema, errCompiledGraphSchema = compiler.Compile(jsonSchemaID)
	})
	return compiledGraphSchema, errCompiledGraphSchema
}
