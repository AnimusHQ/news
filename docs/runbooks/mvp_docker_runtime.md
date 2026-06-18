# Runbook - MVP Docker Runtime

This runbook describes the MVP-Docker-001 local operator flow. The runtime keeps
content inputs in shell variables and keeps provider credentials/configuration in
`.env.mvp.local`.

## 1. Prepare local environment

```bash
cp .env.mvp.example .env.mvp.local
```

Edit only cloud keys, provider choices/endpoints/models, the live-call gate, an
optional non-secret run id, and optional Chatterbox voice/consent metadata.

Do not put prompt, topic, episode id, duration, platforms, style, CTA, local
absolute paths, FFmpeg paths, or wrapper paths in `.env.mvp.local`.

Minimum live values:

```bash
ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
ANTHROPIC_API_KEY=...
SEEDANCE_API_KEY=...
```

## 2. Start Chatterbox

```bash
docker compose -f docker-compose.mvp.yml --profile mvp up -d chatterbox
```

The `chatterbox` service exposes the Chatterbox-compatible contract used by the
existing wrapper:

```text
GET /health
POST /v1/audio/speech
```

If `CHATTERBOX_API_KEY` is set in `.env.mvp.local`, the local service requires
`Authorization: Bearer <value>` for speech requests. `/health` remains available
for the Docker healthcheck.

The container packages the upstream `travisvn/chatterbox-tts-api` FastAPI server,
which documents `/health`, `/v1/audio/speech`, and Docker operation at
`https://chatterboxtts.com/docs`.

## 3. Smoke run

```bash
EPISODE_ID=mvp-smoke-001 \
PROMPT="runtime prompt here" \
LANGUAGE=ru \
DURATION=10s \
PLATFORMS=tiktok \
docker compose -f docker-compose.mvp.yml --profile mvp run --rm animus-news
```

Expected output:

```text
episodes/mvp-smoke-001/dist/mvp-smoke-001-release-candidate.mp4
```

or a precise fail-closed blocker from the CLI, wrapper, Chatterbox healthcheck,
Seedance wrapper, Claude API review, or FFmpeg render step.

## 4. Full MVP run

```bash
EPISODE_ID=mvp-001 \
PROMPT="runtime prompt here" \
LANGUAGE=ru \
DURATION=45s \
PLATFORMS=tiktok,instagram,youtube \
docker compose -f docker-compose.mvp.yml --profile mvp run --rm animus-news
```

Expected output:

```text
episodes/mvp-001/dist/mvp-001-release-candidate.mp4
```

The generation command does not publish publicly. The pilot writes a
`publish_manifest.json` in `release_candidate_only` mode with live publishing
disabled.

## 5. What Docker Provides

- Go toolchain/runtime for `go run ./cmd/animus-news`.
- Python 3 for provider wrapper scripts.
- FFmpeg and ffprobe on `PATH`.
- Stable wrapper paths inside `/app/scripts/providers`.
- Stable input/output roots at `/workspace/episodes`.
- Chatterbox base URL as `http://chatterbox:4123`.
- Docker healthcheck gating before `animus-news` runs.

## 6. Operator-Provided Values

- `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1`.
- `ANTHROPIC_API_KEY`.
- Claude model/base URL/timeout/token overrides when needed.
- `SEEDANCE_API_KEY`.
- Seedance base URL/model/poll timeout when dashboard values differ from repo
  defaults.
- Optional `ANIMUS_LIVE_RUN_ID`.
- Optional `CHATTERBOX_API_KEY`.
- Optional `CHATTERBOX_VOICE` and `CHATTERBOX_VOICE_CONSENT_REFERENCE`.
- Runtime shell variables: `EPISODE_ID`, `PROMPT`, `LANGUAGE`, `DURATION`,
  `PLATFORMS`.

## 7. Optional NVIDIA Runtime

The committed compose file defaults to portable CPU execution. For NVIDIA
workstations, keep the override local and uncommitted, for example:

```yaml
# docker-compose.mvp.nvidia.local.yml
services:
  chatterbox:
    profiles: ["mvp", "nvidia"]
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
```

Run with:

```bash
docker compose \
  -f docker-compose.mvp.yml \
  -f docker-compose.mvp.nvidia.local.yml \
  --profile mvp \
  --profile nvidia \
  up -d chatterbox
```

## 8. Verification

Static verification does not require Docker, live keys, network calls, or paid
provider calls:

```bash
make verify-mvp-docker
```

Full takeover verification:

```bash
git status --porcelain
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
make verify-mvp-docker
go vet ./...
go test ./...
```

## 9. Common Blockers

- Missing `.env.mvp.local`: copy `.env.mvp.example` first.
- Missing `PROMPT`: the entrypoint exits before any provider call.
- `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS` is not `1`: Claude, Seedance, and
  Chatterbox wrappers fail closed.
- Missing `ANTHROPIC_API_KEY`: Claude API review fails closed.
- Missing `SEEDANCE_API_KEY`: visual generation wrapper fails closed.
- Chatterbox healthcheck does not become healthy: inspect
  `docker compose -f docker-compose.mvp.yml --profile mvp logs chatterbox`.
- Provider output outside `/workspace/episodes`: Animus rejects it during
  artifact normalization.
- FFmpeg output missing audio or not `1080x1920`: render validation fails.
