# M3 Status Report - DaVinci Resolve MCP, OmniVoice, and Replay Hardening

Scope: add safe, disabled-by-default provider lanes for DaVinci Resolve MCP and
OmniVoice; improve offline workflow replay evidence; add provider capability
metadata; preserve M1/M2 safety gates and dry-run-only publishing.

## Verification status

Commands run during implementation:

| Command | Result | Notes |
| --- | --- | --- |
| `make verify` | pass | `M3 VERIFY: GREEN`; includes gofmt check, build, vet, full tests, secret scan, provider capability CLI, mock demo success/block, and schema validation. |
| `make verify-m2-local` | pass | M2 local adapter and workflow determinism checks remain green after M3 edits. |
| `make verify-m3` | pass | M3 provider boundary, registry, CLI, and replay checks. |
| `go vet ./...` | pass | No findings. |
| `go test ./...` | pass | All packages green. |
| `go test ./internal/shortform/...` | pass | Includes DaVinci, OmniVoice, MCP, capability registry, schema updates. |
| `go test ./internal/workflows` | pass | Includes M3 replay fixture and workflow-boundary checks. |
| `go test ./cmd/animus-news` | pass | Includes provider capability CLI test. |
| `git status --porcelain` | pass | Clean before and after takeover cleanup. |

## Gate results

| Gate | Result | Evidence |
| --- | --- | --- |
| M3-G1 - M1/M2 preserved | Implemented | Post-edit `make verify` and `make verify-m2-local` are green; M1 demo success/block paths still work; no real publish. |
| M3-G2 - Temporal replay evidence improved | Implemented | Happy-path fixture now includes notes/state transitions; blocked-path fixture hashes added for storyboard rejection, release denial, and render-gate block. |
| M3-G3 - DaVinci Resolve MCP provider boundary safe | Implemented | `internal/shortform/providers/mcp` and `render/davinci`; disabled by default, MCP tool allowlist, dry-run client, path containment, start-render guard, draft outputs, workflow-boundary test. |
| M3-G4 - OmniVoice provider boundary safe | Implemented | `internal/shortform/providers/voice/omnivoice`; disabled by default, missing binary/model fail closed, dry-run/fake sidecar tests, consent metadata enforcement, draft outputs. |
| M3-G5 - provider capability registry safe | Implemented | `internal/shortform/providers/capabilities`; registry lists M3 providers and planned providers, unknown/disabled fail closed, no provider can approve or publish live. |
| M3-G6 - takeover ready | Implemented | Final committed-state takeover commands passed: `git status --porcelain`, `make verify`, `make verify-m2-local`, `go vet ./...`, `go test ./...`, `make verify-m3`, cleanup, and clean-tree check. |

## Implemented

| Component | Evidence |
| --- | --- |
| Replay fixture hardening | `internal/workflows/shortform_test.go`; happy path includes state-transition notes; blocked path fixtures added. |
| Workflow provider-boundary guard | `internal/workflows/provider_boundary_test.go`; workflow source must not reference DaVinci, OmniVoice, MCP, or process execution. |
| DaVinci Resolve MCP allowlist | `internal/shortform/providers/mcp`; non-allowlisted tools are refused. |
| DaVinci Resolve render provider boundary | `internal/shortform/providers/render/davinci`; disabled by default, dry-run/local-MCP modes, start-render guard, draft render manifest. |
| OmniVoice voice provider boundary | `internal/shortform/providers/voice/omnivoice`; disabled by default, dry-run/local-sidecar modes, local model/binary checks, consent checks. |
| Optional artifact metadata | `voiceover_manifest` supports voice prompt/consent/sample-rate metadata; `short_render_manifest` supports provider/timeline/MCP metadata. |
| Provider capability registry | `internal/shortform/providers/capabilities`; CLI `provider-capabilities`; Make target `provider-capabilities`. |
| M3 verification target | `make verify-m3`. |
| ADRs | ADR-0009, ADR-0010, ADR-0011. |

## Partial

| Component | Remaining gap |
| --- | --- |
| Full Temporal JSON-history replay | Offline fixture evidence is stronger, but `worker.WorkflowReplayer` against a committed JSON history still needs a gated Temporal dev-server capture in M4. |
| Real DaVinci Resolve MCP execution | Boundary and dry-run client are implemented; no real Resolve GUI/MCP server is required or exercised by default tests. |
| Real OmniVoice model execution | Boundary and fake sidecar tests are implemented; no real model download or production-quality TTS claim is made. |

## Planned

| Component | Target |
| --- | --- |
| Temporal history recorder | M4 gated `ANIMUS_RECORD_TEMPORAL_HISTORY=1` flow plus committed `WorkflowReplayer` fixture. |
| Real DaVinci Resolve local-MCP integration | M4 or later, gated and workstation-specific. |
| Real OmniVoice local model integration | M4 or later, gated and model-presence-specific. |
| Seedance live provider | Future milestone; no M3 live calls. |
| ElevenLabs live provider | Future milestone; no M3 live calls. |
| Upload-Post live provider | Future milestone with separate ADR; live publish impossible in M3. |

## Security notes

- New providers are disabled by default and fail closed when configuration is
  missing.
- DaVinci MCP tools are allowlisted; arbitrary Python/Lua/script/eval/secret
  tool names are refused.
- DaVinci render start requires both request intent and `AllowStartRender=true`.
- OmniVoice reference voice workflows require explicit consent metadata and
  reference audio hash/allowance.
- Provider capability registry rejects approval authority and live-publish
  claims.
- Default verification requires no network, GUI, DaVinci Resolve install,
  OmniVoice model download, provider credentials, or social upload.

## Runtime/config notes

- DaVinci Resolve MCP: `DaVinciResolveConfig{Enabled: true, Mode: "dry_run" |
  "local_mcp", MCPURL, InputRoot, OutputRoot, ProjectRoot}`. `StartRender`
  requires `AllowStartRender=true`.
- OmniVoice: `OmniVoiceConfig{Enabled: true, Mode: "dry_run" | "local_sidecar",
  BinaryPath, ModelRoot, ModelPath, InputRoot, OutputRoot, RequireConsent}`.
- Provider capabilities: `animus-news provider-capabilities` or `make
  provider-capabilities`.

## Takeover commands

```bash
git status --porcelain
make verify
make verify-m2-local
go vet ./...
go test ./...
make verify-m3
rm -rf build dist
git status --porcelain
```
