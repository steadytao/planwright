# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import hashlib
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).resolve().parents[1] / "prepare_release_assets.py"


class PrepareReleaseAssetsTests(unittest.TestCase):
    def test_prepare_release_assets_copies_binaries_and_writes_checksums(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            dist = root / "dist"
            out = root / "out"
            fixtures = {
                "planwright_windows_amd64_v1/planwright.exe": b"win-amd64",
                "planwright_windows_arm64_v8.0/planwright.exe": b"win-arm64",
                "planwright_linux_amd64_v1/planwright": b"linux-amd64",
                "planwright_linux_arm64_v8.0/planwright": b"linux-arm64",
                "planwright_darwin_amd64_v1/planwright": b"darwin-amd64",
                "planwright_darwin_arm64_v8.0/planwright": b"darwin-arm64",
            }

            for relative_path, contents in fixtures.items():
                path = dist / relative_path
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(contents)

            subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--dist",
                    str(dist),
                    "--out",
                    str(out),
                ],
                check=True,
            )

            expected_assets = {
                "planwright_windows_amd64.exe": b"win-amd64",
                "planwright_windows_arm64.exe": b"win-arm64",
                "planwright_linux_amd64": b"linux-amd64",
                "planwright_linux_arm64": b"linux-arm64",
                "planwright_darwin_amd64": b"darwin-amd64",
                "planwright_darwin_arm64": b"darwin-arm64",
            }

            for name, contents in expected_assets.items():
                self.assertEqual((out / name).read_bytes(), contents)

            self.assertEqual(
                (out / "SHA2-256SUMS").read_text(encoding="utf-8"),
                checksum_text(expected_assets, "sha256"),
            )
            self.assertEqual(
                (out / "SHA2-512SUMS").read_text(encoding="utf-8"),
                checksum_text(expected_assets, "sha512"),
            )

    def test_prepare_release_assets_checksum_only_includes_sboms(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            out = root / "out"
            out.mkdir()
            assets = {
                "planwright_linux_amd64": b"linux-amd64",
                "planwright_sbom.cdx.json": b'{"bomFormat":"CycloneDX"}',
                "planwright_sbom.spdx.json": b'{"spdxVersion":"SPDX-2.3"}',
            }

            for name, contents in assets.items():
                (out / name).write_bytes(contents)

            for excluded in ("SHA2-256SUMS", "SHA2-512SUMS", "SHA2-256SUMS.sig", "public.key"):
                (out / excluded).write_text("ignored\n", encoding="utf-8")

            subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--out",
                    str(out),
                    "--checksums-only",
                ],
                check=True,
            )

            self.assertEqual(
                (out / "SHA2-256SUMS").read_text(encoding="utf-8"),
                checksum_text(assets, "sha256"),
            )
            self.assertEqual(
                (out / "SHA2-512SUMS").read_text(encoding="utf-8"),
                checksum_text(assets, "sha512"),
            )

    def test_prepare_release_assets_fails_when_expected_binary_is_missing(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            dist = root / "dist"
            out = root / "out"
            path = dist / "planwright_windows_amd64_v1" / "planwright.exe"
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b"win-amd64")

            result = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--dist",
                    str(dist),
                    "--out",
                    str(out),
                ],
                check=False,
                stderr=subprocess.PIPE,
                text=True,
            )

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("expected exactly one windows/arm64 binary", result.stderr)


def checksum_text(files, algorithm):
    lines = []
    for name in sorted(files):
        digest = hashlib.new(algorithm, files[name]).hexdigest()
        lines.append(f"{digest}  {name}\n")
    return "".join(lines)


if __name__ == "__main__":
    unittest.main()
