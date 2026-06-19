# REPO-HYGIENE-001 Ledger — Docs, Licensing, Vendor-Name Decoupling, Wiring Tests

Mission: bring the repository to its own stated baseline (project README, explicit
license, honest status banners) and remove commercial-vendor leakage from the
deterministic workflow layer, plus add tests for previously untested wiring
packages. No behavioral change to gates, the publish path, immutability,
self-approval, AI-disclosure, multi-verifier, or the safety model.

## Initial state

Date: 2026-06-19.

Worktree: `/tmp/news-mvp-docker-001` (linked git worktree;
`gitdir: /home/guest/projects/news/.git/worktrees/news-mvp-docker-001`).

Branch: `main`. Head at start: `a4ef737` (Merge PR #3, containerized MVP runtime).

Baseline before edits:

| Command | Result | Notes |
| --- | --- | --- |
| `go build ./...` (GOFLAGS=-buildvcs=false) | pass | VCS stamping fails in this linked worktree under `/tmp`; `-buildvcs=false` is a local workaround only (same finding as CFG-001). |
| `go test ./...` (GOFLAGS=-buildvcs=false) | pass | All packages green. Four packages reported `[no test files]`: `internal/activities`, `internal/models`, `internal/models/adapters`, `internal/worker`. |

## VCS / verification note

As with CFG-001, Go's build VCS stamping fails in this `/tmp` linked worktree
(`error obtaining VCS status: exit status 128`), even though `git status`
succeeds. Local verification used `GOFLAGS=-buildvcs=false`. This is an
environment artifact, not a repository defect; a normal checkout with a regular
`.git` directory does not need the flag and `make verify` runs unmodified.

`make verify` also scans the working tree for secrets. The working tree contained
a pre-existing, gitignored, untracked local file `.env.mvp.local` (created by the
earlier MVP-Docker run) with placeholder API-key assignments; the secret scan
flags it. It is not tracked, not part of this change set, and absent from a clean
checkout, so `make verify` is green on the committed state. Local verification was
confirmed green with that untracked file temporarily moved aside (see status
report for the exact GREEN output).

## Work items

### WI-1 — Project README + profile/ resolution + CONTRIBUTING reconcile

- Rewrote `README.md` to describe Animus News specifically: what-it-is (content
  compiler, not an AI content farm), honest status (pre-production scaffold on
  mock / fail-closed providers; ~Level 3 with Level 4 partial per
  `docs/PRODUCTION_READINESS.md`), non-goals, offline quickstart
  (`make verify` / `make demo` / `make demo-blocked` / pilot CLI), and doc
  pointers. Removed all "organization defaults" framing.
- Removed `profile/README.md` (org GitHub profile content that only renders in a
  repo named `.github`; inert and misleading here). Decision recorded in
  ADR-0014.
- Reconciled `CONTRIBUTING.md` to the proprietary posture: external contributions
  not accepted by default; if invited, only under a written rights-assignment to
  Animus.

### WI-2 — Proprietary LICENSE

- Added top-level `LICENSE`: proprietary, all rights reserved to Animus, public
  visibility grants no license, scoped to this repository only (separately
  published Animus OSS community projects keep their own terms). No SPDX OSS
  identifier or permissive text anywhere.
- Added a licensing line to `README.md` and brief proprietary clarifications to
  `SECURITY.md` and `SUPPORT.md` so no file implies an open-source grant.
- TODO (owner): confirm the exact registered legal entity name. The LICENSE uses
  "Animus" per the task default; if the legal entity differs, the owner must
  update `LICENSE`. (Tracked here, not in `LICENSE`.)

### WI-3 — Status banners on aspirational docs

- Prepended a status banner (target design ≠ current implementation; links to
  `docs/PRODUCTION_READINESS.md`; names the implemented short-form M1–L2 slice on
  mock/fail-closed providers) to: `docs/SYSTEM_BLUEPRINT.md`,
  `docs/MULTIMODEL_STRATEGY.md`, `docs/ROADMAP.md`, `docs/DEVELOPMENT_PLAN.md`,
  `docs/CODEX_MASTER_PLAN.md`, `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md`. Design
  content was not otherwise modified.

### WI-4 — Capability-named workflow activities (semantics-preserving rename)

Renamed vendor-named, workflow-visible short-form activity methods (registered by
method name via `RegisterActivity(NewMockActivities())`):

| Old | New |
| --- | --- |
| `GenerateMockVisualShots` | `GenerateVisualShotsMock` |
| `GenerateSeedanceShots` | `GenerateVisualShotsReal` |
| `GenerateElevenLabsVoiceover` | `GenerateVoiceover` |
| `GenerateUploadPostPublishManifest` | `GeneratePublishManifest` |
| `ValidateUploadPostPublishManifest` | `ValidatePublishManifest` |
| `UploadPostDryRun` (activity) | `PublishDryRun` |
| `UploadPostSchedulePublish` | `PublishSchedule` |

- Updated call sites in `internal/workflows/shortform.go` and the in-process
  runner `internal/shortform/runner/runner.go`, plus matching comments and one
  vendor-named runtime note string ("upload-post" → "publishing").
- Updated `internal/shortform/activities/activities_test.go` (method names + test
  name `TestPublishDryRunRefusesNonDryRunMode`).
- The mock-vs-real split and every fail-closed behavior are preserved:
  `GenerateVisualShotsReal` and `PublishSchedule` still refuse; `PublishDryRun`
  still requires dry_run mode and the human/QA/disclosure preconditions.
- The provider-layer interface method `PublishingProvider.UploadPostDryRun` and
  the `providers/uploadpost`, `providers/voice/omnivoice`,
  `providers/review/claude` packages were intentionally NOT renamed: the provider
  layer is the sanctioned vendor boundary.
- The note-string change altered only the happy-path deterministic result fixture
  in `internal/workflows/shortform_test.go`. Updated that one pinned hash from
  `sha256:522c6ed1…` to `sha256:c156bd41…`. Artifact hashes and gate results are
  byte-identical (verified in the test failure diff); the blocked-path fixtures
  are unchanged.

### WI-5 — Wiring-package tests

Added offline, deterministic tests to the four previously untested packages:

- `internal/worker/worker_test.go`:
  - `TestRegisteredShortFormActivityNamesAreVendorNeutral` — reflects over the
    registered activity set and fails if any registered name encodes a commercial
    vendor (permanent WI-4 guard).
  - `TestCapabilityNamedActivitiesExist` — locks the capability activity names.
  - `TestWorkerRegistrationMatchesWorkflowUsage` — runs `ShortFormWorkflow` in the
    Temporal test environment with the worker's exact registration; fails if a
    registered activity name diverges from a name the workflow invokes
    (registration/usage contract).
- `internal/models/adapters/adapter_test.go` — compile-time interface guards
  (`mock.Provider`, `sandbox.Provider` satisfy `adapters.Provider`) plus a run
  through the interface and error-type checks.
- `internal/models/wiring_test.go` — registry + router resolve the expected task
  categories and fail closed on an unsatisfiable capability.
- `internal/activities/episode_test.go` — the registered long-form placeholder
  activities are offline and deterministic.

## ADRs

- ADR-0014: repository hygiene, proprietary licensing, and vendor-name
  decoupling (profile/ removal rationale, rename approach, banner approach).

## Result classification

- WI-1: Implemented. WI-2: Implemented. WI-3: Implemented. WI-4: Implemented.
  WI-5: Implemented.
- Deferred/owner action: confirm legal entity name in `LICENSE` (Partial only in
  that the placeholder "Animus" may need the exact entity).
