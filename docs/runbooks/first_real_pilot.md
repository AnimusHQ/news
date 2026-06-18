# Runbook — First Real Pilot Video

End-to-end flow to produce the first real prompt-driven `release_candidate` MP4
with real providers. Commands use placeholders only. No public publishing.

Related: `docs/REAL_PILOT_V1.md`, `docs/PRODUCTION_DEPLOYMENT.md`,
`docs/providers/CLAUDE_API.md`, `docs/providers/CHATTERBOX_TTS.md`,
`docs/providers/SEEDANCE2.md`, and the wrapper runbooks.

## 1. Prerequisites

- Go toolchain; `ffmpeg` + `ffprobe` on PATH.
- A Seedance API key (or another visual provider wrapper).
- A running Chatterbox TTS server (or another voice wrapper).
- A faster-whisper sidecar (or use `--subtitle-provider script-timing`).
- For `--claude-review api`: `ANTHROPIC_API_KEY`. Otherwise use manual review.
- Secrets live in your shell/secrets-manager, never in the repo.

## 2. Verify the repository

```bash
git status --porcelain
make verify
make verify-real-pilot
make verify-l2-providers
```

## 3. Configure Claude review

Define runtime content in shell variables. Do not put these in `.env`:

```bash
export EPISODE_ID=<episode-id>
export PROMPT='<operator-provided prompt>'
export LANGUAGE=ru
export DURATION=45s
export PLATFORMS=tiktok,instagram,youtube
export EPISODE_DIR="$(pwd)/episodes/${EPISODE_ID}"
```

API mode (automated):

```bash
export ANTHROPIC_API_KEY=...            # required; fails closed if unset
export ANIMUS_CLAUDE_MODEL=claude-opus-4-8   # optional
```

Or manual mode: skip the key and use `--claude-review manual` (you paste Claude's
JSON and import it).

## 4. Start / configure Chatterbox (voice)

```bash
# start your Chatterbox server, then:
export ANIMUS_VOICE_COMMAND="$(pwd)/scripts/providers/chatterbox-voice-wrapper.example.py"
export ANIMUS_VOICE_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_TIMEOUT=10m
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
export CHATTERBOX_BASE_URL=http://localhost:4123
```

See `docs/runbooks/chatterbox_voice_wrapper.md`.

## 5. Configure Seedance wrapper (visual)

```bash
export ANIMUS_VISUAL_COMMAND="$(pwd)/scripts/providers/seedance2-visual-wrapper.example.py"
export ANIMUS_VISUAL_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_TIMEOUT=15m
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
export SEEDANCE_API_KEY=<seedance-api-key>
```

See `docs/runbooks/seedance_visual_wrapper.md`.

## 6. Configure faster-whisper

```bash
export ANIMUS_FASTER_WHISPER_COMMAND=/abs/path/to/faster-whisper-wrapper
export ANIMUS_FASTER_WHISPER_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_FASTER_WHISPER_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_FASTER_WHISPER_TIMEOUT=10m
# Or skip subtitles transcription with --subtitle-provider script-timing.
```

## 7. Run generate-real

```bash
go run ./cmd/animus-news pilot generate-real \
  --episode-id "$EPISODE_ID" \
  --prompt "$PROMPT" \
  --language "$LANGUAGE" \
  --duration "$DURATION" \
  --platforms "$PLATFORMS" \
  --visual-provider external-command \
  --voice-provider external-command \
  --subtitle-provider faster-whisper \
  --render-provider ffmpeg \
  --claude-review api \
  --out "$EPISODE_DIR"
```

With `--claude-review api`, the script review is requested automatically and the
run stops at the next missing/blocking step. With `manual`, it stops at the
script checkpoint.

## 8. If script review blocks, inspect and revise

```bash
go run ./cmd/animus-news pilot status --episode-dir "$EPISODE_DIR"
cat "$EPISODE_DIR/claude_script_review_response.json"
```

A `fail` verdict or blocking issues stop the run. Revise `script.md`, delete the
stale `claude_script_review_response.json`, and resume to re-review.

(Manual mode: send `claude_script_review_request.md` to Claude, save the JSON,
then `pilot import-claude-review --kind script --file <json>`.)

## 9. Resume after script review

```bash
go run ./cmd/animus-news pilot resume --episode-dir "$EPISODE_DIR"
```

## 10. Generate visuals

The visual wrapper runs during resume. Animus verifies shot IDs, root
containment, 1080x1920/30fps, and hashes; it writes `visual_shot_manifest.json`.

## 11. Generate voice

The voice wrapper runs during resume. Animus hashes the audio and writes
`voiceover_manifest.json` (with `voice_consent_reference` if a cloned voice was
used).

## 12. Generate subtitles

faster-whisper (or the explicit `script-timing` fallback) produces
`subtitles/transcript.json`, `captions.srt`, `captions.ass`, and
`subtitle_manifest.json`.

## 13. Render release candidate

FFmpeg renders `dist/${EPISODE_ID}-release-candidate.mp4`; ffprobe validates the
geometry and that audio + burned captions are present
(`short_render_manifest.json`).

## 14. Run / import final Claude QA

API mode: the final QA review is requested automatically on resume. Manual mode:
send `final_review_request.md` to Claude, save the JSON, then:

```bash
go run ./cmd/animus-news pilot import-claude-review \
  --episode-dir "$EPISODE_DIR" --kind final --file <json>
go run ./cmd/animus-news pilot resume --episode-dir "$EPISODE_DIR"
```

## 15. Validate the release candidate

```bash
go run ./cmd/animus-news pilot status   --episode-dir "$EPISODE_DIR"
go run ./cmd/animus-news pilot validate --episode-dir "$EPISODE_DIR"
```

`validate` checks artifact presence, hashes, containment, render properties,
final Claude QA, production QA, and that live/public publishing stays disabled.

## 16. Troubleshooting

- `missing configuration: ANIMUS_*` → export the provider env and retry
  (fail-closed by design).
- `ANTHROPIC_API_KEY` error in api mode → set the key or use `--claude-review
  manual`.
- `hash mismatch` / `escapes configured root` / `expected N` → the wrapper
  returned a bad/contained-violating/short result; fix the wrapper.
- `approved_script_hash does not match` (manual import) → re-review the current
  `script.md`.
- Render fails → confirm `ffmpeg`/`ffprobe` and audio/caption inputs.

## 17. Cleanup

```bash
rm -rf build dist episodes/tmp episodes/test-output
git status --porcelain
```

Generated episode media is gitignored. Never commit real audio/video or secrets.
