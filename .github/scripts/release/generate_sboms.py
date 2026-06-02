#!/usr/bin/env python3
# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

"""Generate release-level SPDX and CycloneDX SBOMs from Go module metadata."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import subprocess
import uuid
from pathlib import Path


PROJECT_NAME = "planwright"
PROJECT_MODULE = "github.com/steadytao/planwright"
PROJECT_SUPPLIER = "Organization: The Planwright Authors"
SPDX_NAMESPACE = uuid.UUID("d3d28e5a-9c09-4c74-9184-808e61b5d586")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True, help="release version or snapshot label")
    parser.add_argument("--out", default="dist/release-assets", help="release asset output directory")
    parser.add_argument("--module-json", help="read go list module JSON from this file instead of running go")
    parser.add_argument("--created", help="creation timestamp for deterministic tests")
    args = parser.parse_args()

    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)

    created = args.created or dt.datetime.now(dt.UTC).replace(microsecond=0).isoformat().replace("+00:00", "Z")
    modules = read_modules(args.module_json)
    release_version = args.version.removeprefix("v")

    write_json(out / "planwright_sbom.spdx.json", spdx_document(modules, args.version, release_version, created))
    write_json(out / "planwright_sbom.cdx.json", cyclonedx_document(modules, args.version, release_version, created))
    return 0


def read_modules(module_json_path: str | None) -> list[dict]:
    if module_json_path:
        raw = Path(module_json_path).read_text(encoding="utf-8")
    else:
        result = subprocess.run(
            ["go", "list", "-m", "-json", "all"],
            check=True,
            stdout=subprocess.PIPE,
            text=True,
        )
        raw = result.stdout
    return parse_json_stream(raw)


def parse_json_stream(raw: str) -> list[dict]:
    decoder = json.JSONDecoder()
    index = 0
    modules = []
    while index < len(raw):
        while index < len(raw) and raw[index].isspace():
            index += 1
        if index >= len(raw):
            break
        value, index = decoder.raw_decode(raw, index)
        modules.append(value)
    if not modules:
        raise SystemExit("go module metadata is empty")
    return modules


def spdx_document(modules: list[dict], version_tag: str, release_version: str, created: str) -> dict:
    packages = [spdx_package(module, release_version) for module in modules]
    document_namespace = f"https://github.com/steadytao/planwright/releases/download/{version_tag}/planwright-sbom"
    return {
        "spdxVersion": "SPDX-2.3",
        "dataLicense": "CC0-1.0",
        "SPDXID": "SPDXRef-DOCUMENT",
        "name": f"{PROJECT_NAME}-{version_tag}",
        "documentNamespace": document_namespace,
        "creationInfo": {
            "created": created,
            "creators": [
                "Tool: Planwright release SBOM generator",
                PROJECT_SUPPLIER,
            ],
        },
        "packages": packages,
        "relationships": [
            {
                "spdxElementId": "SPDXRef-DOCUMENT",
                "relationshipType": "DESCRIBES",
                "relatedSpdxElement": "SPDXRef-Package-planwright",
            }
        ],
    }


def spdx_package(module: dict, release_version: str) -> dict:
    module_path, module_version = module_identity(module, release_version)
    is_project = module_path == PROJECT_MODULE
    return {
        "name": module_path,
        "SPDXID": spdx_id(module_path),
        "downloadLocation": "NOASSERTION",
        "filesAnalyzed": False,
        "versionInfo": module_version,
        "licenseConcluded": "Apache-2.0" if is_project else "NOASSERTION",
        "licenseDeclared": "Apache-2.0" if is_project else "NOASSERTION",
        "copyrightText": "NOASSERTION",
        "externalRefs": [
            {
                "referenceCategory": "PACKAGE-MANAGER",
                "referenceType": "purl",
                "referenceLocator": package_url(module_path, module_version),
            }
        ],
    }


def cyclonedx_document(modules: list[dict], version_tag: str, release_version: str, created: str) -> dict:
    components = [cyclonedx_component(module, release_version) for module in modules]
    serial = uuid.uuid5(SPDX_NAMESPACE, f"{PROJECT_NAME}:{version_tag}")
    return {
        "bomFormat": "CycloneDX",
        "specVersion": "1.6",
        "serialNumber": f"urn:uuid:{serial}",
        "version": 1,
        "metadata": {
            "timestamp": created,
            "tools": {
                "components": [
                    {
                        "type": "application",
                        "name": "Planwright release SBOM generator",
                    }
                ]
            },
            "component": {
                "type": "application",
                "name": PROJECT_NAME,
                "version": release_version,
                "purl": package_url(PROJECT_MODULE, release_version),
            },
        },
        "components": components,
    }


def cyclonedx_component(module: dict, release_version: str) -> dict:
    module_path, module_version = module_identity(module, release_version)
    component = {
        "type": "application" if module_path == PROJECT_MODULE else "library",
        "bom-ref": package_url(module_path, module_version),
        "name": module_path,
        "version": module_version,
        "purl": package_url(module_path, module_version),
    }
    if module_path == PROJECT_MODULE:
        component["licenses"] = [{"license": {"id": "Apache-2.0"}}]
    return component


def module_identity(module: dict, release_version: str) -> tuple[str, str]:
    replacement = module.get("Replace")
    if replacement:
        module = replacement
    module_path = module["Path"]
    module_version = module.get("Version") or release_version
    return module_path, module_version


def package_url(module_path: str, module_version: str) -> str:
    return f"pkg:golang/{module_path}@{module_version}"


def spdx_id(module_path: str) -> str:
    if module_path == PROJECT_MODULE:
        return "SPDXRef-Package-planwright"
    normalised = re.sub(r"[^A-Za-z0-9.-]+", "-", module_path).strip("-")
    return f"SPDXRef-Package-{normalised}"


def write_json(path: Path, document: dict) -> None:
    path.write_text(json.dumps(document, indent=2, sort_keys=True) + "\n", encoding="utf-8", newline="\n")


if __name__ == "__main__":
    raise SystemExit(main())
