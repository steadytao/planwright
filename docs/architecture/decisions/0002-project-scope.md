# Project Scope

<!-- ![Proposed](https://img.shields.io/badge/status-proposed-495977.svg?style=for-the-badge&labelColor=031835) -->
![Accepted](https://img.shields.io/badge/status-accepted-2064FC.svg?style=for-the-badge&labelColor=031835)
<!-- ![Superseded](https://img.shields.io/badge/status-superseded-364457.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Deprecated](https://img.shields.io/badge/status-deprecated-495977.svg?style=for-the-badge&labelColor=031835) -->
<!-- ![Denied](https://img.shields.io/badge/status-denied-031835.svg?style=for-the-badge&labelColor=020F22) -->

## Context

Planwright is presented as a local-first infrastructure planning, compatibility and evidence engine.

Without a clear scope decision, the project could drift into adjacent categories such as a cloud console, visual diagrammer, generic IaC converter, compliance scanner, hosted SaaS product or one-click deployment tool.

That would weaken the product's clarity, increase the verification burden and make its safety claims less defensible.

## Decision

Planwright is defined as a local-first infrastructure planning, compatibility and evidence engine.

Its intended scope is:
- typed infrastructure plan parsing
- architecture graph modelling
- loss-aware import from selected local artefacts
- reviewable report generation
- Terraform/OpenTofu-oriented and other generator output where explicitly supported
- local static analysis for security, cost, deployability, policy and compatibility evidence
- local browser, CLI, future TUI and CI interfaces over the same engine

Planwright is not defined as:
- a cloud console
- a hosted SaaS product
- a one-click deployment tool
- a generic diagrammer
- a vulnerability scanner
- a compliance certification tool
- a lossless universal IaC converter
- a replacement for Terraform, OpenTofu, CloudFormation, SAM, CDK, Pulumi or Kubernetes

The canonical human-facing scope documents are [`README.md`](../../../README.md), [`../design.md`](../design.md), [`../threat-model.md`](../threat-model.md) and [`../../compatibility.md`](../../compatibility.md).

## Consequences

This decision means that:
- feature proposals that widen Planwright into unrelated product categories may be declined
- compatibility work should prioritise explicit support levels, provenance and loss reports over broad conversion claims
- generated output remains review evidence unless deployability is separately tested and documented
- cloud credentials, live scans and mutation require explicit future decisions and threat-model updates
- hosted and browser features must not become a way around the local-first boundary
