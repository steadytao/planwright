# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import shutil
import subprocess
import sys
import tempfile
import unittest
import zipfile
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
WRITE_ICON_SVG = SCRIPTS / "write_icon_svg.py"
PACKAGE_PLATFORM_APPS = SCRIPTS / "package_platform_apps.py"


class IconPackagingTests(unittest.TestCase):
    def test_write_icon_svg_generates_256_viewbox_from_banner_body(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            banner = root / "banner.svg"
            out = root / "planwright.svg"
            banner.write_text(
                '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 527 527"><path d="M1 2h3v4H1z"/></svg>',
                encoding="utf-8",
            )

            subprocess.run(
                [sys.executable, str(WRITE_ICON_SVG), "--banner", str(banner), "--out", str(out)],
                check=True,
            )

            generated = out.read_text(encoding="utf-8")
            self.assertIn('viewBox="0 0 256 256"', generated)
            self.assertIn('transform="translate(31.5 0) scale(0.48577)"', generated)
            self.assertIn('<path d="M1 2h3v4H1z"/>', generated)

    @unittest.skipIf(not shutil.which("magick") and not shutil.which("convert"), "ImageMagick is unavailable")
    def test_package_platform_apps_writes_desktop_and_macos_app_packages(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            dist = root / "dist"
            out = root / "out"
            icon = root / "planwright.svg"
            out.mkdir()
            icon.write_text(
                '<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">'
                '<rect width="256" height="256" fill="#2064FC"/></svg>',
                encoding="utf-8",
            )

            for goos, goarch, binary_name in (
                ("linux", "amd64", "planwright"),
                ("linux", "arm64", "planwright"),
                ("darwin", "amd64", "planwright"),
                ("darwin", "arm64", "planwright"),
            ):
                binary = dist / f"planwright_{goos}_{goarch}_v1" / binary_name
                binary.parent.mkdir(parents=True, exist_ok=True)
                binary.write_bytes(f"{goos}-{goarch}".encode("utf-8"))

            subprocess.run(
                [
                    sys.executable,
                    str(PACKAGE_PLATFORM_APPS),
                    "--dist",
                    str(dist),
                    "--out",
                    str(out),
                    "--icon",
                    str(icon),
                    "--version",
                    "v0.11.0",
                ],
                check=True,
            )

            with zipfile.ZipFile(out / "planwright_linux_amd64_desktop.zip") as archive:
                self.assertIn("bin/planwright", archive.namelist())
                self.assertIn("share/applications/planwright.desktop", archive.namelist())
                self.assertIn("share/icons/hicolor/scalable/apps/planwright.svg", archive.namelist())

            with zipfile.ZipFile(out / "planwright_darwin_arm64_app.zip") as archive:
                self.assertIn("Planwright.app/Contents/MacOS/planwright", archive.namelist())
                self.assertIn("Planwright.app/Contents/Resources/planwright.icns", archive.namelist())
                info_plist = archive.read("Planwright.app/Contents/Info.plist").decode("utf-8")
                self.assertIn("<string>0.11.0</string>", info_plist)

    @unittest.skipIf(not shutil.which("magick") and not shutil.which("convert"), "ImageMagick is unavailable")
    def test_generated_png_icon_preserves_transparent_corners(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            icon = root / "planwright.svg"
            png = root / "planwright.png"
            icon.write_text(
                '<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256">'
                '<rect x="96" y="96" width="64" height="64" fill="#2064FC"/></svg>',
                encoding="utf-8",
            )

            command = shutil.which("magick") or shutil.which("convert")
            subprocess.run(
                [
                    command,
                    "-background",
                    "none",
                    "-alpha",
                    "on",
                    str(icon),
                    "-resize",
                    "256x256",
                    "-define",
                    "png:color-type=6",
                    str(png),
                ],
                check=True,
            )
            result = subprocess.run(
                [command, str(png), "-format", "%[pixel:p{0,0}]", "info:"],
                check=True,
                stdout=subprocess.PIPE,
                text=True,
            )

            pixel = result.stdout.lower()
            self.assertTrue("none" in pixel or pixel.endswith(",0)"), pixel)


if __name__ == "__main__":
    unittest.main()
