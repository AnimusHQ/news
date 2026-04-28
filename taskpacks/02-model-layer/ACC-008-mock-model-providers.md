# ACC-008 — Mock Model Providers

## Goal

Implement deterministic mock model providers for local development, CI, router tests, council tests, and end-to-end dry runs.

Mock providers allow the system to develop without real model credentials or network calls.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/SECURITY_AND_SAFETY.md`

## Dependencies

- ACC-006
- ACC-007 may be implemented before or after this task, but router integration tests depend on both.

## Scope

Allowed files:

- `src/models/mock/**`
- `src/models/adapters/**`
- `tests/models/mock/**`
- `tests/models/router/**` only if adding integration fixtures

## Non-goals

Do not implement:

- real provider SDKs;
- real API calls;
- real credentials;
- real model evaluation;
- production provider health checks.

## Requirements

1. Implement at least three mock providers:
   - technical reviewer;
   - editorial reviewer;
   - safety reviewer.
2. Mock outputs must be deterministic.
3. Support configurable verdicts:
   - `approve`;
   - `approve_with_suggestions`;
   - `request_revision`;
   - `block`.
4. Support configurable failure modes:
   - timeout;
   - invalid output;
   - provider unavailable;
   - policy blocked.
5. Include normalized metadata:
   - model ID;
   - provider;
   - latency;
   - cost estimate;
   - confidence;
   - notes.
6. No network calls.

## Acceptance criteria

- Mock providers implement the adapter interface.
- Tests can simulate approval, dissent, blocker, and provider failure.
- Mock outputs are stable across test runs.
- No real API key handling exists.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- adding real provider dependencies;
- making network calls;
- hard-coding mock approval as universal success;
- bypassing router/council abstractions.

## PR summary requirements

Include:

- mock providers implemented;
- failure modes supported;
- tests run;
- limitations.
