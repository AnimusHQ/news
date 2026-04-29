# ACC-029 — End-to-End Dry Run for Pilot Episode

## Goal

Prove that the Animus News MVP pipeline can execute end-to-end on the pilot episode using local fixtures, mock providers, deterministic render/preview output, dry-run publishing, and no real secrets or network-dependent provider calls.

This is the final integration task for the first production-system slice.

## Required reading

- `AGENTS.md`
- `README.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SCHEMAS.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/OPERATIONS.md`
- `docs/CODEX_MASTER_PLAN.md`
- all previous task packs that produced implemented modules

## Dependencies

- ACC-004
- ACC-005
- ACC-006
- ACC-007
- ACC-008
- ACC-009
- ACC-010
- ACC-011
- ACC-012
- ACC-013
- ACC-014
- ACC-015
- ACC-016
- ACC-017
- ACC-018
- ACC-019
- ACC-020
- ACC-021
- ACC-022
- ACC-023
- ACC-024 recommended
- ACC-025 recommended
- ACC-026 recommended
- ACC-027 recommended

## Scope

Allowed files:

- `src/pipeline/**`
- `src/cli/**`
- `tests/e2e/**`
- `episodes/0001-after-git-push/**` only for generated dry-run artifacts or fixture alignment
- `examples/**`
- `docs/OPERATIONS.md` only if documenting dry-run command
- `package.json` only for scripts

## Non-goals

Do not implement:

- real provider calls;
- real YouTube upload;
- public publishing;
- real credentials;
- production UI;
- fully polished visual identity;
- uncontrolled web ingestion.

## Required dry-run flow

Implement a command or documented command sequence that performs:

1. Validate pilot episode artifacts.
2. Validate source registry and research pack.
3. Extract or load claims.
4. Verify claims using mock model council.
5. Generate multimodel approval report.
6. Generate human QA decision packet or consume approved fixture.
7. Enforce workflow transition requirements.
8. Generate storyboard.
9. Generate render/preview output or deterministic placeholder render.
10. Generate render manifest.
11. Run production QA checks.
12. Generate publish pack.
13. Run dry-run publishing adapter.
14. Import fixture analytics.
15. Generate analytics insight report.
16. Generate final dry-run summary.

## Suggested command

```bash
pnpm animus-news dry-run episodes/0001-after-git-push
```

or equivalent.

## Final dry-run summary requirements

The summary must include:

- episode ID;
- artifacts validated;
- workflow states reached;
- model council verdicts;
- human QA status;
- production QA status;
- publish visibility;
- analytics window;
- cost summary if available;
- audit event count if available;
- generated output paths;
- blockers;
- warnings;
- remaining production gaps.

## Acceptance criteria

- Dry run completes without real network/provider credentials.
- All generated artifacts validate.
- Direct public publishing is not possible.
- Mock model dissent is preserved if fixture includes dissent.
- Missing required artifact causes dry run failure.
- Secret fixture causes failure if security scanning is included.
- Final summary is readable and machine-readable if possible.
- Tests cover happy path and at least one blocked path.

## Validation commands

```bash
pnpm test
pnpm typecheck
pnpm animus-news dry-run episodes/0001-after-git-push
```

If the exact command differs, document the command used.

## Mutation policy

Forbidden:

- adding real credentials;
- making real platform uploads;
- bypassing validation failures;
- bypassing human QA fixture/approval requirement;
- treating mock model approval as human approval;
- weakening any previous task's safety gates.

## PR summary requirements

Include:

- dry-run command;
- generated artifacts;
- tests run;
- blockers/warnings observed;
- production gaps remaining;
- follow-up tasks for real provider integration, UI, and deployment.

## Definition of first MVP completion

The first Animus News MVP is considered technically demonstrated when this task passes on the pilot episode and all critical safety, verification, publishing, and artifact gates are exercised at least once.
