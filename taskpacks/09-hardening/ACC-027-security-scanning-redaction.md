# ACC-027 — Security Scanning and Redaction Utilities

## Goal

Implement local security scanning and redaction utilities to prevent secrets, credentials, private data, and sensitive values from entering artifacts, logs, prompts, generated assets, or publish packs.

## Required reading

- `AGENTS.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/QUALITY_GATES.md`
- `docs/OPERATIONS.md`
- `SECURITY.md`

## Dependencies

- ACC-001
- ACC-024 if audit integration is desired

## Scope

Allowed files:

- `src/security/**`
- `scripts/**`
- `tests/security/**`
- `.github/workflows/**` only for CI integration
- `package.json` only for scripts/dependencies

## Non-goals

Do not implement paid external scanning services, real secret manager integration, or production DLP service integration.

## Requirements

1. Implement deterministic secret pattern scanner.
2. Detect common token/key patterns.
3. Support configurable redaction patterns.
4. Scan:
   - artifacts;
   - logs/test fixtures;
   - publish packs;
   - generated descriptions;
   - prompt-like text files if applicable.
5. Provide redaction helper for logs and audit events.
6. Fail release/security checks on high-confidence secret findings.
7. Add fake-token fixtures for tests.
8. Integrate scan command into package scripts and CI if available.

## Acceptance criteria

- Fake token fixture is detected.
- Redaction removes sensitive value while preserving useful context.
- High-confidence finding causes non-zero exit.
- Low-confidence finding can be warning if policy defines it.
- Tests are deterministic.

## Validation commands

```bash
pnpm test
pnpm typecheck
pnpm scan:secrets
```

## Mutation policy

Forbidden:

- committing real secrets;
- logging full detected secrets;
- allowing release gate to pass with high-confidence secret finding;
- sending sensitive findings to external services.

## PR summary requirements

Include scanner patterns, redaction behavior, tests run, CI integration, and limitations.
