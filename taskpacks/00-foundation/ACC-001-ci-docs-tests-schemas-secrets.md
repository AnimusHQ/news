# ACC-001 — CI for Docs, Tests, Schemas, and Secrets

## Goal

Add a GitHub Actions CI baseline that verifies the repository can be safely developed and reviewed.

This task should create the first automated guardrail layer for future Codex changes.

## Required reading

Before implementation, read:

- `AGENTS.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/QUALITY_GATES.md`

## Dependencies

- ACC-000 should be complete or implemented in the same initial setup branch only if explicitly approved.

## Scope

Allowed files:

- `.github/workflows/**`
- `package.json`
- `scripts/**`
- `docs/OPERATIONS.md` only if documenting CI commands

## Non-goals

Do not implement:

- production deployment;
- publishing automation;
- provider integrations;
- real secret scanning SaaS integration requiring paid services;
- rendering checks unless tooling already exists.

## Requirements

1. Add GitHub Actions workflow for pull requests and pushes to `main`.
2. CI should run:
   - dependency install;
   - typecheck;
   - tests;
   - validation command if present;
   - a basic secret scan script if feasible.
3. CI should be deterministic and not require real provider credentials.
4. Add a simple local script for secret scanning if no dependency is chosen.
5. Make CI fail on obvious fake secret patterns only if the scanner is deterministic and tested.

## Suggested checks

```bash
pnpm install --frozen-lockfile
pnpm typecheck
pnpm test
pnpm validate
pnpm scan:secrets
```

If some commands are not yet available, either add minimal placeholders or document why they are omitted.

## Acceptance criteria

- Workflow YAML is valid.
- CI uses Node.js LTS.
- CI uses pnpm cache if practical.
- Local commands match CI commands.
- No secrets are committed.
- CI does not require network access except package installation.

## Mutation policy

Forbidden:

- adding real deployment credentials;
- adding direct publishing jobs;
- adding provider API keys;
- weakening existing project invariants;
- rewriting unrelated documentation.

## PR summary requirements

Include:

- workflow files added;
- commands run locally;
- limitations;
- any CI checks intentionally deferred.
