# ACC-021 — Safe Publishing Adapter Interface

## Goal

Define safe provider-agnostic publishing adapter interfaces for private upload, scheduled release, and dry-run publication workflows.

This task must not add real credentials or direct public publishing.

## Required reading

- `AGENTS.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/OPERATIONS.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-020

## Scope

Allowed files:

- `src/publishing/adapters/**`
- `src/publishing/**`
- `tests/publishing/**`

## Non-goals

Do not implement:

- real YouTube OAuth;
- real uploads;
- real platform credentials;
- public upload path;
- analytics import.

## Requirements

1. Define publishing adapter interface.
2. Support dry-run adapter.
3. Support methods conceptually:
   - `uploadPrivateDraft`;
   - `scheduleDraft`;
   - `validateMetadata`;
   - `getDraftStatus`.
4. Public visibility must require explicit human release approval and should still be blocked in dry-run MVP unless explicitly scoped.
5. Normalize adapter errors:
   - auth missing;
   - upload failed;
   - processing failed;
   - policy blocked;
   - visibility not allowed;
   - metadata invalid.
6. Persist adapter response into publish manifest or publish result artifact.

## Acceptance criteria

- Dry-run adapter works offline.
- Direct public visibility fails.
- Missing human approval fails schedule/public transition.
- Adapter tests are deterministic.
- No real credential code is added.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- real credentials;
- direct public publishing;
- platform-specific logic in core publishing abstractions;
- bypassing release approval.

## PR summary requirements

Include adapter methods, dry-run behavior, safety blockers, tests run, and follow-up tasks for real platform adapters.
