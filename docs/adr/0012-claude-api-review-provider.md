# ADR-0012: Claude API Review Provider (stdlib HTTP)

Status: accepted.

## Context

L1 added manual Claude review checkpoints: the pilot writes request markdown,
the operator pastes Claude's JSON, and `pilot import-claude-review` validates it.
L2 needs an automated path (`--claude-review api`) so the pilot can request the
script and final QA reviews itself.

Two constraints shape the implementation:

- The repository enforces a no-new-dependencies invariant (ADR-0003), and
  `make verify` must run fully offline. Adding `anthropic-sdk-go` (the skill
  default) would pull in a dependency tree and is not installable offline.
- A review provider must never become an approval authority. The existing gates
  (`scriptReviewPassed`, `finalReviewPassed`) and the script-hash binding must
  stay authoritative.

## Decision

Add `internal/shortform/providers/review/claude`, a Go standard-library client
for the Anthropic Messages API.

- Wire contract: `POST {base}/v1/messages`, headers `x-api-key` +
  `anthropic-version: 2023-06-01`; body `{model, max_tokens, system,
  thinking:{type:"adaptive"}, messages}`; model default `claude-opus-4-8`. No
  `output_format`/prefill (removed on Opus 4.8).
- `Review(ctx, kind, episodeID, prompt)` returns the validated JSON object: text
  blocks are concatenated, the first JSON object is extracted (code fences
  tolerated), required keys + `schema_version` + `episode_id` are checked, and a
  markdown-only/incomplete response is rejected.
- Fails closed without `ANTHROPIC_API_KEY`. Retries only 429/5xx/transport
  errors (max 2); never retries 4xx or validation failures. Redacts the key from
  every error.
- Config via env: `ANTHROPIC_API_KEY`, `ANIMUS_CLAUDE_MODEL`,
  `ANIMUS_CLAUDE_TIMEOUT`, `ANIMUS_CLAUDE_MAX_TOKENS`, `ANIMUS_CLAUDE_BASE_URL`
  (the last for fake-server tests).

Pilot wiring (`internal/shortform/pilot/review.go`): a `ReviewClient` interface
(injected in tests, built from env in production), and `ensureAPIScriptReview` /
`ensureAPIFinalReview` steps that run only when `--claude-review api` is selected
and the response file is absent. The model owns the editorial verdict; the pilot
rewrites `approved_script_hash` to bind the review to `script.md`. Validation and
the gate decision remain in the pilot.

## Consequences

- The pilot can run Claude review automatically without a network dependency in
  tests (an `httptest` fake server pins the contract; pilot tests inject a fake
  `ReviewClient`). No real API call occurs in any test.
- Claude is not a hidden authority: a non-pass verdict is written through and
  blocks at the existing gate; a transport/JSON failure fails closed.
- The deviation from the SDK default is deliberate and documented; the wire shape
  stays faithful to the Messages API.
- Manual review (`--claude-review manual`) remains the offline default and is
  unchanged.
