# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import os
from pathlib import Path
import sys
import tempfile
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

import check_action_pins


class CheckActionPinsTests(unittest.TestCase):
  def test_iter_workflow_uses_reads_yaml_and_quoted_uses(self) -> None:
    with tempfile.TemporaryDirectory() as temp_dir:
      root = Path(temp_dir)
      workflows = root / ".github" / "workflows"
      workflows.mkdir(parents=True)
      (workflows / "ci.yaml").write_text(
        "jobs:\n"
        "  test:\n"
        "    steps:\n"
        "      - uses: \"actions/checkout@de0fac2e4500dabe0009e67214ff5f5447ce83dd\" # v6.0.2\n",
        encoding="utf-8",
      )

      old_cwd = Path.cwd()
      try:
        os.chdir(root)
        uses = check_action_pins.iter_workflow_uses()
      finally:
        os.chdir(old_cwd)

    self.assertEqual(len(uses), 1)
    workflow, line_number, action_path, owner_repo, ref, tag = uses[0]
    self.assertEqual(workflow, Path(".github/workflows/ci.yaml"))
    self.assertEqual(line_number, 4)
    self.assertEqual(action_path, "actions/checkout")
    self.assertEqual(owner_repo, "actions/checkout")
    self.assertEqual(ref, "de0fac2e4500dabe0009e67214ff5f5447ce83dd")
    self.assertEqual(tag, "v6.0.2")

  def test_latest_semver_tag_ignores_major_minor_prerelease_and_deref_refs(self) -> None:
    latest = check_action_pins.latest_semver_tag([
      "refs/tags/v6",
      "refs/tags/v6.0",
      "refs/tags/v6.0.2",
      "refs/tags/v6.0.2^{}",
      "refs/tags/v6.1.0-alpha.1",
      "refs/tags/v6.1.0",
      "refs/tags/not-a-version",
    ])

    self.assertEqual(latest, "v6.1.0")

  def test_latest_semver_tag_returns_none_without_full_semver_tags(self) -> None:
    latest = check_action_pins.latest_semver_tag([
      "refs/tags/v6",
      "refs/tags/v6.0",
      "refs/tags/v6.1.0-alpha.1",
      "refs/tags/not-a-version",
    ])

    self.assertIsNone(latest)


if __name__ == "__main__":
  unittest.main()
