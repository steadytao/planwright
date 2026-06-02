# Policy Profile Boundary

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright includes built-in policy profiles for local graph review.

The current profiles are useful because they turn graph facts into review findings for common safety expectations. They are also easy to overstate. A policy profile can look like a compliance engine, an organisational policy platform or an OPA/Rego replacement if the boundary is not explicit.

Policy execution is security-sensitive because future custom policy features could involve user-supplied rules, downloaded bundles or third-party code.

## Decision

Planwright's current policy profiles are built-in static checks over local `planwright.graph.v1` JSON.

The current policy profile boundary is:
- local graph JSON is validated before policy review
- profiles are compiled into Planwright
- profile checks are deterministic and offline
- policy output may be written as Markdown and SARIF
- policy findings are review evidence, not deployment authorisation
- policy findings are not compliance certification

The current policy profile boundary excludes:
- custom policy files
- downloaded policy bundles
- OPA/Rego execution
- Conftest-compatible execution
- organisation-specific policy inheritance
- live cloud or Kubernetes inspection
- automatic deployment approval or rejection

Future custom policy packs, OPA/Rego execution and organisation-specific inheritance require a separate decision before implementation.

## Consequences

This decision means that:
- documentation must describe built-in profiles as static local checks
- policy checks must not read credentials, call cloud APIs or inspect live clusters
- policy output should explain scope and limitations plainly
- compliance language must stay conservative and avoid certification claims
- future policy-pack work needs an explicit trust model for user-supplied rules
- future OPA/Rego work must be opt-in and must not silently run third-party policy code
