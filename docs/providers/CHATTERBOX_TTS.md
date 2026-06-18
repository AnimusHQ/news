# Chatterbox TTS — Voice Provider

Status: **External-command only** (no native Go provider).

Chatterbox is connected through the existing `external_command_voice` boundary
(`ANIMUS_VOICE_COMMAND`). Animus never calls Chatterbox directly; an operator
wrapper does. Default tests use the Go fake voice provider and have no Chatterbox
dependency.

Studied: https://chatterboxtts.com/docs (2026-06-17).

## What Chatterbox is

A self-hosted FastAPI TTS server (OpenAI-compatible speech API). Default base URL
`http://localhost:4123`, **no API key** by default. GPU recommended; model files
are downloaded locally. ~22 languages including Russian and English. Supports
voice cloning from uploaded samples.

| Endpoint | Method | Purpose |
| --- | --- | --- |
| `/v1/audio/speech` | POST | generate complete audio (binary WAV) |
| `/v1/audio/speech/stream` | POST | streaming generation (SSE) |
| `/v1/audio/speech/upload` | POST | generate with an uploaded voice sample |
| `/voices` | GET/POST | manage the voice library |
| `/languages` | GET | list languages |
| `/health` | GET | health check |

Key `/v1/audio/speech` body fields (verbatim): `input` (text), `voice`,
`exaggeration` (0.25–2.0), `cfg_weight` (0.0–1.0), `temperature` (0.05–5.0),
`stream_format` (`audio` or `sse`).

> Field names and the default sample rate may vary by server version. The wrapper
> owns the live call and should be verified against the running server.

## How Animus uses it

The pilot sends the external voice request JSON on stdin and expects the external
voice response JSON on stdout (see `docs/REAL_PILOT_V1.md` → External Voice
Protocol). The wrapper:

1. reads `{schema_version, episode_id, language, text, output_dir}`;
2. calls Chatterbox (`POST /v1/audio/speech`, optionally `/upload` for cloning);
3. writes a WAV into `output_dir`;
4. prints `{schema_version, episode_id, provider, output_path, duration_sec,
   sample_rate, voice_consent_reference?}`.

Animus then independently hashes the file, checks root containment, and validates
the resulting `voiceover_manifest.json`.

## Configuration

```bash
export EPISODE_ID=<episode-id>
export EPISODE_DIR=/abs/path/episodes/${EPISODE_ID}
export ANIMUS_VOICE_COMMAND=/abs/path/scripts/providers/chatterbox-voice-wrapper.example.py
export ANIMUS_VOICE_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_TIMEOUT=10m
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
# wrapper-only (never committed):
export CHATTERBOX_BASE_URL=http://localhost:4123
```

Missing `ANIMUS_VOICE_COMMAND`/roots or the live-call guard → the pilot or
wrapper fails closed.

## Voice cloning & consent

If a reference/cloned voice is used, the wrapper must set
`voice_consent_reference` to an auditable consent record id. Animus carries it
into `voiceover_manifest.json` (`voice_consent_reference`). Do not clone a voice
without consent.

## Security

- No secrets in the repo. If a deployment adds auth, keep the token in the wrapper
  environment.
- stdout is the JSON channel only — never print request text, tokens, or audio to
  stdout; send diagnostics to stderr.
- The wrapper must exit non-zero on any provider error.
- Do not commit generated audio.

See the runbook: `docs/runbooks/chatterbox_voice_wrapper.md` and the sample
wrapper `scripts/providers/chatterbox-voice-wrapper.example.py` (non-production).
