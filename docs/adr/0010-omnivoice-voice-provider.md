# ADR-0010: OmniVoice Voice Provider Boundary

Status: accepted.

## Context

M3 adds an optional local/multilingual voice generation lane. OmniVoice may help
future offline or self-hosted TTS experiments, multilingual narration, and
provider-independence work.

OmniVoice is not the default production voice provider and must not self-approve
generated audio, bypass voice QA, or clone/reference voices without explicit
consent metadata.

## Decision

Add a disabled-by-default OmniVoice provider under
`internal/shortform/providers/voice/omnivoice`.

The provider implements `providers.VoiceProvider` and supports:

- `disabled`;
- `dry_run`;
- `local_sidecar`.

Both enabled modes require configured local binary and local model paths. Default
verification uses fake local fixtures only; no model is downloaded and no network
is required.

Reference voice or voice-prompt use requires consent metadata when
`RequireConsent` or request-level consent is enabled:

- `voice_consent_reference`;
- reference audio explicitly allowed;
- reference audio hash recorded when reference audio is used.

OmniVoice emits draft `voiceover_manifest` artifacts with provider metadata,
language, optional voice prompt/consent references, output path/hash, duration,
format, and sample rate.

## Consequences

- No voice cloning or prompt-conditioned voice path can proceed without recorded
  consent metadata.
- Generated voice remains draft and requires downstream approval/gates.
- Workflow code remains provider-agnostic and does not execute OmniVoice.
- Real model execution is a future optional local integration, not a default
  verification requirement.

