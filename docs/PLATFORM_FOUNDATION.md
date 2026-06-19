# Animus Media Engine Platform Foundation

> Status: target architecture and current-slice roadmap. This document captures the current product and engineering north star after CFG-001 and MVP-Docker-001. It is documentation only; implementation status remains code-backed by tests, milestone reports, and runbooks.

## 1. What we are building

`AnimusHQ/news` is evolving from the Animus News short-form pipeline into the **Animus Media Engine**: a production-grade AI media operating system for strategy-driven, platform-aware, high-quality video production.

It is not a prompt-to-video toy, not a single-provider wrapper, and not a content farm. It is a typed, auditable production system that turns a runtime creative task into a validated media artifact set:

```text
Task / campaign goal / creative objective
  -> platform analysis
  -> audience strategy
  -> competitive/reference analysis
  -> creative direction
  -> research and source state
  -> script
  -> verification
  -> multi-agent critique
  -> human QA
  -> storyboard and shot plan
  -> OTIO timeline
  -> visual generation
  -> voice generation
  -> subtitles
  -> render
  -> production QA
  -> release candidate
  -> human release gate
  -> publishing and analytics later
```

The immediate visible output remains:

```text
episodes/<episode-id>/dist/<episode-id>-release-candidate.mp4
```

The production output is broader: typed artifacts, media objects, provenance, timelines, agent/model decisions, QA reports, release decisions, and audit metadata.

## 2. Current completed checkpoints

- **M1-M3**: typed contracts, validators, gates, mock providers, Temporal skeleton, local execution boundaries, DaVinci/OmniVoice boundaries, provider registry and replay hardening.
- **L1-L2**: real CLI pilot framework, Claude API review lane, external-command Seedance/Chatterbox lanes, FFmpeg render path, provider docs and runbooks.
- **CFG-001**: content-agnostic provider configuration. Runtime content is not stored in `.env`; `.env` is provider/service config only.
- **MVP-Docker-001**: containerized MVP runtime. Docker Compose starts local Chatterbox, provides stable wrapper paths, FFmpeg/ffprobe, Python, Go runtime, and `/workspace/episodes` roots.

These are necessary checkpoints, not the final platform.

## 3. Immediate next checkpoints

### MVP-001 Live Smoke

From a clean clone, `.env.mvp.local`, and Docker Compose, produce:

```text
episodes/mvp-smoke-001/dist/mvp-smoke-001-release-candidate.mp4
```

If it fails, record an exact blocker. Do not fake success.

### MVP-001 Full Live

After smoke succeeds, produce:

```text
episodes/mvp-001/dist/mvp-001-release-candidate.mp4
```

### PLATFORM-001: Local Production Platform Foundation

Bootstrap the local production platform through Docker Compose:

```text
Postgres
Temporal
MinIO / S3-compatible object storage
Keycloak
Animus API
Animus worker
Animus CLI
Chatterbox
optional faster-whisper
optional observability collector
```

The platform must start without manual host dependency setup.

## 4. Platform services and responsibilities

### Docker

Docker is the local production bootstrap mechanism. It must start all local infrastructure and provide stable internal paths. Operators should not manually install FFmpeg, launch Chatterbox, configure wrapper paths, create buckets, create Keycloak realms, or initialize Temporal by hand.

### Postgres

Postgres stores canonical metadata and lineage, not large media files.

Core domains:

```text
projects
episodes
workflow_runs
agent_runs
provider_attempts
artifacts
object_refs
quality_reports
human_reviews
release_decisions
timeline_versions
publish_attempts
analytics_events
```

### MinIO / S3-compatible object storage

MinIO is the local S3-compatible artifact and media store. Large media and artifact payloads should be stored as objects; Postgres stores object references, hashes, schema versions, and status.

Initial bucket layout:

```text
animus-artifacts/projects/<project-id>/
animus-artifacts/episodes/<episode-id>/artifacts/
animus-media/episodes/<episode-id>/shots/
animus-media/episodes/<episode-id>/audio/
animus-media/episodes/<episode-id>/subtitles/
animus-media/episodes/<episode-id>/renders/
animus-release-candidates/<episode-id>/
```

Every stored object should carry metadata such as `episode_id`, `project_id`, `artifact_type`, `schema_version`, `content_hash`, `provider`, `created_by_agent`, `workflow_id`, `run_id`, and `validation_status`.

### Keycloak

Keycloak is the identity and authorization layer for API, console, CLI/service accounts, human QA, and release approval.

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

Hard rules:

- no anonymous release approval;
- no public publishing without authenticated human release gate;
- no provider secrets in frontend;
- no production API bypassing authorization;
- service workers use service accounts, not shared user credentials.

### Temporal

Temporal remains the durable orchestration engine. Provider calls, storage I/O, rendering, model calls, and publishing happen in activities. Workflows remain deterministic and replay-safe.

Core future workflows:

```text
EpisodeProductionWorkflow
PlatformAnalysisWorkflow
ResearchWorkflow
CreativeBriefWorkflow
ScriptWorkflow
VerificationWorkflow
StoryboardWorkflow
MediaGenerationWorkflow
TimelineAssemblyWorkflow
RenderWorkflow
ProductionQAWorkflow
ReleaseWorkflow
PublishWorkflow
```

### OTIO / OpenTimelineIO

