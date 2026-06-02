# Testing

Testing is part of Planwright's compatibility and evidence model.
- [Current Checks](#current-checks)
- [Required Coverage](#required-coverage)
- [Fuzzing](#fuzzing)
- [Version Gate Review](#version-gate-review)

# Current Checks

Use:
```bash
go test ./...
go vet ./...
go build ./cmd/planwright
go test -race ./...
go test ./internal/plan -run=^$ -fuzz=FuzzParse -fuzztime=30s
golangci-lint config verify --config .github/config/golangci.yml
golangci-lint fmt --config .github/config/golangci.yml --diff
golangci-lint run --config .github/config/golangci.yml
staticcheck ./...
gosec ./...
govulncheck ./...
go mod verify
go mod tidy -diff
go run ./cmd/planwright docs check .
npx --yes cspell@10.0.1 lint "**/*.md" ".github/**/*.yml" ".github/**/*.yaml" --config .github/config/cspell.json --no-progress --no-summary --no-must-find-files
python3 .github/scripts/ci/check_action_pins.py
python3 .github/scripts/ci/check_file_headers.py
python3 .github/scripts/release/generate_contributors.py --check
actionlint
pre-commit run --all-files
```

# Required Coverage

Behaviour changes should include tests for:
- valid input
- invalid input
- boundary values
- unsupported constructs
- loss reports for unsupported or ambiguous imported constructs
- Kubernetes rendered-manifest import including Secret value redaction and unresolved relationship reporting
- AWS scan bundle import including local-only file handling, identity redaction and public security-group ingress inference
- graph diff review including deterministic added/removed/changed reporting and invalid graph refusal
- graph JSON Schema validation including schema export, structural failures and semantic validation after schema validation
- fuzz targets for untrusted parsers and reviewers including plan YAML, graph JSON, CloudFormation/SAM, Kubernetes manifests, AWS scan JSON and Terraform plan JSON
- policy profile review including built-in profile listing, blocking policy exit codes, Markdown output and SARIF output
- Terraform plan review findings
- SARIF output structure where relevant
- diagnostics
- CLI exit status where relevant
- local web server security headers, Host handling, body limits and malformed browser input where relevant
- documentation style, British-English spelling and spellchecker configuration
- GitHub Actions pinning, script quality, file headers and contributor metadata

Security-sensitive changes should include negative tests. Examples include path traversal, archive safety, credential handling, generated scripts, local HTTP boundaries and live scan controls.

# Fuzzing

Planwright uses native Go fuzzing for parser and importer boundaries. Short bounded fuzzing runs in CI; longer local runs are useful before changing YAML, JSON or relationship-inference code. Use:
```bash
go test ./internal/plan -run=^$ -fuzz=FuzzParse -fuzztime=2m
go test ./internal/graph -run=^$ -fuzz=FuzzValidateJSON -fuzztime=2m
go test ./internal/importers/cloudformation -run=^$ -fuzz=FuzzImportCloudFormation -fuzztime=2m
go test ./internal/importers/cloudformation -run=^$ -fuzz=FuzzImportSAM -fuzztime=2m
go test ./internal/importers/kubernetes -run=^$ -fuzz=FuzzParseSource -fuzztime=2m
go test ./internal/importers/awsscan -run=^$ -fuzz=FuzzRejectDuplicateJSONKeys -fuzztime=2m
go test ./internal/review/terraformplan -run=^$ -fuzz=FuzzReviewBytes -fuzztime=2m
```

ClusterFuzzLite and OSS-Fuzz are intentionally deferred until Planwright has a public repository, a clearer fuzz corpus policy and stable long-running resource limits.

# Version Gate Review

Before starting the next roadmap version, perform a full pass over:
- source code
- tests
- examples
- documentation
- public compatibility claims
- dependencies
- generated or accidental artefacts

Record the result in the final change summary for that version.
