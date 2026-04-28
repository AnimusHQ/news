# ACC-001 — Continuous Integration Baseline

## Status

Ready after ACC-000.

## Priority

P0.

## Risk

Low.

## Objective

Add a GitHub Actions CI baseline that runs the same checks developers run locally. CI must verify the repository foundation without requiring external providers, secrets, paid services, model APIs, rendering infrastructure, or publishing credentials.

## Required reading

- `AGENTS.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/OPERATIONS.md`
- `docs/SECURITY_AND_SAFETY.md`

## Dependencies

- ACC-000.

## Allowed paths

- `.github/workflows/**`
- `package.json`
- `pnpm-lock.yaml`
- `docs/codex/tasks/ACC-001-ci.md` only for status notes if needed

## Forbidden paths

- `docs/SYSTEM_BLUEPRINT.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `src/**` unless a script hook is strictly necessary and explained
- `episodes/**`
- `schemas/**`

## Non-goals

- Do not add deployment.
- Do not add publishing.
- Do not add provider credentials.
- Do not add real secret scanning services that require account setup.
- Do not add rendering workers.
- Do not add database services.

## Functional requirements

1. Add a CI workflow for pull requests and pushes to `main`.
2. Use Node.js LTS.
3. Enable pnpm via Corepack or official pnpm setup.
4. Install dependencies with a frozen lockfile where feasible.
5. Run:
   - `pnpm typecheck`
   - `pnpm test`
6. Include a simple repository hygiene/security step, such as:
   - checking for obvious placeholder secrets with a local script if available;
   - or adding a clearly marked placeholder job that will be expanded in ACC-027.
7. Keep the workflow minimal and deterministic.

## Security requirements

- Do not require secrets.
- Do not upload artifacts containing source, logs, or private data unless explicitly needed.
- Do not add deployment permissions.
- Set minimal GitHub token permissions where possible.

## Acceptance criteria

- CI YAML is syntactically valid.
- CI commands match local scripts.
- CI has no external secret dependency.
- CI does not deploy or publish.
- CI does not weaken repository security.

## Required validation commands

Codex must run or explain inability to run:

```bash
pnpm test
pnpm typecheck
```

If a workflow syntax validation tool is available, run it. Otherwise state that YAML was reviewed structurally.

## Mutation policy

Forbidden:

- deployment;
- publishing;
- real credentials;
- weakening permissions;
- broad unrelated edits.

Allowed:

- minimal workflow files;
- package script alignment if needed.

## Codex prompt

```text
You are implementing ACC-001 — Continuous Integration Baseline for Animus News.

Read first:
- AGENTS.md
- docs/CODEX_USAGE.md
- docs/CODEX_MASTER_PLAN.md
- docs/OPERATIONS.md
- docs/SECURITY_AND_SAFETY.md

Implement only this task. Add minimal GitHub Actions CI for typecheck and tests. Do not add deployment, publishing, secrets, provider integrations, or rendering.

Run relevant checks and return changed files, commands run, assumptions, risks, and follow-ups.
```
