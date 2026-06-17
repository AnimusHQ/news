# L2 Provider Integration Status Report

Scope: study the five provider documentation sources, map each provider to its
correct Animus role, implement the safest high-value integrations (native Claude
API review + external-command wrappers), update the capability registry and docs,
and keep every M1/M2/M3/L1 gate green — with no live calls, spend, secrets, or
public publishing.

## Verification status

Commands run before implementation (inherited from the L1 commit):

| Command | Result | Notes |
| --- | --- | --- |
| `git status --porcelain` | pass | clean (AGENTS.md already committed in `5f7648d`) |
| `git branch --show-current` | pass | `m1-openshorts-integration` |
| `make verify` | pass | `M3 VERIFY: GREEN` |
| `make verify-real-pilot` | pass | L1 pilot + docs presence |
| `make verify-m2-local` | pass | M2 adapters + determinism |
| `make verify-m3` | pass | M3 boundary + replay |
| `go vet ./...` | pass | no findings |
| `go test ./...` | pass | all packages green |

Commands run after L2 implementation:

| Command | Result | Notes |
| --- | --- | --- |
| `make verify` | pass | `M3 VERIFY: GREEN`; secret scan `findings: null` |
| `make verify-real-pilot` | pass | L1 flow preserved |
| `make verify-m2-local` | pass | unchanged |
| `make verify-m3` | pass | unchanged |
| `make verify-l2-providers` | pass | `L2 PROVIDERS VERIFY: GREEN` (fake HTTP + fake external-command) |
| `go vet ./...` | pass | no findings |
| `go test ./...` | pass | exit 0, no FAIL/panic |
| `rm -rf build dist episodes/tmp episodes/test-output` | pass | cleanup completed |
| final `git status --porcelain` | pass | only intended L2 changes; no stray artifacts |

## Gate results

| Gate | Result | Evidence |
| --- | --- | --- |
| L2-G1 Repository takeover clean | Pass | AGENTS.md committed in `5f7648d` (production-agent/parallelization policy); tree clean before edits; branch `m1-openshorts-integration`. |
| L2-G2 Provider docs studied and captured | Pass | `docs/providers/PROVIDER_RESEARCH_L2.md` covers all five with the full template; unknowns recorded; OpenAI 403 stated; no invented contracts. |
| L2-G3 Claude API review provider integrated | Pass (Implemented) | `internal/shortform/providers/review/claude` + `httptest` tests; `--claude-review api` wired; missing key fails closed; structured JSON validated. |
| L2-G4 Chatterbox voice path ready | Pass | `docs/providers/CHATTERBOX_TTS.md`, `docs/runbooks/chatterbox_voice_wrapper.md`, sample wrapper; external-command contract documented; fake provider tests green. |
| L2-G5 Seedance visual path ready | Pass | `docs/providers/SEEDANCE2.md`, `docs/runbooks/seedance_visual_wrapper.md`, sample wrapper; no native API invented (deferred); fake tests + path-traversal test green. |
| L2-G6 OpenAI provider decision made | Pass (deferred with reason) | `docs/providers/OPENAI_API.md`; role = storyboard/reference image; official docs 403, no storyboard stage → clearly planned for L3; capability entry `openai_image` planned; no real calls. |
| L2-G7 Claude Code MCP documented correctly | Pass | `docs/providers/CLAUDE_CODE_MCP.md`; explicit Claude API vs Claude Code MCP distinction; security/prompt-injection; future operator role; forbidden patterns. |
| L2-G8 Provider capability registry updated | Pass | 5 new providers in `provider-capabilities`; honest statuses; `Registry.Validate` + tests forbid approval/publish authority. |
| L2-G9 First real pilot runbook exists | Pass | `docs/runbooks/first_real_pilot.md` covers Claude API, Chatterbox, Seedance wrapper, faster-whisper, FFmpeg, validation, troubleshooting, cleanup. |
| L2-G10 Existing L1 flow preserved | Pass | `make verify-real-pilot` green; `generate-real` still works with fake providers; no silent mock fallback (api mode fails closed without a key). |
| L2-G11 Security preserved | Pass | secret scan clean; API-key redaction test + documented redaction; missing config fails closed; no public publish path; `.env` gitignored; no repo secrets. |
| L2-G12 Final verification green | Pass | all targets above pass; clean working tree (no build/dist/episode artifacts). |

## Implemented

- Native Claude API review provider (`--claude-review api`), stdlib HTTP, fake-HTTP
  tested, fail-closed, transient-only retry, API-key redaction.
- Pilot wiring: `ReviewClient` interface, env/inject construction, script + final
  API review steps; pilot binds `approved_script_hash`.
- Provider capability registry: `claude_api_review`, `chatterbox_tts_external`,
  `seedance2_visual_external`, `openai_image`, `claude_code_mcp_operator`.
- `make verify-l2-providers`; `.env.example`; `.gitignore` hardening.
- Provider research, per-provider docs, runbooks, sample wrappers, production
  deployment guide, ADRs 0012–0013, ledger.

## Partial

- Chatterbox and Seedance are connected through the sanctioned external-command
  boundary with documented wrapper contracts and sample wrappers; native adapters
  are not implemented.
- faster-whisper remains a sidecar protocol (from L1).

## Planned

- Native Seedance adapter (after auth/job lifecycle verified with fake-server
  tests).
- Native OpenAI image provider + a storyboard stage and
  `storyboard_image_manifest` wiring (L3).
- Optional OpenAI review/council and TTS-fallback roles.
- Claude Code MCP operator tooling and Review Room (L4); it remains a
  non-runtime connector.

## How to run fake-provider verification

```bash
make verify-l2-providers   # fake HTTP server + fake external-command; no real calls
go test ./internal/shortform/providers/review/claude ./internal/shortform/pilot \
        ./internal/shortform/providers/capabilities
```

## How to run the first real pilot with providers

See `docs/runbooks/first_real_pilot.md` and `docs/PRODUCTION_DEPLOYMENT.md`.
Summary: export `ANTHROPIC_API_KEY` (or use `--claude-review manual`), point
`ANIMUS_VOICE_COMMAND`/`ANIMUS_VISUAL_COMMAND` at the wrappers, configure
faster-whisper or `--subtitle-provider script-timing` and FFmpeg, then run
`pilot generate-real ... --claude-review api ...`, `resume`, and `validate`.

## Security notes

- No secrets in the repo; `.env` gitignored; `.env.example` placeholders only.
- Claude provider redacts `ANTHROPIC_API_KEY` from every error; tested.
- Missing provider config fails closed; `generate-real` never silently uses mocks.
- Provider outputs are path-contained, hashed, schema-validated; mismatch /
  missing / traversal rejected.
- No public publishing; publishing stays `release_candidate_only` / private /
  live disabled.
- No live provider calls, spend, infrastructure deployment, or credential handling
  occurred in this session; CI requires none.

## Known risks

- Chatterbox/Seedance contracts are from vendor docs; the operator wrapper owns the
  live call and must be verified against the running service.
- OpenAI's official API contract could not be verified here (403); native work is
  deferred rather than built on assumptions.

## Next milestone recommendation

L3: add a storyboard stage with a native OpenAI image provider (fake-server
tested), verify the Seedance auth/job lifecycle and add a native Seedance adapter,
and begin source-grounded research/claims hardening per the connector roadmap.

## Exact takeover commands

```bash
git status --porcelain
git branch --show-current
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
go vet ./...
go test ./...
rm -rf build dist episodes/tmp episodes/test-output
git status --porcelain
```
