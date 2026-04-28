# ACC-017 — Storyboard Generator MVP

## Goal

Convert an approved script into a structured storyboard artifact with scenes, narration, mascot direction, visual plan, captions, and timing targets.

## Required reading

- `AGENTS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/SCHEMAS.md`
- `docs/QUALITY_GATES.md`

## Dependencies

- ACC-016

## Scope

Allowed files:

- `src/storyboard/**`
- `tests/storyboard/**`
- `examples/**` only for deterministic fixtures

## Non-goals

Do not implement rendering, real model calls, TTS, generated video, or publishing.

## Requirements

1. Input approved script and optional human QA packet.
2. Segment script into scenes.
3. Assign each scene:
   - scene ID;
   - narration;
   - timing target;
   - mascot mode;
   - mascot emotion/action;
   - visual type;
   - on-screen text;
   - caption plan;
   - source/claim references where relevant.
4. Support deterministic MVP behavior without real model calls.
5. Output canonical `storyboard.yaml`.

## Acceptance criteria

- Storyboard validates against schema.
- Every scene has narration and visual plan.
- Every scene has stable ID.
- No scene silently drops technical claims.
- Tests use deterministic script fixture.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden: bypassing human QA, generating final video, making real provider calls, or hiding missing visual requirements.
