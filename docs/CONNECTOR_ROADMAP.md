# Connector Roadmap

This roadmap sequences connector work after Launch Slice L1. It avoids turning
L1 into unfinished platform work while preserving the final architecture.

## L1 - Real CLI pilot

Implemented in this slice:

- `pilot generate-real`, `resume`, `status`, `validate`;
- manual Claude script and final review import;
- `external_command_visual`;
- `external_command_voice`;
- `faster_whisper` external sidecar protocol;
- explicit `script_timing_fallback`;
- FFmpeg release-candidate render;
- non-live publish manifest.

Acceptance focus:

- real media file output;
- fail-closed missing provider configuration;
- provider output hashing and root containment;
- final Claude QA required before release-candidate readiness;
- no public publishing.

## L2 - First real provider wrappers

Recommended next milestone:

- Add documented wrapper examples for one visual provider and one voice provider.
- Keep secrets outside the repository.
- Record provider cost/latency metadata in manifests.
- Add fixture-based tests for provider response normalization.
- Add manual operator runbook for first real episode.

Candidate connectors:

- Seedance or another visual provider through `external_command_visual`;
- OmniVoice local sidecar or a paid TTS through `external_command_voice`;
- faster-whisper local wrapper with documented model install path.

## L3 - Source-grounded pilot hardening

Add the missing source-grounding depth for repeatable editorial quality:

- manual source upload;
- documentation site connector;
- GitHub repository/release connector;
- research pack builder for pilot prompts;
- claim extraction and verification connected to the pilot script;
- stronger Claude request packets that include sources and claims.

Gate additions:

- no high-risk unsupported claims;
- source list required for publish metadata;
- license and AI disclosure checks.

## L4 - Review and revision ergonomics

Improve human-in-the-loop flow without opening publishing:

- Review Room UI task pack;
- artifact browser;
- video preview;
- script/shot/voice/subtitle panels;
- Claude review import panel;
- revision plan artifacts;
- approve/reject controls.

The UI must call backend validation; it must not bypass gates.

## L5 - Native provider adapters

Add native adapters only after wrappers have produced real operating evidence:

- native Seedance visual connector;
- native OpenAI/Claude/OpenAI image providers where policy allows;
- native ElevenLabs/Cartesia connectors;
- WhisperX or stronger subtitle alignment;
- visual artifact and watermark detectors.

Each native adapter needs:

- ADR or task pack;
- explicit credentials policy;
- mocked API tests;
- cost/rate-limit behavior;
- no model/provider release authority.

## L6 - Durable production orchestration

Move the pilot into durable workflow execution:

- Temporal workflow for the real pilot lifecycle;
- activities for provider calls, rendering, validation, and imports;
- human signals or updates for Claude/human QA;
- idempotent activity design;
- workflow queries for status.

Do not move side effects into workflow code.

## L7 - Storage and archive

Introduce durable storage after the artifact contracts stabilize:

- Postgres state store;
- S3-compatible object storage;
- artifact archive/export;
- backup/export connector;
- object-level hashes and immutable references.

## L8 - Private/scheduled publishing

Publishing remains later work:

- Upload-Post private/scheduled connector;
- YouTube/TikTok/Instagram native connectors;
- publish status polling;
- release checklist UI;
- human release approval artifact.

Public publishing must remain disabled until private/scheduled staging is proven.

## L9 - Analytics and feedback loop

Add analytics only after safe publishing exists:

- Upload-Post analytics;
- YouTube/TikTok/Instagram analytics;
- retention/hook performance reports;
- episode feedback connector;
- correction workflow triggers.

Analytics cannot override source, safety, or editorial gates.

## Status Discipline

Do not mark a connector `Implemented` unless:

- code exists in the repository;
- default verification or a named make target tests it;
- failure modes are covered;
- security posture is documented;
- it cannot bypass human QA or release approval.

Use `Partial` for boundaries, dry-runs, fake sidecars, or manual import paths.
Use `Planned` for everything else.
