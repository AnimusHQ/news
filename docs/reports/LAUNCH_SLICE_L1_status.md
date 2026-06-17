# Launch Slice L1 Status Report

Scope: add a real CLI pilot that can create a prompt-driven
`release_candidate` MP4 through safe real-provider boundaries, plus full
connector and final workflow documentation.

## Verification status

Commands run before implementation:

| Command | Result | Notes |
| --- | --- | --- |
| `git status --porcelain` | pass | clean before edits |
| `git branch --show-current` | pass | `m1-openshorts-integration` |
| `git log --oneline -n 20` | pass | M1-M3 commits present |
| `make verify` | pass | `M3 VERIFY: GREEN` |
| `make verify-m2-local` | pass | M2 local adapter checks and workflow determinism passed |
| `make verify-m3` | pass | M3 provider boundary and replay checks passed |
| `go vet ./...` | pass | no findings |
| `go test ./...` | pass | all packages green |

Commands run after L1 implementation:

| Command | Result | Notes |
| --- | --- | --- |
| `make verify-real-pilot` | pass | L1 pilot CLI/fake external-command provider tests and documentation presence checks passed. |
| `go vet ./...` | pass | no findings |
| `go test ./...` | pass | all packages green |
| `make verify` | pass | existing M3 single-signal verification remains green |
| `make verify-m2-local` | pass | M2 local adapter and workflow determinism checks remain green |
| `make verify-m3` | pass | M3 provider boundary and replay checks remain green |
| `rm -rf build dist episodes/tmp episodes/test-output` | pass | cleanup command completed |
| final `git status --porcelain` | dirty | L1 changes are committed; only `AGENTS.md` remains modified outside the L1 patch set |

Final takeover commands from committed state:

```bash
git status --porcelain
make verify
make verify-real-pilot
go vet ./...
go test ./...
make verify-m2-local
make verify-m3
rm -rf build dist episodes/tmp episodes/test-output
git status --porcelain
```

## Gate results

| Gate | Result | Evidence |
| --- | --- | --- |
| L1-G1 - Existing milestones preserved | Implemented | Baseline and post-edit `make verify`, `make verify-m2-local`, `make verify-m3`, `go vet ./...`, and `go test ./...` passed. |
| L1-G2 - Real CLI pilot exists | Implemented | `animus-news pilot generate-real`, `resume`, `status`, `validate`, review import, visual import, and voice import in `cmd/animus-news/main.go`; logic in `internal/shortform/pilot`. |
| L1-G3 - Claude script gate enforced | Implemented | Missing script review stops before visual generation; import validates `approved_script_hash`; mismatch test rejects. |
| L1-G4 - Real visual provider boundary works | Implemented | `external_command_visual` JSON stdin/stdout boundary; missing config fails closed; fake provider happy path; hash mismatch and missing shot rejected. |
| L1-G5 - Real voice provider boundary works | Implemented | `external_command_voice` JSON boundary; missing config fails closed; fake provider happy path; output hashed in `voiceover_manifest.json`. |
| L1-G6 - Subtitle/render path creates release candidate | Implemented | faster-whisper sidecar protocol and explicit script-timing fallback; FFmpeg renders `dist/<episode-id>-release-candidate.mp4`; ffprobe validates 1080x1920/audio. |
| L1-G7 - Claude final QA gate exists | Implemented | `final_review_request.md` is generated; release readiness remains blocked until valid final review or explicit override. |
| L1-G8 - Full connector documentation exists | Implemented | `docs/CONNECTORS.md`, `docs/CONNECTOR_ROADMAP.md`, `docs/PROVIDER_CAPABILITY_MODEL.md`. |
| L1-G9 - Final workflow documentation exists | Implemented | `docs/WORKFLOW_FINAL.md` includes current launch workflow, future production workflow, DaVinci lane, and Review Room. |
| L1-G10 - No unsafe publishing opened | Implemented | L1 writes `publish_manifest.json` with `mode=release_candidate_only`, `visibility=private`, and `live_publishing_enabled=false`; no live upload command added. |
| L1-G11 - Takeover ready | Implemented except clean-tree caveat | Verification commands passed from the L1 commit; final clean-tree status depends on resolving the unrelated `AGENTS.md` modification noted in handoff. |

## Implemented

- Real pilot CLI surface.
- Manual Claude script and final review artifacts.
- External-command visual and voice provider protocols.
- Faster-whisper sidecar protocol and script-timing fallback.
- FFmpeg release-candidate render path.
- ffprobe technical validation.
- Content hashing and root containment checks for provider outputs.
- Fake external-command provider integration tests.
- `make verify-real-pilot`.
- Full connector/workflow/provider docs.

## Partial

- Real provider wrappers are operator-supplied. L1 provides the safe boundary and
  tests it with fake providers.
- Faster-whisper is represented as a sidecar command protocol; model
  installation and native wrapper implementation remain operator-managed.
- Production QA is L1 technical QA plus Claude final QA, not the full future
  multimodel/human Review Room flow.

## Planned

- Native Seedance/fal/Kling/Runway/Pika adapters.
- Native voice provider adapters.
- Source-grounded research pack and claims graph for every prompt.
- Review Room UI.
- Temporal durable real-pilot workflow.
- Postgres/S3 artifact/state stores.
- Private/scheduled publishing, then public publishing after a separate release
  approval milestone.
- Analytics connectors and feedback loop.

## Security notes

- Missing provider configuration fails closed.
- Mock providers are not used silently by `generate-real`.
- Provider outputs must stay under the episode root.
- Provider hashes are checked against local file hashes.
- Unknown or missing visual shot IDs are rejected.
- Public publishing remains unavailable.
- Generated episode media is ignored by git.
