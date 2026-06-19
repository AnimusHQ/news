# Production Deployment and Platform Boundary

This guide describes the operational boundary for real provider execution and the target local production platform foundation.

Nothing in this document makes live provider calls by itself. CI and `make verify*` remain safe-by-default: no real secrets, no paid API calls, no social upload, no public publishing.

## 1. Current deployment layers

### Current implemented lane: MVP Docker runtime

`docker-compose.mvp.yml` is the current local operator path for live smoke/full runs.

It provides:

```text
animus-news container
chatterbox service
FFmpeg / ffprobe
Python wrappers
Go runtime/tooling
stable wrapper paths
/workspace/episodes roots
```

It does not provide the full platform stack yet: no Postgres, MinIO, Keycloak, Animus API, Animus worker split, or platform UI.

Runbook: [`docs/runbooks/mvp_docker_runtime.md`](runbooks/mvp_docker_runtime.md).

### Target next lane: PLATFORM-001

`PLATFORM-001` should introduce a Docker-bootstrapped local production platform:

```text
postgres
temporal
temporal-ui
minio
minio-init
keycloak
keycloak-init
animus-api
animus-worker
animus-cli
chatterbox
optional faster-whisper
optional observability collector
```

The target command should be:

```bash
docker compose -f docker-compose.platform.yml up -d
make verify-platform-local
```

## 2. Provider execution architecture

The provider boundary remains typed and provider-agnostic:

```text
Animus core / workflow activity
  -> model router or provider registry
  -> native adapter or sanctioned external-command wrapper
  -> real provider / local service
  -> normalized output
  -> schema validation
  -> hashing / object storage
  -> quality gate
  -> next workflow state
```

Explicitly not the architecture:

```text
Animus core
  -> fake MCP wrapper around every HTTP API
  -> uncontrolled tools
  -> live calls in CI
  -> secrets in repo
  -> direct public publishing
```

MCP stays reserved for true MCP tools such as DaVinci Resolve finishing or Claude Code operator tooling. HTTP/API providers use native adapters or external-command wrappers.

## 3. Provider matrix

| Provider / service | Mode | Runtime config | Current status |
| --- | --- | --- | --- |
| Claude API | Native Go adapter | `ANTHROPIC_API_KEY`, `ANIMUS_CLAUDE_*` | Implemented for review/QA lane |
| OpenAI / ChatGPT | Native adapter planned | `OPENAI_API_KEY`, model router config | Target for multi-agent routing |
| Seedance 2 | External-command wrapper | `SEEDANCE_API_KEY`, `SEEDANCE_BASE_URL`, `SEEDANCE_MODEL` | Implemented boundary |
| Chatterbox | Docker/local HTTP service + wrapper | `CHATTERBOX_*` optional config | Implemented in MVP Docker lane |
| faster-whisper | External-command sidecar | sidecar command/model config | Partial / optional |
| FFmpeg / ffprobe | Container/local binary | container-provided in Docker lane | Implemented |
| Remotion | Future render/UI lane | Node/TS runtime later | Deferred |
| DaVinci Resolve | True MCP finishing lane | operator workstation | Boundary documented, optional |
| Upload-Post / social APIs | Publishing adapter | future provider secrets | Dry-run/release-candidate only |

## 4. Environment and secrets

Environment files are configuration boundaries, not content storage.

Allowed in `.env.mvp.local` or future `.env.platform.local`:

```text
provider keys
provider endpoints
provider model ids
timeouts
live-call gate
local dev infrastructure credentials
non-secret run id
```

Forbidden:

```text
PROMPT
TOPIC
EPISODE_ID
VIDEO_STYLE
CTA
LANGUAGE
DURATION
PLATFORMS
any runtime creative content
```

Runtime content enters via CLI/API request fields only.

Do not print secrets:

```text
cat .env
printenv
env
set
```

Provider adapters and wrappers must redact keys, auth headers, and signed URLs from errors and reports.

## 5. MinIO / S3-compatible object storage target

PLATFORM-001 should make MinIO the local S3-compatible object store for media and large artifacts.

Object storage holds:

```text
raw/redacted provider responses
research packs
scripts and manifests
voiceover WAV
visual shots MP4
subtitle files
OTIO timelines
render outputs
release candidates
QA reports
```

Postgres stores object references, lineage, hashes, schema versions, status, and access metadata.

Initial bucket layout:

```text
animus-artifacts
animus-media
animus-release-candidates
```

The application should never assume local filesystem paths are the production source of truth once PLATFORM-001 lands. Local `episodes/` remains a developer/operator convenience and Docker volume.

## 6. Keycloak authorization target

PLATFORM-001 should add Keycloak as the local identity provider and authorization layer.

Initial roles:

```text
admin
operator
creative_director
reviewer
publisher
viewer
service_worker
service_provider
```

Initial clients:

```text
animus-api
animus-console
animus-cli
animus-worker
```

Hard rules:

- API validates JWTs for protected routes.
- Human QA and release approval require authenticated human roles.
- Workers use service accounts.
- Frontend never receives provider credentials.
- Publishing cannot bypass authenticated release gates.

## 7. OTIO timeline target

OpenTimelineIO becomes the canonical timeline interchange layer for edit structure.

Generated/rendered episodes should eventually store:

```text
timeline.otio
timeline_manifest.json
edit_decision_list.json
shot_graph.yaml
scene_graph.yaml
render_plan.json
```

OTIO references media in MinIO and allows the same episode primitives to scale from short-form clips to scenes, sequences, trailers, cinematic reels, and full productions.

## 8. Live-call gate

Real provider calls require:

```bash
ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1
```

Missing gate means fail-closed before provider spend or network media calls.

No live provider execution is part of CI.

## 9. Failure behavior

All real execution failures must be explicit and classified.

Initial failure classes:

```text
runtime_config_missing
claude_api_failure
openai_api_failure
model_router_failure
seedance_auth_failure
seedance_generation_failure
seedance_download_failure
chatterbox_health_failure
chatterbox_voice_failure
subtitle_failure
ffmpeg_render_failure
otio_timeline_failure
storage_failure
auth_failure
validation_failure
quality_gate_failure
release_gate_blocked
unknown
```

No command may report success when the requested target artifact does not exist and validate.

## 10. Publishing posture

Publishing remains dry-run / release-candidate only:

```text
release_candidate_only: true
live_publishing_enabled: false
public_publish_enabled: false
human_release_required: true
```

Public publishing is a separate milestone after private release, QA, authorization, and incident/correction runbooks are ready.

## 11. Verification policy

CI-safe targets must require no real provider keys, no paid APIs, no local GPU, and no social credentials.

Current targets:

```bash
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
make verify-mvp-docker
go vet ./...
go test ./...
```

Target PLATFORM-001 verification:

```bash
make verify-platform-static
make verify-platform-compose
make verify-auth-config
make verify-storage-config
make verify-agent-registry
make verify-otio
make verify-platform-local
```

Live smoke/full runs should remain explicit operator actions.

## 12. External references

- MinIO container documentation: https://min.io/docs/minio/container/index.html
- Keycloak documentation: https://www.keycloak.org/documentation
- OpenTimelineIO documentation: https://opentimelineio.readthedocs.io/
