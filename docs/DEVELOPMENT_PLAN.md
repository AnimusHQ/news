# Animus News Development Plan

> **Status: target design, not current implementation.** This document describes
> the intended target system. For authoritative, code-backed status see
> [`PRODUCTION_READINESS.md`](PRODUCTION_READINESS.md). **Implemented today:** the
> short-form (M1–L2) typed-contract / gate / `ShortFormWorkflow` slice running
> end-to-end on mock and fail-closed (disabled-by-default) providers — no live
> calls, no spend, no public publishing.

## 1. Objective

Build Animus News into a source-grounded, multimodel, artifact-driven production system for educational IT media.

The near-term target is not public publishing. The next target is a reliable local and Temporal-backed MVP that can:

- validate canonical episode artifacts;
- build or update artifacts through deterministic activities;
- route model tasks through provider-agnostic adapters;
- preserve multimodel dissent;
- require human QA and release approval;
- generate deterministic render/preview outputs;
- run production QA;
- generate a safe private/scheduled publish pack;
- complete an end-to-end dry run without real credentials or public upload.

## 2. Current Baseline

The repository currently supports a local dry-run scaffold:

- Go CLI and Make/PowerShell start paths;
- pilot episode artifact bundle;
- strict per-artifact runtime validation;
- source registry and research audit;
- deterministic script claim extraction;
- deterministic human QA decision packet generation;
- deterministic storyboard generator gated by human QA;
- deterministic local render preview/manifest generator gated by storyboard;
- deterministic production QA checks gated by render output;
- strict release pack generation with storyboard chapters and disclosure fields;
- offline analytics import and advisory insight reports;
- approved-fixture end-to-end dry-run coverage for storyboard, render preview, production QA, publish draft, and analytics insight generation;
- structured audit events with workflow transition tracing;
- cost aggregation with warn/approval/block budget policy;
- provider health and fallback policy integrated into model routing;
- deterministic research pack builder from supplied source records and snippets;
- lifecycle transition dependency validation with source artifact hash checks;
- canonical episode state machine with explicit gates and block/unblock transitions;
- local persistence and content-addressed artifact store interfaces with filesystem implementation;
- storage runtime credential-reference resolver and backend factory;
- sandbox model provider and private/scheduled publishing adapters behind provider-agnostic interfaces;
- typed Postgres/S3-compatible backend configuration and deterministic migration plan;
- standard-library sandbox HTTP model client behind the provider client interface;
- repository-local architecture conformance tests for workflow, adapter, publishing, and analytics boundaries;
- model registry, router, mock providers, and council aggregation;
- deterministic claim verification;
- safe publish pack and dry-run publishing adapter;
- cost tracking and secret scanning;
- Temporal workflow skeleton and worker command;
- operational runbooks;
- taskpack release audit.

The current dry run intentionally reports `revision_required` / `request_revision` for the pilot because evidence locators are placeholders. That is correct behavior and must not be bypassed.

## 3. Development Principles

Every task must preserve these constraints:

- Go is the canonical backend implementation language.
- Temporal is the canonical orchestration layer for durable workflows.
- Workflow code must stay deterministic.
- Side effects belong in activities, not workflows.
- Every pipeline stage must produce typed artifacts.
- No claim may proceed without source evidence.
- No single model may be final authority.
- Model outputs are evidence, not approval.
- Human QA and release approval are mandatory for publishable output.
- Public upload must never be the default path.
- Provider-specific code must stay behind adapters.
- Secrets must not enter artifacts, logs, prompts, generated assets, or examples.

## 4. Phase Plan

### Phase 0 - Stabilize Local Developer Experience

Related taskpacks: ACC-000, ACC-001, ACC-003, ACC-027, ACC-029.

Goals:

- Keep `go test ./...`, `go vet ./...`, secret scan, validation, and dry-run green.
- Keep `scripts/smoke.ps1` and `make smoke` aligned.
- Ensure a new developer can run the project without a Temporal service.
- Keep `go.sum` committed and CI reproducible.

Deliverables:

- documented start commands;
- local smoke scripts;
- CI parity with local smoke checks;
- clear dry-run output and failure modes.

Exit criteria:

```bash
go test ./...
go vet ./...
go run ./cmd/animus-news scan-secrets .
go run ./cmd/animus-news validate-episode episodes/0001-after-git-push
go run ./cmd/animus-news dry-run episodes/0001-after-git-push
```

### Phase 1 - Complete Artifact Contracts

Related taskpacks: ACC-002, ACC-003, ACC-004, ACC-012.

Goals:

- Replace minimal artifact validation with strict per-artifact validators.
- Validate required metadata, status values, dependency references, and content hashes.
- Detect stale artifacts when upstream dependencies change.
- Keep validation reusable from CLI and Temporal activities.

Deliverables:

- validators for every canonical artifact;
- valid and invalid fixtures for every artifact type;
- dependency graph validation;
- machine-readable validation reports;
- tests for missing, malformed, rejected, superseded, and stale artifacts.

