# Final Workflow

This document records the current MVP Docker workflow, the next live smoke/full run, and the target production workflow for Animus Media Engine.

The workflow is artifact-first, gate-first, provider-agnostic, and release-safe.

## 1. Current MVP Docker workflow

Current implemented local operator path:

```text
runtime shell input
  -> Docker Compose MVP runtime
  -> animus-news CLI
  -> episode workspace
  -> script generation / import gates
  -> Claude review / QA lane
  -> visual shot requests
  -> Seedance external-command wrapper
  -> Chatterbox voice wrapper / Docker service
  -> subtitles or script-timing fallback
  -> FFmpeg render
  -> final QA
  -> release_candidate MP4
  -> dry-run publish manifest only
```

Expected smoke output:

```text
episodes/mvp-smoke-001/dist/mvp-smoke-001-release-candidate.mp4
```

Expected full MVP output:

```text
episodes/mvp-001/dist/mvp-001-release-candidate.mp4
```

Rules:

- Runtime content comes from shell variables or CLI/API fields, never `.env`.
- `.env.mvp.local` contains provider/service config only.
- Real provider calls require `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1`.
- No mock/fake provider can count as live success.
- Public publishing is disabled.

## 2. Target production workflow

Production begins with strategy, not with a script.

```text
Creative/Product Task
  -> Claude/OpenAI discovery if needed
  -> Platform Analysis
  -> Audience Strategy
  -> Competitive / Reference Analysis
  -> Format Strategy
  -> Creative Direction
  -> Success Criteria
  -> Research Pack
  -> Claims / Source State
  -> Editorial Brief
  -> Script Draft
  -> Script QA
  -> Verification / Fact Check
  -> Multi-agent Council Review
  -> Human Editorial QA
  -> Storyboard / Shot Plan
  -> OTIO Timeline Plan
  -> Visual Generation
  -> Voice Generation
  -> Subtitles / Captions
  -> Timeline Assembly
  -> Render
  -> Production QA
  -> Human Release Review
  -> Release Candidate
  -> Private/Dry-run Publish later
  -> Public Publish later
  -> Analytics Feedback later
```

This is intentionally not:

```text
prompt -> script -> video -> publish
```

## 3. Creative intelligence stages

Before script generation, the system must create and validate:

```text
platform_analysis_report.json
audience_strategy.json
competitive_reference_report.json
format_strategy.json
creative_direction.md
success_criteria.json
quality_benchmark.json
```

Purpose:

- optimize for the target platform;
- avoid generic content;
- define success criteria before production;
- keep topic, style, pacing, platform, voice and CTA as runtime creative decisions;
- support quality-first revision loops.

## 4. Multi-agent and model routing

The workflow is multi-agent. Claude is not the only possible model. ChatGPT/OpenAI and future providers must be routable per stage.

Canonical role structure:

```text
AgentRole
  -> TaskContract
  -> AllowedModels
  -> DefaultModel
  -> SelectionPolicy
  -> OutputSchema
  -> ValidationGate
```

Example stages where model selection matters:

| Stage | Possible policy |
| --- | --- |
| Platform Analysis | Claude or ChatGPT, single expert |
| Competitive Reference Analysis | Claude + web/research tools, or OpenAI for second opinion |
| Creative Direction | high-quality creative model, human override allowed |
| Script Draft | writer model selected by language/genre |
| Script Critique | model different from writer, or council |
| Fact Check | verifier model plus deterministic source checks |
| Production QA | council or strict reviewer policy |
| Release Review | human authority required |

Hard rules:

- model choice is recorded in artifact metadata;
- critical artifacts are not self-approved by the same model that created them;
- council reports preserve dissent;
- fallback models cannot silently change authority;
- model output is schema-validated before the next stage.

## 5. Target Temporal orchestration

Durable orchestration target:

```text
Temporal Workflow
  -> Activity: create or load project/episode
  -> Activity: run platform analysis
  -> Activity: run audience strategy
  -> Activity: run competitive/reference analysis
  -> Activity: produce creative direction
  -> Activity: build research pack
  -> Activity: extract claims/source state
  -> Activity: draft script
  -> Activity: run verification
  -> Activity: run multi-agent council
  -> Wait: human editorial QA signal
  -> Activity: generate storyboard / shot plan
  -> Activity: generate OTIO timeline plan
  -> Activity: generate/import visual assets
  -> Activity: generate voice
  -> Activity: generate subtitles
  -> Activity: assemble timeline
  -> Activity: render
  -> Activity: production QA
  -> Wait: human release approval signal
  -> Activity: create release candidate package
  -> Activity: private/dry-run publish later
  -> Activity: analytics import later
```

