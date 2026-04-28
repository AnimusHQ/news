# ACC-020 — Publish Pack Generator

## Goal

Generate a safe publication metadata package for an approved episode without uploading it publicly.

The publish pack turns a rendered episode into a reviewable release bundle: title candidates, description, sources, chapters, pinned comment, community post, and publish manifest draft.

## Required reading

- `AGENTS.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/SCHEMAS.md`

## Dependencies

- ACC-019

## Scope

Allowed files:

- `src/publishing/**`
- `tests/publishing/**`
- `examples/**` only for deterministic fixtures

## Non-goals

Do not implement real platform upload, OAuth, credentials, direct public publishing, or platform analytics.

## Requirements

1. Input approved production QA report, storyboard, render manifest, claims, and source list.
2. Generate:
   - title candidates;
   - description draft;
   - source list section;
   - chapters;
   - pinned comment draft;
   - community post draft;
   - publish manifest draft.
3. `publish_manifest.visibility` must default to `private` or `scheduled`, never `public`.
4. Include CTA aligned with editorial standard.
5. Include disclosure fields where required.
6. Output must be deterministic in tests.

## Acceptance criteria

- Publish manifest validates.
- Public visibility without explicit approval is not generated.
- Sources are included when claims exist.
- Chapters are generated from storyboard timing where available.
- Tests cover safe defaults and invalid public visibility.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- direct upload;
- public-by-default manifest;
- hiding source list;
- bypassing production QA;
- generating misleading titles or clickbait as recommended output.

## PR summary requirements

Include generated metadata fields, default visibility behavior, tests run, and known limitations.
