#!/usr/bin/env bash
# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

root="$(git rev-parse --show-toplevel)"
generated="${root}/.github/generated"
icon_svg="${generated}/planwright.svg"
icon_ico="${generated}/planwright.ico"
winres_config="${generated}/winres.json"
resource_out="${root}/cmd/planwright/rsrc"

rm -rf "${generated}"
mkdir -p "${generated}"
trap 'rm -f "${icon_ico}"' EXIT

python3 "${root}/.github/scripts/release/write_icon_svg.py" \
  --banner "${root}/.github/banner.svg" \
  --out "${icon_svg}"

if command -v magick >/dev/null 2>&1; then
  image_convert=(magick)
elif command -v convert >/dev/null 2>&1; then
  image_convert=(convert)
else
  echo "ImageMagick is required to generate the temporary Windows icon." >&2
  exit 1
fi

"${image_convert[@]}" \
  -background none \
  -alpha on \
  -define png:color-type=6 \
  "${icon_svg}" \
  -define icon:auto-resize=256,128,64,48,32,16 \
  "${icon_ico}"

cat > "${winres_config}" <<EOF
{
  "RT_GROUP_ICON": {
    "APP": {
      "0000": "planwright.ico"
    }
  }
}
EOF

go run github.com/tc-hib/go-winres@v0.3.3 make \
  --in "${winres_config}" \
  --arch amd64,arm64 \
  --out "${resource_out}"
