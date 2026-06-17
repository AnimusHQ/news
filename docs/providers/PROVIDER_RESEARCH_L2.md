# Provider Research — L2

This document records the L2 provider study and the integration decision for each
provider. It is the source of truth for what was verified, what was assumed, and
what remains unknown. Statuses are honest: a provider is only `Implemented` when
code **and** tests exist in the repository.

Status vocabulary: `Implemented`, `Partial`, `Planned`, `Blocked`,
`External-command only`, `Native candidate`.

Documentation was studied on 2026-06-17. Where a documentation page could not be
reached or rendered, that is stated explicitly and no request/response contract
was invented.

| Provider | Category | Decision | Status |
| --- | --- | --- | --- |
| Claude API | Review / QA | Native Go review provider (stdlib HTTP) behind `--claude-review api` | Implemented |
| Chatterbox TTS | Voice | External-command wrapper via `ANIMUS_VOICE_COMMAND` | External-command only |
| Seedance 2 | Visual video | External-command wrapper via `ANIMUS_VISUAL_COMMAND`; native deferred | External-command only / Native candidate |
| OpenAI API | Storyboard / reference image | Documented; official docs unverifiable here; defer native to L3 | Planned / Native candidate |
| Claude Code MCP | Operator / developer connector | Operator tooling only; not a runtime pilot provider | Planned (operator connector) |

---

## 1. Chatterbox TTS

- **provider_id:** `chatterbox_tts_external`
- **category:** voice / TTS
- **intended role in Animus:** real voiceover generation for the pilot, behind the
  existing `external_command_voice` boundary.
- **current integration status:** External-command only (no native Go provider).
- **recommended first integration mode:** `ANIMUS_VOICE_COMMAND` external-command
  wrapper.
- **native provider now?** No.
- **external-command wrapper now?** Yes.
- **requires credentials:** No (self-hosted server has no API key by default).
- **requires local server:** Yes (self-hosted FastAPI server, default
  `http://localhost:4123`).
- **requires local model:** Yes (model files downloaded locally; GPU recommended).
- **requires network:** Only to the local server (no public egress required).
- **requires GUI/session:** No.
- **input artifacts:** Animus external voice request JSON (`ExternalVoiceInput`):
  `schema_version`, `episode_id`, `language`, `text`, `output_dir`.
- **output artifacts:** WAV audio file under the episode `audio/` dir + Animus
  external voice response JSON (`ExternalVoiceResponse`): `provider`,
  `output_path`, `output_hash` (optional), `duration_sec`, `sample_rate`,
  `voice_consent_reference` (optional).
- **request contract (Chatterbox):** `POST /v1/audio/speech` with body fields
  `input` (text), `voice` (optional), `exaggeration`, `cfg_weight`, `temperature`,
  `stream_format`. Voice upload via `POST /v1/audio/speech/upload` (`voice_file`).
  Voice library via `GET/POST /voices`. `GET /languages`, `GET /health`.
- **response contract (Chatterbox):** binary WAV by default; SSE streaming emits
  `{"type":"speech.audio.delta","audio":"<base64>"}` then
  `{"type":"speech.audio.done","usage":{...}}`.
- **auth method:** none documented (local server).
- **rate limit / quota notes:** none documented; bounded by local GPU/CPU.
- **streaming support:** yes (`/v1/audio/speech/stream`, SSE).
- **async job support:** no explicit job queue; synchronous generation plus
  `/status`/`/status/progress` monitoring.
- **file upload/download behavior:** voice samples uploaded (≤10MB; MP3/WAV/FLAC/
  M4A/OGG); generated audio returned inline.
- **expected media formats:** WAV output (primary).
- **supported languages / modalities:** ~22 languages incl. Russian and English;
  text-to-speech and voice cloning.
- **safety risks:** voice cloning enables impersonation. Reference/cloned voices
  must carry consent metadata (`voice_consent_reference`). Docs do not enforce
  consent — the wrapper and Animus must.
- **secret handling requirements:** none required; if a deployment adds auth, keep
  the token in the wrapper environment, never in the repo.
- **logging/redaction requirements:** wrapper must not print request text or tokens
  to stdout (stdout is the JSON channel); diagnostics go to stderr.
- **retry behavior:** owned by the wrapper; Animus applies an external-command
  timeout and fails closed.
- **timeout behavior:** `ANIMUS_VOICE_TIMEOUT` (default 2m).
- **idempotency strategy:** deterministic output path per episode; Animus hashes
  the file independently and rejects on mismatch.
- **test strategy:** default tests use the Go fake external-command voice provider
  (no Chatterbox dependency). Missing `ANIMUS_VOICE_COMMAND` fails closed.
