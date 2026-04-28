# ACC-010 — Claim Verification Workflow

## Goal

Implement the claim verification workflow that checks extracted claims against source evidence and multimodel reviewer outputs.

This task turns the principle “no claim without a source” into enforceable pipeline behavior.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/QUALITY_GATES.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/SECURITY_AND_SAFETY.md`

## Dependencies

- ACC-009
- ACC-002
- ACC-008

## Scope

Allowed files:

- `src/verification/**`
- `src/council/**`
- `src/claims/**` only for shared claim types/helpers
- `tests/verification/**`

## Non-goals

Do not implement:

- claim extraction from scripts;
- source ingestion;
- web browsing;
- real model provider calls;
- human QA UI;
- rendering.

## Requirements

1. Load a claim set and source/evidence references.
2. Verify every claim has at least one source reference.
3. For each claim, determine:
   - supported;
   - partially supported;
   - unsupported;
   - contradicted;
   - needs human review.
4. Use mock verifier panel through council interfaces for MVP.
5. Generate `verification_report` artifact.
6. Block high-risk claims if unsupported, contradicted, or needs human review.
7. Preserve reviewer notes.
8. Include evidence locator references in report.

## Acceptance criteria

- Supported claim passes.
- Missing source fails.
- Unsupported high-risk claim blocks verification.
- Low-risk claim can be marked for revision without blocking entire episode if policy allows.
- Contradicted claim blocks.
- Verification report validates.
- Tests are deterministic and offline.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- treating model confidence as source evidence;
- marking unlinked high-risk claims as supported;
- hiding unresolved claims;
- making real provider calls;
- weakening claim status semantics.

## PR summary requirements

Include:

- verification rules implemented;
- blocker policy;
- tests run;
- limitations of MVP verification;
- follow-up tasks for real source-grounded verification.
