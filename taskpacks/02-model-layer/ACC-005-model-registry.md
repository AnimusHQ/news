# ACC-005 — Model Registry Schema and Configuration

## Goal

Implement a provider-agnostic model registry that records available models, capabilities, privacy tiers, cost/latency profiles, benchmark scores, health status, and known failure modes.

The registry is the foundation of Animus News multimodel architecture.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `docs/CODEX_MASTER_PLAN.md`

## Dependencies

- ACC-002

## Scope

Allowed files:

- `src/models/registry/**`
- `src/models/types.ts`
- `config/model-registry.example.yaml`
- `tests/models/registry/**`
- `docs/MULTIMODEL_STRATEGY.md` only if implementation clarifies documented fields

## Non-goals

Do not implement:

- real provider API calls;
- task routing;
- model council;
- provider health polling;
- secrets management;
- real production model credentials.

## Requirements

1. Define a model registry schema.
2. Support model status:
   - `active`
   - `degraded`
   - `disabled`
3. Support modalities:
   - `text`
   - `vision`
   - `audio`
   - `video`
   - `code`
4. Support task capabilities:
   - research synthesis;
   - technical verification;
   - script writing;
   - editorial review;
   - storyboard planning;
   - visual reasoning;
   - safety review;
   - analytics interpretation;
   - TTS/voice if represented.
5. Support privacy tiers:
   - `public`
   - `internal_approved`
   - `restricted`
   - `local_only`
6. Support cost profile.
7. Support latency profile.
8. Support benchmark history.
9. Support known failure modes.
10. Add example config with mock model entries only.

## Acceptance criteria

- Example model registry validates.
- Invalid model records fail validation.
- Disabled models are represented clearly.
- Privacy tier values are validated.
- Tests cover missing required fields, invalid capability, invalid status, invalid privacy tier.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- adding real provider keys;
- adding real provider calls;
- designating any model as universal authority;
- removing provider independence requirements.

## PR summary requirements

Include:

- registry fields implemented;
- example models added;
- validation behavior;
- tests run;
- limitations and follow-up tasks.