Workflow code must remain deterministic. Provider calls, file I/O, MinIO/S3 access, rendering, publishing, model calls, and Keycloak/service-account operations belong in activities or external services, not workflow decision code.

## 6. Artifact graph

Current short-form artifacts remain valid:

```text
topic.yaml
research_pack.json
claims.json
editorial_brief.md
script.md
verification_report.json
multimodel_approval_report.json
human_qa_report.json
storyboard.yaml
asset_manifest.json
render_manifest.json
production_qa_report.json
publish_manifest.json
analytics_report.json
```

New target artifacts for the current platform slice:

```text
platform_analysis_report.json
audience_strategy.json
competitive_reference_report.json
format_strategy.json
creative_direction.md
success_criteria.json
quality_benchmark.json
timeline.otio
timeline_manifest.json
edit_decision_list.json
shot_graph.yaml
scene_graph.yaml
render_plan.json
strategy_qa_report.json
script_qa_report.json
visual_qa_report.json
voice_qa_report.json
timeline_qa_report.json
render_qa_report.json
release_review_report.json
```

Every structured artifact must include schema version, episode/project ids, provenance, agent/model metadata, validation status, and content hash where applicable.

## 7. OTIO timeline lane

OpenTimelineIO becomes the canonical edit/timeline interchange artifact.

Short-form path:

```text
storyboard.yaml
  -> shot_graph.yaml
  -> generated visual/audio assets
  -> timeline.otio
  -> render_plan.json
  -> FFmpeg/Remotion render
  -> release_candidate.mp4
```

Cinematic path:

```text
Project
  -> Sequence
  -> Scene
  -> Shot
  -> Take
  -> Asset
  -> OTIO timeline
  -> finishing/export
```

OTIO references media stored in MinIO/S3-compatible storage. It is not itself the media store.

## 8. Platform services in workflow context

| Service | Workflow role |
| --- | --- |
| Postgres | metadata, lineage, status, artifact index |
| MinIO | media and large artifact object storage |
| Keycloak | user/service identity and authorization |
| Temporal | durable workflow execution and human wait states |
| Animus API | project/episode/artifact/review/release interface |
| Animus worker | activities and provider orchestration |
| Chatterbox | local voice service |
| faster-whisper | optional subtitles/STT sidecar |

## 9. Human gates

Human authority is required for:

```text
editorial approval
high-risk claim approval
creative direction override
production QA override
release approval
public publishing approval later
```

No generated output may go directly public.

Future public publishing must pass:

```text
release_candidate
  -> production QA approved
  -> authenticated human release approval
  -> private/scheduled upload
  -> metadata/status validation
  -> explicit public release action
```

## 10. DaVinci final studio lane

Optional finishing path:

```text
Animus package
  -> DaVinci Resolve project
  -> human edit
  -> Resolve export
  -> Animus import
  -> hash/ffprobe validation
  -> production QA
  -> release_candidate
```

DaVinci Resolve never becomes workflow authority, QA authority, or release authority. Resolve MCP tools must be allowlisted.

## 11. Current vs target authority

| Authority | Current MVP Docker | Target platform |
| --- | --- | --- |
| Strategy | runtime prompt only | platform/audience/creative intelligence artifacts |
| Script | current pilot path | research-backed writer via model router |
| Review | Claude review lane | multi-agent council plus human QA |
| Visuals | external-command Seedance wrapper | provider registry plus visual QA |
| Voice | Chatterbox wrapper/service | provider registry plus consent/loudness gates |
| Subtitles | script timing or sidecar | STT providers plus timing/safe-zone QA |
| Timeline | render manifest | OTIO timeline + timeline QA |
| Render | FFmpeg release candidate | FFmpeg/Remotion/DaVinci lanes |
| Storage | local episodes directory | MinIO objects + Postgres metadata |
| Auth | local operator only | Keycloak-authenticated API/UI/service accounts |
| Publish | non-live manifest only | gated private/public release process |
