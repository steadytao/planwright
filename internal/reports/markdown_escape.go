// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package reports

import "strings"

func markdownCode(value string) string {
	clean := markdownInline(value)
	backticks := longestBacktickRun(clean) + 1
	fence := strings.Repeat("`", backticks)
	if strings.Contains(clean, "`") {
		return fence + " " + clean + " " + fence
	}
	return fence + clean + fence
}

func markdownText(value string) string {
	clean := markdownInline(value)
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(clean)
}

func markdownInline(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}

func longestBacktickRun(value string) int {
	longest := 0
	current := 0
	for _, char := range value {
		if char == '`' {
			current++
			if current > longest {
				longest = current
			}
			continue
		}
		current = 0
	}
	return longest
}