- **docs links studied:** https://chatterboxtts.com/docs (rendered successfully).
- **unknowns:** exact default sample rate; precise field names may vary by server
  version; whether `language` is a top-level field on `/v1/audio/speech`. The
  wrapper owns the live call, so these are operator-verifiable.
- **decision:** external-command wrapper now; native HTTP provider only later if
  the local API contract proves stable. See `docs/providers/CHATTERBOX_TTS.md` and
  `docs/runbooks/chatterbox_voice_wrapper.md`.

---

## 2. Claude API

- **provider_id:** `claude_api_review`
- **category:** review / QA (not a media generator).
- **intended role in Animus:** structured script review and final QA, producing the
  existing review response artifacts that the pilot gates on.
- **current integration status:** Implemented (native Go provider + tests).
- **recommended first integration mode:** native Go provider behind
  `--claude-review api`.
- **native provider now?** Yes (`internal/shortform/providers/review/claude`).
- **external-command wrapper now?** No (not needed).
- **requires credentials:** Yes — `ANTHROPIC_API_KEY`. Fails closed when unset.
- **requires local server:** No.
- **requires local model:** No.
- **requires network:** Yes (to the Messages API, or to `ANIMUS_CLAUDE_BASE_URL`).
- **requires GUI/session:** No.
- **input artifacts:** `claude_script_review_request.md`, `final_review_request.md`
  (the pilot already writes these).
- **output artifacts:** `claude_script_review_response.json`,
  `final_review_response.json` (`ClaudeReviewResponse` shape).
- **request contract:** `POST {base}/v1/messages` with headers `x-api-key`,
  `anthropic-version: 2023-06-01`, `content-type: application/json`. Body:
  `{model, max_tokens, system, thinking:{type:"adaptive"}, messages:[{role:"user",
  content:<request md>}]}`. Model default `claude-opus-4-8`.
- **response contract:** `{content:[{type,text}], stop_reason, model, usage}`. The
  provider concatenates `text` blocks (skipping `thinking` blocks), extracts the
  JSON object, and validates required keys + `schema_version` + `episode_id`.
- **auth method:** `x-api-key` header.
- **rate limit / quota notes:** 429 is retried with backoff; other 4xx are not.
- **streaming support:** not used (4096 max_tokens, non-streaming is within SDK
  timeout guidance).
- **async job support:** no.
- **file upload/download behavior:** none (text in, JSON out).
- **expected media formats:** none (JSON only).
- **supported languages / modalities:** text reasoning; the review prompt content
  may be Russian or English.
- **safety risks:** a model could "approve" content. Mitigated: Claude is **not** an
  approval authority — the pilot's `scriptReviewPassed`/`finalReviewPassed` gates
  decide, the pilot binds `approved_script_hash`, and the only publish path is
  unchanged. No `output_format`/prefill is used (removed on Opus 4.8).
- **secret handling requirements:** `ANTHROPIC_API_KEY` from env only; never in
  repo; redacted from all error text via `localexec.Redact`.
- **logging/redaction requirements:** the key is redacted from every returned
  error; response-body snippets in errors are also redacted and truncated.
- **retry behavior:** retry only 429 and 5xx and transport errors (max 2 retries,
  linear backoff); never retry 4xx or JSON-validation failures.
- **timeout behavior:** per-request context timeout, `ANIMUS_CLAUDE_TIMEOUT`
  (default 60s).
- **idempotency strategy:** the pilot skips the call when the response file already
  exists; re-running `resume` does not re-bill.
- **test strategy:** `httptest` fake server covers success (script + final), code
  fences, markdown-only rejection, schema mismatch, episode mismatch, refusal,
  transient retry, no-retry on 4xx, key redaction. Pilot tests inject a fake
  `ReviewClient`. No real API calls anywhere in the test suite.
- **docs links studied:** the bundled `claude-api` skill (authoritative Messages
  API contract, model catalog, Opus 4.8 behavior). `platform.claude.com` was not
  fetched; the skill is the cached source of truth.
- **unknowns:** none material for this provider.
- **decision:** implemented as a native review provider. See
  `docs/providers/CLAUDE_API.md` and `docs/adr/0012-claude-api-review-provider.md`.

---

## 3. Seedance 2

- **provider_id:** `seedance2_visual_external`
- **category:** visual video generation.
- **intended role in Animus:** real per-shot 9:16 video generation behind the
  existing `external_command_visual` boundary.
- **current integration status:** External-command only. Native = `Native candidate`
  (deferred).
- **recommended first integration mode:** `ANIMUS_VISUAL_COMMAND` external-command
  wrapper.
