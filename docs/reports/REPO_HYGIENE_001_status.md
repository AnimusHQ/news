# REPO-HYGIENE-001 Status Report

Task: repository hygiene — project README, proprietary LICENSE, status banners on
aspirational docs, capability-named workflow activities (vendor-name removal), and
tests for previously untested wiring packages.

No change to gate logic, the publish-path invariant, immutability, self-approval,
AI-disclosure, multi-verifier behavior, or the safety model.

## Per work item

### WI-1 — Project README + profile/ resolution + CONTRIBUTING — Implemented

- Changed: `README.md` (full rewrite), `CONTRIBUTING.md` (proprietary posture),
  removed `profile/README.md`.
- Evidence: `README.md` opens with Animus News scope/status/non-goals/quickstart;
  contains no "organization defaults" framing; status is tied to
  `docs/PRODUCTION_READINESS.md` (pre-production scaffold on mock / fail-closed
  providers, ~Level 3 with Level 4 partial). `profile/` removed (`git status`
  shows `D profile/README.md`).

### WI-2 — Proprietary LICENSE — Implemented

- Changed: added `LICENSE`; `README.md` licensing line; clarifying notes in
  `SECURITY.md` and `SUPPORT.md`.
- Evidence: `LICENSE` states proprietary / all-rights-reserved to Animus, that
  public visibility grants no license, and scopes itself to this repository. No
  SPDX OSS identifier or permissive license text exists in the repo.
- Owner action (Partial): confirm exact legal entity name in `LICENSE`
  (placeholder "Animus").

### WI-3 — Status banners — Implemented

- Changed: `docs/SYSTEM_BLUEPRINT.md`, `docs/MULTIMODEL_STRATEGY.md`,
  `docs/ROADMAP.md`, `docs/DEVELOPMENT_PLAN.md`, `docs/CODEX_MASTER_PLAN.md`,
  `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md`.
- Evidence: each opens with a banner distinguishing target design from current
  implementation, linking to `docs/PRODUCTION_READINESS.md`, and naming the
  implemented short-form M1–L2 slice. Design content otherwise unchanged.

### WI-4 — Capability-named activities — Implemented

- Changed: `internal/shortform/activities/activities.go`,
  `internal/workflows/shortform.go`, `internal/shortform/runner/runner.go`,
  `internal/shortform/activities/activities_test.go`,
  `internal/workflows/shortform_test.go` (one pinned fixture hash).
- Evidence: no workflow-layer symbol or registered activity name encodes a
  commercial vendor; all fail-closed methods still refuse; e2e demo behavior
  unchanged (see verification output below). Registered Temporal activity types in
  test logs now read `GenerateVisualShotsMock`, `GenerateVoiceover`,
  `GeneratePublishManifest`, `PublishDryRun`.

### WI-5 — Wiring tests — Implemented

- Added: `internal/worker/worker_test.go`,
  `internal/models/adapters/adapter_test.go`, `internal/models/wiring_test.go`,
  `internal/activities/episode_test.go`.
- Evidence: `go test ./internal/worker/... ./internal/models/... ./internal/activities/...`
  → all `ok`. The registration/usage contract is guarded by
  `TestWorkerRegistrationMatchesWorkflowUsage` (Temporal test env) and the
  vendor-name invariant by `TestRegisteredShortFormActivityNamesAreVendorNeutral`.

## Final verification (exact output)

Run with `GOFLAGS=-buildvcs=false` (linked-worktree workaround) and with the
pre-existing untracked, gitignored `.env.mvp.local` moved aside (it is not part of
this change set and is absent from a clean checkout):

```
==> [6/9] provider capability registry
==> [7/9] compile CLI -> build/animus-news
==> [8/9] end-to-end mock demo (success + failure-injected)
episode:     episode-0001
run dir:     build/verify-demo/success/episode-0001
state:       published_dry_run_complete
blocked:     false
artifacts:   8  gates evaluated: 8
expectation met: terminal
---
episode:     episode-0001
run dir:     build/verify-demo/blocked/episode-0001
state:       blocked
blocked:     true
block reason: storyboard_image: image_not_approved
artifacts:   1  gates evaluated: 2
expectation met: blocked:storyboard_image
==> [9/9] schema validation of produced short-form artifacts

M3 VERIFY: GREEN
```

`go test ./...` → all packages `ok` (the four formerly `[no test files]`
packages — `internal/activities`, `internal/models`, `internal/models/adapters`,
`internal/worker` — now report `ok`).

## Behavior preserved

- Success path reaches `published_dry_run_complete` with 8 artifacts and 8 gates.
- Injected-failure path yields `blocked` at `storyboard_image`
  (`image_not_approved`).
- `GenerateVisualShotsReal` and `PublishSchedule` still refuse; `PublishDryRun`
  still requires `dry_run` mode plus human/QA/disclosure preconditions.

## Classification

Implemented: WI-1, WI-2, WI-3, WI-4, WI-5.
Partial (owner action): exact legal entity name in `LICENSE`.
Planned: none for this task.
