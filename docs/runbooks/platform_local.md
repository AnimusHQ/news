# Runbook - Local Platform Foundation

> Status: target runbook for PLATFORM-001. The currently implemented local runtime is still [`mvp_docker_runtime.md`](mvp_docker_runtime.md). This document defines the next production-foundation operator flow.

## 1. Goal

Bring up the local Animus Media Engine platform without manual host dependency setup:

```bash
docker compose -f docker-compose.platform.yml up -d
make verify-platform-local
```

The platform should provide:

```text
Postgres
Temporal
Temporal UI
MinIO
Keycloak
Animus API
Animus Worker
Animus CLI
Chatterbox
optional faster-whisper
optional observability collector
```

## 2. Environment file

Expected local file:

```bash
cp .env.platform.example .env.platform.local
```

Allowed values:

```text
POSTGRES_*
TEMPORAL_*
MINIO_*
KEYCLOAK_*
ANIMUS_*
ANTHROPIC_*
OPENAI_*
SEEDANCE_*
CHATTERBOX_*
```

Forbidden values:

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

Runtime creative content must enter via CLI/API request fields, not env files.

## 3. Bootstrap responsibilities

The platform compose should automatically bootstrap:

```text
Postgres schema
Temporal namespace / worker connection
MinIO buckets
MinIO service accounts and policies
Keycloak realm
Keycloak clients
Keycloak roles
Keycloak local users for development
Keycloak service accounts
Animus default agent/model registry
healthchecks
```

Manual UI setup in MinIO or Keycloak should not be required for local development.

## 4. MinIO readiness

Expected buckets:

```text
animus-artifacts
animus-media
animus-release-candidates
```

Expected object families:

```text
projects/<project-id>/
episodes/<episode-id>/artifacts/
episodes/<episode-id>/shots/
episodes/<episode-id>/audio/
episodes/<episode-id>/subtitles/
episodes/<episode-id>/renders/
release-candidates/<episode-id>/
```

Readiness checks should verify:

```text
MinIO health
bucket existence
service account access
write/read/delete test object in dev bucket
no public anonymous access unless explicitly configured for local dev
```

## 5. Keycloak readiness

Expected realm:

```text
animus-local
```

Expected roles:

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

Expected clients:

```text
animus-api
animus-console
animus-cli
animus-worker
```

Readiness checks should verify:

```text
Keycloak health
realm exists
clients exist
roles exist
local operator user exists
service account token can be obtained
Animus API rejects unauthenticated protected requests
Animus API accepts valid token with expected role
```

## 6. API and worker readiness

Expected checks:

```text
animus-api /health is healthy
animus-api /ready confirms database/object-store/auth dependencies
animus-worker connects to Temporal
animus-worker registers expected activities
Temporal UI shows worker polling
```

## 7. Model router readiness

The local platform should load an agent/model registry containing at least:

```text
Claude / Anthropic lane
OpenAI / ChatGPT lane
mock lane for CI/dev
human reviewer lane
```

Readiness checks should verify:

```text
agent roles load
allowed models load
per-step default model exists
manual override validation works
model choice is recorded in a dry-run artifact
no model can self-approve critical output
```

## 8. OTIO readiness

The platform should provide a typed timeline boundary:

```text
timeline.otio
timeline_manifest.json
edit_decision_list.json
shot_graph.yaml
scene_graph.yaml
render_plan.json
```

Initial readiness may use deterministic stub assets, but the output must be typed, validated, and stored through the storage abstraction.

Checks:

```text
OTIO package/build dependency available where needed
timeline artifact generated or stubbed through typed interface
timeline references media object refs, not untracked host paths
render plan validates before render
```

## 9. Local smoke target

After PLATFORM-001, a local platform smoke should be possible:

```bash
EPISODE_ID=platform-smoke-001 \
PROMPT="runtime creative task here" \
LANGUAGE=ru \
DURATION=10s \
PLATFORMS=tiktok \
docker compose -f docker-compose.platform.yml run --rm animus-cli generate
```

Expected result:

```text
release candidate MP4 or precise classified blocker
Postgres metadata rows
MinIO artifacts/media objects
agent/model metadata
OTIO timeline artifact
authenticated API visibility
```

## 10. Verification targets

Target make commands:

```bash
make verify-platform-static
make verify-platform-compose
make verify-auth-config
make verify-storage-config
make verify-agent-registry
make verify-otio
make verify-platform-local
```

CI-safe verification must not require live provider keys or paid API calls.

## 11. Failure classes

Initial platform failure classes:

```text
platform_compose_invalid
postgres_unhealthy
temporal_unhealthy
minio_unhealthy
minio_bootstrap_failure
keycloak_unhealthy
keycloak_bootstrap_failure
auth_token_failure
api_unhealthy
worker_unhealthy
storage_failure
model_registry_failure
model_router_failure
otio_timeline_failure
provider_live_gate_missing
validation_failure
quality_gate_failure
unknown
```

Every failure should include stage, sanitized error, suspected cause, minimal fix, and whether retry is safe.

## 12. Security rules

Never commit:

```text
.env.platform.local
.env.mvp.local
provider keys
Keycloak local admin secrets
MinIO root credentials
generated media
Docker volumes
signed URLs
raw auth headers
```

Never log:

```text
cat .env
printenv
env
set
provider auth headers
signed URLs
```

## 13. Relationship to MVP Docker

`docker-compose.mvp.yml` remains the minimal live media proof path.

`docker-compose.platform.yml` is the next production foundation path.

MVP Docker proves:

```text
runtime prompt -> provider boundaries -> local render -> release-candidate MP4
```

Platform foundation proves:

```text
runtime task -> authenticated platform -> durable workflow -> object storage -> metadata -> timeline -> agent/model routing -> release candidate
```
