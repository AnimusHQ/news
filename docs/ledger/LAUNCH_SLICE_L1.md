# Launch Slice L1 Ledger - Real CLI Pilot and Connector Blueprint

Mission: make Animus News produce a real prompt-driven short-form
`release_candidate` MP4 through safe provider boundaries, while documenting the
future connector architecture and final workflow.

## Initial state

Date: 2026-06-17

Current branch: `m1-openshorts-integration`

Recent baseline commits:

- `597c90b M3: finalize takeover evidence`
- `8826f0f M3: document provider expansion milestone`
- `f63d9c5 M3: add provider capability registry and verification target`
- `47c0454 M3: add OmniVoice voice provider boundary`
- `3dab1bb M3: add DaVinci Resolve MCP render boundary`
- `930228b M2: document local adapter milestone`
- `cfb1366 M1 Phase 6: make verify, M1 status report, CLAUDE.md, takeover`

Pre-change verification:

| Command | Result | Evidence |
| --- | --- | --- |
| `git status --porcelain` | pass | clean output before edits |
| `git branch --show-current` | pass | `m1-openshorts-integration` |
| `git log --oneline -n 20` | pass | M1-M3 commits visible |
| `make verify` | pass | `M3 VERIFY: GREEN` |
| `make verify-m2-local` | pass | local adapter and workflow determinism checks passed |
| `make verify-m3` | pass | M3 provider boundary and replay checks passed |
| `go vet ./...` | pass | no output |
| `go test ./...` | pass | all packages green |

## Scope

Implemented files are limited to:

- L1 pilot CLI/package;
- tests;
- Makefile and `.gitignore`;
- required L1 docs, ledger, status report, and Claude guidance.

Non-goals preserved:

- no Review Room UI;
- no Postgres/S3 implementation;
- no analytics ingestion;
- no live Upload-Post/public publishing;
- no native Seedance/OmniVoice/DaVinci production execution;
- no native social platform APIs.

## Implementation notes

Added `internal/shortform/pilot`:

- episode workspace creation;
- deterministic local script generation;
- manual Claude script review request/import;
- visual shot request generation;
- external-command visual provider protocol;
- external-command voice provider protocol;
- faster-whisper external sidecar protocol;
- explicit script-timing subtitle fallback;
- FFmpeg render to `dist/<episode-id>-release-candidate.mp4`;
- ffprobe render validation;
- manual Claude final QA request/import;
- production QA report and non-live publish manifest;
- status and validation reporting.

Added CLI commands:

```text
animus-news pilot generate-real
animus-news pilot resume
animus-news pilot status
animus-news pilot validate
animus-news pilot import-claude-review
animus-news pilot import-visual-shot
animus-news pilot import-voice
```

Added `make verify-real-pilot`.

## Gate mapping

| Gate | Evidence |
| --- | --- |
| L1-G1 existing milestones preserved | pre-change `make verify`, `make verify-m2-local`, `make verify-m3`, `go vet ./...`, `go test ./...`; final commands rerun in status report |
| L1-G2 real CLI pilot exists | `cmd/animus-news/main.go`, `internal/shortform/pilot` |
| L1-G3 Claude script gate enforced | `scriptReviewPassed`, import validation, tests for missing/mismatched hash |
| L1-G4 visual provider boundary works | `external_command_visual`, fake provider tests, hash/missing-shot tests |
| L1-G5 voice provider boundary works | `external_command_voice`, fake provider tests, missing-config tests |
| L1-G6 subtitle/render path creates release candidate | faster-whisper fake sidecar, FFmpeg/ffprobe test path |
| L1-G7 Claude final QA gate exists | final request/import validation, readiness blocked before final review |
| L1-G8 connector documentation exists | `docs/CONNECTORS.md` |
| L1-G9 final workflow documentation exists | `docs/WORKFLOW_FINAL.md` |
| L1-G10 unsafe publishing not opened | `publish_manifest.json` uses `release_candidate_only`, private visibility, live publishing disabled |
| L1-G11 takeover ready | final takeover commands in status report |

## Risks and limits

- L1 script generation is deterministic and local; it does not claim Claude wrote
  the script.
- L1 provider quality depends on external wrappers configured by the operator.
- L1 faster-whisper support is a sidecar protocol; the real model installation
  remains operator-managed.
- L1 does not add native provider APIs or live publishing.
- L1 creates a release candidate, not a public release.
