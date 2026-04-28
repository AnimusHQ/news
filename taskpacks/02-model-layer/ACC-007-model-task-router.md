# ACC-007 — Model Task Router

## Goal

Implement the model task router that selects the best model or model panel for a task based on task category, risk, modality, privacy tier, benchmark scores, cost, latency, and provider health.

This task makes multimodel support operational.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-005
- ACC-006

## Scope

Allowed files:

- `src/models/router/**`
- `src/models/registry/**`
- `src/models/types.ts`
- `tests/models/router/**`

## Non-goals

Do not implement:

- real provider API calls;
- multimodel council aggregation;
- human QA;
- cost billing integration;
- provider health polling beyond static registry status.

## Requirements

1. Accept a `ModelTaskRequest`.
2. Determine required capabilities.
3. Filter candidate models by:
   - status;
   - modality;
   - task capability;
   - privacy tier;
   - context requirements if represented;
   - cost/latency constraints if represented.
4. Select:
   - single model for low-risk tasks;
   - primary + reviewer for medium-risk tasks;
   - model panel for high-risk tasks.
5. Return a routing decision object that includes:
   - selected models;
   - rejected models and reasons;
   - risk policy applied;
   - privacy policy applied;
   - cost/latency notes;
   - deterministic explanation.
6. Ensure disabled models are never selected.
7. Ensure privacy-incompatible models are blocked.

## Acceptance criteria

- Low-risk task selects one model.
- Medium-risk task selects at least two distinct roles where possible.
- High-risk task selects a panel where possible.
- Disabled model is rejected.
- Privacy mismatch is rejected.
- Routing decision is deterministic for same input.
- Tests cover no-candidate failure.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- selecting one hard-coded global default model;
- ignoring privacy tier;
- silently falling back to lower-trust model for restricted data;
- making network calls.

## PR summary requirements

Include:

- routing policy implemented;
- test scenarios;
- assumptions about risk levels;
- limitations and follow-ups.
