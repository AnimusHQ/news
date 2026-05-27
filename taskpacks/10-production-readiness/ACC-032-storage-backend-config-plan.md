# ACC-032 - Storage Backend Configuration and Migration Plan

## Goal

Add typed configuration and migration planning for future Postgres and S3-compatible storage backends without introducing live service dependencies or credentials.

## Dependencies

- ACC-030
- ACC-031

## Scope

Allowed files:

- `internal/storage/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- real Postgres connections;
- real S3 clients;
- migrations applied to a live database;
- credential loading;
- deployment-specific config files with secrets.

## Requirements

1. Define typed storage backend configuration.
2. Support local filesystem mode and future Postgres/S3-compatible mode.
3. Require credential references, not credential values.
4. Reject raw secret-looking values in config fields.
5. Provide deterministic migration plan steps for the future Postgres schema.
6. Keep the plan reviewable and testable without external services.

## Acceptance Criteria

- Default local config validates.
- Postgres/S3 config validates only with required reference fields.
- Raw-looking secrets are rejected.
- Config can be loaded from YAML.
- Migration plan includes episode state and artifact reference tables.
- Migration plan contains no credential material.

## Validation Commands

```bash
go test ./internal/storage
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- Config uses references such as `env:POSTGRES_DSN`, not secret values.
- No live backend is contacted.
- No credentials are committed.
