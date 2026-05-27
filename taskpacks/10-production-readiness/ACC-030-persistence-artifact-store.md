# ACC-030 - Persistence and Artifact Store Interfaces

## Goal

Introduce durable application state and content-addressed artifact storage behind Go interfaces without changing the existing local dry-run defaults.

This task prepares the project for Postgres and S3-compatible backends while keeping CI and developer workflows offline and deterministic.

## Dependencies

- ACC-002
- ACC-011
- ACC-012
- ACC-024
- ACC-029

## Scope

Allowed files:

- `internal/storage/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- real Postgres connectivity;
- real S3 connectivity;
- credentials or environment-specific configuration;
- workflow migration to persistent storage;
- public publishing.

## Requirements

1. Define provider-neutral Go interfaces for episode state persistence and artifact storage.
2. Provide a local filesystem implementation for tests and offline development.
3. Store artifacts by content hash.
4. Validate canonical artifacts before storing when requested.
5. Preserve immutability for an episode artifact name once written.
6. Make repeated writes of identical content idempotent.
7. Reject path traversal and unsafe artifact names.
8. Keep stored references auditable with episode ID, artifact name, hash, size, URI, and timestamp.

## Acceptance Criteria

- Local filesystem artifact store writes and reads content-addressed artifacts.
- Canonical artifact validation failures prevent storage.
- Rewriting an existing artifact name with different content fails.
- Rewriting identical content returns the same reference.
- Episode state records can be saved, loaded, and linked to stored artifacts.
- Tests cover success, invalid artifact, immutability, idempotency, missing artifact, and path traversal.

## Validation Commands

```bash
go test ./internal/storage
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- No secrets are introduced.
- Local filesystem paths are constrained to the configured root.
- Artifact names and episode IDs are validated before filesystem operations.
- Storage does not bypass canonical artifact validation.

## Documentation Updates

Update the release audit and development plan to note that local interfaces exist, while real Postgres/S3 backends remain future work.
