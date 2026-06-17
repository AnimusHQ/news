# Claude API — Review / QA Provider

Status: **Implemented** (`internal/shortform/providers/review/claude`).

The Claude API provider performs the pilot's two review tasks — `script_review`
and `final_qa` — through the Anthropic Messages API and returns the model's
structured JSON verdict. It is selected with `--claude-review api`.

It is a **transport + strict JSON validator only**. It is never an approval
authority: the pilot's gates (`scriptReviewPassed`, `finalReviewPassed`) decide
pass/fail, and the pilot — not the model — binds `approved_script_hash` to the
script on disk.

## Why a stdlib HTTP client (not the SDK)

The skill default is the official `anthropic-sdk-go`. This repository instead
uses a Go standard-library client because:

- the repo enforces a **no-new-dependencies** invariant (`docs/adr/0003`); and
- `make verify` must run **fully offline** (no `go get`).

The wire contract is faithful to the documented Messages API. See
`docs/adr/0012-claude-api-review-provider.md`.

## Configuration

```bash
export ANTHROPIC_API_KEY=...           # required; fails closed when unset
export ANIMUS_CLAUDE_MODEL=claude-opus-4-8   # default
export ANIMUS_CLAUDE_TIMEOUT=60s       # default
export ANIMUS_CLAUDE_MAX_TOKENS=4096   # default
export ANIMUS_CLAUDE_BASE_URL=https://api.anthropic.com   # default; override for tests
```

## Wire contract

Request: `POST {base}/v1/messages`

```
headers:
  x-api-key: $ANTHROPIC_API_KEY
  anthropic-version: 2023-06-01
  content-type: application/json
body:
  {
    "model": "claude-opus-4-8",
    "max_tokens": 4096,
    "system": "<review instructions: JSON-only, required keys>",
    "thinking": {"type": "adaptive"},
    "messages": [{"role": "user", "content": "<request markdown>"}]
  }
```

The provider concatenates the `text` content blocks (skipping `thinking`
blocks), extracts the first JSON object (tolerating code fences), and validates:
required keys present, `schema_version == "1.0"`, `episode_id` matches, and a
non-empty `verdict`. A markdown-only or schema-incomplete response is rejected.

## Expected response JSON

Script review:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "verdict": "pass",
  "production_readiness": 85,
  "blocking_issues": [],
  "suggested_revisions": [],
  "approved_script_hash": "sha256:...",
  "can_continue_to_visual_generation": true
}
```

Final QA:

```json
{
  "schema_version": "1.0",
  "episode_id": "animus-oss-001",
  "verdict": "pass",
  "production_readiness": 85,
  "blocking_issues": [],
  "suggested_revisions": [],
  "can_release_candidate": true
}
```

## Hash binding

For `script` reviews, the pilot **rewrites** `approved_script_hash` to the actual
sha256 of `script.md` before writing the response. The model owns the editorial
verdict; the pilot owns the content binding — exactly as it does everywhere else
in the codebase. If the script later changes, the gate recomputes the hash and
re-blocks. (The manual import path still enforces a matching hash, so the gate's
hash check remains load-bearing and tested.)

## Failure and retry semantics

- Missing `ANTHROPIC_API_KEY` → fails closed; no response file is written.
- 429 / 5xx / transport errors → retried (max 2, linear backoff).
- 4xx (other than 429) and JSON-validation failures → **not** retried.
- `stop_reason: "refusal"` → rejected with a clear error.
- The API key is redacted (`[REDACTED]`) from every returned error; response-body
  snippets are also redacted and truncated.

## Idempotency

`resume` skips the API call when the response file already exists, so re-runs do
not re-bill. To force a fresh review, delete the response file.

## CLI

```bash
go run ./cmd/animus-news pilot generate-real ... --claude-review api ...
go run ./cmd/animus-news pilot resume --episode-dir ./episodes/animus-oss-001
```

`--claude-review manual` (operator pastes Claude JSON, then
`pilot import-claude-review`) remains fully supported and is the offline default.

## Tests

`internal/shortform/providers/review/claude/claude_test.go` uses an `httptest`
fake server: success (script + final), code fences, markdown-only rejection,
schema mismatch, episode mismatch, refusal, transient retry, no-retry on 4xx,
and API-key redaction. Pilot wiring tests inject a fake `ReviewClient`. No real
API calls occur in any test.
