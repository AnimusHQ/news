# ACC-025 — Cost Tracking

## Goal

Implement cost tracking for model usage, rendering, publishing dry runs/adapters, and production tasks per episode and per stage.

Cost tracking prevents multimodel execution from becoming economically uncontrolled.

## Required reading

- `AGENTS.md`
- `docs/OPERATIONS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`

## Dependencies

- ACC-007
- ACC-024 if audit integration is desired

## Scope

Allowed files:

- `src/cost/**`
- `src/models/**` only for cost metadata integration
- `tests/cost/**`

## Non-goals

Do not implement real billing provider integrations or paid API calls.

## Requirements

1. Define cost event type.
2. Support cost dimensions:
   - episode ID;
   - stage;
   - provider;
   - model ID;
   - operation type;
   - estimated input units;
   - estimated output units;
   - estimated cost;
   - currency;
   - timestamp.
3. Aggregate costs by:
   - episode;
   - stage;
   - provider;
   - model;
   - day/week if easy.
4. Support budget policy:
   - warn;
   - require approval;
   - block non-critical automation.
5. Add tests for aggregation and budget exceed behavior.

## Acceptance criteria

- Cost events are validated.
- Aggregation by episode works.
- Budget exceeded can block non-critical task.
- Cost report can be included in analytics or QA packet.
- Tests are deterministic.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- hiding cost overruns;
- making external billing calls;
- treating cost optimization as permission to reduce quality gates.

## PR summary requirements

Include cost model, budget behavior, tests run, and limitations.
