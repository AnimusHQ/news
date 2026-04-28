# ACC-011 — Episode State Machine

## Goal

Implement the episode lifecycle state machine with explicit allowed transitions.

The state machine prevents silent forward progress and ensures episodes move through required quality gates.

## Required reading

- `AGENTS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/QUALITY_GATES.md`
- `docs/OPERATIONS.md`
- `docs/SCHEMAS.md`

## Dependencies

- ACC-003

## Scope

Allowed files:

- `src/workflow/**`
- `src/artifacts/**` only for shared artifact references
- `tests/workflow/**`

## Non-goals

Do not implement:

- durable workflow engine;
- database persistence;
- model calls;
- rendering;
- publishing;
- artifact dependency enforcement beyond transition metadata. That belongs to ACC-012.

## Required states

Implement states:

- `backlog`
- `candidate`
- `approved_topic`
- `researching`
- `research_ready`
- `drafting`
- `verifying`
- `human_qa`
- `storyboarding`
- `asset_production`
- `rendering`
- `production_qa`
- `scheduled`
- `published`
- `monitored`
- `archived`
- `blocked`

## Requirements

1. Define allowed transitions.
2. Reject invalid transitions.
3. Include transition reason metadata.
4. Include actor metadata:
   - human;
   - system;
   - model;
   - workflow.
5. Include timestamp metadata.
6. Make blocked/unblocked transitions explicit.
7. Add tests for happy path and invalid paths.

## Acceptance criteria

- Full happy path from backlog to archived is testable.
- Invalid transition fails.
- `blocked` cannot silently continue.
- Human-required transitions are labeled.
- Tests cover at least 5 invalid transitions.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- allowing direct transition from draft to published;
- allowing publish states without QA states;
- collapsing human QA and production QA;
- making blocked state bypassable.

## PR summary requirements

Include:

- states implemented;
- transition table;
- tests run;
- known limitations;
- follow-up for artifact dependency enforcement.
