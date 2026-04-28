# ACC-002 — Canonical Artifact Schemas

## Status

Ready after ACC-000.

## Priority

P0.

## Risk

Medium.

## Objective

Implement the canonical artifact schemas that define the Animus News production pipeline. These schemas are the system's contracts. Future workflow, verification, rendering, publishing, analytics, and QA logic must depend on these contracts rather than ad hoc data shapes.

This task must be schema-first. Do not implement business workflows yet.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-000.

## Allowed paths

- `src/schemas/**`
- `schemas/**`
- `examples/artifacts/**`
- `tests/schemas/**`
- `package.json` only if adding schema-related scripts/dependencies
- `docs/SCHEMAS.md` only for alignment notes if necessary
- `docs/codex/tasks/ACC-002-canonical-schemas.md` only for status notes if needed

## Forbidden paths

- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `.github/**`
- `src/models/**`
- `src/workflow/**`
- `src/render/**`
- `src/publishing/**`

## Non-goals

- Do not implement validation CLI. That is ACC-003.
- Do not implement model registry. That is ACC-005.
- Do not implement workflow state machine. That is ACC-011.
- Do not implement rendering.
- Do not implement publishing.
- Do not add provider SDKs.
- Do not add database persistence.

## Functional requirements

Implement schemas and exported types for these canonical artifacts:

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

Each artifact schema must include common metadata:

- `schema_version`
- `episode_id`
- `artifact_id`
- `created_at`
- `created_by`
- `source_artifacts`
- `content_hash` if available or nullable with documented semantics
- `status`

Status enum should include at least:

- `draft`
- `approved`
- `rejected`
- `superseded`

Model-produced artifacts must include model/provider metadata where relevant.

Publishing schemas must default to safe visibility semantics and must not imply direct public publishing.

## Security requirements

- Schemas must not include fields intended for real secrets.
- If tokens or credentials are represented conceptually, they must be references only, never raw values.
- Provider metadata must not require API keys.

## Reliability requirements

- Schemas should be strict enough to catch missing required fields.
- Unknown fields policy must be deliberate and documented in tests or comments.
- Enums must be explicit.

## Acceptance criteria

- All 13 artifact schemas exist.
- TypeScript types are exported.
- Valid examples pass schema parsing.
- Invalid examples fail schema parsing.
- Tests cover required field failures.
- Tests cover enum failures.
- Tests cover publish visibility safety.
- No workflow/business logic is implemented.

## Required tests

Add tests for:

- a valid minimal artifact per schema;
- missing common metadata;
- invalid artifact status;
- invalid claim verification status;
- invalid council verdict;
- invalid publish visibility or public visibility without approval semantics if represented.

## Required validation commands

Codex must run or explain inability to run:

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- changing canonical artifact names from the documentation without explicit explanation;
- adding runtime workflows;
- adding providers;
- adding rendering/publishing logic;
- weakening source/provenance requirements.

Allowed:

- adding schema dependencies such as Zod;
- adding JSON Schema export if straightforward;
- adding examples under `examples/artifacts/**`.

## Codex prompt

```text
You are implementing ACC-002 — Canonical Artifact Schemas for Animus News.

Read first:
- AGENTS.md
- docs/SCHEMAS.md
- docs/SYSTEM_BLUEPRINT.md
- docs/MULTIMODEL_STRATEGY.md
- docs/QUALITY_GATES.md
- docs/SECURITY_AND_SAFETY.md
- docs/ARCHITECTURE_DECISIONS.md

Implement only canonical artifact schemas and schema tests. Do not implement CLI, workflow, model providers, rendering, or publishing logic.

Run:
- pnpm test
- pnpm typecheck

Return changed files, commands run, assumptions, risks, and follow-ups.
```
