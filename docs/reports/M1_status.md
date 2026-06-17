# M1 Status Report — OpenShorts → AnimusHQ/news Integration

Scope: prove the **architecture, contracts, and gates** for the short-form video
pipeline run end-to-end on mock providers, with every gate enforced in code and the
whole thing reproducible by one command. Real provider execution is M2/M3.

## Verify command and result

```
make verify
```

Runs: `go build ./...` → `go vet ./...` → `go test ./...` → compile CLI → end-to-end
mock demo (success **and** failure-injected) → schema-validate every produced artifact.
**Result: `M1 VERIFY: GREEN` (exit 0).** No network, no secrets, no spend.

Last full run: `go test ./...` all green (37 packages; ~249 passing tests, ~68 in the
short-form + workflow packages); demo success reaches `published_dry_run_complete` with 8
stamped artifacts; demo `--inject unapproved_storyboard` halts at the `storyboard_image`
gate.

---

## Implemented (committed + passing test/runnable demo)

| Component | Evidence |
| --- | --- |
| Phase 0 audit, delta report, ADRs | `docs/ledger/M1.md` §Phase 0; `docs/adr/0001..0005`. |
| Common envelope + status model (`in_review`, `locked`, `IsTerminalImmutable`) | `internal/artifacts/types.go`; existing artifact tests still green. |
| Deterministic content hashing (canonical JSON, excludes hash field) | `internal/shortform/contenthash` — `TestComputeIsStableAcrossRuns`, `...ExcludesContentHashField`, `...IsKeyOrderIndependent`, `...RoundTripThroughHashField`, `TestVerifyDetectsTampering`. |
| JSON Schema (8 docs) + dependency-free interpreter (fails closed) | `internal/shortform/schemas/*.schema.json`; `internal/shortform/schema` — `TestCompileRejectsUnknownKeyword` + 8 keyword tests. |
| 8 short-form artifacts: types + validation + round-trip + hash + accept/reject | `internal/shortform/artifacts.go`, `validate.go`; `TestArtifactsValidateStampAndRoundTrip`, `TestArtifactHashIsDeterministicAndExcludesHashField`, `TestArtifactsRejectInvalidAgainstSchema`, `TestAllSchemasCompile`. |
| 6 provider interfaces + deterministic mocks + failure injection | `internal/shortform/providers` — `TestMocksEmitSchemaValidDraftArtifacts`, `TestMocksAreDeterministic`, `TestVisualShotsReferenceStoryboardImageHashes`, `TestDefectErrorIsInjectable`, `TestDomainDefectsStaySchemaValid`. |
| 10 gates (5 content/release + 5 invariant) as pure functions | `internal/shortform/gates` — see the gate table below. 98.5% statement coverage. |
| §9 activities on mock providers; M2/M3 paths disabled | `internal/shortform/activities` — `TestActivitiesAreIdempotent`, `TestProductionQADependsOnRenderQuality`, `TestDeferredActivitiesNeverRunInM1`, `TestUploadPostDryRunRefusesNonDryRunMode`. |
| `ShortFormWorkflow` + signals + replay/determinism | `internal/workflows/shortform.go` — `TestShortFormWorkflowHappyPath`, `...StoryboardRejectedBlocks`, `...ReleaseDeniedBlocks`, `...RenderDefectBlocksAtRenderGate`, `...ProviderErrorPropagates`, `...ReplayIsDeterministic`. Registered in `internal/worker/worker.go`. |
| End-to-end demo runner + `animus-news demo` CLI | `internal/shortform/runner` — `TestRunnerHappyPathReachesTerminalState`, `TestRunnerUnapprovedStoryboardHaltsAtStoryboardGate`, `TestRunnerRenderNoAudioHaltsAtRenderGate`, `TestRunnerReleaseDeniedBlocks`, `TestRunnerIsDeterministic`. |
| Immutable, content-addressed artifact store (reused) | `internal/storage` (`ErrImmutableConflict`); pre-existing, green. |
| `make verify` single pass/fail + `make demo`/`demo-blocked` | `Makefile`; runnable, exit 0. |

### Gates — positive and negative tests

| Gate (§) | Positive test | Negative test (each blocking condition) |
| --- | --- | --- |
| StoryboardImageGate (§8) | `TestStoryboardImageGatePasses` | `TestStoryboardImageGateBlocks` (8: manifest_missing, source_not_chatgpt_manual, operator_approval_missing, scene_image_missing, image_path_missing, image_hash_missing, image_not_approved, visual_review_failed) |
| VisualShotGate (§8) | `TestVisualShotGatePasses` | `TestVisualShotGateBlocks` (9) |
| SubtitleGate (§8) | `TestSubtitleGatePasses` | `TestSubtitleGateBlocks` (7) |
| RenderGate (§8) | `TestRenderGatePasses` | `TestRenderGateBlocks` (9) |
| ReleaseGate (§8) | `TestReleaseGatePasses` | `TestReleaseGateBlocks` (8) |
| SelfApprovalGate (§4.6) | `TestSelfApprovalGatePasses` | `TestSelfApprovalGateBlocks` (3) |
| ImmutabilityGate (§4.9) | `TestImmutabilityGatePasses` | `TestImmutabilityGateBlocks` (approved + locked) |
| AIDisclosureGate (§4.8) | `TestAIDisclosureGatePasses` | `TestAIDisclosureGateBlocks` (2) |
| MultiVerifierGate (§4.7) | `TestMultiVerifierGatePasses` | `TestMultiVerifierGateBlocks` (3) |
| ProvenanceGate (§4.1/§4.2) | `TestProvenanceGatePasses` | `TestProvenanceGateBlocks` (2) |

