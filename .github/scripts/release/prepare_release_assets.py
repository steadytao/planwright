#!/usr/bin/env python3
# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

"""Prepare stable release asset names and checksum manifests."""

from __future__ import annotations

import argparse
import hashlib
import shutil
from pathlib import Path


TARGETS = (
    ("windows", "amd64", "planwright.exe", "planwright_windows_amd64.exe"),
    ("windows", "arm64", "planwright.exe", "planwright_windows_arm64.exe"),
    ("linux", "amd64", "planwright", "planwright_linux_amd64"),
    ("linux", "arm64", "planwright", "planwright_linux_arm64"),
    ("darwin", "amd64", "planwright", "planwright_darwin_amd64"),
    ("darwin", "arm64", "planwright", "planwright_darwin_arm64"),
)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--dist", default="dist", help="GoReleaser dist directory")
    parser.add_argument("--out", default="dist/release-assets", help="release asset output directory")
    parser.add_argument("--skip-checksums", action="store_true", help="copy binaries without writing checksum manifests")
    parser.add_argument("--checksums-only", action="store_true", help="write checksum manifests for an existing asset directory")
    args = parser.parse_args()

    out = Path(args.out)
    if args.skip_checksums and args.checksums_only:
        raise SystemExit("--skip-checksums and --checksums-only cannot be used together")

    if args.checksums_only:
        if not out.is_dir():
            raise SystemExit(f"release asset output directory does not exist: {out}")
        write_release_checksums(out)
        return 0

    dist = Path(args.dist)
    if not dist.is_dir():
        raise SystemExit(f"dist directory does not exist: {dist}")

    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)

    for goos, goarch, binary_name, asset_name in TARGETS:
        source = find_built_binary(dist, goos, goarch, binary_name)
        destination = out / asset_name
        shutil.copy2(source, destination)

    if not args.skip_checksums:
        write_release_checksums(out)
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


def write_checksums(path: Path, files: list[Path], algorithm: str) -> None:
    lines = []
    for file_path in sorted(files, key=lambda item: item.name):
        digest = file_digest(file_path, algorithm)
        lines.append(f"{digest}  {file_path.name}\n")
    path.write_text("".join(lines), encoding="utf-8", newline="\n")


def write_release_checksums(out: Path) -> None:
    files = checksum_inputs(out)
    write_checksums(out / "SHA2-256SUMS", files, "sha256")
    write_checksums(out / "SHA2-512SUMS", files, "sha512")


def checksum_inputs(out: Path) -> list[Path]:
    files = []
    for path in out.iterdir():
        if not path.is_file() or should_exclude_from_checksums(path.name):
            continue
        files.append(path)
    if not files:
        raise SystemExit(f"no release assets found for checksums in {out}")
    return files


def should_exclude_from_checksums(name: str) -> bool:
    return (
        name in {"SHA2-256SUMS", "SHA2-512SUMS", "public.key"}
        or name.endswith(".sig")
    )


def file_digest(path: Path, algorithm: str) -> str:
    digest = hashlib.new(algorithm)
    with path.open("rb") as file:
        for chunk in iter(lambda: file.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


if __name__ == "__main__":
    raise SystemExit(main())
