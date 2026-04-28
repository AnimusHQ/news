# ACC-024 — Audit Logging

## Goal

Implement structured audit logging for critical decisions, approvals, state transitions, model routing decisions, and release gates.

Audit logs are required for replayability, accountability, incident response, and production trust.

## Required reading

- `AGENTS.md`
- `docs/OPERATIONS.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/QUALITY_GATES.md`
- `docs/SYSTEM_BLUEPRINT.md`

## Dependencies

- ACC-011
- ACC-007

## Scope

Allowed files:

- `src/audit/**`
- `src/workflow/**` only for audit integration
- `src/models/**` only for routing audit integration
- `tests/audit/**`

## Non-goals

Do not implement centralized log storage, external SIEM integration, or real cloud logging.

## Requirements

1. Define audit event schema/type.
2. Support event categories:
   - artifact validation;
   - state transition;
   - model routing;
   - council decision;
   - human QA decision;
   - production QA decision;
   - release approval;
   - publishing adapter action;
   - security finding;
   - incident/correction action.
3. Include actor, timestamp, episode ID, artifact ID where applicable, decision, reason, and correlation ID.
4. Redact secrets and sensitive values.
5. Provide in-memory or filesystem audit sink for MVP.
6. Add tests for event creation and redaction.

## Acceptance criteria

- Audit events are structured.
- Secret-like values are redacted.
- Release approval event requires human actor metadata.
- Workflow transition can emit audit event.
- Tests are deterministic.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- logging secrets;
- silently ignoring audit failures for critical release events;
- using audit log as approval substitute.

## PR summary requirements

Include event categories, redaction behavior, tests run, and limitations.
