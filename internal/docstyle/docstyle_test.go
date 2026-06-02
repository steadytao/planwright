// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package docstyle

import "testing"

func TestCheckMarkdownFlagsBlankLineBeforeBullet(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("This is text:\n\n- This is a dot point.\n"))
	assertFinding(t, findings, "PW-DOC-LIST-001")
}

func TestCheckMarkdownAllowsBulletHuggingText(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("This is text:\n- This is a dot point.\n"))
	assertNoFinding(t, findings, "PW-DOC-LIST-001")
}

func TestCheckMarkdownFlagsBlankLineBeforeCodeFence(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("Run this:\n\n```bash\ngo test ./...\n```\n"))
	assertFinding(t, findings, "PW-DOC-CODE-001")
}

func TestCheckMarkdownAllowsCodeFenceHuggingText(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("Run this:\n```bash\ngo test ./...\n```\n"))
	assertNoFinding(t, findings, "PW-DOC-CODE-001")
}

func TestCheckMarkdownFlagsSimpleCommaBeforeAnd(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("This, and this should not use that comma.\n"))
	assertFinding(t, findings, "PW-DOC-COMMA-002")
}

func TestCheckMarkdownFlagsForbiddenCommaBeforeWords(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"This is mostly right, but the comma should not be there.\n",
		"This applies to inputs, including YAML.\n",
		"This is acceptable, however it should be rewritten.\n",
		"Use this path, or rewrite the sentence.\n",
		"Verify the signature, then check the downloaded file.\n",
	} {
		findings := CheckMarkdown("README.md", []byte(input))
		assertFinding(t, findings, "PW-DOC-COMMA-001")
	}
}

func TestCheckMarkdownAllowsTopicSeparatingCommaBeforeAnd(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("This and this, and that can use the comma for separation.\n"))
	assertNoFinding(t, findings, "PW-DOC-COMMA-002")
}

func TestCheckMarkdownFlagsAmericanEnglish(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("The current behavior should use British English.\n"))
	assertFinding(t, findings, "PW-DOC-EN-GB-001")
}

func TestCheckMarkdownIgnoresFencedCodeBlocks(t *testing.T) {
	t.Parallel()

	findings := CheckMarkdown("README.md", []byte("```text\nThis behavior is inside code, but should be ignored.\n\n- So should this list.\n```\n"))
	for _, code := range []string{"PW-DOC-EN-GB-001", "PW-DOC-COMMA-001", "PW-DOC-LIST-001"} {
		assertNoFinding(t, findings, code)
	}
}

func assertFinding(t *testing.T, findings []Finding, code string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Code == code {
			return
		}
	}
	t.Fatalf("findings = %#v, want %s", findings, code)
}

func assertNoFinding(t *testing.T, findings []Finding, code string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Code == code {
			t.Fatalf("findings = %#v, did not want %s", findings, code)
		}
	}
}
