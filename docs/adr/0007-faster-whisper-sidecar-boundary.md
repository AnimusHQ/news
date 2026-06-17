# ADR-0007: faster-whisper Sidecar Boundary

Status: accepted.

## Context

M2 needs a production-quality subtitle-generation boundary without making Python
the backend stack and without implicitly downloading models during verification.
The core system must remain Go and Temporal, and workflow code must not execute
Python or perform filesystem side effects directly.

## Decision

Add a faster-whisper sidecar adapter under
`internal/shortform/providers/subtitles`. The adapter implements
`providers.SubtitleProvider` and is disabled by default.

When explicitly enabled, the adapter requires:

- a configured executable;
- an optional configured script path;
- a controlled input root for voiceover audio;
- a controlled output root for subtitle artifacts;
- a controlled local model root and existing model path;
- an explicit timeout.

The sidecar contract is JSON over stdout. The sidecar writes draft artifacts only
under the configured output directory and returns:

- provider metadata;
- language;
- transcript JSON path;
- SRT path;
- optional ASS path;
- word timestamp, safe-zone, and sync check booleans.

The Go adapter validates that returned files stay under the output directory,
hashes the transcript/caption outputs, verifies transcript JSON syntax, and
normalizes the response into a `subtitle_manifest`.

## Consequences

- No Python code becomes authoritative workflow or artifact logic.
- No model is downloaded implicitly by default tests or verification.
- Missing binary/script/model configuration fails closed.
- Word timestamps are explicit in the request contract and enforced when
  required.
- Safe-zone and sync results remain gate inputs, so a sidecar can produce a
  schema-valid draft that is later blocked by `SubtitleGate`.
- pyannote diarization and Hugging Face gated models are out of scope for M2.

