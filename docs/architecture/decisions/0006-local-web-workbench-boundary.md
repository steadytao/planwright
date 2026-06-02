# Local Web Workbench Boundary

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright includes `planwright serve` as a local browser workbench over the current typed-plan path.

The eventual product vision includes a richer visual planning interface; the current implementation is a text-and-table workbench. It validates posted plan text in memory and previews graph data, reports and generated artefacts.

Browser surfaces can easily blur local-only review into deployment, credential handling or GUI-only engine behaviour if the boundary is not explicit.

## Decision

The current local web workbench is an interface over the existing Go engine, not a separate product engine.

The workbench:
- runs only when explicitly started with `planwright serve`
- defaults to `127.0.0.1:5786`
- rejects non-loopback bind addresses
- rejects unexpected Host headers
- sets restrictive browser security headers
- validates browser-posted plan text in memory
- previews graph data, diagnostics, reports, Terraform/OpenTofu-oriented files and Mermaid output
- does not write project files from browser actions
- does not load credentials or contact cloud APIs

The workbench is not:
- the future drag-and-drop visual planner
- a canvas editor
- a resource-card graph editor
- a deployment console
- a hosted application

## Consequences

This decision means that:
- local web changes must preserve Host, loopback, request-size and content-security boundaries
- browser state-changing features require explicit CSRF design before implementation
- UI features must call engine packages rather than duplicating validation or generation logic
- docs must describe `planwright serve` as a current text-and-table workbench until the visual planner exists
- hosted demo work requires a separate safety review because it changes distribution and execution context
