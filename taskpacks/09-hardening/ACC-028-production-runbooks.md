# ACC-028 — Production Runbooks

## Goal

Add operational runbooks for releases, incidents, failures, corrections, provider outages, model council disagreements, and budget issues.

Runbooks make the system operable under real production pressure.

## Required reading

- `AGENTS.md`
- `docs/OPERATIONS.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/QUALITY_GATES.md`
- `SECURITY.md`

## Dependencies

- ACC-019
- ACC-021
- ACC-024 recommended

## Scope

Allowed files:

- `docs/runbooks/**`
- `docs/OPERATIONS.md` only to link runbooks

## Non-goals

Do not implement code, automation, provider integration, or incident tooling in this task.

## Required runbooks

Create:

1. `docs/runbooks/release-checklist.md`
2. `docs/runbooks/factual-correction.md`
3. `docs/runbooks/private-data-exposure.md`
4. `docs/runbooks/provider-outage.md`
5. `docs/runbooks/render-failure.md`
6. `docs/runbooks/publishing-failure.md`
7. `docs/runbooks/model-council-disagreement.md`
8. `docs/runbooks/cost-budget-exceeded.md`
9. `docs/runbooks/security-finding.md`

## Required structure for each runbook

Each runbook must include:

- purpose;
- severity guidance;
- detection signals;
- immediate containment;
- diagnosis steps;
- resolution steps;
- communication guidance;
- prevention/follow-up;
- artifacts/logs to inspect;
- owner role.

## Acceptance criteria

- All required runbooks exist.
- Each runbook follows the required structure.
- `docs/OPERATIONS.md` links to runbook directory or individual runbooks.
- Runbooks preserve human release authority and safety gates.

## Validation commands

If docs validation exists, run it. Otherwise run available tests/typecheck if codebase exists.

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- weakening incident response;
- recommending hiding public corrections when needed;
- removing human approval requirements;
- adding operational steps that expose secrets.

## PR summary requirements

Include runbooks added, linked docs, and operational gaps remaining.
