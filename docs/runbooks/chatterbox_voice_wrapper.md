# Runbook — Chatterbox Voice Wrapper

How to make Chatterbox TTS the real voice provider for the pilot through the
`external_command_voice` boundary. No native dependency is added; default tests
keep using the Go fake provider.

See also: `docs/providers/CHATTERBOX_TTS.md`.

## 1. Run a Chatterbox server

Follow the Chatterbox docs to run the self-hosted server (Docker or local
Python). GPU is recommended. Confirm it is up:

```bash
curl -s "${CHATTERBOX_BASE_URL:-http://localhost:4123}/health"
```

## 2. Point Animus at the wrapper

```bash
export EPISODE_ID=<episode-id>
export EPISODE_DIR="$(pwd)/episodes/${EPISODE_ID}"
export ANIMUS_VOICE_COMMAND="$(pwd)/scripts/providers/chatterbox-voice-wrapper.example.py"
export ANIMUS_VOICE_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VOICE_TIMEOUT=10m
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1

# wrapper-only env (never committed)
export CHATTERBOX_BASE_URL=http://localhost:4123
export CHATTERBOX_VOICE=<voice-name>            # optional
export CHATTERBOX_VOICE_CONSENT_REFERENCE=<id>  # required if cloning a voice
chmod +x "$ANIMUS_VOICE_COMMAND"
```

## 3. Wrapper contract

The wrapper reads the Animus voice request JSON on **stdin**:

```json
{ "schema_version": "1.0", "episode_id": "<episode-id>",
  "language": "ru", "text": "Voiceover text...",
  "output_dir": "/abs/.../episodes/<episode-id>/audio" }
```

It writes a WAV into `output_dir` and prints the Animus voice response JSON on
**stdout**:

```json
{ "schema_version": "1.0", "episode_id": "<episode-id>", "provider": "chatterbox",
  "output_path": "/abs/.../audio/voiceover.wav", "duration_sec": 44.7,
  "sample_rate": 24000, "voice_consent_reference": "consent-001" }
```

Rules the wrapper must follow:

- read the Animus request JSON from stdin;
- call Chatterbox per its docs;
- write the audio into the requested `output_dir`;
- return the Animus voice response JSON on stdout (stdout = JSON channel only);
- never print secrets or audio to stdout; diagnostics go to stderr;
- set `voice_consent_reference` whenever a reference/cloned voice is used;
- exit non-zero on any provider error.

## 4. Quick local test (no Chatterbox)

A stub proves the boundary without a server. Create a script that reads the
request and writes a tiny WAV (e.g. via `ffmpeg -f lavfi -i sine=... out.wav`) and
prints the response JSON, then point `ANIMUS_VOICE_COMMAND` at it. Animus will
hash the file, check containment, and validate `voiceover_manifest.json`. This is
exactly what the default Go fake provider does in
`internal/shortform/pilot/pipeline_test.go`.

## 5. Verify fail-closed behavior

```bash
unset ANIMUS_VOICE_COMMAND
go run ./cmd/animus-news pilot resume --episode-dir "$EPISODE_DIR"
# -> error: voice provider external-command missing configuration: ANIMUS_VOICE_COMMAND
```

## Troubleshooting

- **Empty/short audio** → check Chatterbox `input` and server logs (stderr).
- **`hash mismatch`** → the wrapper changed the file after reporting; let Animus
  hash, or report the post-write hash.
- **`escapes configured root`** → write only inside `output_dir`.
- **consent error** → set `CHATTERBOX_VOICE_CONSENT_REFERENCE` for cloned voices.
