# Final Workflow

This document records both the current Launch Slice L1 workflow and the future
production workflow. It is intentionally artifact-first and gate-first.

## Current launch workflow

```text
prompt
  -> CLI episode workspace
  -> script generation
  -> Claude script review
  -> visual shot requests
  -> external visual generation
  -> voice generation
  -> subtitles
  -> FFmpeg render
  -> Claude final QA
  -> release_candidate
  -> manual or dry-run publish manifest only
```

Current L1 details:

1. `pilot generate-real` creates `topic.yaml`, `episode_manifest.json`,
   `script.md`, `script_manifest.json`, and `audit.log`.
2. The CLI writes `claude_script_review_request.md` and stops.
3. `pilot import-claude-review --kind script` validates Claude JSON and the
   current `script.md` hash.
4. `pilot resume` creates `visual_shot_requests.json`.
5. The visual `external-command` connector generates real shot video files under
   `visual/`.
6. The voice `external-command` connector generates real voiceover audio under
   `audio/`.
7. `faster-whisper` sidecar or explicit `script-timing` fallback creates
   transcript and captions.
8. FFmpeg renders `dist/<episode-id>-release-candidate.mp4`.
9. The CLI writes `final_review_request.md` and stops.
10. `pilot import-claude-review --kind final` validates Claude final QA.
11. `pilot resume` writes `production_qa_report.json` and
    `publish_manifest.json` with live public publishing disabled.

The current launch workflow is a real media-generation path when real wrappers
are configured, but it is not a public publishing workflow.

## Future full production workflow

```text
topic intelligence
  -> source ingestion
  -> research pack
  -> claims graph
  -> editorial brief
  -> script
  -> multimodel verification
  -> human QA
  -> storyboard generation
  -> ChatGPT / image provider storyboard images
  -> Claude storyboard review
  -> Seedance / visual provider video shots
  -> voice provider
  -> subtitles
  -> FFmpeg/Remotion render
  -> optional DaVinci Resolve final studio lane
  -> production QA
  -> release approval
  -> private/scheduled publish
  -> public publish
  -> analytics
  -> feedback loop
```

Future production requirements:

- Every factual claim links to source evidence.
- Research packs and claims graphs precede authoritative scripts.
- Critical scripts pass multimodel verification and human QA.
- Provider outputs remain untrusted until normalized, validated, hashed, and
  reviewed.
- Public publishing is reachable only after production QA, human release
  approval, private/scheduled staging, and final status validation.
- Analytics can influence future topic and format decisions but cannot override
  trust gates.

## Temporal production orchestration

Target durable workflow:

```text
Temporal Workflow
  -> Activity: ingest/normalize sources
  -> Activity: build research pack
  -> Activity: extract claims
  -> Activity: run multimodel council
  -> Wait: human QA signal
  -> Activity: generate storyboard
  -> Activity: generate/import visual assets
  -> Activity: generate voice
  -> Activity: generate subtitles
  -> Activity: render
  -> Activity: production QA
  -> Wait: release approval signal
  -> Activity: private/scheduled publish
  -> Activity: poll status
  -> Activity: import analytics
  -> Activity: feedback report
```

Workflow code must remain deterministic. Provider calls, file I/O, rendering,
publishing, and model calls belong in activities.

## DaVinci final studio lane

```text
Animus package
  -> DaVinci Resolve project
  -> human edit
  -> Resolve export
  -> Animus import
  -> hash/validate
  -> Claude QA
  -> release_candidate
```

Rules for the DaVinci lane:

- The Animus package contains approved script, shot manifests, voiceover,
  captions, source metadata, and render instructions.
- DaVinci Resolve never becomes workflow authority, QA authority, or release
  authority.
- Resolve MCP tools must be allowlisted.
- Exported files return to Animus through import, hash validation, ffprobe
  validation, and Claude final QA.
- Public publishing remains outside the workstation lane.

## Future Review Room

```text
artifact browser
video preview
script view
shot manifest view
voice/subtitle controls
Claude review panel
blocking issues
revision plan
approve/reject controls
release checklist
publish checklist
```

Review Room responsibilities:

- Show the full artifact graph and current gate state.
- Let reviewers inspect source provenance, script, claims, visuals, voice,
  captions, render manifests, QA reports, and publish manifests.
- Import Claude/manual review responses through structured forms.
- Record human decisions as typed artifacts or workflow signals.
- Prevent approval when blocking artifacts are missing or invalid.
- Keep publish controls separate from generation controls.

## Current vs future authority

| Authority | Current L1 | Future production |
| --- | --- | --- |
| Script generation | Deterministic local pilot script | Research-pack-backed writer and model routing |
| Script review | Manual Claude JSON gate | Multimodel council plus human QA |
| Visual generation | External-command wrapper | Native providers plus manual/import lanes |
| Voice generation | External-command wrapper | Provider registry with consent/loudness gates |
| Subtitles | Faster-whisper wrapper or explicit fallback | STT providers plus style/safe-zone validators |
| Render | FFmpeg release candidate | FFmpeg/Remotion/DaVinci lanes |
| Final QA | Manual Claude JSON gate | Multimodel QA plus human production QA |
| Publish | Non-live manifest only | Private/scheduled first, then public after approval |

## Public publishing invariant

No generated output may go directly public. Future public publishing must pass:

```text
release_candidate
  -> production QA approved
  -> human release approval
  -> private/scheduled upload
  -> metadata/status validation
  -> explicit public release action
```
