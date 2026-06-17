# OpenAI API — Storyboard / Reference Image (Planned)

Status: **Planned / Native candidate**. Not implemented in this repo.

## Decision and reason

The intended first role for OpenAI in Animus is a **native storyboard / reference
image provider**. It is **not implemented** in L2 for two reasons:

1. **The official API documentation could not be verified in this environment.**
   - `https://developers.openai.com/api/docs` → HTTP 403.
   - `https://platform.openai.com/docs/api-reference/images/create` → HTTP 403.
   A web search corroborated the contract via secondary sources (Microsoft Learn,
   apidog, aimlapi), but the primary docs were not renderable here. Per the L2
   anti-hallucination rule, an unverified contract is not implemented.
2. **There is no storyboard stage in the L1/L2 pilot.** Wiring an image provider
   would force a new pipeline stage and artifact flow. Per the L2 rules, large
   schema changes are not forced now; this is planned for L3.

A `openai_image` entry is registered in the provider capability registry with
status `planned` (disabled, requires paid API, no approval/publish authority).

## Contract (from secondary sources — to be verified before implementing)

- Endpoint: `POST /v1/images/generations`
- Auth: `Authorization: Bearer $OPENAI_API_KEY`
- Request: `{ "model": "gpt-image-1", "prompt": "...", "n": 1, "size": "1024x1536" }`
- Response: `{ "created": ..., "data": [ { "b64_json": "<base64 png>" } ] }`
  (GPT image models always return base64; `response_format` is not supported for
  them.)

## Planned L3 implementation (mirror the Claude provider)

`internal/shortform/providers/image/openai`:

- `FromEnv()` reading `OPENAI_API_KEY`, `ANIMUS_OPENAI_MODEL`,
  `ANIMUS_OPENAI_BASE_URL`, `ANIMUS_OPENAI_TIMEOUT`; fail closed without a key.
- `GenerateImage(ctx, prompt, size, outPath)` → decode base64, write the file
  under the episode root, return the sha256 hash; redact the key from errors.
- Save images under the episode root; create/update a
  `storyboard_image_manifest.json` (the artifact already exists in
  `internal/shortform`).
- `httptest` fake-server tests; **no real API calls** in tests.

## Other possible OpenAI roles (later)

- multimodel review/council provider (alongside Claude; never the sole truth);
- TTS fallback (only after quality evaluation);
- script/rewrite provider.

OpenAI must never be a hidden source of truth or an approval/publish authority.

Studied: official docs returned 403 here; contract corroborated via web search of
secondary sources on 2026-06-17.
