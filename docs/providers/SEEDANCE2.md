# Seedance 2 — Visual Video Provider

Status: **External-command only** now; native provider is a **Native candidate**
(deferred). No native Seedance API is implemented in this repo.

Seedance is connected through the existing `external_command_visual` boundary
(`ANIMUS_VISUAL_COMMAND`). Default tests use the Go fake visual provider and have
no Seedance dependency.

Studied: https://seedance2.ai/ru/api-docs (2026-06-17). The contract below is as
documented on that page; it has **not** been independently re-verified against
live responses, which is why the native provider is deferred (per the L2
strategy and gate L2-G5).

## Documented API contract

Auth: `Authorization: Bearer <seedance-api-key>`. Base
`https://api.seedance2.ai`. **Asynchronous**: submit a task, then poll or receive
a webhook, then download.

Create task — `POST /v1/videos/generations`:

```json
{
  "model": "seedance-2-0",
  "callback_url": "https://...optional",
  "input": {
    "prompt": "text describing the video",
    "generation_type": "text-to-video",
    "image_urls": [],
    "duration": 5,
    "aspect_ratio": "9:16",
    "resolution": "1080p",
    "seed": -1
  }
}
```

Response: `{"taskId": "...", "credits": 60}`.

Poll — `GET /v1/tasks/:id`:

```json
{
  "id": "...",
  "status": "completed",
  "data": {
    "results": ["https://cdn.seedance2.ai/.../x.mp4"],
    "video_expires_at": "2026-06-13T10:00:00Z",
    "processing_time": 48
  }
}
```

`status` ∈ `queued | generating | completed | failed`. On `failed`, credits are
refunded. Rate limit ~60 req/min; 429 carries `Retry-After`. Duration 4–15s;
resolution 480p/720p/1080p; aspect ratios include `9:16`. Output is MP4 at a CDN
URL that expires.

## How Animus uses it

The pilot sends the external visual request JSON on stdin and expects the
external visual response JSON on stdout (see `docs/REAL_PILOT_V1.md` → External
Visual Protocol). The wrapper, per shot:

1. reads each shot `{shot_id, prompt, negative_prompt, duration_sec, width,
   height, fps}` and `output_dir`;
2. submits a Seedance task, polls until `completed`;
3. downloads the MP4 into `output_dir` as `<shot_id>.mp4`;
4. prints `{schema_version, episode_id, provider, shots:[{shot_id, status,
   output_path, duration_sec, width, height, fps}]}`.

Animus then independently verifies shot IDs, root containment, file existence,
1080x1920/30fps properties, hashes, and `visual_shot_manifest.json` validity.
Missing or unknown shot IDs, hash mismatches, and paths that escape the episode
root are rejected (`TestExternalVisualPathTraversalRejected`).

## Configuration

```bash
export EPISODE_ID=<episode-id>
export EPISODE_DIR=/abs/path/episodes/${EPISODE_ID}
export ANIMUS_VISUAL_COMMAND=/abs/path/scripts/providers/seedance2-visual-wrapper.example.py
export ANIMUS_VISUAL_INPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_OUTPUT_ROOT="$EPISODE_DIR"
export ANIMUS_VISUAL_TIMEOUT=15m   # video generation is slow; raise as needed
export ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
# wrapper-only (never committed):
export SEEDANCE_API_KEY=<seedance-api-key>
export SEEDANCE_BASE_URL=https://api.seedance2.ai
```

Missing `ANIMUS_VISUAL_COMMAND`/roots or the live-call guard → the pilot or
wrapper fails closed.

## Why native is deferred

The job lifecycle, exact image-to-video field names, credit costs, and live auth
behavior are not independently verified here. Per the L2 strategy, the native
provider is implemented only after these are confirmed against live responses. A
`planned_seedance` and a `seedance2_visual_external` entry both appear in the
provider capability registry.

## Security

- Bearer key in the wrapper environment only; never in the repo.
- stdout is the JSON channel only; never print the token.
- Exit non-zero on missing output; never write outside `output_dir`; never
  publish.
- Do not commit generated video.

See the runbook: `docs/runbooks/seedance_visual_wrapper.md` and the sample wrapper
`scripts/providers/seedance2-visual-wrapper.example.py` (non-production).
