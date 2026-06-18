# CFG-001 Content-Agnostic Provider Configuration Report

Scope: make the L1/L2 real pilot content-agnostic before a live MVP run. Provider
credentials, endpoints, models, command paths, timeouts, and live-call guards stay
in `.env`/environment. Episode content stays in runtime CLI flags.

## Result

```yaml
content_source: cli_flags
env_contains_content: false
provider_keys_from_env: true
hardcoded_topic_in_code: false
cfg_result: complete
mvp_result: not_run
live_provider_calls: none
release_candidate_created: false
reason: CFG-001 was configuration/policy cleanup only
```

MVP status:

```text
MVP-001 was not run in CFG-001. No live providers were called and no release
candidate MP4 was created.
```

## Implemented

| Area | Result | Evidence |
| --- | --- | --- |
| Runtime content input | Implemented | `pilot generate-real` already supports `--episode-id`, `--prompt`, `--language`, `--duration`, `--platforms`, provider flags, Claude review mode, and `--out`; no request-file feature added. |
| `.env` policy | Implemented | `.env.example` contains provider/operation placeholders only, including live guard, Claude, visual, voice, faster-whisper, and FFmpeg variables. |
| Pilot content neutrality | Implemented | Pilot script and visual-shot generation no longer hardcode open-source, Animus, ecosystem, audience, CTA, or fixed episode themes. |
| Wrapper policy | Implemented | Seedance and Chatterbox examples read request JSON from stdin, require `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1`, and enforce output-root containment. |
| Claude API live-call guard | Implemented | Env-built Claude API review client fails closed unless `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1` and `ANTHROPIC_API_KEY` are present. |
| Regression check | Implemented | `internal/shortform/pilot/content_config_separation_test.go` checks `.env.example`, wrappers, and pilot production code for obvious content/config violations. |
| Documentation | Implemented | Runbooks/provider docs use runtime shell variables and placeholders rather than fixed episode/topic examples. |

## Partial

| Area | Remaining gap |
| --- | --- |
| Real MVP smoke run | Not run in CFG-001. |
| Full MVP run | Not run in CFG-001. |

## Planned

| Area | Planned condition |
| --- | --- |
| Request file | Add only if operators need runtime inputs beyond current CLI flags, such as audience, style, CTA, or per-shot prompt overrides. |
| Richer content compiler | Future source-grounded research/claims flow remains separate from this MVP-001 config/content separation fix. |

## Verification

Baseline before edits:

- `make verify` blocked at `go build ./...` with `error obtaining VCS status: exit status 128`.
- `make verify-real-pilot` passed.
- `make verify-m2-local` passed.
- `make verify-m3` passed.
- `make verify-l2-providers` passed.
- `go vet ./...` passed.
- `go test ./...` passed.

VCS correction:

- Root cause: the initial CFG worktree was a linked worktree under `/tmp`; Go
  VCS stamping invoked `git status --porcelain` from `/tmp`, where an invalid
  `/tmp/.git` directory existed.
- Fix: moved the CFG patch set to `/home/guest/projects/news-cfg-001`, a full
  checkout with a normal `.git` directory, and kept VCS stamping enabled.
- `go list -x ./cmd/animus-news` now invokes Git from
  `/home/guest/projects/news-cfg-001` and succeeds.
- `-buildvcs=false` was not used.

Implementation checks:

- `python3 -m py_compile scripts/providers/chatterbox-voice-wrapper.example.py scripts/providers/seedance2-visual-wrapper.example.py` passed.
- `go test ./internal/shortform/pilot ./internal/shortform/providers/review/claude` passed.

Final verification:

| Command | Result | Notes |
| --- | --- | --- |
| `git status --porcelain` | pass | CFG-001 source/doc/test changes only before commit. |
| `make verify` | pass | VCS stamping succeeds from the full checkout. |
| `make verify-real-pilot` | pass | Pilot tests include content/config separation regression. |
| `make verify-m2-local` | pass | M2 adapter and workflow checks passed. |
| `make verify-m3` | pass | M3 provider boundary checks passed. |
| `make verify-l2-providers` | pass | L2 fake provider checks passed. |
| `go vet ./...` | pass | No findings. |
| `go test ./...` | pass | All packages green. |
| no-prompt runtime invocation | blocked as expected | `missing required flags: --prompt`; no generated episode directory. |

## Security notes

- No `.env` file was committed.
- No secrets or real API keys were added.
- No generated media was committed.
- No public publishing path was added.
- Live provider calls now require an explicit `ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1`
  guard.
- Provider outputs remain untrusted and are still hashed, path-contained, and
  validated by Animus.
