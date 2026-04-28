# ACC-014 — Research Pack Builder MVP

## Goal

Build structured research packs from explicitly supplied source records and extracted source text.

This is the controlled first step toward source-grounded episode generation.

## Required reading

- `AGENTS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/SCHEMAS.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/SECURITY_AND_SAFETY.md`

## Dependencies

- ACC-013
- ACC-007
- ACC-002

## Scope

Allowed files:

- `src/research/**`
- `src/sources/**`
- `tests/research/**`
- `examples/**` only for deterministic fixtures

## Non-goals

Do not implement:

- uncontrolled web crawling;
- arbitrary internet browsing;
- real model provider calls unless mock adapters are used;
- script generation;
- publishing.

## Inputs

- topic artifact;
- source registry entries;
- extracted source snippets;
- optional manual editorial constraints.

## Output

- canonical `research_pack.json` artifact.

## Requirements

1. Build research pack with:
   - core question;
   - audience;
   - learning objectives;
   - sources;
   - required terms;
   - claim candidates;
   - unresolved questions;
   - known controversies;
   - forbidden simplifications;
   - visual opportunities;
   - recommended CTA.
2. Use source trust ranking.
3. Flag insufficient primary source coverage.
4. Preserve source IDs and locators.
5. Produce deterministic output in tests.

## Acceptance criteria

- Research pack validates against schema.
- Missing primary source for high-risk topic is flagged.
- Source IDs are preserved.
- Forbidden simplifications are included.
- Tests use deterministic fixtures and no network.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- fabricating sources;
- promoting model-generated statements to source evidence;
- performing uncontrolled network access;
- marking insufficient research as approved.

## PR summary requirements

Include:

- builder inputs/outputs;
- source coverage behavior;
- tests run;
- known limitations;
- follow-up tasks.
