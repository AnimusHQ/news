# ACC-013 — Source Registry and Trust Ranking

## Goal

Implement source metadata management and trust ranking for research packs and claim verification.

## Required reading

- `AGENTS.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/SCHEMAS.md`

## Dependencies

- ACC-003

## Scope

Allowed files:

- `src/sources/**`
- `tests/sources/**`
- `src/schemas/**` only for source schema alignment

## Non-goals

Do not implement web crawling, arbitrary browsing, real connector access, or model calls.

## Requirements

1. Define source registry records with ID, title, URI, type, trust level, content hash, license notes, retrieved timestamp, and locator support.
2. Support source types: official docs, specification/RFC, source code, release notes, maintainer statement, book, engineering blog, community discussion, comment, unknown.
3. Rank sources by trust level.
4. Prevent community-only sources from satisfying high-risk technical claims.
5. Surface missing license/provenance warnings.
6. Add deterministic tests.

## Acceptance criteria

- Primary sources outrank secondary/community sources.
- Missing source ID fails validation.
- High-risk authority check fails with community-only evidence.
- License/provenance warnings are exposed.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden: arbitrary network ingestion, source trust auto-escalation, or weakening source hierarchy.
