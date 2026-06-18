# Runbook — Seedance Visual Wrapper

How to make Seedance 2 the real visual provider for the pilot through the
`external_command_visual` boundary. No native Seedance API is added; default
tests keep using the Go fake provider.

See also: `docs/providers/SEEDANCE2.md`.

## 1. Get a Seedance API key

Obtain a key from Seedance and keep it in the wrapper environment only — never in
the repository.

## 2. Point Animus at the wrapper

```bash
export EPISODE_ID=<episode-id>
export EPISODE_DIR="$(pwd)/episodes/${EPISODE_ID}"
export ANIMUS_VISUAL_COMMAND="$(pwd)/scripts/providers/seedance2-visual-wrapper.example.py"
export ANIMUS_VISUAL_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_TIMEOUT=15m    # video generation is slow
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1

# wrapper-only env (never committed)
export SEEDANCE_API_KEY=<seedance-api-key>
export SEEDANCE_BASE_URL=https://api.seedance2.ai
export SEEDANCE_MODEL=seedance-2-0
export SEEDANCE_POLL_TIMEOUT=600
chmod +x "$ANIMUS_VISUAL_COMMAND"
```

## 3. Wrapper contract

The wrapper reads the Animus visual request JSON on **stdin** (one entry per
shot with `shot_id`, `prompt`, `negative_prompt`, `duration_sec`, `width`,
`height`, `fps`, plus `output_dir`). For each shot it must:

- submit a Seedance task (`POST /v1/videos/generations`);
- poll `GET /v1/tasks/:id` (or use a webhook) until `completed`;
- download the MP4 into `output_dir` as `<shot_id>.mp4`;
- never write outside `output_dir`; never publish.

It prints the Animus visual response JSON on **stdout**:

```json
{ "schema_version": "1.0", "episode_id": "<episode-id>", "provider": "seedance2",
  "shots": [ { "shot_id": "shot-001", "status": "generated",
    "output_path": "/abs/.../visual/shot-001.mp4",
    "duration_sec": 5, "width": 1080, "height": 1920, "fps": 30 } ] }
```

Rules the wrapper must follow:

- one output file per `shot_id`; report all requested shots;
- report `width=1080`, `height=1920`, `fps=30` (Animus requires exactly this —
  configure Seedance for 9:16 1080p);
- exit non-zero on missing output or generation failure;
- never print the Bearer token (stdout = JSON channel only).

## 4. What Animus enforces (untrusted output)

After the wrapper returns, Animus independently:

- maps every returned shot to a request and rejects unknown/missing shot IDs;
- resolves `output_path` under the episode root (rejects traversal/escape —
  `TestExternalVisualPathTraversalRejected`);
- hashes each file and rejects a provider-reported hash mismatch;
- requires 1080x1920 / 30fps;
- validates `visual_shot_manifest.json` and marks shots `in_review` (operator
  approval still required downstream).

## 5. Verify fail-closed behavior

```bash
unset ANIMUS_VISUAL_COMMAND
go run ./cmd/animus-news pilot resume --episode-dir "$EPISODE_DIR"
# -> error: visual provider external-command missing configuration: ANIMUS_VISUAL_COMMAND
```

## Troubleshooting

- **Timeout** → raise `ANIMUS_VISUAL_TIMEOUT` and `SEEDANCE_POLL_TIMEOUT`.
- **`expected 3` shot mismatch** → the wrapper dropped a shot; return all.
- **`invalid properties`** → Seedance produced non-1080x1920/30; fix the request.
- **Expired CDN URL** → download immediately after `completed`.
- **Spend** → Seedance charges credits on success; test with a sandbox key and a
  single short shot first.
