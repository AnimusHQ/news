# Taskpack Release Audit

Date: 2026-04-29

## Release Position

The repository is ready for a local MVP dry-run release: it builds, validates the pilot episode, runs the safe dry-run pipeline, and blocks public publishing by design.

It is not ready for public launch. Public launch remains blocked by placeholder source locators, real-provider absence, missing persistence/object storage, and missing final human release approval for a real episode.

## Status Legend

- Complete: implemented and covered by local checks.
- Partial: safe scaffold exists, but the taskpack is not fully implemented.
- Missing: no concrete implementation beyond fixtures or documentation.

## Taskpack Status

| Task | Status | Notes |
|---|---|---|
| ACC-000 | Complete | Go module, CLI, Make targets, and package baseline exist. |
| ACC-001 | Complete | GitHub Actions runs Go checks, secret scan, pilot validation, and dry-run. |
| ACC-002 | Partial | Core structs and validation exist, but schemas are still minimal and not strict per artifact. |
| ACC-003 | Complete | CLI supports `validate`, `validate --json`, and `validate-episode`. |
| ACC-004 | Complete | Pilot episode artifact bundle exists and validates. |
| ACC-005 | Complete | Model registry config and loader exist. |
| ACC-006 | Complete | Provider-agnostic adapter interface and normalized errors exist. |
| ACC-007 | Complete | Deterministic router selects by capability, modality, privacy, risk, status, and degraded policy. |
| ACC-008 | Complete | Deterministic mock providers support verdicts and failure injection. |
| ACC-009 | Complete | Council aggregation preserves dissent and blockers. |
| ACC-010 | Complete | Claim verification enforces source coverage and high-risk blockers. |
| ACC-011 | Partial | Temporal workflow skeleton exists, but the full documented state table is not implemented as a reusable state machine. |
| ACC-012 | Partial | Required artifact and release-safety validation exists; stale dependency/hash enforcement is not implemented. |
| ACC-013 | Complete | Source registry validates, ranks trust, and blocks community-only high-risk authority. |
| ACC-014 | Partial | Research audit exists; a full research pack builder activity is not implemented. |
| ACC-015 | Complete | Deterministic script claim extractor package, activity, CLI command, tests, and dry-run integration exist. |
| ACC-016 | Complete | Deterministic human QA packet generator, activity, tests, and dry-run recommendation summary exist; it does not mark operator approval. |
| ACC-017 | Complete | Deterministic storyboard generator, activity, validation-backed tests, and dry-run gate check exist; current pilot correctly skips generation until human QA approval. |
| ACC-018 | Complete | Deterministic local HTML preview generator, placeholder asset provenance, render manifest generation, activity, tests, and dry-run render gate check exist. |
| ACC-019 | Complete | Deterministic production QA package, activity, render/output/provenance/policy/verification/human-QA checks, tests, and dry-run gate check exist. |
| ACC-020 | Partial | Publish pack generator exists with safe defaults, but chapters/disclosure handling is minimal. |
| ACC-021 | Complete | Safe dry-run publishing adapter blocks public upload and requires human approval for scheduling. |
| ACC-022 | Missing | Analytics import interface/package is not implemented beyond fixture artifact. |
| ACC-023 | Missing | Analytics insight generator is not implemented beyond fixture artifact. |
| ACC-024 | Partial | Structured audit events and memory sink exist; workflow/router integration is minimal. |
| ACC-025 | Complete | Cost events, aggregation, and budget decisions exist. |
| ACC-026 | Partial | Router handles degraded/disabled model policy; standalone provider health/fallback package is not implemented. |
| ACC-027 | Complete | Local secret scanner, redaction helper, CLI, tests, and CI integration exist. |
| ACC-028 | Complete | Required operational runbooks exist and are linked from Operations. |
| ACC-029 | Partial | End-to-end local dry-run passes, but it does not generate every downstream artifact from scratch. |

## Current Release Checks

These checks must pass for the local MVP dry-run release:

```bash
go test ./...
go vet ./...
go run ./cmd/animus-news scan-secrets .
go run ./cmd/animus-news validate episodes/0001-after-git-push/research_pack.json
go run ./cmd/animus-news validate --json episodes/0001-after-git-push/research_pack.json
go run ./cmd/animus-news validate-episode episodes/0001-after-git-push
go run ./cmd/animus-news dry-run episodes/0001-after-git-push
```

## Public Launch Blockers

- Pilot `claims.json` still uses placeholder evidence ranges and `needs_human_review` statuses.
- `verification_report.json`, `multimodel_approval_report.json`, and `human_qa_report.json` intentionally request revision.
- Real model/provider adapters are not implemented.
- Rendering, analytics import, and insight generation are fixtures/scaffolds.
- Durable Postgres/object storage and immutable evidence bundles are not implemented.
- No real private/scheduled platform adapter is configured.

## Recommended Next Slice

Implement ACC-020 through ACC-023 as bounded Go task packs:

1. publish pack chapters and disclosure handling;
2. safe private/scheduled publishing refinements;
3. analytics fixture import;
4. advisory analytics insight reports.
