# ADR-0011: Provider Capability Registry

Status: accepted.

## Context

M1 and M2 added provider interfaces, mocks, and local adapter boundaries. M3 adds
more optional provider lanes. Future provider selection needs a central safety
description so provider availability is explicit and fail-closed.

The registry must not become an authorization bypass. Artifact validation, gates,
production QA, human approval, and release checks remain authoritative.

## Decision

Add `internal/shortform/providers/capabilities` with a default in-memory
capability registry.

Capability records include:

- provider name and type;
- supported modes;
- enabled/disabled state;
- network, binary, GPU, GUI, MCP, paid API, and consent requirements;
- draft/approval/publish authority flags;
- dry-run support;
- supported artifact types;
- known limitations.

The registry lists:

- `mock`;
- `ffmpeg`;
- `faster_whisper`;
- `upload_post_dry_run`;
- `davinci_resolve_mcp`;
- `omnivoice`;
- `planned_seedance`;
- `planned_elevenlabs`;
- `planned_uploadpost_live`.

Provider selection fails closed when a provider is unknown, disabled, has the
wrong type, claims approval authority, or claims live publish authority.

## Consequences

- Tests can assert provider safety posture independently of provider execution.
- CLI users can inspect provider capabilities via `animus-news
  provider-capabilities`.
- Live-capable future providers remain disabled and cannot publish in M3.
- The registry is descriptive; it does not override gates or approval checks.

