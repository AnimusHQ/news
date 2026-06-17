# L2 Provider Integration Ledger

Mission: study each provider's documentation, connect each to its correct Animus
role, implement the safest high-value integrations, and leave the project ready
to run the first real prompt-driven release-candidate video through the CLI —
without compromising architecture, gates, or security.

## Initial state

Date: 2026-06-17 (work completed 2026-06-18).

Current branch: `m1-openshorts-integration`.

Recent baseline commits:

- `5f7648d docs: production agent workflow and parallelization policy`
- `a16da8b L1: add real CLI pilot`
- `597c90b M3: finalize takeover evidence`

AGENTS.md takeover resolution: the L1 caveat (AGENTS.md modified, uncommitted)
was already resolved before L2 — commit `5f7648d` committed the 692-line
production-agent/parallelization policy. The working tree was clean before L2
edits began.

Pre-change verification: `make verify` (`M3 VERIFY: GREEN`), `make
verify-real-pilot`, `make verify-m2-local`, `make verify-m3`, `go vet ./...`,
`go test ./...` all green (inherited from the L1 commit).

## Direction (confirmed by operator)

"Production" means native typed adapters where the API contract is verified, and
sanctioned external-command wrappers where fast real execution is needed first.
Do not wrap plain HTTP APIs in MCP. No live calls, spend, infrastructure
deployment, or real credentials in this session; runtime secrets are supplied
outside it. No public publishing. All providers opt-in, fail-closed, fake-tested.

## Documentation studied

- Chatterbox TTS — https://chatterboxtts.com/docs (rendered).
- Seedance 2 — https://seedance2.ai/ru/api-docs (rendered; treated as
  documented-not-independently-verified).
- Claude Code MCP — https://code.claude.com/docs/ru/mcp (rendered).
- OpenAI API — https://developers.openai.com/api/docs and
  https://platform.openai.com/docs/api-reference/images/create both returned
  **HTTP 403**; contract corroborated via web search of secondary sources only.
- Claude API — the bundled `claude-api` skill (authoritative Messages API
  contract and Opus 4.8 behavior).

Captured in `docs/providers/PROVIDER_RESEARCH_L2.md`.

## Provider decisions

| Provider | Decision | Status |
| --- | --- | --- |
| Claude API | native Go review provider (stdlib HTTP), `--claude-review api` | Implemented |
| Chatterbox TTS | external-command wrapper via `ANIMUS_VOICE_COMMAND` | External-command only |
| Seedance 2 | external-command wrapper via `ANIMUS_VISUAL_COMMAND`; native deferred | External-command only / Native candidate |
| OpenAI | documented; native deferred to L3 (docs unverifiable; no storyboard stage) | Planned / Native candidate |
| Claude Code MCP | operator/developer connector only | Planned (operator connector) |

## Implementation notes

Code:

- `internal/shortform/providers/review/claude/` — native Messages API client
  (`FromEnv`, `Review(kind, episodeID, prompt)`), stdlib only, fail-closed on
  missing key, strict JSON extraction/validation, transient-only retry, API-key
  redaction; `claude_test.go` uses an `httptest` fake server.
- `internal/shortform/pilot/review.go` — `ReviewClient` interface, env/inject
  constructor, `ensureAPIScriptReview`/`ensureAPIFinalReview`; pilot binds
  `approved_script_hash`, model owns the verdict.
- `internal/shortform/pilot/pipeline.go` — `Runner.ReviewClient`; `Resume` wired;
  `validateGenerateRequest` accepts `api`.
- `internal/shortform/providers/capabilities/registry.go` — `TypeReview`,
  `TypeImage`, `TypeOperator`; 5 new records; tests extended.
- `cmd/animus-news/main.go` — usage/flags mention `--claude-review api`.
- `.gitignore` — ignore real `.env` (keep `.env.example`).
- `Makefile` — `verify-l2-providers` (fake HTTP + fake external-command; no real
  calls).

Tests added (no real calls):

- claude: success (script+final), fences, markdown-only rejection, schema
  mismatch, episode mismatch, refusal, transient retry, no-retry on 4xx, key
  redaction, missing-key fail-closed.
- pilot: api script review passes + binds hash; fail verdict blocks at gate;
  api final review writes validated response; missing key fails closed;
  unsupported `--claude-review` rejected; external visual path-traversal rejected.
- capabilities: L2 providers registered with honest posture (no
  approval/publish).

Docs added/updated: see the status report.

## Gate mapping

| Gate | Evidence |
| --- | --- |
| L2-G1 takeover clean | `5f7648d` committed AGENTS.md; clean tree; branch identified |
| L2-G2 docs studied | `docs/providers/PROVIDER_RESEARCH_L2.md`; OpenAI 403 noted; no invented contracts |
| L2-G3 Claude API review | `providers/review/claude` + fake-HTTP tests; `--claude-review api`; fail-closed |
| L2-G4 Chatterbox path | `docs/providers/CHATTERBOX_TTS.md`, runbook, sample wrapper; fake tests green |
| L2-G5 Seedance path | `docs/providers/SEEDANCE2.md`, runbook, sample wrapper; native deferred; traversal test |
| L2-G6 OpenAI decision | documented native candidate; planned (reasoned); no real calls; capability entry |
| L2-G7 Claude Code MCP | `docs/providers/CLAUDE_CODE_MCP.md`; API-vs-MCP distinction; security |
| L2-G8 registry updated | 5 entries in `provider-capabilities`; `Validate` + tests forbid approval/publish |
| L2-G9 first pilot runbook | `docs/runbooks/first_real_pilot.md` (Claude API, Chatterbox, Seedance, fw, ffmpeg, validate) |
| L2-G10 L1 preserved | `make verify-real-pilot` green; no silent mock fallback |
| L2-G11 security preserved | secret scan clean; redaction test; fail-closed; no public publish; `.env` ignored |
| L2-G12 final verification | `make verify`, `verify-real-pilot`, `verify-m2-local`, `verify-m3`, `verify-l2-providers`, `go vet`, `go test` green |

## Risks and limits

- Chatterbox/Seedance field names are from vendor docs; the wrapper owns the live
  call and must be verified against the running service.
- Seedance native provider is deferred until auth/job lifecycle are verified.
- OpenAI native provider is deferred to L3 (docs unverifiable here; no storyboard
  stage).
- The Claude review uses a stdlib HTTP client by deliberate choice (no-new-deps +
  offline verification), documented in ADR-0012.
- This session performs no live provider calls, spend, deployment, or credential
  handling.