- **native provider now?** No (native API not implemented).
- **external-command wrapper now?** Yes.
- **requires credentials:** Yes — a Seedance API key in the wrapper environment.
- **requires local server:** No (cloud API).
- **requires local model:** No.
- **requires network:** Yes (cloud API + CDN download).
- **requires GUI/session:** No.
- **input artifacts:** Animus external visual request JSON (`ExternalVisualInput`):
  per-shot `shot_id`, `scene_id`, `duration_sec`, `prompt`, `negative_prompt`,
  `width`, `height`, `fps`, `output_dir`.
- **output artifacts:** per-shot MP4 under episode `visual/` dir + Animus external
  visual response JSON (`ExternalVisualResponse`): `shot_id`, `status`,
  `output_path`, `output_hash` (optional), `duration_sec`, `width`, `height`,
  `fps`.
- **request contract (Seedance, as documented):** `POST /v1/videos/generations`,
  `Authorization: Bearer sk_live_...`, body `{model:"seedance-2-0"|"seedance-2-0-fast",
  input:{prompt, generation_type, image_urls?, duration(4-15), aspect_ratio(incl
  "9:16"), resolution("480p"|"720p"|"1080p"), seed, ...}, callback_url?}`. Returns
  `{taskId, credits}`.
- **response contract (Seedance, as documented):** poll `GET /v1/tasks/:id` →
  `{id, status:"queued"|"generating"|"completed"|"failed", data:{results:[mp4url],
  video_expires_at, processing_time}}`. Webhook to `callback_url` on completion.
- **auth method:** Bearer token.
- **rate limit / quota notes:** ~60 req/min for generation; 429 carries
  `Retry-After`.
- **streaming support:** no.
- **async job support:** yes — submit task, poll or webhook, then download.
- **file upload/download behavior:** image-to-video reference images via URLs;
  output video downloaded from a CDN URL that expires (`video_expires_at`).
- **expected media formats:** MP4.
- **supported durations / resolutions:** 4–15s; 480p/720p/1080p; aspect ratios
  include `9:16` (the Animus short-form target).
- **safety risks:** generated media is untrusted; cost/spend; expiring URLs; the
  third-party API contract is summarized, not independently re-verified by reading
  raw responses.
- **secret handling requirements:** Bearer key in the wrapper environment only;
  never in the repo.
- **logging/redaction requirements:** wrapper must never print the Bearer token;
  stdout is reserved for the response JSON.
- **retry behavior:** wrapper owns polling/backoff; Animus applies the
  external-command timeout and fails closed on missing output.
- **timeout behavior:** `ANIMUS_VISUAL_TIMEOUT` (default 2m; raise for video).
- **idempotency strategy:** one output file per `shot_id`; Animus hashes each file
  and rejects on mismatch, missing shot, or unknown shot.
- **test strategy:** default tests use the Go fake external-command visual provider
  (no Seedance dependency). Missing config fails closed; hash mismatch, missing
  shot, and path traversal are rejected (`TestExternalVisualPathTraversalRejected`).
- **docs links studied:** https://seedance2.ai/ru/api-docs (rendered; contract
  summarized above). Treated as documented-but-not-independently-verified.
- **unknowns:** exact credit costs; precise image-to-video field names; whether
  9:16 1080p is always available; live auth behavior. These are why the native
  provider is deferred.
- **decision:** external-command wrapper now; native provider only after the auth
  and job lifecycle are verified against live responses. See
  `docs/providers/SEEDANCE2.md` and `docs/runbooks/seedance_visual_wrapper.md`.

---

## 4. OpenAI API

- **provider_id:** `openai_image`
- **category:** storyboard / reference image generation (optional review/council and
  TTS roles deferred).
- **intended role in Animus:** generate reference/storyboard images for shots, as a
  native image provider, in a future storyboard stage.
- **current integration status:** Planned / Native candidate (not implemented).
- **recommended first integration mode:** native Go image provider — **deferred**.
- **native provider now?** No.
- **external-command wrapper now?** No.
- **requires credentials:** Yes — `OPENAI_API_KEY`.
- **requires local server / model:** No / No.
- **requires network:** Yes.
- **requires GUI/session:** No.
- **input artifacts:** (planned) a storyboard image request derived from shot
  prompts.
- **output artifacts:** (planned) PNG images + a `storyboard_image_manifest.json`
  (the artifact already exists in `internal/shortform`).
- **request contract (from secondary sources, not officially verified here):**
  `POST /v1/images/generations`, `Authorization: Bearer $OPENAI_API_KEY`, body
  `{model:"gpt-image-1", prompt, n, size}`. GPT image models always return
  base64 (`response_format` is not supported for them).
