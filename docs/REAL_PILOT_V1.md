# Real Pilot V1 - Launch Slice L1

Launch Slice L1 adds a real CLI pilot path that can produce a prompt-driven
short-form `release_candidate` MP4 when real provider wrappers are configured.
It does not publish publicly.

## Command

```bash
go run ./cmd/animus-news pilot generate-real \
  --episode-id animus-oss-001 \
  --prompt "Объясни, почему open-source разработчикам нужна устойчивая экосистема" \
  --language ru \
  --duration 45s \
  --platforms tiktok,instagram,youtube \
  --visual-provider external-command \
  --voice-provider external-command \
  --subtitle-provider faster-whisper \
  --render-provider ffmpeg \
  --claude-review manual \
  --out ./episodes/animus-oss-001
```

Final successful output:

```text
episodes/animus-oss-001/dist/animus-oss-001-release-candidate.mp4
```

The output state and file name use `release_candidate`. The CLI does not call
this output a draft.

## Workflow

Current L1 workflow:

```text
prompt
  -> episode workspace
  -> deterministic local script
  -> manual Claude script review request/response
  -> visual shot requests
  -> external-command visual provider
  -> external-command voice provider
  -> faster-whisper sidecar or explicit script-timing fallback
  -> FFmpeg render
  -> manual Claude final QA request/response
  -> production QA report
  -> non-live publish manifest
  -> release_candidate MP4
```

Manual Claude mode is intentional. The CLI writes request files, validates JSON
responses, and resumes only after the response artifacts pass gates.

## Episode Layout

L1 writes:

```text
topic.yaml
research_pack.json
episode_manifest.json
script.md
script_manifest.json
claude_script_review_request.md
claude_script_review_response.json
visual_shot_requests.json
visual_shot_manifest.json
visual/shot-001.mp4
visual/shot-002.mp4
visual/shot-003.mp4
audio/voiceover.wav
voiceover_manifest.json
subtitles/transcript.json
subtitles/captions.srt
subtitles/captions.ass
subtitle_manifest.json
dist/<episode-id>-release-candidate.mp4
short_render_manifest.json
final_review_request.md
final_review_response.json
production_qa_report.json
publish_manifest.json
audit.log
```

Generated episode directories are ignored by `.gitignore` except for the
committed sample episode.

## Provider Configuration

Visual external command:

```bash
export ANIMUS_VISUAL_COMMAND=/path/to/visual-provider-wrapper
export ANIMUS_VISUAL_INPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_VISUAL_OUTPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_VISUAL_TIMEOUT=10m
```

Voice external command:

```bash
export ANIMUS_VOICE_COMMAND=/path/to/voice-provider-wrapper
export ANIMUS_VOICE_INPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_VOICE_OUTPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_VOICE_TIMEOUT=10m
```

Faster-whisper sidecar command:

```bash
export ANIMUS_FASTER_WHISPER_COMMAND=/path/to/faster-whisper-wrapper
export ANIMUS_FASTER_WHISPER_INPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_FASTER_WHISPER_OUTPUT_ROOT=/absolute/path/to/episodes/animus-oss-001
export ANIMUS_FASTER_WHISPER_TIMEOUT=10m
```

FFmpeg:

```bash
export ANIMUS_FFMPEG_BINARY=/usr/bin/ffmpeg
export ANIMUS_FFPROBE_BINARY=/usr/bin/ffprobe
export ANIMUS_FFMPEG_TIMEOUT=10m
```

If required configuration is missing, the CLI fails closed and reports the
missing environment variables.

## External Visual Protocol

The command receives JSON on stdin and returns JSON on stdout.

Input:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "provider": "external-command",
  "shots": [
    {
      "shot_id": "shot-001",
      "scene_id": "scene-001",
      "duration_sec": 5,
      "prompt": "Vertical cinematic 9:16 video...",
      "negative_prompt": "watermark, distorted text...",
      "width": 1080,
      "height": 1920,
      "fps": 30
    }
  ],
  "output_dir": "./episodes/animus-oss-001/visual"
}
```

Output:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "provider": "seedance-wrapper",
  "shots": [
    {
      "shot_id": "shot-001",
      "status": "generated",
      "output_path": "./episodes/animus-oss-001/visual/shot-001.mp4",
      "output_hash": "sha256:...",
      "duration_sec": 5,
      "width": 1080,
      "height": 1920,
      "fps": 30
    }
  ]
}
```

The CLI independently verifies shot IDs, root containment, file existence,
properties, hashes, and manifest schema validity.

## External Voice Protocol

Input:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "language": "ru",
  "text": "Voiceover text...",
  "output_dir": "./episodes/animus-oss-001/audio"
}
```

Output:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "provider": "omnivoice-wrapper",
  "output_path": "./episodes/animus-oss-001/audio/voiceover.wav",
  "output_hash": "sha256:...",
  "duration_sec": 44.7,
  "sample_rate": 48000,
  "voice_consent_reference": "optional-consent-record"
}
```

## Claude Review Workflow

1. Run `pilot generate-real`.
2. Send `claude_script_review_request.md` to Claude.
3. Save Claude JSON as a local file.
4. Import it:

```bash
go run ./cmd/animus-news pilot import-claude-review \
  --episode-dir ./episodes/animus-oss-001 \
  --kind script \
  --file ./claude_script_review_response.json
```

5. Resume:

```bash
go run ./cmd/animus-news pilot resume --episode-dir ./episodes/animus-oss-001
```

6. After render, send `final_review_request.md` to Claude.
7. Import final review:

```bash
go run ./cmd/animus-news pilot import-claude-review \
  --episode-dir ./episodes/animus-oss-001 \
  --kind final \
  --file ./final_review_response.json
```

8. Resume and validate.

Claude script approval requires `verdict=pass`,
`can_continue_to_visual_generation=true`, no blocking issues, and an
`approved_script_hash` matching current `script.md`.

Claude final review requires `verdict=pass`, `can_release_candidate=true`, and
no blocking issues, unless an explicit operator override with reason is recorded.

## Validation

```bash
go run ./cmd/animus-news pilot status --episode-dir ./episodes/animus-oss-001
go run ./cmd/animus-news pilot validate --episode-dir ./episodes/animus-oss-001
make verify-real-pilot
```

`pilot validate` checks artifact presence, content hashes, provider output
containment, render properties, final Claude QA, production QA, and that live
public publishing remains disabled.

## Security Notes

- No provider output is trusted automatically.
- External outputs must remain under the episode root.
- Provider-returned hashes are checked against local file hashes.
- Missing shot IDs and unknown shot IDs are rejected.
- Missing provider configuration fails closed.
- Public publishing is not implemented in L1.
- `publish_manifest.json` uses `mode=release_candidate_only`,
  `visibility=private`, and `live_publishing_enabled=false`.
