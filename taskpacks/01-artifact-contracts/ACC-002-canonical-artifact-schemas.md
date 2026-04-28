# ACC-002 — Canonical Artifact Schemas

## Goal

Implement runtime schemas and TypeScript types for all canonical Animus News pipeline artifacts.

The schemas are the contract that prevents uncontrolled mutation across the production system.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/QUALITY_GATES.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-000

## Scope

Allowed files:

- `src/schemas/**`
- `schemas/**`
- `tests/schemas/**`
- `examples/valid/**`
- `examples/invalid/**`
- `package.json` only for schema dependencies/scripts
- `docs/SCHEMAS.md` only if clarifying schema implementation details

## Non-goals

Do not implement:

- model routing;
- provider adapters;
- workflow state machine;
- rendering;
- publishing;
- database persistence.

## Required artifacts

Implement schemas for:

1. `topic`
2. `source`
3. `research_pack`
4. `claims`
5. `verification_report`
6. `multimodel_approval_report`
7. `human_qa_report`
8. `storyboard`
9. `asset_manifest`
10. `render_manifest`
11. `production_qa_report`
12. `publish_manifest`
13. `analytics_report`

## Common metadata requirements

Every artifact schema must include:

- `schema_version`;
- `episode_id` where applicable;
- `artifact_id` where applicable;
- `created_at` where applicable;
- `created_by` where applicable;
- `status` where applicable;
- source/dependency references where applicable.

## Implementation requirements

1. Use Zod or equivalent runtime validation.
2. Export TypeScript types from schemas.
3. Keep schemas strict enough to reject malformed artifacts.
4. Define enums centrally where helpful:
   - artifact status;
   - claim status;
   - claim risk;
   - model verdict;
   - QA decision;
   - publish visibility.
5. Provide valid and invalid examples for every major artifact.
6. Add tests that prove examples validate or fail correctly.

## Critical schema constraints

- `publish_manifest.visibility` must not default to `public`.
- `multimodel_approval_report` must preserve dissent.
- `claims` must support evidence locators.
- `asset_manifest` must support provenance and license fields.
- `human_qa_report` must include explicit decision.
- `production_qa_report` must represent blocking issues.

## Acceptance criteria

- All schemas compile.
- All TypeScript types export successfully.
- Valid examples pass.
- Invalid examples fail.
- Tests cover every schema.
- Unknown critical fields are not silently accepted unless explicitly documented.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- renaming canonical artifacts from documentation without ADR;
- weakening publication safety fields;
- allowing unsupported high-risk claims to be represented as approved;
- introducing model/provider-specific schemas in core artifact contracts.

## PR summary requirements

Include:

- schema list implemented;
- examples added;
- tests run;
- unresolved schema design questions;
- any deviations from `docs/SCHEMAS.md`.
