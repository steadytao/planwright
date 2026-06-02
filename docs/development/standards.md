# Standards

Planwright should accumulate clarity before breadth.
- [General Expectations](#general-expectations)
- [Documentation Style](#documentation-style)
- [Source Headers](#source-headers)
- [Public Contracts](#public-contracts)

# General Expectations

Contributions should:
- keep changes narrow
- prefer explicit behaviour over cleverness
- preserve local-first defaults
- surface trust and compatibility implications
- use precise errors and diagnostics
- avoid hidden network calls
- avoid generated artefact churn unless the generated output is part of the contract

# Documentation Style

Public documentation uses British English.

House style:
- bullet lists must hug the text above them
- write `Text:` followed immediately by `- item`
- fenced code blocks must hug the text above them
- write `Text:` followed immediately by the opening fence when showing commands or examples
- avoid commas before `but`
- avoid commas before `including`, `however` and `or`
- avoid simple comma-before-`and` phrasing unless the comma separates larger topics
- prefer semicolons or commas where they improve readability, without making sentences grammatically weaker

Check documentation with:
```bash
go run ./cmd/planwright docs check .
npx --yes cspell@10.0.1 lint "**/*.md" ".github/**/*.yml" ".github/**/*.yaml" --config .github/config/cspell.json --no-progress --no-summary --no-must-find-files
```

Check Go formatting and linting with:
```bash
golangci-lint config verify --config .github/config/golangci.yml
golangci-lint fmt --config .github/config/golangci.yml --diff
golangci-lint run --config .github/config/golangci.yml
```

Check workflow metadata with:
```bash
python3 .github/scripts/ci/check_action_pins.py
actionlint
```

# Source Headers

New Planwright-owned source files should include:
```go
// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0
```

The CI file-header check covers Go, Python and shell source files.

Use the comment syntax appropriate for the file type.

# Public Contracts

The following are public contracts once released:
- command names
- flag names
- exit codes
- diagnostic codes
- `planwright.v1`
- `planwright.graph.v1`
- compatibility level meanings

Changes to these contracts need tests and documentation updates.
