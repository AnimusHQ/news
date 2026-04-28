# ACC-006 — Model Provider Adapter Interface

## Goal

Define a stable provider-agnostic interface for model providers.

Provider-specific logic must live behind adapters. Core pipeline logic must never depend directly on a single model provider.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-005

## Scope

Allowed files:

- `src/models/adapters/**`
- `src/models/types.ts`
- `tests/models/adapters/**`

## Non-goals

Do not implement:

- real provider SDKs;
- real API calls;
- real credentials;
- task routing;
- multimodel council;
- model benchmarking.

## Requirements

1. Define provider-agnostic request types.
2. Define provider-agnostic response types.
3. Include task metadata:
   - task category;
   - risk level;
   - required output schema if any;
   - data classification;
   - episode/artifact IDs where applicable.
4. Include normalized response metadata:
   - model ID;
   - provider;
   - latency;
   - estimated cost;
   - token or unit usage if applicable;
   - structured output validation status;
   - safety metadata if available.
5. Define normalized error classes:
   - provider unavailable;
   - timeout;
   - rate limited;
   - invalid output;
   - policy blocked;
   - privacy blocked;
   - unknown provider error.
6. Add mock adapter contract tests.

## Acceptance criteria

- Adapter interface compiles.
- Mock adapter implements the interface.
- Normalized errors are testable.
- No provider-specific dependency is required.
- No API key or credential handling is added.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- hard-coding a default authority provider;
- importing real provider SDKs;
- adding secret loading;
- bypassing model registry.

## PR summary requirements

Include:

- interfaces added;
- normalized errors added;
- tests run;
- known limitations;
- follow-up tasks for real adapters.
