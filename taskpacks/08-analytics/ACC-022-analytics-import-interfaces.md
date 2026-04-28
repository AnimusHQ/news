# ACC-022 — Analytics Import Interfaces

## Goal

Define provider-agnostic analytics ingestion interfaces for post-publication performance data.

Analytics must improve future content decisions without bypassing editorial or quality gates.

## Required reading

- `AGENTS.md`
- `docs/OPERATIONS.md`
- `docs/ROADMAP.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/QUALITY_GATES.md`

## Dependencies

- ACC-021

## Scope

Allowed files:

- `src/analytics/**`
- `tests/analytics/**`

## Non-goals

Do not implement real YouTube API integration, OAuth, credentials, or model-based optimization.

## Requirements

1. Define analytics provider adapter interface.
2. Support dry-run/fixture adapter.
3. Support metric categories:
   - CTR;
   - impressions;
   - views;
   - average view duration;
   - first 30 seconds retention;
   - completion rate;
   - subscribers gained;
   - comments count;
   - shares/saves if available;
   - community conversion clicks;
   - cost per episode.
4. Normalize provider responses into canonical analytics input.
5. Produce or update `analytics_report` artifact from fixture data.
6. Keep analytics advisory; do not trigger automatic publishing or editorial override.

## Acceptance criteria

- Fixture analytics import works offline.
- Analytics report validates.
- Missing metric fields are handled explicitly.
- Provider-specific data is normalized.
- Tests cover import success and malformed provider data.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- adding real credentials;
- direct YouTube dependency in core analytics;
- allowing analytics to bypass quality gates;
- recommending misleading clickbait as default optimization.

## PR summary requirements

Include adapter interface, normalized metrics, tests run, limitations, and follow-up tasks for real providers.
