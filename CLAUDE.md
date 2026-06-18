# CLAUDE.md

Guidance for Claude/agents working in this repository. Read `AGENTS.md` first —
it owns the canonical stack rules (Go + Temporal + Postgres + S3; TypeScript only
for console/Remotion/UI). This file adds the short-form video integration (M1).

## What this repo is

Animus News is a source-grounded, multimodel, **artifact-driven** control plane for
educational IT media. It is not an AI content generator: every pipeline stage emits a
typed, validated, content-hashed artifact, and every quality/release decision is a
**code-enforced gate**, not a model instruction.

## Short-form integration (Milestones M1-M3)

M1 integrates the OpenShorts execution capabilities (subtitles, FFmpeg render, 9:16
normalization, Upload-Post publishing) as **typed, validated, gated contracts** that run
end-to-end **on mock providers**.

M2 adds local/dry-run execution boundaries without weakening the safety model:
FFmpeg local render, faster-whisper sidecar contract, and Upload-Post dry-run request
construction. These adapters are disabled by default and must be enabled explicitly in
activity/local-runner configuration. There is still no live Seedance, ElevenLabs, real
Upload-Post scheduling, public social upload, browser automation, or provider spend.
See `docs/reports/M2_status.md` and ADRs `0006` through `0008`.

M3 adds optional professional finishing and local voice lanes: DaVinci Resolve MCP and
OmniVoice. Both are disabled by default, tested through dry-run/fake boundaries, and
remain execution providers only. Animus News still owns workflow state, artifact
validation, gates, production QA, release approval, and publish authority. See
`docs/reports/M3_status.md` and ADRs `0009` through `0011`.

L1 adds the first real CLI pilot path:

- `animus-news pilot generate-real`
- `animus-news pilot resume`
- `animus-news pilot status`
- `animus-news pilot validate`
- `animus-news pilot import-claude-review`
- optional `import-visual-shot` and `import-voice`

The L1 target is a real `release_candidate` MP4 at
`dist/<episode-id>-release-candidate.mp4`. It uses manual Claude JSON review
checkpoints, external-command visual and voice provider boundaries,
faster-whisper or explicit script-timing subtitles, and FFmpeg rendering. It
does not publish publicly. See `docs/REAL_PILOT_V1.md`,
`docs/CONNECTORS.md`, `docs/WORKFLOW_FINAL.md`, and
`docs/reports/LAUNCH_SLICE_L1_status.md`.

L2 integrates real providers in their correct mode: a **native Claude API review
provider** (`--claude-review api`, stdlib HTTP, fake-server tested) plus
documented **external-command wrappers** for Chatterbox (voice) and Seedance
(visual). OpenAI is documented as a native candidate (deferred to L3); Claude
Code MCP is an operator/developer connector only — never a runtime pilot
provider, and HTTP APIs are never wrapped in MCP. Providers are opt-in,
fail-closed, and grant no approval/publish authority. No live calls, spend, or
secrets occur in the repo or CI. See `docs/providers/PROVIDER_RESEARCH_L2.md`,
`docs/PRODUCTION_DEPLOYMENT.md`, ADRs `0012`–`0013`,
`docs/reports/L2_PROVIDER_INTEGRATION_status.md`, and `make verify-l2-providers`.

Key packages (all under `internal/shortform`):

- `contenthash/` — deterministic sha256 over canonical JSON, excluding the hash field.
- `schema/` — dependency-free JSON-Schema (Draft 2020-12 subset) validator; fails closed
  on unsupported keywords. Schemas live in `internal/shortform/schemas/*.schema.json`.
- root (`artifacts.go`, `validate.go`, `approve.go`) — the 8 short-form artifacts, their
  validation, and the operator/human approval + candidate-assembly transforms.
- `providers/` — 6 provider interfaces + deterministic mocks (failure injection).
- `providers/capabilities/` — provider safety/capability registry; descriptive only,
  not a gate bypass.
- `providers/review/claude/` — L2 native Claude API review provider (stdlib HTTP);
  transport + strict JSON validation only, never an approval authority.
- `providers/mcp/` — constrained DaVinci Resolve MCP tool allowlist and dry-run MCP
  client.
- `providers/render/` — FFmpeg local render adapter, disabled by default.
- `providers/render/davinci/` — DaVinci Resolve MCP render/finishing boundary, disabled
  by default.
