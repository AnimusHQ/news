# ADR-0006: FFmpeg Local Render Adapter

Status: accepted.

## Context

M1 proved the short-form render contract with deterministic mock providers only.
M2 needs a real local media-processing path without allowing workflow-side
process execution, network calls, provider spend, or direct publish.

## Decision

Add a local FFmpeg render provider under `internal/shortform/providers/render`.
The provider implements the existing `providers.RenderProvider` interface and is
disabled by default. It may run only when explicitly configured with:

- `Enabled: true`;
- an FFmpeg binary path or a discoverable `ffmpeg` on `PATH`;
- a controlled input root;
- a controlled output root;
- an explicit timeout.

The adapter uses `exec.CommandContext` with argv slices and does not invoke a
shell. Inputs are resolved under the configured input root. Outputs are written
under the configured output root. Captions are copied into the controlled render
work directory before the FFmpeg subtitles filter is invoked.

Target render properties:

- `1080x1920`;
- `9:16`;
- `30fps`;
- H.264 video where supported by local FFmpeg;
- AAC audio;
- burned subtitles;
- raw output file hash recorded in `short_render_manifest`.

## Consequences

- Default `make verify` remains safe and offline; no real provider is enabled by
  default.
- Tests cover disabled behavior, missing binary, path containment, traversal,
  timeout behavior, command construction without shell invocation, and local
  fixture normalization when FFmpeg is available.
- Output byte identity may vary across FFmpeg builds, so tests assert manifest
  and gate properties rather than cross-machine media byte identity.
- Workflow code remains deterministic; FFmpeg execution stays behind provider
  and activity boundaries.

