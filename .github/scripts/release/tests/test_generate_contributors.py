# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import unittest
from pathlib import Path
import sys

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

import generate_contributors


class GenerateContributorsTests(unittest.TestCase):
  def test_normalize_contributors_aggregates_by_email_and_filters_bots(self) -> None:
    contributors = generate_contributors.normalize_contributors(
      [
        ("Zen Dodd", "mail@steadytao.com"),
        ("Zen Dodd", "mail@steadytao.com"),
        ("Z Dodd", "mail@steadytao.com"),
        ("github-actions[bot]", "41898282+github-actions[bot]@users.noreply.github.com"),
        ("Other Person", "other@example.com"),
      ]
    )

    self.assertEqual(
      contributors,
      [
        generate_contributors.Contributor(name="Zen Dodd", email="mail@steadytao.com", commits=3),
        generate_contributors.Contributor(name="Other Person", email="other@example.com", commits=1),
      ],
    )

  def test_render_contributors_emits_expected_header_and_entries(self) -> None:
    rendered = generate_contributors.render_contributors(
      [
        generate_contributors.Contributor(name="Zen Dodd", email="mail@steadytao.com", commits=3),
        generate_contributors.Contributor(name="Other Person", email="", commits=1),
      ]
    )

    self.assertTrue(
      rendered.startswith(
        "# This file is generated from Planwright's reachable non-bot commit history.\n"
        "# In an unborn repository, it is seeded from AUTHORS so contributor recognition\n"
      )
    )
    self.assertIn("Zen Dodd <mail@steadytao.com>\n", rendered)
    self.assertTrue(rendered.endswith("Other Person\n"))

  def test_collect_author_seed_rows_reads_curated_authors(self) -> None:
    import tempfile

    with tempfile.TemporaryDirectory() as temp_dir:
      root = Path(temp_dir)
      (root / "AUTHORS").write_text(
        "# comments are ignored\n\nZen Dodd (mail@steadytao.com - github.com/steadytao)\n",
        encoding="utf-8",
      )

      self.assertEqual(
        generate_contributors.collect_author_seed_rows(root),
        [("Zen Dodd", "mail@steadytao.com")],
      )


if __name__ == "__main__":
  unittest.main()