OpenTimelineIO becomes the canonical editorial timeline interchange layer.

For each rendered episode, the system should generate and store:

```text
timeline.otio
timeline_manifest.json
edit_decision_list.json
shot_graph.yaml
scene_graph.yaml
render_plan.json
```

OTIO references media objects in MinIO. It is the bridge between generated shots, edit decisions, FFmpeg/Remotion rendering, DaVinci Resolve finishing, and future cinematic workflows.

## 5. Creative intelligence layer

A production episode does not start with a script. It starts with strategy.

Before script/storyboard/render, the system must run:

```text
PlatformAnalysisStage
AudienceAnalysisStage
CompetitiveReferenceStage
FormatStrategyStage
CreativeDirectionStage
SuccessCriteriaStage
```

Required artifacts:

```text
platform_analysis_report.json
audience_strategy.json
competitive_reference_report.json
format_strategy.json
creative_direction.md
success_criteria.json
quality_benchmark.json
```

This keeps the system topic-agnostic and style-adaptive. No repo code or env file may hard-code one niche, topic, platform, tone, voice, pacing pattern, or visual style.

## 6. Multi-agent and model routing

Claude is an important provider, not the only authority. The platform must support multiple agent roles and per-step model selection across Claude, ChatGPT/OpenAI, local models later, specialized providers, and humans.

Initial agent roles:

```text
Platform Analyst
Audience Strategist
Researcher
Fact Checker
Creative Director
Scriptwriter
Script Critic
Storyboard Director
Visual Prompt Engineer
Continuity Supervisor
Voice Director
Editor
Production QA
Release Reviewer
Analytics Analyst
```

Every agent role should have:

```text
AgentRole
TaskContract
AllowedModels
DefaultModel
SelectionPolicy
OutputSchema
ValidationGate
```

The model router must support default routing, manual per-run overrides, cost/latency-aware policy, quality-aware routing, fallback routing, council mode, A/B model comparison, and human-selected models for specific stages.

Hard rules:

- model choice is recorded in artifact metadata;
- model output is schema-validated;
- one model cannot self-approve its own critical output;
- council reports preserve dissent;
- fallback cannot silently change final authority.

## 7. Quality-first production model

The target is highest-quality output, not fastest output.

Every major stage should support revision loops:

```text
strategy revision
script revision
claim correction
storyboard revision
visual shot regeneration
voice retry
subtitle correction
timeline edit revision
render revision
QA rejection
human review rejection
```

Quality artifacts:

```text
strategy_qa_report.json
script_qa_report.json
visual_qa_report.json
voice_qa_report.json
timeline_qa_report.json
render_qa_report.json
production_qa_report.json
release_review_report.json
```

Each report should include score, blocking issues, non-blocking issues, revision requirement, recommended fix, reviewing agent/model, and optional human decision.

## 8. Cinematic future compatibility

Short videos are the first production unit, not the final ambition. The architecture must scale toward cinematic and film production:

```text
Project
  -> Campaign / Film / Series
  -> Episode
  -> Sequence
  -> Scene
  -> Shot
  -> Take
  -> Asset
  -> Timeline
  -> Render
```

Future artifacts include:

```text
project_bible.yaml
world_bible.yaml
character_bible.yaml
location_bible.yaml
cinematic_style_bible.yaml
continuity_map.json
scene_graph.yaml
shot_graph.yaml
asset_library_manifest.json
timeline.otio
edit_decision_list.json
sound_design_plan.yaml
music_direction.md
color_pipeline_manifest.json
```

The short-form pipeline must use the same primitives so it can grow into trailers, ads, cinematic reels, series, and full production pictures.

## 9. Security and release gates

The system must preserve existing invariants:

- no claim without source;
- no script without research/source-state;
- no generated output self-approval;
- no render without verification;
- no provider output trusted without validation;
- no public publish without human release approval;
- no secrets in repo/logs/artifacts;
- no fake-live success;
- no direct upload path;
- no provider lock-in.

Public publishing remains out of scope until an explicit release milestone enables it.

## 10. PLATFORM-001 definition of done

PLATFORM-001 is complete when:

1. Docker Compose boots the local platform: Postgres, Temporal, MinIO, Keycloak, Animus API, Animus worker, Animus CLI, Chatterbox, and optional faster-whisper.
2. MinIO buckets and service accounts are initialized automatically.
3. Keycloak realm, clients, roles, service accounts, and local test users are initialized automatically.
4. Postgres schema exists for projects, episodes, workflows, agents, artifacts, object refs, reviews, releases, and timelines.
5. Animus API validates Keycloak JWTs.
6. Animus worker can run Temporal activities.
7. Artifact metadata persists to Postgres and large artifacts/media persist to MinIO.
8. Model registry supports at least Claude and OpenAI/ChatGPT lanes.
9. Per-step model selection is documented and recorded in artifacts.
10. OTIO timeline artifact is generated or explicitly stubbed behind a typed interface for rendered episodes.
11. `make verify-platform-static` and `make verify-platform-compose` exist.
12. Live provider execution remains opt-in and fail-closed.

## 11. References

- MinIO container documentation: https://min.io/docs/minio/container/index.html
- Keycloak documentation: https://www.keycloak.org/documentation
- OpenTimelineIO documentation: https://opentimelineio.readthedocs.io/
