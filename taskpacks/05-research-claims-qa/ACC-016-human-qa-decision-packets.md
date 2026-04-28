# ACC-016 — Human QA Decision Packets

## Goal

Generate concise, high-signal decision packets for the human operator before an episode proceeds to storyboarding or later release gates.

The human operator should receive the final candidate, approvals, dissent, blocking issues, unresolved risks, and recommended decision without manually digging through every artifact.

## Required reading

- `AGENTS.md`
- `docs/QUALITY_GATES.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SCHEMAS.md`
- `docs/EDITORIAL_STANDARD.md`

## Dependencies

- ACC-010
- ACC-015
- ACC-009

## Scope

Allowed files:

- `src/qa/**`
- `src/council/**` only for shared report types/helpers
- `src/verification/**` only for shared report summaries
- `tests/qa/**`

## Non-goals

Do not implement:

- UI;
- real provider calls;
- rendering;
- publishing;
- final human decision persistence beyond artifact output unless already supported.

## Inputs

- research pack summary;
- script metadata or script path;
- claims;
- verification report;
- multimodel council report;
- quality gate status;
- optional operator notes.

## Output

- `human_qa_packet` or draft `human_qa_report` artifact.

## Requirements

1. Summarize episode purpose and format.
2. Include claim risk summary.
3. Include unsupported/contradicted/needs-review claims.
4. Include multimodel approvals and dissent.
5. Include safety and policy blockers.
6. Include recommended human decision:
   - `approve`;
   - `approve_with_minor_edits`;
   - `request_revision`;
   - `block`.
7. Preserve unresolved risks.
8. Never hide dissent.
9. Output must be deterministic for same inputs.

## Acceptance criteria

- Packet includes dissenting model notes.
- Packet includes unsupported claims.
- Blocking issues cannot be hidden by positive summary.
- Tests cover approve/revise/block recommendations.
- Output validates or can be converted into valid `human_qa_report`.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- auto-approving human QA;
- dropping dissent;
- suppressing blockers;
- marking operator approval without explicit input.

## PR summary requirements

Include packet fields, recommendation rules, tests run, and known limitations.
