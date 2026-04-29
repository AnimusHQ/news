# ACC-026 — Provider Health and Fallback Policy

## Goal

Implement provider health tracking and fallback policy for model providers and other external providers.

The system must degrade safely when a provider is unavailable, degraded, too expensive, privacy-incompatible, or producing invalid outputs.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/OPERATIONS.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-007
- ACC-024 if audit integration is desired

## Scope

Allowed files:

- `src/providers/**`
- `src/models/router/**`
- `src/models/registry/**`
- `tests/providers/**`
- `tests/models/router/**`

## Non-goals

Do not implement real provider API health polling unless explicitly scoped. Use deterministic/mock health state for MVP.

## Requirements

1. Define provider health states:
   - `healthy`
   - `degraded`
   - `disabled`
   - `unknown`
2. Define fallback policy:
   - when fallback is allowed;
   - when fallback is blocked by privacy;
   - when fallback requires human approval;
   - when task should fail closed.
3. Integrate health state into model router selection.
4. Ensure disabled providers/models are never selected.
5. Ensure degraded providers are selected only if policy allows.
6. Record fallback reason.
7. Emit audit event if audit layer exists.
8. Add tests for healthy, degraded, disabled, no fallback, and privacy-blocked fallback.

## Acceptance criteria

- Disabled provider is rejected.
- Degraded provider selection follows policy.
- Privacy-incompatible fallback is blocked.
- No-candidate case fails with explicit error.
- Fallback decision is deterministic and explainable.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- silently falling back to lower-privacy provider;
- choosing disabled provider;
- hiding fallback reason;
- weakening router privacy enforcement.

## PR summary requirements

Include health states, fallback policy, tests run, and limitations.
