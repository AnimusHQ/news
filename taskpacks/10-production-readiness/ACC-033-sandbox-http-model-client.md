# ACC-033 - Sandbox HTTP Model Client

## Goal

Add a standard-library HTTP client behind the sandbox model provider client interface.

This prepares provider execution for private sandbox testing while preserving provider-agnostic routing, fail-closed privacy policy, and secret-free configuration.

## Dependencies

- ACC-006
- ACC-026
- ACC-031
- ACC-032

## Scope

Allowed files:

- `internal/models/sandbox/**`
- `taskpacks/10-production-readiness/**`
- `docs/DEVELOPMENT_PLAN.md`
- `docs/TASKPACK_RELEASE_AUDIT.md`

## Non-goals

Do not implement:

- provider-specific SDK integrations;
- real credential loading or secret resolution;
- direct model calls from workflows;
- public publishing behavior;
- storage backend clients.

## Requirements

1. Implement an HTTP client that satisfies `sandbox.Client`.
2. Use only Go standard-library HTTP primitives.
3. Require `http` or `https` endpoints and reject unsupported schemes.
4. Send normalized provider request data as JSON.
5. Treat credential configuration as a reference only, never as a credential value.
6. Do not send credential references as authorization credentials.
7. Reject non-2xx responses.
8. Reject malformed JSON responses.
9. Keep model output validation in the sandbox provider normalization path.
10. Harden sandbox provider credential references to require `env:`, `secretref:`, or `file:` prefixes.

## Acceptance Criteria

- HTTP client sends the expected JSON request shape.
- HTTP client does not set an `Authorization` header from credential references.
- HTTP client rejects unsupported endpoint schemes.
- HTTP client returns an error for non-2xx responses.
- HTTP client returns an error for malformed JSON responses.
- Sandbox provider rejects missing, unprefixed, or secret-like credential references before client execution.
- Existing local mock and dry-run paths remain unaffected.

## Validation Commands

```bash
go test ./internal/models/sandbox
go test ./...
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

## Security Considerations

- No credentials are committed or loaded.
- Credential references are labels such as `env:ANIMUS_SANDBOX_PROVIDER_TOKEN`.
- Credential references must not be used as bearer tokens or API keys.
- Restricted and local-only data remains blocked before any client call.
- Secret-like prompts remain blocked before any client call.
