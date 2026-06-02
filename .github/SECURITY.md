# Security Policy

Please do not report security vulnerabilities in public issues.
- [Private Reporting](#private-reporting)
- [Current Security Boundary](#current-security-boundary)

# Private Reporting

Use GitHub Security Advisories for private vulnerability reports when the repository is hosted on GitHub:
```text
https://github.com/steadytao/planwright/security/advisories/new
```

# Current Security Boundary

Planwright is local-only infrastructure planning software.

It does not:
- read cloud credentials
- contact cloud APIs
- mutate infrastructure
- execute imported infrastructure definitions
- execute Terraform plans
- run a network listener unless `planwright serve` is explicitly started

When `planwright serve` is started, the local web server:
- defaults to `127.0.0.1:5786`
- refuses non-loopback bind addresses
- rejects unexpected Host headers
- sets a restrictive Content Security Policy
- does not set permissive CORS headers
- validates posted plan text in memory
- does not write files from browser actions

Security-sensitive future work includes broader importers, archives, generated scripts, live cloud scans, credential handling and CI output. Those changes require explicit threat-model updates and negative tests.
