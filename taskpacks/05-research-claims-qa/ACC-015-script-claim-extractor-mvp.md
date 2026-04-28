# ACC-015 — Script Claim Extractor MVP

## Goal

Extract factual claims from script drafts and produce a canonical claim set for verification.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/QUALITY_GATES.md`
- `docs/EDITORIAL_STANDARD.md`

## Dependencies

- ACC-014

## Scope

Allowed files:

- `src/claims/**`
- `tests/claims/**`
- `examples/**` only for deterministic script fixtures

## Non-goals

Do not implement full script writing, real model providers, source ingestion, or final verification.

## Requirements

1. Parse script markdown.
2. Extract candidate factual claims.
3. Classify claim type:
   - technical;
   - historical;
   - product;
   - safety;
   - editorial/opinion;
   - CTA/community.
4. Classify risk level:
   - low;
   - medium;
   - high;
   - critical.
5. Link claims to research pack source IDs when explicitly referenced or inferable from structured metadata.
6. Flag unlinked technical/high-risk claims.
7. Output canonical `claims.json`.

## Acceptance criteria

- Test script produces expected claims.
- Unlinked high-risk technical claim is flagged.
- Opinion statements are not misclassified as hard facts unless phrased factually.
- Output validates against schema.
- Tests are deterministic and offline.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- fabricating evidence locators;
- marking unlinked claims as supported;
- relying on model memory as source;
- weakening claim schema.

## PR summary requirements

Include claim extraction method, limitations, tests run, and follow-up tasks for model-assisted extraction.
