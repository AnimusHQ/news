# M2 Status Report - Real Local Execution Adapters Behind Safety Gates

Scope: add real local and dry-run execution adapters while preserving M1 safety:
no live paid providers, no real social upload, no secrets, no direct public
publishing path, and no workflow-side side effects.

## Verification status

Commands run during implementation:

| Command | Result | Notes |
| --- | --- | --- |
| `make verify` | pass | `M2 VERIFY: GREEN`; includes gofmt check, build, vet, full tests, secret scan, CLI compile, mock demo success/block, and short-form schema validation. |
| `make verify-m2-local` | pass | Adapter contract tests and workflow determinism checks. |
| `go test ./internal/shortform/...` | pass | Includes local adapter packages. |
| `go test ./internal/workflows` | pass | Includes determinism fixture. |
| `go test ./cmd/animus-news` | pass | CLI still compiles/tests. |
| `go test ./internal/shortform/providers/render` | pass | FFmpeg integration ran locally when `ffmpeg` was available. |
| `go test ./internal/shortform/providers/subtitles` | pass | Fake sidecar only; no Python/model download. |
| `go test ./internal/shortform/providers/uploadpost` | pass | Dry-run contract only. |
| `go test ./internal/shortform/providers/localexec` | pass | Path containment helpers. |

## Gate results

| Gate | Result | Evidence |
| --- | --- | --- |
| M2-G1 - M1 preserved | Implemented | Post-edit `make verify` is green; mock success reaches `published_dry_run_complete`; injected storyboard failure blocks at `storyboard_image`; no real publish. |
| M2-G2 - FFmpeg adapter safe | Implemented | `internal/shortform/providers/render`; tests cover disabled default, missing binary, invalid path, traversal, timeout, no shell command construction, manifest validation, and local FFmpeg fixture normalization when available. |
| M2-G3 - subtitle adapter boundary safe | Implemented | `internal/shortform/providers/subtitles`; disabled by default, missing binary/model fail closed, fake sidecar contract tests, transcript/caption hashes, word timestamp requirement, invalid path block, safe-zone gate preservation. |
| M2-G4 - Upload-Post dry-run only | Implemented | `internal/shortform/providers/uploadpost`; live mode refused in M2, dry-run needs no API key, render/QA/release/disclosure/platform checks enforced, API key redaction tested. |
| M2-G5 - Temporal replay/determinism improved | Implemented | Added canonical deterministic result fixture hash test in `internal/workflows/shortform_test.go`; full `worker.WorkflowReplayer` JSON-history replay remains Partial. |
| M2-G6 - security and takeover | Implemented | gofmt check, `go vet ./...`, `go test ./...`, `make verify`, and secret scan are green; clean-tree takeover check remains the final handoff step after commit. |

## Implemented

| Component | Evidence |
| --- | --- |
| Shared local adapter safety helpers | `internal/shortform/providers/localexec`; tests reject traversal/outside paths and validate safe output segments. |
| FFmpeg local render adapter | `internal/shortform/providers/render`; disabled by default and uses `exec.CommandContext` with explicit timeouts and controlled roots. |
| faster-whisper sidecar boundary | `internal/shortform/providers/subtitles`; requires configured executable, local model directory, controlled roots, and JSON sidecar response. |
| Upload-Post dry-run adapter | `internal/shortform/providers/uploadpost`; dry-run manifest construction with release, QA, disclosure, and platform checks. |
| Stronger activity dry-run guard | `internal/shortform/activities.UploadPostDryRun`; refuses non-dry-run and obvious missing release/QA/disclosure fields. |
| Workflow determinism fixture | `TestShortFormWorkflowDeterministicResultFixture`; pins terminal state, artifact hashes, and gate sequence to a canonical hash. |
| M2 verify target shape | `Makefile`; `verify` now includes gofmt check and secret scan; `verify-m2-local` is discoverable. |
| ADRs | `docs/adr/0006-ffmpeg-render-adapter.md`, `0007-faster-whisper-sidecar-boundary.md`, `0008-uploadpost-dry-run-adapter.md`. |

## Partial

| Component | Remaining gap |
| --- | --- |
| Temporal JSON-history replay | Offline test environment determinism is stronger and fixture-pinned, but `worker.WorkflowReplayer` against recorded JSON history still needs a Temporal dev-server history capture. |
| FFmpeg media byte determinism | Manifest properties and output hashes are recorded, but exact bytes can vary across FFmpeg builds/codecs. Tests assert contract properties, not cross-machine byte identity. |
| faster-whisper real model execution | Boundary is implemented and fake-sidecar tested. Running a real faster-whisper sidecar requires an explicitly configured local executable and model directory outside default verification. |

## Planned

| Component | Target |
| --- | --- |
| Seedance visual provider | M3 or later; no live calls in M2. |
| ElevenLabs voice provider | M3 or later; no live calls in M2. |
| Upload-Post live provider | M3 or later with new safety ADR and release gates. |
| S3 artifact store | Future persistence milestone. |
| Operator console | Future UI/editorial console milestone. |
| C2PA/disclosure verifier | Future offline/verifiable policy milestone. |

## Security notes

- Real adapters are disabled by default and fail closed when configuration is
  missing.
- Local media adapters resolve inputs under configured roots and write outputs
  only under configured output roots.
- FFmpeg and sidecar execution use `exec.CommandContext` with argv slices and
  explicit timeouts.
- Upload-Post live mode is refused in M2.
- Dry-run tests require no API key; API key redaction is tested.
- No binary fixtures are committed; FFmpeg integration fixtures are generated in
  temporary directories.

## Runtime/config notes

- FFmpeg local execution requires `FFmpegConfig{Enabled: true, InputRoot,
  OutputRoot}` and either `FFmpegPath` or `ffmpeg` on `PATH`.
- faster-whisper sidecar execution requires `FasterWhisperConfig{Enabled: true,
  BinaryPath, InputRoot, OutputRoot, ModelRoot, ModelPath}`.
- Upload-Post dry-run execution requires `DryRunConfig{Enabled: true,
  Mode: "dry_run"}`.
- Default `make verify` remains mock/dry-run and does not require network,
  provider credentials, or real model downloads.
