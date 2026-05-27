# ACC-034 - Architecture Conformance Tests

## Goal

Add automated Go tests that check core architecture boundaries described by the taskpacks and repository instructions.

These tests should catch forbidden workflow side effects and provider/workflow dependency leaks before they reach review.

## Dependencies

- ACC-006
- ACC-011
- ACC-024
- ACC-026
- ACC-031
- ACC-033

## Scope

Allowed files:

- `internal/architecture/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- a full static analysis framework;
- runtime Temporal replay checks;
- provider-specific policy logic;
- broad refactors of workflow, provider, publishing, or storage code.

## Requirements

1. Add a Go test package for architecture conformance.
2. Fail if production workflow code imports direct side-effect packages such as filesystem, network, database, random, or provider SDK packages.
3. Fail if production workflow code calls nondeterministic direct time helpers such as `time.Now`.
4. Fail if model, provider, or publishing adapter packages import workflow packages.
5. Keep the checks deterministic and repository-local.
6. Do not require network, credentials, Temporal service, Postgres, or object storage.

## Acceptance Criteria

- Architecture tests pass against the current repository.
- A future direct workflow import of `os`, `net/http`, `database/sql`, random packages, or known provider SDK roots fails the test.
- A future direct workflow call to `time.Now`, `time.Sleep`, or timer helpers fails the test.
- A future adapter import of `internal/workflows` fails the test.
- `go test ./...` includes the conformance checks.

## Validation Commands

```bash
go test ./internal/architecture
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- Tests operate on local source files only.
- No secrets are read or resolved.
- The checks support the invariant that workflows orchestrate activities instead of performing side effects directly.