- `providers/subtitles/` — faster-whisper sidecar boundary, disabled by default.
- `providers/uploadpost/` — Upload-Post dry-run adapter, disabled by default.
- `providers/voice/omnivoice/` — OmniVoice local/sidecar voice boundary, disabled by
  default and consent-gated for reference voice use.
- `providers/localexec/` — path containment, hashing, and redaction helpers for local
  adapters.
- `gates/` — the §8 content/release gates and §4 invariant gates as pure
  `func(input) Result`.
- `activities/` — the §9 pipeline activities. Side effects belong here or in local
  runners, never in workflow code.
- `runner/` — in-process demo driver (shares activities + gates with the workflow).

The durable workflow is `internal/workflows/shortform.go` (`ShortFormWorkflow`), with
human approval signals `StoryboardImageApproval` and `ReleaseApproval`. It is registered
in `internal/worker/worker.go`.

## Non-negotiable invariants (enforced in code, proven by tests)

1. No artifact self-approves (`gates.SelfApprovalGate`); a model/system creator needs a
   distinct human approver.
2. Approved/locked artifacts are immutable (`gates.ImmutabilityGate`,
   `ArtifactStatus.IsTerminalImmutable`); produce a new version instead.
3. AI disclosure is a blocking release gate (`gates.AIDisclosureGate`).
4. The only publish path is `release_approval → publish_manifest → Validate → dry-run →
   release gate`. No generate→publish shortcut. `UploadPostSchedulePublish` errors in M1.
5. ≥2 distinct verifiers for final sign-off (`gates.MultiVerifierGate`).
6. Every artifact is typed, schema-validated, content-hashed, and links upstream
   `source_artifacts`.

## How to run it

```bash
make verify        # single green/red signal: fmt + build + vet + test + scan + schema + e2e demo
make verify-m2-local # M2 adapter contracts + workflow determinism checks
make verify-m3     # M3 provider boundary, registry, replay, and CLI checks
make provider-capabilities # Print provider capability registry JSON
make demo          # short-form mock demo, success path
make demo-blocked  # short-form mock demo with an injected gate failure
go test ./...      # all unit/integration tests (no network, no secrets)

# CLI
go run ./cmd/animus-news demo --episode episode-0001 --expect terminal
go run ./cmd/animus-news validate-shortform <artifact>.json
export EPISODE_ID=<episode-id>
export PROMPT='<runtime prompt>'
export LANGUAGE=ru
export DURATION=45s
export PLATFORMS=tiktok
export EPISODE_DIR="./episodes/${EPISODE_ID}"
go run ./cmd/animus-news pilot generate-real --episode-id "$EPISODE_ID" --prompt "$PROMPT" --language "$LANGUAGE" --duration "$DURATION" --platforms "$PLATFORMS" --visual-provider external-command --voice-provider external-command --subtitle-provider faster-whisper --render-provider ffmpeg --claude-review manual --out "$EPISODE_DIR"
go run ./cmd/animus-news pilot resume --episode-dir "$EPISODE_DIR"
go run ./cmd/animus-news pilot status --episode-dir "$EPISODE_DIR"
go run ./cmd/animus-news pilot validate --episode-dir "$EPISODE_DIR"
make verify-real-pilot
make verify-l2-providers # L2 provider checks: fake HTTP + fake external-command, no real calls

# L2: automated Claude review (fails closed without live-call guard + ANTHROPIC_API_KEY)
go run ./cmd/animus-news pilot generate-real ... --claude-review api ...
```

## Working agreements

- No real external API calls, no spend, no uploads, no secrets in the repo.
- M2 local adapters are opt-in only. Missing binary/model/configuration must fail
  closed; default verification must remain mock/dry-run and offline.
- M3 provider lanes are opt-in only. DaVinci MCP may call only allowlisted tools;
  OmniVoice reference-voice workflows require consent metadata; no provider can approve
  artifacts or publish live.
- L1 real pilot providers are opt-in only. `generate-real` must not silently use
  mocks. Missing visual, voice, subtitle, FFmpeg, or Claude review artifacts must
  fail closed or stop at an explicit manual checkpoint.
- Record non-trivial decisions as ADRs under `docs/adr/NNNN-*.md`; keep the work ledger
  (`docs/ledger/M1.md`, `docs/ledger/M2.md`, `docs/ledger/M3.md`,
  `docs/ledger/LAUNCH_SLICE_L1.md`) current. State must
  be reconstructable from those files.
- Prefer fewer, correct, tested components. A new gate needs a positive test **and** a
  failing-input test per blocking condition.
