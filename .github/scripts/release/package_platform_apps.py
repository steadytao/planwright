#!/usr/bin/env python3
# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

"""Create platform icon packages from GoReleaser output."""

from __future__ import annotations

import argparse
import shutil
import subprocess
import zipfile
from pathlib import Path


MACOS_TARGETS = (
    ("amd64", "planwright_darwin_amd64_app.zip"),
    ("arm64", "planwright_darwin_arm64_app.zip"),
)

LINUX_TARGETS = (
    ("amd64", "planwright_linux_amd64_desktop.zip"),
    ("arm64", "planwright_linux_arm64_desktop.zip"),
)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dist", default="dist", help="GoReleaser dist directory")
    parser.add_argument("--out", default="dist/release-assets", help="release asset output directory")
    parser.add_argument("--icon", default=".github/generated/planwright.svg", help="generated Planwright SVG icon")
    parser.add_argument("--version", required=True, help="package metadata version")
    args = parser.parse_args()

    dist = Path(args.dist)
    out = Path(args.out)
    icon = Path(args.icon)

    if not dist.is_dir():
        raise SystemExit(f"dist directory does not exist: {dist}")
    if not out.is_dir():
        raise SystemExit(f"release asset output directory does not exist: {out}")
    if not icon.is_file():
        raise SystemExit(f"generated icon does not exist: {icon}")

    for arch, asset_name in LINUX_TARGETS:
        binary = find_built_binary(dist, "linux", arch, "planwright")
        write_linux_desktop_package(binary, icon, out / asset_name)

    for arch, asset_name in MACOS_TARGETS:
        binary = find_built_binary(dist, "darwin", arch, "planwright")
        write_macos_app_package(binary, icon, out / asset_name, package_version(args.version))

    return 0


def find_built_binary(dist: Path, goos: str, goarch: str, binary_name: str) -> Path:
    matches = [
        path
        for path in dist.glob(f"*_{goos}_{goarch}*/{binary_name}")
        if path.is_file()
    ]
    if len(matches) != 1:
        formatted = ", ".join(str(match) for match in matches) or "none"
        raise SystemExit(
            f"expected exactly one {goos}/{goarch} binary named {binary_name}; found {formatted}"
        )
    return matches[0]


def write_linux_desktop_package(binary: Path, icon: Path, destination: Path) -> None:
    stage = destination.with_suffix("")
    if stage.exists():
        shutil.rmtree(stage)
    stage.mkdir(parents=True)

    bin_dir = stage / "bin"
    icon_dir = stage / "share" / "icons" / "hicolor" / "scalable" / "apps"
    desktop_dir = stage / "share" / "applications"
    bin_dir.mkdir(parents=True)
    icon_dir.mkdir(parents=True)
    desktop_dir.mkdir(parents=True)

    copy_executable(binary, bin_dir / "planwright")
    shutil.copy2(icon, icon_dir / "planwright.svg")
    (desktop_dir / "planwright.desktop").write_text(linux_desktop_entry(), encoding="utf-8", newline="\n")

    write_zip(destination, stage)
    shutil.rmtree(stage)


def write_macos_app_package(binary: Path, icon: Path, destination: Path, version: str) -> None:
    stage = destination.with_suffix("")
    if stage.exists():
        shutil.rmtree(stage)
    app = stage / "Planwright.app"
    contents = app / "Contents"
    macos = contents / "MacOS"
    resources = contents / "Resources"
    macos.mkdir(parents=True)
    resources.mkdir(parents=True)

    copy_executable(binary, macos / "planwright")
    write_icns(icon, resources / "planwright.icns")
    (contents / "Info.plist").write_text(macos_info_plist(version), encoding="utf-8", newline="\n")

    write_zip(destination, stage)
    shutil.rmtree(stage)


def copy_executable(source: Path, destination: Path) -> None:
    shutil.copy2(source, destination)
    mode = destination.stat().st_mode
    destination.chmod(mode | 0o755)


def write_icns(icon: Path, destination: Path) -> None:
    image_convert = image_converter()

    tmp = destination.parent / "planwright.iconset"
    if tmp.exists():
        shutil.rmtree(tmp)
    tmp.mkdir()
    try:
        sizes = (16, 32, 128, 256, 512)
        for size in sizes:
            write_png(image_convert, icon, tmp / f"icon_{size}x{size}.png", size)
            write_png(image_convert, icon, tmp / f"icon_{size}x{size}@2x.png", size * 2)

        write_icns_from_pngs(destination, tmp)
    finally:
        shutil.rmtree(tmp)


def image_converter() -> str:
    for command in ("magick", "convert"):
        resolved = shutil.which(command)
        if resolved:
            return resolved
    raise SystemExit("ImageMagick is required to generate platform icons")


def write_png(command: str, icon: Path, destination: Path, size: int) -> None:
    subprocess.run(
        [
            command,
            "-background",
            "none",
            "-alpha",
            "on",
            str(icon),
            "-resize",
            f"{size}x{size}",
            "-define",
            "png:color-type=6",
            str(destination),
        ],
        check=True,
    )


def write_icns_from_pngs(destination: Path, iconset: Path) -> None:
    entries = [
        ("icp4", iconset / "icon_16x16.png"),
        ("icp5", iconset / "icon_32x32.png"),
        ("icp6", iconset / "icon_32x32@2x.png"),
        ("ic07", iconset / "icon_128x128.png"),
        ("ic08", iconset / "icon_256x256.png"),
        ("ic09", iconset / "icon_512x512.png"),
        ("ic10", iconset / "icon_512x512@2x.png"),
    ]
    chunks = []
    for kind, path in entries:
        data = path.read_bytes()
        chunks.append(kind.encode("ascii") + (len(data) + 8).to_bytes(4, "big") + data)

    total_size = 8 + sum(len(chunk) for chunk in chunks)
    destination.write_bytes(b"icns" + total_size.to_bytes(4, "big") + b"".join(chunks))


def write_zip(destination: Path, root: Path) -> None:
    if destination.exists():
        destination.unlink()
    with zipfile.ZipFile(destination, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        for path in sorted(root.rglob("*")):
            if path.is_dir():
                continue
            archive_path = path.relative_to(root).as_posix()
            archive.write(path, archive_path)


def linux_desktop_entry() -> str:
    return """[Desktop Entry]
Type=Application
Name=Planwright
Comment=Local-first infrastructure planning engine
Exec=planwright
Icon=planwright
Terminal=true
Categories=Development;System;
"""


def package_version(version: str) -> str:
    if version == "snapshot":
        return "0.0.0"
    return version.removeprefix("v")


def macos_info_plist(version: str) -> str:
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "https://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>planwright</string>
  <key>CFBundleIconFile</key>
  <string>planwright</string>
  <key>CFBundleIdentifier</key>
  <string>dev.steadytao.planwright</string>
  <key>CFBundleName</key>
  <string>Planwright</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>{version}</string>
  <key>CFBundleVersion</key>
  <string>{version}</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
"""


if __name__ == "__main__":
    raise SystemExit(main())
