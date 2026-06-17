# ADR-0009: DaVinci Resolve MCP Render Provider Boundary

Status: accepted.

## Context

M3 adds an optional professional finishing lane for flagship short-form videos.
DaVinci Resolve may be useful for timeline assembly, color correction, audio
finishing, project archive export, and operator-polished final renders.

DaVinci Resolve must not become workflow authority, artifact authority, QA
authority, release authority, or publish authority.

## Decision

Add a disabled-by-default DaVinci Resolve MCP render provider under
`internal/shortform/providers/render/davinci` and a constrained MCP helper under
`internal/shortform/providers/mcp`.

The provider implements `providers.RenderProvider` and supports:

- `disabled`;
- `dry_run`;
- `local_mcp` with an explicit client.

M3 default tests use only dry-run/stub clients. No DaVinci Resolve install, GUI
session, MCP server, or network access is required for default verification.

The MCP boundary allows only these conceptual Resolve tools:

- `resolve.healthcheck`;
- `resolve.create_project_from_manifest`;
- `resolve.import_assets_from_manifest`;
- `resolve.create_vertical_timeline`;
- `resolve.queue_render_from_manifest`;
- `resolve.start_render_if_allowed`;
- `resolve.get_render_status`;
- `resolve.collect_render_result`;
- `resolve.export_project_archive`.

Any non-allowlisted tool is refused. Render start is refused unless the request
sets `StartRender` and config sets `AllowStartRender=true`.

## Consequences

- DaVinci output remains a draft `short_render_manifest`.
- Output cannot bypass production QA, release approval, or publish gates.
- Workflow code does not call MCP or DaVinci directly.
- Real DaVinci Resolve execution remains a future gated integration task.
- Capability metadata marks DaVinci as optional, GUI/MCP-dependent, disabled by
  default, dry-run capable, and unable to publish or approve artifacts.