### Code-enforced invariants (each proven by a test)

- No self-approval — `gates.SelfApprovalGate` + `TestSelfApprovalGateBlocks`; mocks never
  author as a human (`TestMocksEmitSchemaValidDraftArtifacts`).
- Approved/locked immutability — `gates.ImmutabilityGate` + `TestImmutabilityGateBlocks`;
  store `ErrImmutableConflict`.
- AI-disclosure-required → release blocked — `gates.AIDisclosureGate` /
  `ReleaseGate` `disclosure_missing` case.
- Only the guarded publish path — `UploadPostSchedulePublish` errors
  (`TestDeferredActivitiesNeverRunInM1`); `UploadPostDryRun` refuses non-dry-run.

---

## Partial (exists but incomplete)

| Component | What's there / what's missing |
| --- | --- |
| Workflow replay test | The signal-driven, multi-task workflow is replayed across workflow-task boundaries by the Temporal **test environment**, and `TestShortFormWorkflowReplayIsDeterministic` asserts byte-identical results across runs. **Missing:** a `worker.WorkflowReplayer` test against a CLI-recorded JSON history (needs a Temporal dev server to record history; offline M1 cannot). Deferred to M2. |
| Production QA | Implemented as a deterministic activity (`RunProductionQA`) tied to render quality. It is **not** the full `production_qa_report.json` canonical artifact pipeline; it returns a decision consumed by the render/release gates. Full production-QA artifact wiring remains future work. |

---

## Planned (not started — deferred)

| Component | Deferred to |
| --- | --- |
| Real FFmpeg render via `exec` (real render gate) | M2 |
| faster-whisper real transcription (sidecar or Go-exec; no pyannote) | M2 |
| Upload-Post dry-run against the real sandbox API | M2 |
| Optional Remotion adapter (flagged; Company License note) | M2 |
| Seedance via fal.ai + fallback provider (real image-to-video) | M3 |
| ElevenLabs Creator+ production audio | M3 |
| Live sanctions/disclosure enforcement, real C2PA signing | M3 |
| Actual scheduled/public publishing, live analytics ingestion | M3 |
| Dashboard/console UI, public gallery/SEO | Out of scope until production path stabilizes |

---

## Definition of Done — M1 checklist

- [x] `go build ./...` passes.
- [x] `go test ./...` passes; every gate has ≥1 positive and ≥1 negative test per blocking condition.
- [x] Every §7 artifact has schema, validator, round-trip, deterministic-hash, accept/reject tests.
- [x] Every provider interface has a mock emitting schema-valid artifacts.
- [x] Temporal workflow passes integration + replay (determinism) tests; signals gate progression. (Replay via test-env cross-task replay; JSON-history replayer is M2 — see Partial.)
- [x] `make verify` runs build + tests + schema validation + e2e demo and returns one pass/fail.
- [x] `demo --episode episode-0001` reaches a terminal state; the failure-injected variant halts at the correct gate. Both tested.
- [x] Code enforces no self-approval, approved-artifact immutability, AI-disclosure-required → release blocked, and no publish path except the guarded one. Each proven by a test.
- [x] No secrets in repo; no test requires network/credentials.
- [x] `docs/ledger/M1.md`, ADRs, `CLAUDE.md`, and this report committed; takeover reproducible from a clean checkout via `make verify`.

---

## Open questions / assumptions (from `docs/ledger/M1.md`)

- **A-01:** `episode-0001` demo inputs are synthesized from a small in-code scene fixture;
  the demo writes to a run dir and does not require the legacy `episodes/0001-after-git-push`
  bundle to contain short-form artifacts (delta D-05).
- **A-02:** "≥2 distinct verifiers" is enforced by `MultiVerifierGate` (storyboard + release
  approvers + production-QA identity), complementing the existing multimodel `model_panel >= 2`.
- **A-03:** Mock outputs use fake-but-valid hashes/paths; render target fixed at
  1080×1920 / 9:16 / 30fps / h264.
- **A-04:** No new Go module dependency (ADR-0003) to guarantee offline takeover reproducibility.
- **D-01:** the driver doc's `cmd/news demo` maps to `cmd/animus-news demo` (the real binary).
- **D-06:** new artifacts use `ai_disclosure_required` (M1/OpenShorts naming); the legacy
  `synthetic_disclosure_required` field in `internal/publishing` is unchanged.
