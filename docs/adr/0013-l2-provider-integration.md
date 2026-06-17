# ADR-0013: L2 Provider Integration Strategy

Status: accepted.

## Context

L2 maps five providers into Animus: Claude API, Chatterbox TTS, Seedance 2,
OpenAI, and Claude Code MCP. The providers we need are HTTP/API services or local
servers — not MCP servers. The project keeps a typed, provider-agnostic core with
code-enforced gates, no repo secrets, no spend, and no live calls in
verification.

## Decision

Integrate each provider in its correct mode; do not force a uniform mechanism.

- **Native typed adapter** where the API contract is verified and fake-server
  tested. Claude API is implemented this way (ADR-0012) because structured review
  is central to the pipeline.
- **Sanctioned external-command boundary** for fast, real execution before native
  adapters are stable. Seedance (visual) and Chatterbox (voice) ride the existing
  `external_command_visual` / `external_command_voice` contracts via documented
  wrappers. Their outputs are path-contained, hashed, schema-validated, and
  gated. Native adapters are added only after each contract is verified with
  fake-server tests.
- **Do not wrap plain HTTP APIs in MCP.** Seedance, Chatterbox, OpenAI, and Claude
  API are HTTP/API providers and must not go behind MCP unless they expose a real
  MCP server. MCP stays reserved for true MCP tools — the DaVinci Resolve
  finishing lane (already bounded) and Claude Code operator tooling. Claude Code
  MCP is an operator/developer connector, never a runtime pilot provider.
- **OpenAI is deferred.** The official API docs could not be verified in this
  environment (HTTP 403); per the anti-hallucination rule, no native provider is
  built on an unverified contract. `openai_image` is registered as `planned`; the
  storyboard stage and native provider are planned for L3 (no schema change
  forced now).
- **Capability registry** records every provider's honest posture; no entry may
  claim approval or live-publish authority (enforced by `Registry.Validate`).
- **Operational boundary:** runtime secrets, API keys, local services, and
  infrastructure are supplied by the operator outside this session
  (`docs/PRODUCTION_DEPLOYMENT.md`). No live calls, spend, deployment, or real
  credentials occur in this session or in CI.

## Consequences

- The first real video can be produced through native review + external-command
  media providers, all behind gates, with no new dependency and no repo secrets.
- Statuses stay honest: Implemented (Claude review), External-command only
  (Seedance, Chatterbox), Planned (OpenAI native, Claude Code MCP runtime use).
- `make verify-l2-providers` exercises docs presence, capability entries, fake
  HTTP provider tests, and fake external-command tests — no paid providers.
- Public publishing remains absent; publishing stays dry-run / release-candidate
  only.
