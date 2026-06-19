# Animus Media Engine

Animus Media Engine is a source-grounded, multi-agent, artifact-driven production system for strategy-aware AI media generation.

The repository is currently named `AnimusHQ/news` because the first product slice is short-form Animus News, but the target is broader: a production-grade AI media operating system that can generate verified short-form release candidates today and evolve toward cinematic and film-scale production later.

It is not a prompt-to-video toy, a single-provider wrapper, or a generate-to-publish content farm. It is a media production control plane: every important stage emits typed artifacts, every provider output is untrusted until validated, and every release decision passes explicit gates.

## Current product framing

The system transforms a runtime creative task into a validated production artifact set:

```text
Task / campaign goal / creative objective
  -> platform analysis
  -> audience strategy
  -> competitive/reference analysis
  -> creative direction
  -> research and source state
  -> script
  -> verification
  -> multi-agent critique
  -> human QA
  -> storyboard and shot plan
  -> OTIO timeline
  -> visual generation
  -> voice generation
  -> subtitles
  -> render
  -> production QA
  -> release candidate
  -> human release gate
  -> publishing and analytics later
```

The immediate observable output for the MVP lane is:

```text
episodes/<episode-id>/dist/<episode-id>-release-candidate.mp4
```

The production output is broader: typed artifacts, hashes, provenance, agent/model decisions, media object references, timelines, QA reports, and release decisions.

## Status

The repository is a **pre-production production foundation**, not a public media platform. It is safe-by-default: no real provider calls, no credentials, no spend, no uploads, and no public publishing occur in CI or default local verification.

Completed checkpoints:

- **M1-M3**: typed contracts, validators, gates, mock providers, Temporal skeleton, local execution boundaries, provider registry, replay hardening, DaVinci/OmniVoice boundaries.
- **L1-L2**: real CLI pilot framework, native Claude API review lane, external-command Seedance/Chatterbox lanes, FFmpeg render path, provider docs and runbooks.
- **CFG-001**: provider configuration is content-agnostic. Runtime content is not stored in `.env`; environment files contain provider/service configuration only.
- **MVP-Docker-001**: local MVP runtime is containerized. Docker Compose starts Chatterbox and provides stable FFmpeg/ffprobe, Python, Go, wrapper paths, and episode roots.

Immediate next checkpoints:

1. **MVP-001 Live Smoke**: from clean clone + `.env.mvp.local` + Docker Compose, produce `episodes/mvp-smoke-001/dist/mvp-smoke-001-release-candidate.mp4` or a precise blocker.
2. **MVP-001 Full Live**: produce `episodes/mvp-001/dist/mvp-001-release-candidate.mp4` or a precise blocker.
3. **PLATFORM-001**: local production platform foundation with Docker-bootstrapped Postgres, Temporal, MinIO, Keycloak, Animus API, Animus worker, model router, OTIO timeline layer, and multi-agent execution boundaries.

## Non-goals

- No shortcut from generation to public publishing.
- No public publishing, scheduled public upload, browser automation, or social upload until an explicit release milestone enables it.
- No real provider spend, live API calls, or secrets in CI.
- No single model is a final authority; generated output does not self-approve.
- No hardcoded topic, style, platform, voice, duration, or creative format.
- No provider-specific code path may bypass the model router or validation gates.
- No TypeScript-first backend: Go + Temporal + Postgres + S3-compatible object storage is the canonical stack. TypeScript is reserved for console, review room, Remotion, and UI surfaces.

## Local verification

Offline verification requires no credentials:

```bash
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
make verify-mvp-docker
go vet ./...
go test ./...
```

MVP Docker runtime is documented in [`docs/runbooks/mvp_docker_runtime.md`](docs/runbooks/mvp_docker_runtime.md).

## Documentation

- [`docs/PLATFORM_FOUNDATION.md`](docs/PLATFORM_FOUNDATION.md) — current north star and PLATFORM-001 scope.
- [`docs/SYSTEM_BLUEPRINT.md`](docs/SYSTEM_BLUEPRINT.md) — target system design and artifact-first architecture.
- [`docs/PRODUCTION_READINESS.md`](docs/PRODUCTION_READINESS.md) — readiness levels, current status, and blockers.
- [`docs/WORKFLOW_FINAL.md`](docs/WORKFLOW_FINAL.md) — current MVP workflow and target production workflow.
- [`docs/PRODUCTION_DEPLOYMENT.md`](docs/PRODUCTION_DEPLOYMENT.md) — provider and platform deployment boundaries.
- [`docs/runbooks/mvp_docker_runtime.md`](docs/runbooks/mvp_docker_runtime.md) — MVP Docker smoke/full local operator flow.
- [`AGENTS.md`](AGENTS.md) — canonical stack and repository-wide rules.
- [`CLAUDE.md`](CLAUDE.md) — current short-form integration rules and CLI usage.

## License

Proprietary. © Animus. All rights reserved. See [`LICENSE`](LICENSE). The public visibility of this repository grants no license to use, copy, modify, or distribute the software.
