# ACC-009 — Multimodel Council Reports

## Goal

Implement the multimodel council report generator that aggregates reviewer model outputs, preserves dissent, computes consensus, and produces canonical approval reports.

This task ensures no single model can become final authority.

## Required reading

- `AGENTS.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SCHEMAS.md`
- `docs/ARCHITECTURE_DECISIONS.md`

## Dependencies

- ACC-007
- ACC-008
- ACC-002

## Scope

Allowed files:

- `src/council/**`
- `src/models/**`
- `src/schemas/**` only if schema alignment is required
- `tests/council/**`

## Non-goals

Do not implement:

- real provider calls;
- claim verification workflow;
- human QA UI;
- workflow state machine;
- publishing.

## Requirements

1. Accept reviewer outputs from multiple model roles.
2. Preserve each model verdict and notes.
3. Preserve dissenting opinions explicitly.
4. Compute consensus:
   - `approved`;
   - `approved_with_suggestions`;
   - `revision_required`;
   - `blocked`.
5. Blocking safety or technical verdict must prevent approval.
6. Include model/provider metadata.
7. Include confidence scores if available.
8. Include final recommendation for human operator.
9. Prevent self-approval where generator and reviewer identity conflict under policy.

## Acceptance criteria

- Unanimous approval produces approved report.
- One non-blocking suggestion produces approved-with-suggestions.
- Technical blocker produces blocked report.
- Safety blocker produces blocked report.
- Dissent is preserved in output.
- Report validates against schema.
- Tests use mock providers only.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- discarding dissent;
- treating confidence as proof;
- letting the generating model self-approve critical artifacts;
- making network calls;
- bypassing human QA.

## PR summary requirements

Include:

- consensus rules implemented;
- dissent handling;
- blocker handling;
- tests run;
- unresolved policy questions.
