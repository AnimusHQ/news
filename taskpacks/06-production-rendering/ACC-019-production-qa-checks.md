# ACC-019 — Production QA Checks

## Goal

Implement automated production QA checks for rendered outputs before publishing.

Production QA must catch missing artifacts, broken manifests, missing provenance, unsafe publish defaults, and unresolved verification issues.

## Required reading

- `AGENTS.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/SCHEMAS.md`
- `docs/OPERATIONS.md`

## Dependencies

- ACC-018
- ACC-012

## Scope

Allowed files:

- `src/production-qa/**`
- `tests/production-qa/**`
- `src/artifacts/**` only for shared helpers

## Non-goals

Do not implement real upload, human UI, video content analysis, or real platform policy API integration.

## Checks

Implement checks for:

1. Required render outputs exist.
2. `render_manifest` validates.
3. `asset_manifest` validates.
4. All assets have provenance.
5. All generated/imported assets have license status.
6. Subtitle file exists when required by render manifest.
7. Publish manifest does not request direct public visibility.
8. Synthetic disclosure status exists when required by metadata.
9. Verification report has no unresolved high-risk blockers.
10. Human QA approval exists before production QA approval.

## Output

Canonical `production_qa_report.json`.

## Acceptance criteria

- Missing render output fails.
- Missing asset provenance fails.
- Direct public publish intent fails.
- Unresolved high-risk claim fails.
- Passing fixture produces approved QA report.
- Tests are deterministic and offline.

## Validation commands

```bash
pnpm test
pnpm typecheck
```

## Mutation policy

Forbidden:

- treating warnings as approval for blocking checks;
- bypassing human QA;
- allowing direct public publishing;
- ignoring missing provenance.

## PR summary requirements

Include QA checks implemented, blocker policy, tests run, and limitations.
