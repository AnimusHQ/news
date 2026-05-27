# ACC-031 - Provider and Platform Sandbox Adapters

## Goal

Add sandbox adapters behind the existing provider-agnostic model and publishing interfaces.

These adapters prepare the system for real provider/platform integrations without adding credentials, public uploads, or direct network calls in core logic.

## Dependencies

- ACC-006
- ACC-007
- ACC-021
- ACC-026
- ACC-029
- ACC-030

## Scope

Allowed files:

- `internal/models/sandbox/**`
- `internal/publishing/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- direct HTTP calls to model providers;
- direct YouTube/public publishing;
- real credentials;
- provider-specific SDK dependencies;
- environment-specific deployment config.

## Requirements

1. Add a model provider sandbox adapter that implements `adapters.Provider`.
2. Keep actual provider execution behind an injected client interface.
3. Require explicit enablement and credential reference metadata, but never store a credential value.
4. Enforce allowed privacy tiers before any client call.
5. Reject secret-like prompts before any client call.
6. Normalize model review output and reject invalid verdict/confidence data.
7. Add a private/scheduled publishing sandbox adapter that implements `publishing.Adapter`.
8. Store sandbox draft state locally in memory for tests.
9. Refuse public visibility.
10. Require human release approval for scheduled visibility.

## Acceptance Criteria

- Disabled sandbox model provider fails closed.
- Missing credential reference fails closed.
- Restricted or local-only task data is blocked before client execution.
- Secret-like prompt text is blocked before client execution.
- Invalid model output is rejected.
- Valid model output is normalized.
- Publishing sandbox creates private draft status.
- Publishing sandbox schedules only approved scheduled drafts.
- Publishing sandbox refuses public upload.

## Validation Commands

```bash
go test ./internal/models/sandbox ./internal/publishing
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- No secrets are committed.
- Credential references are labels such as environment variable names, not credential values.
- Public publishing remains impossible through the sandbox adapter.
- Sandbox adapters keep local/mock dry-run behavior as the default path.
