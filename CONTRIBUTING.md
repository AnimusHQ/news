# Contributing to AnimusHQ

AnimusHQ accepts contributions that improve correctness, security posture, operability, documentation quality, test coverage, or release safety.

Repositories may define project-specific contribution rules. If a repository has its own `CONTRIBUTING.md`, follow that file first.

## Before contributing

Before opening a pull request:

1. Read the repository README.
2. Check project status and non-goals.
3. Check open issues and existing pull requests.
4. Avoid changing security-sensitive behavior without a design note or issue discussion.
5. Do not include secrets, private keys, tokens, customer data, internal logs, or confidential material.

## Accepted contribution areas

Useful contributions include:

- tests and reproducible failure cases;
- documentation corrections;
- CI, linting, formatting, and release-safety improvements;
- architecture notes and threat-model clarifications;
- observability and diagnostics improvements;
- bug fixes with clear reproduction steps;
- examples that use safe test data and local-only configuration.

## Security-sensitive changes

Changes involving identity, sessions, authorization, relay behavior, cryptography, transport security, secret handling, audit logs, or service exposure require extra care.

For these changes, include:

- what behavior changes;
- what threat or failure mode is affected;
- how the change is tested;
- what remains unsupported;
- any migration or compatibility impact.

Do not submit public proof-of-concept exploits for unresolved vulnerabilities. Report vulnerabilities through `SECURITY.md`.

## Pull request checklist

Before submitting:

- [ ] The change has a clear scope.
- [ ] Tests or validation steps are included.
- [ ] Documentation is updated when behavior changes.
- [ ] Security implications are stated when relevant.
- [ ] No secrets, tokens, private data, or internal logs are included.
- [ ] The change does not expand project claims beyond documented status.
- [ ] New runtime behavior is observable or diagnosable where practical.

## Review expectations

Maintainers review for:

- correctness;
- maintainability;
- security impact;
- compatibility with repository scope;
- operational clarity;
- testability;
- documentation quality.

A pull request may be declined if it expands scope, weakens security boundaries, introduces unclear behavior, removes validation, or creates unsupported production claims.