- **response contract (from secondary sources):** `{created, data:[{b64_json}]}`.
- **auth method:** Bearer token.
- **rate limit / quota notes:** not verified.
- **streaming support / async job support:** n/a for image generation / no.
- **file upload/download behavior:** base64 image in the response, decoded to a file.
- **expected media formats:** PNG (and others).
- **supported languages / modalities:** image generation from a text prompt.
- **safety risks:** generated images are untrusted; must be hashed, path-contained,
  and reviewed; OpenAI must never be a hidden source of truth.
- **secret handling requirements:** `OPENAI_API_KEY` from env only; never in repo.
- **logging/redaction requirements:** redact the key from logs (as the Claude
  provider does) when implemented.
- **retry behavior / timeout behavior / idempotency:** to be defined with the native
  implementation (mirror the Claude provider patterns).
- **test strategy:** when implemented, `httptest` fake server; no real calls.
- **docs links studied:** `https://developers.openai.com/api/docs` →
  **HTTP 403 (could not render)**; `https://platform.openai.com/docs/api-reference/images/create`
  → **HTTP 403 (could not render)**. A web search corroborated the contract via
  secondary sources (Microsoft Learn, apidog, aimlapi), but the **official primary
  docs were not verifiable in this environment**.
- **unknowns:** exact official request/response schema and limits could not be
  confirmed against the primary source.
- **decision:** **Do not implement a native provider on an unverified contract.**
  Document OpenAI as a native candidate, register `openai_image` as `planned`, and
  plan pipeline wiring for L3 alongside a storyboard stage. See
  `docs/providers/OPENAI_API.md`. There is no L1/L2 storyboard stage, so no schema
  change is forced now.

---

## 5. Claude Code MCP

- **provider_id:** `claude_code_mcp_operator`
- **category:** operator / developer connector (not a runtime model provider).
- **intended role in Animus:** local operator tooling and (future) Review Room and
  DaVinci MCP automation. It is **not** used to generate or review pilot content.
- **current integration status:** Planned (operator connector); documented only.
- **recommended first integration mode:** operator documentation now; no runtime
  wiring.
- **native provider now? / external-command wrapper now?** No / No.
- **requires credentials:** depends on the connected MCP server (kept by the
  operator, never in repo).
- **requires local server:** stdio servers are local processes; HTTP/SSE servers are
  remote.
- **requires local model:** No.
- **requires network:** for remote HTTP/SSE servers.
- **requires GUI/session:** No (CLI/IDE operator session).
- **input/output artifacts:** none in the pilot artifact graph.
- **request/response contract:** MCP (Model Context Protocol) over stdio, HTTP
  (`streamable-http`), SSE (deprecated), or WebSocket. Configured via `claude mcp
  add --transport <http|sse|stdio> <name> <url|-- cmd>`; scopes `local`/`project`/
  `user`; project scope stored in `.mcp.json`; `--env KEY=value`; `MCP_TIMEOUT`.
- **auth method:** per-server (headers/OAuth for HTTP; env for stdio).
- **rate limit / quota notes:** per-server.
- **streaming support / async job support:** transport-dependent.
- **file upload/download behavior:** server-dependent.
- **expected media formats:** n/a.
- **supported languages / modalities:** tool/data access for the operator's Claude
  Code session.
- **safety risks:** **prompt injection** from servers that fetch external content;
  over-broad tool access. Mitigation: trust each server, allowlist tools, and never
  treat MCP output as authoritative for claims or approvals.
- **secret handling requirements:** server credentials live in operator config
  (`~/.claude.json`) or env, never in the repo; do not commit `.mcp.json` with
  secrets.
- **logging/redaction requirements:** operator responsibility.
- **retry behavior / timeout behavior:** Claude Code reconnect/backoff;
  per-server `timeout` and `MCP_TIMEOUT`.
- **idempotency strategy:** n/a.
- **test strategy:** none in this repo (documentation only). The constrained DaVinci
  Resolve MCP tool allowlist (`internal/shortform/providers/mcp`) remains the only
  in-repo MCP boundary and is unrelated to operator tooling.
- **docs links studied:** https://code.claude.com/docs/ru/mcp (rendered).
- **unknowns:** none material for the operator-connector decision.
- **decision:** document as an operator/developer connector only. It must never be a
  hidden runtime review or generation provider for the pilot. See
  `docs/providers/CLAUDE_CODE_MCP.md`.

---

## Cross-cutting decisions

- **No new Go dependencies.** Native HTTP providers use the standard library, per
  `docs/adr/0003` and offline verification. See `docs/adr/0012`.
- **Claude is not a hidden authority.** Review providers produce JSON; gates decide.
- **Generated assets are untrusted.** Every provider output is path-contained,
  hashed, schema-validated, and rejected on mismatch/missing/escape.
- **No public publish.** Publishing stays dry-run / release-candidate only.
