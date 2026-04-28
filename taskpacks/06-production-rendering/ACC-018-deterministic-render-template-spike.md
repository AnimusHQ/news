# ACC-018 — Deterministic Render Template Spike

## Goal

Create the first deterministic render or preview template that can transform a storyboard into a local video preview or HTML preview using controlled placeholder assets.

This is a spike, not the final visual system.

## Required reading

- `AGENTS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/SCHEMAS.md`

## Dependencies

- ACC-017

## Scope

Allowed files:

- `src/render/**`
- `render/**`
- `tests/render/**`
- `package.json` only for required render dependencies
- `docs/ARCHITECTURE_DECISIONS.md` only if proposing render stack deviation

## Non-goals

Do not implement:

- final mascot rig;
- real TTS;
- real AI video generation;
- YouTube upload;
- production asset store;
- large binary assets committed to repository.

## Requirements

1. Accept storyboard input.
2. Generate deterministic preview output.
3. Support placeholder scene types:
   - title card;
   - mascot placeholder;
   - terminal/code scene;
   - diagram placeholder;
   - caption text.
4. Generate `render_manifest.json`.
5. Preserve asset provenance for generated placeholders.
6. Fail explicitly if required scene data is missing.
7. Do not commit generated video binaries unless explicitly approved and tiny fixtures are necessary.

## Acceptance criteria

- Sample storyboard renders/previews locally.
- Render manifest validates.
- Missing required scene field fails.
- Render output path is deterministic.
- No real secrets or provider calls.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

If render command exists, run it and report output path.

## Mutation policy

Forbidden:

- adding uncontrolled AI video generation;
- committing large generated binaries;
- making network calls;
- bypassing storyboard validation.

## PR summary requirements

Include render approach, commands run, generated outputs excluded/included, limitations, and follow-up tasks for production renderer.
