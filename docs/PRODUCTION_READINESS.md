# Production Readiness Checklist

## Current repository status

The repository is now prepared for Go + Temporal production implementation. It includes:

- architecture documentation;
- Go module baseline;
- CLI skeleton;
- structural episode validation;
- safe dry-run skeleton;
- Temporal workflow skeleton;
- Temporal activity stubs;
- pilot episode artifact bundle;
- CI workflow;
- Codex task packs and agent instructions.

The current implementation is a **production-start scaffold**, not a fully production-ready running news-generation platform. It is intentionally safe-by-default: no real model calls, no provider credentials, no uploads, and no public publishing.

## Launch readiness levels

```mermaid
flowchart TD
  L0[Level 0: Architecture Ready] --> L1[Level 1: Go Scaffold Ready]
  L1 --> L2[Level 2: Local Dry Run Ready]
  L2 --> L3[Level 3: Temporal Local Ready]
  L3 --> L4[Level 4: Provider Sandbox Ready]
  L4 --> L5[Level 5: Private Production Ready]
  L5 --> L6[Level 6: Public Launch Ready]
```

## Level 0 — Architecture Ready

Complete when:

- system blueprint exists;
- multimodel strategy exists;
- security model exists;
- quality gates exist;
- artifact schemas are documented;
- Codex execution plan exists.

Status: complete.

## Level 1 — Go Scaffold Ready

Complete when:

- Go module exists;
- CLI entrypoint exists;
- core packages exist;
- basic tests exist;
- CI exists;
- pilot artifact bundle exists.

Status: substantially complete.

## Level 2 — Local Dry Run Ready

Complete when:

- `go test ./...` passes;
- `go run ./cmd/animus-news validate-episode episodes/0001-after-git-push` passes;
- `go run ./cmd/animus-news dry-run episodes/0001-after-git-push` passes;
- no network or secrets required.

Status: pending verification in CI/Codex environment.

## Level 3 — Temporal Local Ready

Complete when:

- Temporal workflow tests pass with the Go SDK test environment;
- activities are registered and tested;
- workflow waits for human QA signal;
- workflow waits for release approval signal;
- invalid transitions block;
- replay/determinism constraints are documented and tested.

Status: scaffold exists; tests and worker runtime pending.

## Level 4 — Provider Sandbox Ready

Complete when:

- model registry exists;
- mock providers exist;
- model router exists;
- multimodel council exists;
- claim verification uses mock providers;
- provider health and fallback policy exists;
- cost tracking exists.

Status: planned in task packs.

## Level 5 — Private Production Ready

Complete when:

- Postgres persistence exists;
- object storage exists;
- artifacts are immutable/versioned;
- provider credentials are managed through secrets;
- private/scheduled publishing adapter is implemented;
- production QA blocks unsafe release;
- audit logging is enforced;
- incident runbooks exist;
- private upload is possible but public release is gated.

Status: future implementation.

## Level 6 — Public Launch Ready

Complete when:

- real source ingestion is safe and provenance-preserving;
- real multimodel council is active;
- real human QA console or workflow is available;
- real rendering pipeline exists;
- real publishing dry-run and scheduled release are tested;
- security scanning passes;
- correction workflow is rehearsed;
- first episode passes final human editorial approval.

Status: future implementation.

## Non-negotiable launch blockers

The system must not publicly launch if any of these are true:

- any high-risk claim is unsupported;
- source locators are placeholders;
- human QA is missing;
- production QA is missing;
- publish manifest requests public visibility by default;
- secrets are present in artifacts/logs/descriptions;
- synthetic media disclosure is unresolved;
- provider-specific model code bypasses the router;
- Temporal workflow can publish without release approval;
- incident correction process is missing.

## Immediate next implementation sequence

1. Run Codex on Go-corrected ACC-000/ACC-001 cleanup if CI exposes issues.
2. Implement deep Go artifact validators.
3. Implement model registry and mock providers.
4. Implement model router.
5. Implement multimodel council.
6. Implement claim extraction and verification.
7. Implement Temporal workflow tests.
8. Implement local worker command.
9. Implement dry-run publish pack.
10. Implement real provider sandbox behind adapters.

## Required local verification commands

```bash
go test ./...
go vet ./...
go run ./cmd/animus-news validate-episode episodes/0001-after-git-push
go run ./cmd/animus-news dry-run episodes/0001-after-git-push
```

## Current safety posture

Safe by default:

- no real provider calls;
- no credentials;
- no public publishing;
- pilot episode is draft/dry-run only;
- placeholder claims are not represented as production-approved;
- release approval is modeled as a workflow signal.
