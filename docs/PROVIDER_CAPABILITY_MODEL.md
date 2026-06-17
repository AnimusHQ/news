# Provider Capability Model

Provider capability metadata is descriptive evidence for routing and safety. It
is not an authorization bypass. Artifact validation, gates, human QA, and
release approval remain authoritative.

## Capability Record

Every provider or connector should define:

```yaml
provider_id: string
connector_id: string
category: source | reasoning | image | visual_video | voice | subtitle | render | publishing | storage | qa | analytics | operator
status: implemented | partial | planned | disabled
execution_modes:
  - manual
  - mock
  - dry_run
  - local_sidecar
  - local_mcp
  - api
  - live
default_enabled: false
requires_network: false
requires_credentials: false
requires_local_binary_or_model: false
requires_gui_or_session: false
input_artifact_kinds: []
output_artifact_kinds: []
can_create_draft_artifacts: false
can_create_release_candidate: false
can_approve_artifacts: false
can_publish_publicly: false
supports_dry_run: false
supports_fake_fixture: false
cost_tracking_required: false
privacy_tier: public | internal-approved | restricted
known_failure_modes: []
security_notes: []
validation_gates: []
test_targets: []
```

Required invariants:

- `can_approve_artifacts` must be false for model/media providers.
- `can_publish_publicly` must be false unless a dedicated publishing milestone
  implements private/scheduled/public gates.
- `default_enabled` should be false for any provider with network, credential,
  local model, GUI, spend, or public side effects.
- Provider health must never override artifact validation.
- Capability metadata cannot mark model output as final truth.

## Current L1 Capabilities

| provider_id | connector_id | status | modes | default | network | credentials | local binary/model | GUI/session | outputs | authority |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `external_command_visual` | `external_command_visual` | Implemented | local_sidecar/api wrapper | false | wrapper-dependent | wrapper-owned | wrapper true | false | visual files, `visual_shot_manifest` | no approval, no publish |
| `external_command_voice` | `external_command_voice` | Implemented | local_sidecar/api wrapper | false | wrapper-dependent | wrapper-owned | wrapper true | false | audio file, `voiceover_manifest` | no approval, no publish |
| `faster_whisper_l1_sidecar` | `faster_whisper` | Partial | local_sidecar | false | false | false | command/model true | false | transcript, SRT, ASS, `subtitle_manifest` | no approval, no publish |
| `script_timing_fallback` | `script_timing_fallback` | Implemented | local | false | false | false | false | false | transcript, SRT, ASS, `subtitle_manifest` | explicit fallback only |
| `ffmpeg_l1_render` | `ffmpeg_render` | Implemented | local binary | false | false | false | ffmpeg/ffprobe true | false | `release_candidate` MP4, `short_render_manifest` | no final approval, no publish |
| `claude_manual_script_review` | `claude_manual_review` | Implemented | manual | selected by CLI | operator external | not stored | false | operator | review JSON | gate evidence only |
| `claude_manual_final_qa` | `claude_final_qa` | Implemented | manual | selected by CLI | operator external | not stored | false | operator | review JSON | gate evidence only |
| `local_filesystem_artifact_store` | `local_filesystem_artifact_store` | Implemented | local | true | false | false | filesystem | false | artifact files | storage only |
| `upload_post_dry_run` | `upload_post_dry_run` | Implemented | dry_run | false | false | false | false | false | dry-run publish manifest | no live publish |

## Current L2 Capabilities

L2 maps real providers to their correct mode: native typed adapters where the API
contract is verified and fake-server tested, sanctioned external-command wrappers
where fast real execution is needed first. HTTP APIs are not wrapped in MCP.

| provider_id | category | status | modes | default | network | credentials | authority |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `claude_api_review` | review/qa | Implemented | api | enabled (key-gated) | true | `ANTHROPIC_API_KEY` | review JSON only; no approval, no publish |
| `chatterbox_tts_external` | voice | External-command only | external_command | false | local server | wrapper-owned | draft voiceover; no approval, no publish |
| `seedance2_visual_external` | visual_video | External-command only | external_command | false | true | wrapper-owned | draft visuals; no approval, no publish |
| `openai_image` | image | Planned | planned (L3) | false | true | `OPENAI_API_KEY` | none yet (no storyboard stage) |
| `claude_code_mcp_operator` | operator | Planned (operator connector) | operator | false | server-dependent | operator-owned | none; not a runtime pilot provider |

The native Claude review provider lives in
`internal/shortform/providers/review/claude`; the registry is in
`internal/shortform/providers/capabilities`. L2 evidence:

```text
make verify-l2-providers
```

## Routing Rules

Provider selection should evaluate:

1. Required artifact type and lifecycle stage.
2. Required modality: text, image, video, audio, subtitles, render, publishing.
3. Privacy tier and data classification.
4. Current provider status and health.
5. Execution mode allowed by the task pack.
6. Credentials and local binary/model availability.
7. Cost and timeout policy.
8. Gate requirements before and after execution.

The router must fail closed if no provider matches. It must not silently fall
back from a real provider request to a mock provider.

## Status Semantics

`Implemented`:

- code path exists;
- default or named verification target covers happy and blocked paths;
- no live side effect occurs unless explicitly scoped;
- docs describe configuration and failure modes.

`Partial`:

- boundary or dry-run exists;
- fake sidecar or manual import exists;
- real production execution is not fully proven.

`Planned`:

- documented future connector;
- no implementation claim.

`Disabled`:

- code may exist but current config or policy prevents use.

## Failure Behavior

Provider failures are classified as:

- missing configuration;
- unavailable binary/model;
- invalid request;
- timeout;
- provider error;
- invalid response JSON;
- missing output;
- output outside root;
- hash mismatch;
- schema mismatch;
- policy block.

All classes fail closed. Retry behavior belongs in a workflow/activity layer and
must be bounded and idempotent.

## Audit Requirements

Provider execution should audit:

- episode ID;
- connector ID;
- provider ID and version/model where available;
- request artifact hash;
- response artifact hash;
- output file hashes;
- local command path or API provider name;
- timeout/cost metadata where available;
- operator identity for manual imports;
- gate result after normalization.

Logs must not contain secrets, tokens, credentials, private voice references, or
raw sensitive prompts.

## Test Strategy

Each implemented or partial provider needs:

- disabled/missing-config fail-closed test;
- invalid response test;
- output containment test where files are involved;
- hash mismatch test where files are involved;
- happy path with fake provider or fixture;
- gate integration test proving the provider cannot self-approve or publish.

L1 evidence lives in:

```text
make verify-real-pilot
go test ./internal/shortform/pilot ./cmd/animus-news
```