Architecture conformance checks:

- all artifacts include schema version and episode ID where required;
- publish manifest cannot default to public;
- human QA report requires explicit decision;
- claims require source IDs and evidence locators where risk requires them.

### Phase 2 - Research, Claims, and Human QA

Related taskpacks: ACC-014, ACC-015, ACC-016.

Goals:

- Implement a deterministic research pack builder from explicitly supplied source records and snippets.
- Implement script claim extraction.
- Generate human QA decision packets from research, claims, verification, and council output.

Deliverables:

- `internal/research` builder activity;
- `internal/claims` extractor package;
- `internal/qa` packet generator;
- fixtures for approved, revision-required, and blocked paths;
- CLI or dry-run integration that can regenerate these artifacts.

Testing:

- unit tests for source coverage and forbidden simplifications;
- claim extractor tests for factual, opinion, technical, high-risk, and unlinked claims;
- QA packet tests proving dissent and blockers cannot be hidden.

Architecture conformance checks:

- no source fabrication;
- no model memory as evidence;
- no auto-approval of human QA;
- high-risk unsupported claims block progress.

### Phase 3 - Temporal Workflow Completion

Related taskpacks: ACC-011, ACC-012, ACC-024, ACC-026.

Goals:

- Convert the current workflow skeleton into the canonical episode lifecycle.
- Add complete signals and queries.
- Integrate artifact validation, council status, cost summary, and blocking issue queries.
- Integrate structured audit events.
- Add provider health and fallback policy as explicit workflow/activity inputs.

Deliverables:

- explicit state machine helper;
- complete Temporal workflow stages;
- activity interfaces for each side-effecting stage;
- query handlers for state, artifacts, blockers, council status, and cost;
- signal handlers for QA, release, block, revision, and correction;
- workflow tests for happy path, blocked path, revision path, and release-denied path.

Testing:

- Temporal SDK workflow tests;
- deterministic replay-oriented tests where practical;
- invalid transition tests;
- signal validation tests;
- activity retry/fail-closed tests.

Architecture conformance checks:

- workflow code performs no direct file I/O, provider calls, random, wall-clock calls, rendering, or publishing;
- all side effects are activities;
- human waits are explicit signals;
- blocked states cannot silently continue.

### Phase 4 - Storyboard, Render Preview, and Production QA

Related taskpacks: ACC-017, ACC-018, ACC-019.

Goals:

- Generate storyboard artifacts from approved script and QA packet.
- Generate deterministic preview/render outputs from storyboard and assets.
- Run automated production QA over render manifest, asset provenance, captions, policy fields, and publish intent.

Deliverables:

- `internal/storyboard` generator;
- `internal/render` deterministic preview generator;
- `internal/productionqa` checks;
- render manifest generation;
- production QA report generation;
- fixtures for missing assets, missing provenance, unsafe visibility, and unresolved claims.

Testing:

- storyboard segmentation tests;
- scene ID stability tests;
- missing scene field failures;
- render manifest validation tests;
- production QA blocker tests.

Architecture conformance checks:

- no final render without verified claims and human QA;
- no asset without provenance and license status;
- no public publishing intent from production QA;
- deterministic render output paths.

### Phase 5 - Publishing and Analytics Loop

Related taskpacks: ACC-020, ACC-021, ACC-022, ACC-023.

Goals:

- Complete publish pack generation.
- Keep private/scheduled publishing safe by default.
- Add analytics import interfaces and advisory insight reports.

Deliverables:

- chapter generation from storyboard timing;
- disclosure field handling;
- publish manifest validation;
- analytics fixture adapter;
- analytics insight report generator;
- tests proving analytics cannot override editorial gates.

Testing:

- safe visibility tests;
- missing source list tests;
- malformed analytics fixture tests;
- low retention recommendation tests;
- low CTR recommendation tests without clickbait;
- factual correction signal tests.

Architecture conformance checks:

- no direct public upload;
- public/scheduled transitions require human release approval;
- analytics recommendations are advisory only;
- no automatic metadata mutation after publication.

### Phase 6 - Provider Sandbox

Related taskpacks: ACC-005 through ACC-010, ACC-026.

Goals:

- Introduce real provider adapters only behind existing interfaces.
- Keep mock providers available for CI and local dry runs.
- Add provider health, fallback policy, privacy filters, and cost tracking.

Deliverables:

- provider health package;
- fallback policy tests;
- real adapter skeletons with no committed credentials;
- provider-neutral sandbox HTTP client;
- request/response normalization;
- provider error mapping;
- privacy gate before external calls.

Testing:

- mock-provider contract tests;
- provider unavailable, timeout, rate-limit, invalid-output, policy-blocked tests;
- privacy-blocked fallback tests;
- budget exceeded tests.

Architecture conformance checks:

- no provider becomes global authority;
- disabled providers are never selected;
- restricted data cannot fall back to lower privacy providers;
- all real provider calls happen in activities.

