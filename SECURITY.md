# Security policy

> **Licensing note:** Animus News (this repository) is proprietary software; all
> rights are retained by Animus (see [`LICENSE`](LICENSE)). Nothing in this
> security policy grants any license to use, copy, or modify the software. The
> vulnerability-reporting process below still applies.

AnimusHQ projects may include components related to secure access, service exposure, identity, sessions, relays, transport, control planes, and management planes.

Do not treat any AnimusHQ project as production-certified security software unless that repository explicitly documents a supported release, deployment model, threat model, and security review status.

## Reporting a vulnerability

Report security issues privately by email:

```text
rewanderer@proton.me
```

Do not open public GitHub issues for vulnerabilities, exploit paths, authentication bypasses, cryptographic weaknesses, secret exposure, or deployment configurations that could put users at risk.

Include:

- affected repository and commit or release;
- description of the issue;
- reproduction steps or proof of concept when safe to share;
- expected impact;
- affected configuration;
- any logs or traces that do not contain secrets.

## Response expectations

AnimusHQ is currently a small open-source organization. Response time is best-effort, not guaranteed by SLA.

Expected handling process:

1. Acknowledge the report when received.
2. Reproduce or assess the issue.
3. Classify impact and affected scope.
4. Prepare a fix, mitigation, or public advisory when applicable.
5. Credit the reporter if they request credit and the report is valid.

## Supported versions

Unless a repository states otherwise, public repositories are early-stage implementations and do not have a supported-version matrix.

A project becomes supported only when its repository defines:

- release artifacts;
- versioning policy;
- changelog;
- supported deployment model;
- security review status;
- vulnerability handling process.

## Security boundaries

Security-sensitive repositories should document:

- trust assumptions;
- authentication and authorization model;
- identity and session model;
- relay and transport responsibilities;
- secret handling;
- logging and audit behavior;
- non-goals and unsupported configurations.

## Safe handling rules

Do not include secrets, private keys, access tokens, customer data, production logs, or internal infrastructure details in public issues or pull requests.

Use test keys, isolated environments, and minimal reproduction cases whenever possible.
