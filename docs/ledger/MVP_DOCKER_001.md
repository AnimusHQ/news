# MVP-Docker-001 Ledger - Containerized MVP Runtime

Task ID: MVP-Docker-001

Title: Containerized MVP Runtime

Goal: make the local real-pilot MVP runnable through Docker Compose so the
operator no longer manually installs FFmpeg, starts Chatterbox, configures
wrapper paths, or wires local service roots by hand.

Urgency: launch-critical L1/L2 operator workflow hardening.

Scope:

- Add Docker Compose service wiring for `animus-news` and `chatterbox`.
- Add Dockerfiles for the Animus CLI runner and Chatterbox-compatible service.
- Add a fail-closed entrypoint for runtime shell variables.
- Add `.env.mvp.example` with only operator-editable provider config.
- Add static verification for Docker runtime policy.
- Add runbook and implementation report.

Non-goals:

- No public publishing.
- No Seedance containerization.
- No MCP provider for HTTP APIs.
- No faster-whisper requirement when `script-timing` subtitles are used.
- No fake-provider fallback for live MVP success.
- No schema, workflow, or provider-interface rewrite.

Dependencies:

- CFG-001 merged into `main`.
- Existing `pilot generate-real` CLI.
- Existing Seedance external-command visual wrapper.
- Existing Chatterbox external-command voice wrapper.
- Existing Claude API review provider.
- Existing FFmpeg render path.

Parallelization plan:

- Lane A, runtime owner: Docker Compose, Dockerfiles, entrypoint, Chatterbox
  launcher, `.env.mvp.example`.
- Lane B, verification owner: `.gitignore`, `Makefile`, static Go tests.
- Lane C, docs owner: runbook, ledger, report.
- Lane G, integration owner: final merge, verification, honest blocker report.

Separate write worktrees were not used for the final patch because the runtime,
verification, and docs lanes share one small environment contract and the
integration risk is lower with one local patch owner. Explorer sub-agents were
used for read-only CLI/provider contract and verification-scope inspection.

Allowed files:

- `docker-compose.mvp.yml`
- `docker/**`
- `.dockerignore`
- `.env.mvp.example`
- `.gitignore`
- `Makefile`
- `internal/shortform/pilot/*mvp_docker*`
- `docs/runbooks/mvp_docker_runtime.md`
- `docs/ledger/MVP_DOCKER_001.md`
- `docs/reports/MVP_DOCKER_001_CONTAINERIZED_RUNTIME.md`

Forbidden changes:

- No generated media committed.
- No `.env.mvp.local` committed.
- No secrets committed.
- No public upload path.
- No provider mock fallback for real Docker MVP runs.
- No hardcoded video topic, prompt, style, or CTA in env files.
- No arbitrary shell/MCP provider execution added.

Acceptance gates:

- `.env.mvp.example` contains no prompt/topic/episode/duration/platform content
  keys and no local paths.
- `docker-compose.mvp.yml` mounts `./episodes:/workspace/episodes`, loads
  `.env.mvp.local`, and does not hardcode a topic.
- `animus-news` entrypoint fails closed when `PROMPT` is empty.
- `.env.mvp.local` is gitignored.
- Docker files contain no high-risk secret findings.
- `make verify-mvp-docker` passes without live keys.

Validation commands:

```bash
make verify-mvp-docker
git status --porcelain
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
go vet ./...
go test ./...
```

Pilot success command:

```bash
EPISODE_ID=mvp-001 \
PROMPT="runtime prompt here" \
LANGUAGE=ru \
DURATION=45s \
PLATFORMS=tiktok,instagram,youtube \
docker compose -f docker-compose.mvp.yml --profile mvp run --rm animus-news
```

Expected real output:

```text
episodes/mvp-001/dist/mvp-001-release-candidate.mp4
```

Manual checkpoints:

- `.env.mvp.local` copied from `.env.mvp.example`.
- Live provider calls explicitly enabled.
- Anthropic and Seedance credentials present.
- Chatterbox healthcheck healthy.
- Optional voice consent reference present before using a cloned/reference voice.

Provider configuration:

- Claude remains cloud API via `ANTHROPIC_API_KEY`.
- Seedance remains cloud API via `SEEDANCE_API_KEY`.
- Chatterbox runs as a local Docker service at `http://chatterbox:4123`.
- FFmpeg/ffprobe come from the `animus-news` image.

Failure behavior:

- Missing runtime variables block in the entrypoint before provider execution.
- Missing provider keys or live-call gate block before real calls.
- Chatterbox service health failure blocks `animus-news` startup.
- Invalid provider output blocks during Animus artifact validation.
- No command publishes publicly.

Documentation requirements:

- Runbook with smoke/full commands.
- Report with branch/commit range, services, operator values, verification,
  security notes, limitations, and next step.

Final report format:

Use the `MVP-Docker-001 Summary` fields requested in the task prompt.

Takeover commands:

```bash
git status --porcelain
make verify
make verify-real-pilot
make verify-m2-local
make verify-m3
make verify-l2-providers
make verify-mvp-docker
go vet ./...
go test ./...
```
