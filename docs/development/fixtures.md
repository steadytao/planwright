# Compatibility Fixtures

Compatibility fixtures are the evidence boundary for Planwright compatibility claims.
- [Purpose](#purpose)
- [Metadata](#metadata)
- [Command Expectations](#command-expectations)
- [Golden Output Updates](#golden-output-updates)
- [Review Rules](#review-rules)

# Purpose

A fixture connects a public compatibility claim to checked input and expected behaviour.

Use fixtures for:
- supported local source formats
- supported command surfaces
- expected compatibility levels
- visible loss categories
- diagnostics and findings
- generated files where the output is part of the claim

Do not raise a compatibility level in [`../compatibility.md`](../compatibility.md) unless the claim is backed by fixture metadata or explicitly marked as documentation-only.

# Metadata

Example metadata:
```yaml
schema: planwright.fixture.v1
id: aws-webapp-basic
name: AWS web application basic proof path
source_format: planwright.yaml
source_kind: file
source_path: planwright.yaml
compatibility_level: 5
expected_loss_categories: []
commands:
  - name: validate
    args: ["validate", "${source}"]
    want_exit: 0
    want_stdout_contains:
      - validation passed
```

Fields:
- `schema` must be `planwright.fixture.v1`
- `id` must be stable and unique in the fixture suite
- `source_format` names the supported source surface being exercised
- `source_kind` may be `file` or `directory`; omitted means `file`
- `source_path` must be a relative slash-separated path inside the fixture directory
- `compatibility_level` must be between `0` and `8`
- `expected_loss_categories` records visible loss evidence for the fixture
- `commands` records command expectations that the test runner executes

Supported loss categories:
- `preserved`
- `normalised`
- `inferred`
- `ambiguous`
- `unsupported`
- `unsafe`
- `manual-review-required`

# Command Expectations

Command arguments may use these placeholders:
- `${source}`, the source file path resolved from `source_path`
- `${fixture}`, the fixture directory
- `${temp}`, a temporary output directory for the current test run

Directory source fixtures may use `source_path: .` when the fixture directory itself is the input bundle. File source fixtures must name a concrete file.

Command expectations may check:
- exit code
- required stdout text
- required stderr text
- generated files under `${temp}`

Fixtures should test public command surfaces rather than internal helper functions wherever practical. Internal unit tests should still cover parser, graph and report edge cases directly.

# Golden Output Updates

Do not update expected output casually.

A golden or fixture expectation update is acceptable when:
- the implementation deliberately changed
- the new output is more correct or more explicit
- the compatibility impact is understood
- the review explains why the old expectation was wrong or incomplete

Future golden file updates should require an explicit update flag or environment variable. Until that workflow exists, review fixture-output changes manually and keep them small.

# Review Rules

When adding or changing a fixture:
- keep the source input minimal
- include unsupported or lossy constructs when the compatibility claim depends on loss reporting
- avoid credentials, tokens and real account identifiers
- avoid live cloud calls
- avoid generated artefacts unless they are deterministic and intentionally checked in
- update [`../compatibility.md`](../compatibility.md) only when the fixture supports the claim
- run `go test ./...`
- run `go run ./cmd/planwright docs check .`
