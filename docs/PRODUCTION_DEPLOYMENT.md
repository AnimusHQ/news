# Production Deployment — Provider Layer

How real provider execution is configured and run **outside** this coding
session: runtime secrets, API keys, local services, and infrastructure. This is
the operational boundary for the L2 provider layer.

This guide does not perform any deployment. Nothing here runs in CI. No real
secret, network call, paid API, local model, or live deployment is required by
`make verify*` or the test suite.

## Architecture (the production path)

```
Animus core (Go, typed, provider-agnostic)
  -> typed provider adapter  (native, e.g. Claude API review)
     or sanctioned external-command boundary  (e.g. Seedance/Chatterbox wrapper)
  -> real HTTP provider / local service
  -> returned media or structured JSON
  -> Animus validation, hashing, manifests, gates
  -> release candidate
```

Explicitly **not** the architecture:

```
Animus core -> fake MCP wrapper around every HTTP API -> uncontrolled tools
            -> live calls in CI -> secrets in repo -> public publishing
```

Providers are HTTP/API services, not MCP servers, so they are reached by native
adapters or external-command wrappers — never wrapped in MCP to fit a pattern.
MCP stays reserved for true MCP tools (DaVinci Resolve finishing, Claude Code
operator tooling).

## Provider matrix

| Provider | Mode | Secrets / services (runtime) | Status |
| --- | --- | --- | --- |
| Claude API (review/QA) | Native Go adapter | `ANTHROPIC_API_KEY` | Implemented |
| Chatterbox (voice) | External-command wrapper | local Chatterbox server (`CHATTERBOX_BASE_URL`) | External-command only |
| Seedance 2 (visual) | External-command wrapper | `SEEDANCE_API_KEY` (in wrapper env) | External-command only |
| faster-whisper (subtitles) | External-command sidecar | local model/binary | Partial (sidecar) |
| FFmpeg (render) | Local binary | `ffmpeg`/`ffprobe` | Implemented |
| OpenAI (image) | Native (planned) | `OPENAI_API_KEY` | Planned (L3) |

## Where secrets come from at runtime

Secrets are supplied by the operator's environment at run time — never committed:

- **Local / dev:** a gitignored `.env` (template: `.env.example`) or shell exports.
- **Server / CI-adjacent:** a secrets manager (e.g. AWS Secrets Manager, GCP
  Secret Manager, Vault, Kubernetes Secrets) injected as environment variables
  into the process that runs `animus-news`.
- Animus reads only environment variables. It never reads secrets from the repo
  and never writes them to artifacts or logs (the Claude provider redacts its key
  from all errors).

Required runtime variables are listed in `.env.example`. Each provider fails
closed when its variables are missing.

## Per-provider deployment

### Claude API (native)

```bash
export ANTHROPIC_API_KEY=...                 # from your secrets manager
export ANIMUS_CLAUDE_MODEL=claude-opus-4-8   # optional
export ANIMUS_CLAUDE_TIMEOUT=60s             # optional
# then: --claude-review api
```

### Chatterbox (local HTTP service, via wrapper)

1. Deploy a Chatterbox server (GPU recommended); confirm `/health`.
2. Point `ANIMUS_VOICE_COMMAND` at the wrapper and set `CHATTERBOX_BASE_URL`.
3. See `docs/runbooks/chatterbox_voice_wrapper.md`.

### Seedance (cloud API, via wrapper)

1. Provision a Seedance API key in your secrets manager.
2. Point `ANIMUS_VISUAL_COMMAND` at the wrapper; export `SEEDANCE_API_KEY` in the
   wrapper environment only.
3. See `docs/runbooks/seedance_visual_wrapper.md`.

### faster-whisper (sidecar)

Install the model/binary and expose the sidecar via
`ANIMUS_FASTER_WHISPER_COMMAND`, or use `--subtitle-provider script-timing`.

### OpenAI (planned)

Deferred to L3 (no storyboard stage yet; official API contract unverified here).
When implemented it will read `OPENAI_API_KEY` and follow the Claude provider's
fail-closed + redaction patterns.

## Failure behavior (fail-closed)

- Missing provider config → the pilot stops with a clear error naming the missing
  variable. `generate-real` never silently falls back to mocks.
- Native adapters: missing key → fail closed; 4xx / validation failures → no
  retry; 429 / 5xx / transport → bounded retry; keys redacted from errors.
- External-command wrappers: must exit non-zero on provider error; Animus
  enforces a timeout, output hashing, root containment, and schema validation,
  and rejects on mismatch / missing output / path escape.

## No-live-CI policy

- `make verify`, `make verify-real-pilot`, `make verify-m2-local`,
  `make verify-m3`, and `make verify-l2-providers` use mocks, fake HTTP servers,
  and fake external-command providers only.
- CI must not require real secrets, real network calls, paid APIs, local models,
  or live deployment.
- This coding session performs no live provider calls, no spend, no
  infrastructure deployment, and handles no real credentials.

## Publishing

Publishing remains dry-run / `release_candidate_only` with `visibility=private`
and `live_publishing_enabled=false`. Public publishing is a separate, later
milestone and is intentionally absent here.
