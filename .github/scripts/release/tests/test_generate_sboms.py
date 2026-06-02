# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).resolve().parents[1] / "generate_sboms.py"


class GenerateSBOMTests(unittest.TestCase):
    def test_generate_sboms_writes_spdx_and_cyclonedx_documents(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            module_json = root / "modules.json"
            out = root / "out"
            module_json.write_text(
                "\n".join(
                    [
                        json.dumps({"Path": "github.com/steadytao/planwright", "Main": True}),
                        json.dumps({"Path": "gopkg.in/yaml.v3", "Version": "v3.0.1"}),
                    ]
                ),
                encoding="utf-8",
            )

            subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--version",
                    "v0.11.0",
                    "--out",
                    str(out),
                    "--module-json",
                    str(module_json),
                    "--created",
                    "2026-06-01T00:00:00Z",
                ],
                check=True,
            )

            spdx = json.loads((out / "planwright_sbom.spdx.json").read_text(encoding="utf-8"))
            cdx = json.loads((out / "planwright_sbom.cdx.json").read_text(encoding="utf-8"))

            self.assertEqual(spdx["spdxVersion"], "SPDX-2.3")
            self.assertEqual(spdx["creationInfo"]["created"], "2026-06-01T00:00:00Z")
            self.assertEqual(spdx["packages"][0]["versionInfo"], "0.11.0")
            self.assertEqual(spdx["packages"][0]["licenseDeclared"], "Apache-2.0")
            self.assertEqual(spdx["packages"][1]["name"], "gopkg.in/yaml.v3")

            self.assertEqual(cdx["bomFormat"], "CycloneDX")
            self.assertEqual(cdx["specVersion"], "1.6")
            self.assertEqual(cdx["metadata"]["component"]["version"], "0.11.0")
            self.assertEqual(cdx["components"][1]["purl"], "pkg:golang/gopkg.in/yaml.v3@v3.0.1")


if __name__ == "__main__":
    unittest.main()
