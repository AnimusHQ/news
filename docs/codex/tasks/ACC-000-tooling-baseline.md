# ACC-000 — Repository Tooling Baseline

## Status

Ready.

## Priority

P0.

## Risk

Low.

## Objective

Create the minimal TypeScript engineering foundation for Animus News without implementing business logic. This task exists so every future Codex task has a consistent runtime, test runner, type checker, and project layout.

The result should be intentionally small and boring: a clean repository skeleton that future tasks can build on safely.

## Required reading

- `AGENTS.md`
- `README.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

None.

## Allowed paths

- `package.json`
- `pnpm-lock.yaml`
- `tsconfig.json`
- `vitest.config.ts`
- `.gitignore`
- `src/**`
- `tests/**`
- `docs/codex/tasks/ACC-000-tooling-baseline.md` only for status notes if needed

## Forbidden paths

- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `episodes/**`
- `schemas/**`
- `.github/**`

## Non-goals

- Do not implement artifact schemas.
- Do not implement CLI validation.
- Do not implement model registry.
- Do not implement providers.
- Do not implement workflow engine.
- Do not implement rendering.
- Do not implement publishing.
- Do not add database dependencies.
- Do not add real API keys or credentials.

## Functional requirements

1. Initialize a TypeScript project using Node.js LTS assumptions.
2. Use `pnpm` as the package manager.
3. Add scripts:
   - `test`
   - `typecheck`
   - `build` or `check` if useful
4. Add `src/index.ts` exporting a small project identity constant or function.
5. Add initial directories:
   - `src/core/`
   - `src/schemas/`
   - `src/cli/`
   - `src/models/`
   - `src/workflow/`
6. Add a smoke test proving the test runner works.
7. Ensure strict TypeScript settings.

## Security requirements

- Do not add secrets.
- Do not add network calls.
- Do not add provider SDKs.
- Do not add postinstall scripts that fetch remote binaries.

## Reliability requirements

- Tooling must work locally without external services.
- Tests must be deterministic.

## Acceptance criteria

- `pnpm install` succeeds.
- `pnpm test` succeeds.
- `pnpm typecheck` succeeds.
- The repository has a minimal, typed source layout.
- No unrelated documentation is rewritten.

## Required tests

- Smoke test for exported project identity/version.

## Required validation commands

Codex must run or explain inability to run:

```bash
pnpm install
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- changing architecture docs;
- adding business logic;
- adding providers;
- adding rendering/publishing code;
- adding external services.

Allowed:

- minimal package and TypeScript configuration;
- minimal test setup;
- minimal source directories.

## Codex prompt

```text
You are implementing ACC-000 — Repository Tooling Baseline for Animus News.

Read first:
- AGENTS.md
- README.md
- docs/CODEX_USAGE.md
- docs/CODEX_MASTER_PLAN.md
- docs/ARCHITECTURE_DECISIONS.md

Implement only this task. Stay within allowed paths. Do not implement business logic, model providers, schemas, rendering, publishing, or workflow logic.

Run:
- pnpm install
- pnpm test
- pnpm typecheck

Return a PR summary with changed files, commands run, assumptions, risks, and follow-ups.
```
