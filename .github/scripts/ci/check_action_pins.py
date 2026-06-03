# Copyright 2026 The Planwright Authors
# SPDX-License-Identifier: Apache-2.0

import re
import subprocess
import sys
from pathlib import Path

USES_RE = re.compile(r"^\s*(?:-\s*)?uses:\s*[\"']?([A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*)@([0-9A-Za-z._/-]+)[\"']?(?:\s+#\s*(v[^\s]+))?\s*$")
FULL_SHA_RE = re.compile(r"^[0-9a-f]{40}$")
SEMVER_TAG_RE = re.compile(r"^refs/tags/(v(\d+)\.(\d+)\.(\d+))$")

SemverTag = tuple[tuple[int, int, int], str]

def iter_workflow_uses() -> list[tuple[Path, int, str, str, str, str | None]]:
  results: list[tuple[Path, int, str, str, str, str | None]] = []
  workflow_dir = Path(".github/workflows")
  workflows = sorted([*workflow_dir.glob("*.yml"), *workflow_dir.glob("*.yaml")])
  for workflow in workflows:
    for line_number, line in enumerate(workflow.read_text(encoding="utf-8").splitlines(), start=1):
      match = USES_RE.match(line)
      if not match:
        continue
      action_path, ref, tag = match.groups()
      owner_repo = "/".join(action_path.split("/")[:2])
      results.append((workflow, line_number, action_path, owner_repo, ref, tag))
  return results

def resolve_tag(owner_repo: str, tag: str) -> str:
  command = [
    "git",
    "ls-remote",
    f"https://github.com/{owner_repo}",
    f"refs/tags/{tag}",
    f"refs/tags/{tag}^{{}}",
  ]
  result = subprocess.run(command, check=False, capture_output=True, text=True)
  if result.returncode != 0:
    raise RuntimeError(f"failed to resolve {owner_repo}@{tag}: {result.stderr.strip()}")

  lines = [line for line in result.stdout.splitlines() if line.strip()]
  if not lines:
    raise RuntimeError(f"tag {tag} was not found for {owner_repo}")

  refs: dict[str, str] = {}
  for line in lines:
    sha, ref_name = line.split()
    refs[ref_name] = sha

  sha = refs.get(f"refs/tags/{tag}^{{}}", refs.get(f"refs/tags/{tag}"))
  if not sha:
    raise RuntimeError(f"tag {tag} was not found for {owner_repo}")

  if not FULL_SHA_RE.match(sha):
    raise RuntimeError(f"resolved ref for {owner_repo}@{tag} was not a full SHA: {sha}")
  return sha

def parse_semver_tag_ref(ref_name: str) -> SemverTag | None:
  match = SEMVER_TAG_RE.match(ref_name)
  if not match:
    return None
  tag = match.group(1)
  version = (int(match.group(2)), int(match.group(3)), int(match.group(4)))
  return version, tag

def latest_semver_tag(ref_names: list[str]) -> str | None:
  tags = [parsed for ref_name in ref_names if (parsed := parse_semver_tag_ref(ref_name)) is not None]
  if not tags:
    return None
  return max(tags, key=lambda item: item[0])[1]

def resolve_latest_semver_tag(owner_repo: str) -> tuple[str, str]:
  command = ["git", "ls-remote", "--tags", f"https://github.com/{owner_repo}"]
  result = subprocess.run(command, check=False, capture_output=True, text=True)
  if result.returncode != 0:
    raise RuntimeError(f"failed to list tags for {owner_repo}: {result.stderr.strip()}")

  ref_names: list[str] = []
  for line in result.stdout.splitlines():
    parts = line.split()
    if len(parts) != 2:
      continue
    ref_name = parts[1]
    if ref_name.endswith("^{}"):
      continue
    ref_names.append(ref_name)

  tag = latest_semver_tag(ref_names)
  if not tag:
    raise RuntimeError(f"no semver tags were found for {owner_repo}")
  return tag, resolve_tag(owner_repo, tag)

def main() -> int:
  failures: list[str] = []
  latest_cache: dict[str, tuple[str, str]] = {}

  for workflow, line_number, action_path, owner_repo, ref, tag in iter_workflow_uses():
    if action_path.startswith("./"):
      continue

    location = f"{workflow}:{line_number}"

    if not FULL_SHA_RE.match(ref):
      failures.append(f"{location}: {action_path}@{ref} is not pinned to a full 40-character SHA")
      continue

    if not tag:
      failures.append(f"{location}: {action_path}@{ref} is pinned but does not document an expected tag comment")
      continue

    try:
      resolved = resolve_tag(owner_repo, tag)
    except RuntimeError as exc:
      failures.append(f"{location}: {exc}")
      continue

    if resolved != ref:
      failures.append(
        f"{location}: {action_path} pinned to {ref} but {tag} currently resolves to {resolved}"
      )
      continue

    if owner_repo not in latest_cache:
      try:
        latest_cache[owner_repo] = resolve_latest_semver_tag(owner_repo)
      except RuntimeError as exc:
        failures.append(f"{location}: {exc}")
        continue

    latest_tag, latest_sha = latest_cache[owner_repo]
    if tag != latest_tag:
      failures.append(f"{location}: {action_path} documents {tag} but the latest semver tag is {latest_tag}")
      continue
    if ref != latest_sha:
      failures.append(f"{location}: {action_path} pins {ref} but latest {latest_tag} resolves to {latest_sha}")

  if failures:
    print("Action pin/drift check failed:", file=sys.stderr)
    for failure in failures:
      print(f"- {failure}", file=sys.stderr)
    return 1

  return 0

if __name__ == "__main__":
  raise SystemExit(main())