### Phase 7 - Persistence and Artifact Store

Related architecture: Go/Temporal plan, Operations, Security and Safety.

Goals:

- Introduce durable state and artifact storage without changing core architecture.
- Keep artifact hashes and provenance auditable.

Deliverables:

- Postgres-backed application state;
- S3-compatible artifact storage;
- runtime credential-reference resolution;
- local/injected backend factory;
- immutable evidence bundles;
- content-addressed asset storage;
- migration strategy;
- local development compose file if approved.

Testing:

- repository tests with isolated test database;
- artifact hash and immutability tests;
- credential reference redaction tests;
- local backend factory tests;
- retry/idempotency tests;
- backup/restore smoke checks.

Architecture conformance checks:

- approved artifacts are immutable;
- revisions create new versions;
- storage layer does not bypass validation;
- secrets are not stored in artifacts.
- resolved credentials are never logged or serialized.

### Phase 8 - Private Production Readiness

Goals:

- Run the complete workflow on a real but private episode candidate.
- Perform human QA and release approval without public launch.
- Rehearse incident and correction runbooks.

Deliverables:

- private/scheduled platform adapter in sandbox mode;
- operator release checklist;
- first private preview;
- production QA report;
- correction drill;
- post-run audit report.

Exit criteria:

- all local and CI checks pass;
- Temporal workflow completes;
- human QA and release approval are recorded;
- private/scheduled upload path is tested;
- no public upload path exists without explicit release approval;
- runbooks are rehearsed.

### Phase 9 - Public Launch Readiness

Goals:

- Prepare first public episode only after all gates pass.

Exit criteria:

- all high-risk claims are supported;
- no placeholder source locators remain;
- real multimodel council report exists;
- human QA approves;
- production QA approves;
- release approval is recorded;
- publish manifest is private or scheduled before release;
- security scan passes;
- correction plan is ready;
- public metadata is reviewed by a human.

## 5. Testing Strategy

### Required Local Checks

Run before every merge:

```bash
go test ./...
go vet ./...
go run ./cmd/animus-news scan-secrets .
go run ./cmd/animus-news validate-episode episodes/0001-after-git-push
go run ./cmd/animus-news dry-run episodes/0001-after-git-push
```

On Windows:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/smoke.ps1
```

### Test Layers

| Layer | Purpose |
|---|---|
| Unit tests | Validate pure Go helpers, schemas, routing, council logic, QA rules, cost, security. |
| Artifact fixture tests | Prove valid examples pass and invalid examples fail. |
| Activity tests | Test side-effect boundaries with local files and mock providers. |
| Workflow tests | Test Temporal lifecycle, signals, queries, retries, and blocked states. |
| Dry-run tests | Prove the pilot pipeline works offline with no real credentials. |
| Security tests | Prove fake secrets are detected and redacted without committing real secrets. |
| Architecture tests | Detect forbidden dependencies and workflow nondeterminism patterns. |
| CI checks | Re-run core local checks on every pull request. |

### Future Architecture Test Ideas

- Extend workflow boundary checks beyond imports and direct time calls into activity registration coverage.
- Fail if generated artifacts omit source artifacts or content hashes once strict schemas are implemented.
- Fail if generated artifact refs are not linked into storage state.

## 6. Architecture Conformance Review

Every taskpack PR must include a short conformance section:

- Taskpack IDs implemented.
- Files changed and whether they match allowed scope.
- Artifact contracts added or changed.
- Validation examples added or updated.
- Workflow/activity boundary review.
- Provider independence review.
- Human QA/release gate review.
- Security and secret-handling review.
- Tests run.
- Known gaps and follow-up taskpacks.

Review checklist:

- Does this preserve Go/Temporal as the backend/orchestration stack?
- Does it keep workflow code deterministic?
- Does it keep provider-specific logic behind adapters?
- Does it avoid direct public publishing?
- Does it preserve model dissent?
- Does it require human approval where required?
- Does it keep source provenance and claim evidence visible?
- Does it fail closed on uncertainty?
- Does it add tests for both success and blocked paths?

## 7. Definition of Done

A taskpack is complete only when:

- implementation matches the taskpack scope;
- all relevant artifacts validate;
- tests cover success and failure cases;
- no architecture invariant is weakened;
- no secret or credential is introduced;
- dry-run behavior remains safe;
- CI-relevant commands pass;
- documentation or audit status is updated;
- remaining gaps are recorded.

## 8. Immediate Next Milestone

The ACC-002 through ACC-036 taskpack slice is now implemented as deterministic Go packages with safe dry-run gate checks. The next milestone should move beyond the initial taskpacks:

1. Add provider-specific sandbox endpoint contracts outside repository secret material.
2. Add real Postgres/S3-compatible clients behind the storage interfaces using the runtime credential wiring.
3. Extend architecture checks for artifact hash provenance and activity registration coverage.

This next milestone will move the project from local MVP demonstration toward private production readiness.
