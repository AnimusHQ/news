# ACC-000 — Repository Tooling Baseline

## Goal

Create the minimal production-oriented TypeScript project foundation for Animus News without implementing domain business logic.

This task establishes the development substrate Codex and humans will use for all later tasks.

## Required reading

Before implementation, read:

- `AGENTS.md`
- `README.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Scope

Allowed files:

- `package.json`
- `pnpm-lock.yaml`
- `tsconfig.json`
- `vitest.config.*`
- `.gitignore`
- `src/**`
- `tests/**`
- `docs/CODEX_MASTER_PLAN.md` only if updating status notes

## Non-goals

Do not implement:

- model provider integrations;
- rendering;
- publishing;
- database persistence;
- real episode generation;
- source ingestion;
- claim extraction;
- workflow engine.

## Requirements

1. Initialize a TypeScript Node.js project.
2. Use `pnpm` as the package manager unless there is a strong reason to propose an ADR.
3. Add scripts:
   - `test`
   - `typecheck`
   - `build` if applicable
   - `validate` as placeholder or no-op command with clear output
4. Add baseline source structure:
   - `src/index.ts`
   - `src/core/`
   - `src/schemas/`
   - `src/cli/`
   - `src/models/`
   - `src/workflow/`
5. Add first smoke test under `tests/`.
6. Configure TypeScript in strict mode.
7. Do not introduce unnecessary runtime frameworks.

## Suggested implementation

Recommended dependencies:

- `typescript`
- `tsx` or equivalent for local TS execution
- `vitest`
- `@types/node`

Avoid adding heavy frameworks at this stage.

## Acceptance criteria

- `pnpm install` succeeds.
- `pnpm test` passes.
- `pnpm typecheck` passes.
- Project structure exists.
- No unrelated docs are rewritten.
- No real secrets, credentials, or provider keys are added.

## Validation commands

Run if possible:

```bash
pnpm install
pnpm test
pnpm typecheck
```

If a command cannot be run, explain why in the PR summary.

## Mutation policy

Forbidden:

- editing architecture docs except to fix direct references to added tooling;
- changing quality gates;
- changing multimodel strategy;
- adding provider-specific code;
- adding generated binaries.

## PR summary requirements

Include:

- files changed;
- package manager and dependency choices;
- commands run;
- assumptions;
- risks;
- follow-up tasks.
