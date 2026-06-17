# Claude Code MCP — Operator / Developer Connector

Status: **Planned (operator connector)**. Documented only; **not** a runtime
provider for pilot generation or review.

Studied: https://code.claude.com/docs/ru/mcp (2026-06-17).

## Purpose

Claude Code MCP lets an operator connect their Claude Code session to external
tools and data sources (issue trackers, dashboards, databases, design tools) via
the Model Context Protocol. It is **developer/operator automation**, used while
building and running Animus News — not part of the pilot's content pipeline.

## Critical distinction: Claude API vs Claude Code MCP

| | Claude API (`claude_api_review`) | Claude Code MCP (`claude_code_mcp_operator`) |
| --- | --- | --- |
| What it is | An automated, structured **review/QA provider** | An operator/developer **tooling connector** |
| Where it runs | Inside the pilot (`--claude-review api`) | In a developer's Claude Code session |
| Produces | `claude_*_review_response.json` (gated) | No pilot artifacts |
| Authority | None — pilot gates decide | None — never touches pilot gates |
| In this repo | Implemented + tested | Documented only |

The two must not be conflated. **Claude Code MCP is never used as a hidden
runtime model provider** to generate scripts, visuals, voice, or reviews for an
episode.

## How operators configure it (reference)

```bash
# Remote HTTP server (recommended transport)
claude mcp add --transport http <name> <url>

# Remote SSE server (deprecated transport)
claude mcp add --transport sse <name> <url>

# Local stdio server (-- separates Claude flags from the server command)
claude mcp add --env KEY=value --transport stdio <name> -- <command> [args...]
```

- **Scopes** (`--scope`): `local` (default; only you, this project; `~/.claude.json`),
  `project` (shared via `.mcp.json` in the repo root), `user` (you, all projects).
- **Env**: `--env KEY=value`; startup timeout `MCP_TIMEOUT`; per-server tool
  timeout via a `timeout` field in `.mcp.json`; output cap `MAX_MCP_OUTPUT_TOKENS`.
- Project-scope servers from `.mcp.json` appear as `⏸ Pending approval` until the
  operator approves them interactively.

## Security concerns

- **Prompt injection.** Servers that fetch external content can inject
  instructions. Trust each server before connecting; treat its output as data, not
  commands.
- **Tool allowlisting / least privilege.** Grant only the tools a task needs.
  Approve project-scope servers explicitly.
- **Credentials.** Per-server credentials live in operator config or env — **never
  in the repo**. Do not commit a `.mcp.json` containing secrets.
- **Authority boundary.** MCP output must never become the source of truth for a
  claim, nor an approval/publish authority.

## Relationship to in-repo MCP

The only MCP boundary inside this repository is the **constrained DaVinci Resolve
MCP** render lane (`internal/shortform/providers/mcp`, M3): a tool **allowlist**
and a dry-run client, disabled by default. That is an execution provider for
rendering — distinct from operator tooling. Claude Code MCP, by contrast, is the
human operator's developer connector.

## Future role

- **Review Room** (L4): operator tooling to inspect artifacts and import reviews —
  still calling backend validation, never bypassing gates.
- **DaVinci MCP** (M3+): allowlisted, dry-run-first finishing lane.

## Forbidden patterns

- Using Claude Code MCP to generate or "approve" pilot content.
- Treating MCP tool output as authoritative for factual claims.
- Committing MCP server credentials or secret-bearing `.mcp.json`.
- Granting unrestricted tool access to servers that fetch untrusted content.
