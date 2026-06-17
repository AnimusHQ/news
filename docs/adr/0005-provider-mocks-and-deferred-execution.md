# ADR-0005: Provider interfaces, mocks, and deferred real execution (compliance notes)

Status: accepted (M1)

## Context

§10 requires Go provider interfaces with deterministic mocks (failure injection) for
every provider; §12 defers all real execution to M2/M3 with specific compliance
constraints.

## Decision

Define six interfaces in `internal/shortform/providers`: `StoryboardImageProvider`,
`VisualVideoProvider`, `VoiceProvider`, `SubtitleProvider`, `RenderProvider`,
`PublishingProvider`. Each has a deterministic mock that:
- emits schema-valid **draft** artifacts (status `draft`, `created_by = model:*` or
  `system`), never `approved`;
- produces fake-but-valid hashes/paths (no real bytes, no network, no spend);
- supports **failure injection** (a configurable `FailMode`) so gate negative tests are
  exercisable.

Real adapters are **not** implemented in M1. They are stubbed/flagged for M2/M3.

## Compliance notes carried forward (record, do not act on in M1)

- **Seedance** (ByteDance image-to-video) has copyright-dispute history and
  geopolitical/provider risk → `VisualVideoProvider` must stay swappable with a fallback
  (e.g. fal.ai primary + alternate). Deferred to M3.
- **Remotion** requires a paid Company License for orgs >3 employees and for
  prompt-to-video use → default renderer is **FFmpeg** (real = `exec ffmpeg` in M2);
  Remotion is an optional flagged adapter only.
- **Subtitles:** faster-whisper (MIT) + WhisperX (BSD-2) are fine commercially; **skip
  pyannote** (gated models / diarization unneeded here).
- **AI disclosure** is law/policy, not preference: TikTok/YouTube/Meta require labeling
  of synthetic audio/video and EU AI Act Art. 50 transparency applies from Aug 2026.
  Enforced as a blocking release gate (`ai_disclosure_required` → release fails unless
  disclosure set). Live sanctions/disclosure enforcement and real C2PA signing are M3.

## Consequences

- M1 spends nothing, uploads nothing, makes no real external calls.
- Negative-path gate tests have a deterministic mechanism to produce invalid artifacts.
- Real-execution risk and licensing constraints are documented before any M2/M3 wiring.
