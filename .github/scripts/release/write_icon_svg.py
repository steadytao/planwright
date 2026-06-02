#!/usr/bin/env python3
# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

"""Generate the Planwright 256x256 icon SVG from the README banner."""

from __future__ import annotations

import argparse
import re
from pathlib import Path


SVG_OPEN_PATTERN = re.compile(r"<svg\b[^>]*>", re.IGNORECASE)
SVG_CLOSE_PATTERN = re.compile(r"</svg>\s*$", re.IGNORECASE)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--banner", required=True, help="source banner SVG")
    parser.add_argument("--out", required=True, help="generated icon SVG")
    args = parser.parse_args()

    banner = Path(args.banner)
    out = Path(args.out)

    source = banner.read_text(encoding="utf-8")
    body = SVG_OPEN_PATTERN.sub("", source, count=1)
    body = SVG_CLOSE_PATTERN.sub("", body)

    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(icon_svg(body), encoding="utf-8", newline="\n")
    return 0


def icon_svg(body: str) -> str:
    return f"""<svg xmlns="http://www.w3.org/2000/svg" width="256" height="256" viewBox="0 0 256 256" shape-rendering="crispEdges" role="img" aria-labelledby="title desc">
<title id="title">Planwright</title>
<desc id="desc">The Planwright project mark from the README banner.</desc>
<g transform="translate(31.5 0) scale(0.48577)">
{body.strip()}
</g>
</svg>
"""


if __name__ == "__main__":
    raise SystemExit(main())
