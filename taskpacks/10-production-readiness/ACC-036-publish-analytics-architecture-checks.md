# ACC-036 - Publish and Analytics Architecture Checks

## Goal

Extend architecture conformance tests to protect publishing and analytics boundaries.

Publishing must not grow a direct public upload path, and analytics must remain advisory-only rather than mutating editorial or release metadata.

## Dependencies

- ACC-021
- ACC-022
- ACC-023
- ACC-034

## Scope

Allowed files:

- `internal/architecture/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- new publishing adapters;
- analytics provider integrations;
- public upload functionality;
- broad static analysis framework;
- workflow behavior changes.

## Requirements

1. Fail if production publishing code exposes direct public upload, publish, or schedule entrypoints.
2. Fail if publishing adapter code directly creates public `PublishResult` or `DraftStatus` values.
3. Fail if production analytics code imports publishing, workflow, or storage packages.
4. Fail if production analytics code explicitly sets `AdvisoryOnly` to false.
5. Prove canonical analytics report constructors still produce advisory-only reports.
6. Keep tests deterministic and repository-local.

## Acceptance Criteria

- Architecture tests pass against the current repository.
- Future direct public upload-style function names in production publishing code fail tests.
- Future adapter results/statuses with public visibility fail tests.
- Future analytics imports of publishing/workflow/storage packages fail tests.
- Future analytics reports that unset advisory-only fail tests.

## Validation Commands

```bash
go test ./internal/architecture
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- No network, platform, or credential calls are introduced.
- Public publishing remains blocked by design.
- Analytics recommendations remain advisory and cannot bypass editorial gates.
