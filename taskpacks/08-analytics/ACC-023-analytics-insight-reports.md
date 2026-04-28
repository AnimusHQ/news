# ACC-023 — Analytics Insight Reports

## Goal

Generate advisory analytics insight reports from imported performance data.

Analytics should improve future topics, pacing, hooks, visual patterns, and distribution strategy without degrading trust or bypassing editorial gates.

## Required reading

- `AGENTS.md`
- `docs/OPERATIONS.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/QUALITY_GATES.md`
- `docs/SYSTEM_BLUEPRINT.md`

## Dependencies

- ACC-022

## Scope

Allowed files:

- `src/analytics/**`
- `tests/analytics/**`
- `examples/**` only for deterministic fixtures

## Non-goals

Do not implement real platform analytics APIs, real model providers, automatic editorial decisions, or automatic publishing decisions.

## Requirements

1. Consume normalized analytics input.
2. Generate canonical `analytics_report` with:
   - metric summary;
   - retention observations;
   - CTR observations;
   - community conversion observations;
   - cost observations;
   - content recommendations;
   - production recommendations;
   - correction triggers if viewer feedback indicates factual issues.
3. Mark recommendations as advisory.
4. Explicitly prohibit misleading clickbait recommendations.
5. Support 24h, 72h, and 7d report windows.
6. Include confidence/quality of data notes.

## Acceptance criteria

- Low retention produces pacing/hook review recommendation.
- Low CTR produces title/thumbnail review recommendation without clickbait.
- Factual correction signal produces review/correction recommendation.
- Analytics report validates.
- Tests are deterministic and offline.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- allowing analytics to override quality gates;
- generating manipulative clickbait as preferred recommendation;
- auto-changing published metadata;
- auto-selecting future topics without human approval.

## PR summary requirements

Include insight rules, advisory boundaries, tests run, and limitations.
