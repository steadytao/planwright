// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package yamlutil

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// RejectDuplicateMappingKeys rejects YAML documents with repeated mapping keys.
func RejectDuplicateMappingKeys(root *yaml.Node, sourceName string) error {
	return rejectDuplicateMappingKeys(root, sourceName, "$")
}

func rejectDuplicateMappingKeys(node *yaml.Node, sourceName string, path string) error {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := rejectDuplicateMappingKeys(child, sourceName, path); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		seen := map[string]struct{}{}
		for index := 0; index+1 < len(node.Content); index += 2 {
			keyNode := node.Content[index]
			key := keyNode.Value
			if _, ok := seen[key]; ok {
				return fmt.Errorf("parse %s: duplicate mapping key %q at %s", sourceName, key, path)
			}
			seen[key] = struct{}{}
			if err := rejectDuplicateMappingKeys(node.Content[index+1], sourceName, path+"."+key); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for index, child := range node.Content {
			if err := rejectDuplicateMappingKeys(child, sourceName, path+"["+strconv.Itoa(index)+"]"); err != nil {
				return err
			}
		}
	case yaml.ScalarNode, yaml.AliasNode:
		return nil
	}
	return nil
}
