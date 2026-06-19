# Animus News

Animus News is a source-grounded, multimodel, **artifact-driven control plane**
for educational IT media — a "content compiler", not a shallow AI content
generator. Every pipeline stage emits a typed, schema-validated, content-hashed
artifact, and every quality/release decision is a **code-enforced gate**, not a
model instruction. Source provenance, multimodel verification, human QA,
production safety, durable workflow orchestration, and auditable release gates
are first-class.

This is the project repository for Animus News. Organization-wide GitHub
defaults (community profile, baseline policies) belong in the `AnimusHQ/.github`
repository, not here.

## Status

Animus News is a **pre-production scaffold**, not a running production media
platform. It is safe-by-default: **no real provider calls, no credentials, no
spend, no uploads, and no public publishing**.

- The **short-form video integration (milestones M1–L2)** runs **end-to-end on
  mock / fail-closed providers**: typed contracts, schema validation, content
  gates, the durable `ShortFormWorkflow`, and an in-process demo runner. Real
  provider lanes (FFmpeg render, faster-whisper, Upload-Post dry-run, DaVinci
  MCP, OmniVoice, Claude review, external-command visual/voice) exist as
  **opt-in, disabled-by-default boundaries** that fail closed when not
  configured. No live calls, spend, or secrets occur in the repo or CI.
- On the Level 0–6 launch-readiness ladder in
  [`docs/PRODUCTION_READINESS.md`](docs/PRODUCTION_READINESS.md), the system sits
  around **Level 3 (Temporal Local Ready), with Level 4 (Provider Sandbox)
  partially implemented**. Levels 5–6 (private and public production) are future
  work. See that document for the authoritative, per-level status.

For honest, code-backed status by feature, treat
[`docs/PRODUCTION_READINESS.md`](docs/PRODUCTION_READINESS.md) and the milestone
reports under `docs/reports/` as the source of truth — not this summary.

## Non-goals

- Not a generate→publish AI content farm. There is no shortcut from generation
  to publishing; the only publish path is
  `release_approval → publish_manifest → validate → dry-run → release gate`.
- No public publishing, scheduled upload, browser automation, or social upload.
- No real provider spend, no live API calls, no secrets in the repository or CI.
- No single model is a final authority; nothing self-approves.
- Not TypeScript-first: Go + Temporal + Postgres + S3 is the canonical stack
  (TypeScript is reserved for console/Remotion/UI only — see `AGENTS.md`).

## Quickstart

All commands are offline and require no credentials.

```bash
make verify        # single green/red signal: fmt + build + vet + test + scan + schema + e2e demo
make demo          # short-form mock demo, success path
make demo-blocked  # short-form mock demo with an injected gate failure

go test ./...      # all unit/integration tests (no network, no secrets)

# CLI (mock/demo)
go run ./cmd/animus-news demo --episode episode-0001 --expect terminal
go run ./cmd/animus-news validate-shortform <artifact>.json
```

The real CLI pilot (release-candidate MP4 via manual review checkpoints and
opt-in provider boundaries) is documented in
[`docs/REAL_PILOT_V1.md`](docs/REAL_PILOT_V1.md); its commands and configuration
are in [`CLAUDE.md`](CLAUDE.md).

## Documentation

- [`docs/SYSTEM_BLUEPRINT.md`](docs/SYSTEM_BLUEPRINT.md) — target system design.
- [`docs/PRODUCTION_READINESS.md`](docs/PRODUCTION_READINESS.md) — authoritative
  readiness levels and safety posture.
- [`docs/WORKFLOW_FINAL.md`](docs/WORKFLOW_FINAL.md) — the final workflow model.
- [`AGENTS.md`](AGENTS.md) — canonical stack and repository-wide rules.
- [`CLAUDE.md`](CLAUDE.md) — short-form integration rules and CLI usage.

## License

Proprietary. © Animus. All rights reserved. See [`LICENSE`](LICENSE). The public
visibility of this repository grants no license to use, copy, modify, or
distribute the software.
