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
| ACC-002 | Complete | Go runtime validators now cover canonical artifact contracts, common metadata, status/schema enums, source provenance, claims, QA decisions, publish safety, analytics metrics, and valid/invalid fixture tests for major artifact types. |
| ACC-003 | Complete | CLI supports `validate`, `validate --json`, and `validate-episode`. |
| ACC-004 | Complete | Pilot episode artifact bundle exists and validates. |
| ACC-005 | Complete | Model registry config and loader exist. |
| ACC-006 | Complete | Provider-agnostic adapter interface and normalized errors exist. |
| ACC-007 | Complete | Deterministic router selects by capability, modality, privacy, risk, status, and degraded policy. |
| ACC-008 | Complete | Deterministic mock providers support verdicts and failure injection. |
| ACC-009 | Complete | Council aggregation preserves dissent and blockers. |
| ACC-010 | Complete | Claim verification enforces source coverage and high-risk blockers. |
| ACC-011 | Complete | Canonical Go episode state machine implements the full documented state table, allowed transitions, actor/reason/timestamp metadata, explicit block/unblock behavior, human-required labels, and invalid-transition tests. |
| ACC-012 | Complete | Lifecycle transition dependency validator enforces required artifacts, schema validation, rejected/superseded status, source artifact hash checks, and QA/production/release gates with machine-readable issues. |
| ACC-013 | Complete | Source registry validates, ranks trust, and blocks community-only high-risk authority. |
| ACC-014 | Complete | Deterministic research pack builder and activity build draft packs from supplied source records/snippets, preserve locators, rank sources, and flag high-risk topics without primary sources. |
| ACC-015 | Complete | Deterministic script claim extractor package, activity, CLI command, tests, and dry-run integration exist. |
| ACC-016 | Complete | Deterministic human QA packet generator, activity, tests, and dry-run recommendation summary exist; it does not mark operator approval. |
| ACC-017 | Complete | Deterministic storyboard generator, activity, validation-backed tests, and dry-run gate check exist; current pilot correctly skips generation until human QA approval. |
| ACC-018 | Complete | Deterministic local HTML preview generator, placeholder asset provenance, render manifest generation, activity, tests, and dry-run render gate check exist. |
| ACC-019 | Complete | Deterministic production QA package, activity, render/output/provenance/policy/verification/human-QA checks, tests, and dry-run gate check exist. |
| ACC-020 | Complete | Strict release pack generator requires approved production QA, includes storyboard chapters, sources, disclosure fields, and a validating publish manifest draft. |
| ACC-021 | Complete | Safe dry-run publishing adapter blocks public upload, requires human approval for scheduling, validates metadata, exposes draft status, and returns normalized adapter errors. |
| ACC-022 | Complete | Provider-agnostic analytics adapter interface, offline fixture adapter, canonical normalization, missing metric reporting, and validating analytics report generation exist. |
| ACC-023 | Complete | Advisory analytics insight reports cover retention, CTR without clickbait, community conversion, cost, factual correction signals, and data quality notes. |
| ACC-024 | Complete | Structured audit events, redaction, release-approval actor validation, memory sink, JSON Lines output, and workflow transition audit events exist. |
| ACC-025 | Complete | Cost events, aggregation by episode/stage/provider/model/day, and warn/approval/block budget policy decisions exist. |
| ACC-026 | Complete | Provider health states, fallback policy package, router health integration, fallback reasons, disabled/degraded/unknown provider tests, and privacy-blocked fallback tests exist. |
| ACC-027 | Complete | Local secret scanner, redaction helper, CLI, tests, and CI integration exist. |
| ACC-028 | Complete | Required operational runbooks exist and are linked from Operations. |
| ACC-029 | Complete | Default pilot dry-run remains safely gated by placeholder evidence, and an approved local fixture path now exercises storyboard generation, render preview, production QA approval, publish draft generation, fixture analytics import, insight generation, generated output paths, and blocked-path tests without network calls. |
| ACC-030 | Complete | Local persistence and content-addressed artifact store interfaces exist with filesystem implementation, canonical artifact validation before storage, immutable episode artifact refs, idempotent writes, state records, and path traversal protections. |
| ACC-031 | Complete | Sandbox model provider and private/scheduled platform publishing adapters exist behind provider-agnostic interfaces, fail closed without explicit enablement/credential references, block unsafe privacy tiers and secret-like prompts before client execution, normalize model output, and refuse public uploads. |
| ACC-032 | Complete | Typed storage backend configuration and deterministic Postgres migration plan exist for future Postgres/S3-compatible storage, requiring credential references instead of values and rejecting raw secret-like config. |
| ACC-033 | Complete | Standard-library sandbox HTTP model client exists behind the sandbox provider client interface, sends normalized JSON, rejects unsafe endpoint schemes, non-2xx responses, and malformed JSON, avoids authorization credentials, and hardens credential references to `env:`, `secretref:`, or `file:` labels. |
| ACC-034 | Complete | Repository-local architecture conformance tests scan production workflow code for forbidden side-effect imports and direct nondeterministic time calls, and scan adapter packages for workflow dependency leaks. |
| ACC-035 | Complete | Storage runtime credential resolver supports `env:`, `file:`, and injected `secretref:` references with redacted credential values, and backend factory returns local storage or fails closed for `postgres_s3` unless external clients are injected. |
| ACC-036 | Complete | Architecture conformance tests now guard against direct public publishing entrypoints, adapter-created public publish results/statuses, analytics imports of mutation boundaries, and analytics reports that disable advisory-only behavior. |

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
- Real provider-specific model adapters are not configured; provider-neutral sandbox HTTP client boundaries now exist and fail closed without credential references.
- Rendering, analytics import, and insight generation now have deterministic local packages, but they still require real approved inputs before public launch.
- Local persistence, object storage interfaces, runtime credential wiring, and backend factory exist, but real Postgres/S3-compatible clients and applied migrations are not implemented.
- No real private/scheduled platform adapter is configured; local sandbox publishing adapter exists and refuses public visibility.

## Recommended Next Slice

Implement the next post-taskpack production-readiness slice:

1. add provider-specific sandbox endpoint contracts outside repository secret material.
2. add real Postgres/S3-compatible clients behind the storage interfaces using the runtime credential wiring.
3. extend architecture checks for artifact hash provenance and activity registration coverage.
