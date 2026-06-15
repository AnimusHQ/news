# ACC-035 - Storage Credential Runtime Wiring

## Goal

Add runtime credential-reference resolution and storage backend wiring without committing, logging, or serializing credential values.

This prepares Postgres/S3-compatible clients for a later task while keeping local development and CI credential-free.

## Dependencies

- ACC-030
- ACC-032
- ACC-034

## Scope

Allowed files:

- `internal/storage/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- a real Postgres driver dependency;
- a real S3 SDK dependency;
- direct network calls;
- committed credentials;
- public publishing or workflow changes.

## Requirements

1. Resolve `env:` credential references from process environment only at runtime.
2. Resolve `file:` credential references from explicitly referenced local files only at runtime.
3. Treat `secretref:` references as requiring an injected secret resolver.
4. Return credential values through a type that redacts string, Go-string, and JSON representations.
5. Reject empty resolved credentials.
6. Keep raw secret values out of errors.
7. Add a storage backend factory that returns the local filesystem backend for `local` mode.
8. Make `postgres_s3` mode fail closed unless external artifact and episode clients are injected.
9. Keep all checks repository-local and deterministic.

## Acceptance Criteria

- Environment references resolve successfully when set.
- File references resolve successfully from temp files in tests.
- `secretref:` without an injected resolver fails closed.
- Resolved credentials redact in `fmt`, `%#v`, and JSON output.
- Local backend factory returns usable artifact and episode stores.
- `postgres_s3` backend factory refuses to start without injected clients.
- `postgres_s3` backend factory accepts injected clients without resolving or storing secrets.

## Validation Commands

```bash
go test ./internal/storage
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- No credential values are committed.
- Credential values are not logged or serialized by default.
- `secretref:` support is an integration hook, not a repository secret store.
- Local dry-run behavior remains credential-free.
