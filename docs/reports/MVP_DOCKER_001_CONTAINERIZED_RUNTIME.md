# MVP-Docker-001 Containerized Runtime Report

## 1. Branch and Commit Range

- Base branch: `main`
- Base commit after required pull: `5cf67713a3f8dc540d690c61c0d30461ce444106`
- Working branch: `mvp-docker-001-containerized-runtime`
- Commit range: `main..HEAD` after the MVP-Docker-001 commit

## 2. Docker Services Added

- `chatterbox`: Dockerized local Chatterbox-compatible TTS API on port `4123`
  inside the compose network, with `/health` healthcheck.
- `animus-news`: Dockerized Go/Python/FFmpeg CLI runner that executes
  `pilot generate-real` from `/app`.

No public publishing service was added.

## 3. What Operator Still Must Provide

- `.env.mvp.local`, copied from `.env.mvp.example`.
- `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1`.
- `ANTHROPIC_API_KEY`.
- `SEEDANCE_API_KEY`.
- Provider endpoint/model overrides when dashboard values differ from defaults.
- Optional `ANIMUS_LIVE_RUN_ID`.
- Optional `CHATTERBOX_API_KEY`.
- Optional `CHATTERBOX_VOICE` and `CHATTERBOX_VOICE_CONSENT_REFERENCE`.
- Runtime shell variables: `EPISODE_ID`, `PROMPT`, `LANGUAGE`, `DURATION`,
  `PLATFORMS`.

## 4. What Docker Now Provides Automatically

- Go toolchain/runtime for the CLI.
- Python 3 for existing wrapper scripts.
- FFmpeg and ffprobe.
- Stable wrapper paths:
  - `/app/scripts/providers/seedance2-visual-wrapper.example.py`
  - `/app/scripts/providers/chatterbox-voice-wrapper.example.py`
- Stable roots:
  - `ANIMUS_VISUAL_INPUT_ROOT=/workspace/episodes`
  - `ANIMUS_VISUAL_OUTPUT_ROOT=/workspace/episodes`
  - `ANIMUS_VOICE_INPUT_ROOT=/workspace/episodes`
  - `ANIMUS_VOICE_OUTPUT_ROOT=/workspace/episodes`
- Chatterbox URL:
  - `CHATTERBOX_BASE_URL=http://chatterbox:4123`
- Render defaults:
  - `ANIMUS_FFMPEG_BINARY=ffmpeg`
  - `ANIMUS_FFPROBE_BINARY=ffprobe`
  - `ANIMUS_FFMPEG_TIMEOUT=600s`

## 5. Exact Smoke Command

```bash
cp .env.mvp.example .env.mvp.local
# edit .env.mvp.local

docker compose -f docker-compose.mvp.yml --profile mvp up -d chatterbox

EPISODE_ID=mvp-smoke-001 \
PROMPT="runtime prompt here" \
LANGUAGE=ru \
DURATION=10s \
PLATFORMS=tiktok \
docker compose -f docker-compose.mvp.yml --profile mvp run --rm animus-news
```

Expected:

```text
episodes/mvp-smoke-001/dist/mvp-smoke-001-release-candidate.mp4
```

## 6. Exact Full MVP Command

```bash
EPISODE_ID=mvp-001 \
PROMPT="runtime prompt here" \
LANGUAGE=ru \
DURATION=45s \
PLATFORMS=tiktok,instagram,youtube \
docker compose -f docker-compose.mvp.yml --profile mvp run --rm animus-news
```

Expected:

```text
episodes/mvp-001/dist/mvp-001-release-candidate.mp4
```

## 7. Whether a Live MP4 Was Produced

`live_mp4_produced: no`

## 8. If Not Produced, Exact Blocker

`docker_runtime_execution: not_run`

Reason:

- No live Anthropic or Seedance credentials were provided.
- The task forbids fake-provider fallback for live MVP success.
- Docker daemon availability was checked and `docker compose config --quiet`
  passed with a temporary empty `.env.mvp.local`, but containers were not built
  or run.

If Docker itself is unavailable during final verification, the final handoff
must also state:

```text
docker_runtime_execution: not_run
reason: Docker unavailable in this environment
```

## 9. Verification Commands

Baseline on clean `main` before edits:

```bash
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
go vet ./...
go test ./...
```

MVP-Docker-001 verification:

```bash
docker version
cp .env.mvp.example .env.mvp.local
docker compose -f docker-compose.mvp.yml --profile mvp config --quiet
rm -f .env.mvp.local
make verify-mvp-docker
git status --porcelain
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
go vet ./...
go test ./...
```

## 10. Security Notes

- `.env.mvp.local` is gitignored and must not be committed.
- `.env.mvp.example` contains only operator-editable provider configuration,
  live-call gate, optional non-secret run id, and optional Chatterbox
  voice/consent metadata.
- Runtime content stays in shell variables, not env files.
- `.dockerignore` excludes env files, generated episodes, build outputs, Git
  metadata, and media files from Docker build context.
- The `animus-news` entrypoint fails closed on empty `PROMPT`, missing provider
  keys, or missing live-call gate.
- The Chatterbox launcher can require bearer auth when `CHATTERBOX_API_KEY` is
  set and does not log request bodies itself.
- No Docker service can publish publicly.
- Static verification scans Docker/runtime files for high-risk secret patterns.

## 11. Known Limitations

- Chatterbox model startup can be slow on first run because model assets are
  resolved by the upstream Chatterbox API image build/runtime.
- CPU is the committed default for portability; NVIDIA acceleration requires an
  operator-local override documented in the runbook.
- The Chatterbox container build requires network access to install upstream
  Chatterbox API dependencies.
- Docker image build and live provider execution are not part of
  `make verify-mvp-docker`.
- Faster-whisper is intentionally not required for MVP Docker when
  `--subtitle-provider script-timing` is used.

## 12. Next Step

Run the smoke command on an operator machine with Docker, `.env.mvp.local`, live
Anthropic and Seedance credentials, and the selected Chatterbox voice/consent
metadata. If it blocks, keep the generated episode workspace and CLI stderr as
the takeover artifact.

## Classification

Implemented:

- Static Docker MVP runtime files, env template, fail-closed entrypoint,
  Chatterbox launcher, runbook, report, ledger, and static verification target.

Partial:

- Live Docker execution and MP4 production were not run in this coding
  environment.

Planned:

- Operator-machine smoke run with live credentials.
