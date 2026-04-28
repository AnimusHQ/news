# ACC-003 — Artifact Validation CLI

## Goal

Create a CLI that validates individual artifacts and complete episode directories against canonical schemas.

This CLI is the first concrete enforcement mechanism for artifact-driven development.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/QUALITY_GATES.md`
- `docs/CODEX_MASTER_PLAN.md`

## Dependencies

- ACC-002

## Scope

Allowed files:

- `src/cli/**`
- `src/schemas/**`
- `src/artifacts/**`
- `tests/cli/**`
- `examples/**`
- `package.json` only for CLI script wiring

## Non-goals

Do not implement:

- workflow state transitions;
- model calls;
- rendering;
- publishing;
- database storage.

## Commands

Implement:

```bash
animus-news validate <path>
animus-news validate-episode <episode-dir>
```

Package scripts may expose:

```bash
pnpm validate -- <path>
pnpm validate:episode -- <episode-dir>
```

## Requirements

1. Infer artifact type by filename when possible.
2. Validate JSON and YAML where relevant.
3. Validate complete episode directory includes required canonical artifacts.
4. Emit readable human errors.
5. Support `--json` for machine-readable output.
6. Exit with non-zero status on validation failure.
7. Include tests for valid and invalid artifacts.
8. Do not silently skip unknown files unless documented.

## Episode validation requirements

A complete episode should include:

- `topic.yaml` or equivalent topic artifact;
- `research_pack.json`;
- `claims.json`;
- `script.md` may be checked for existence but not schema-validated unless frontmatter is added;
- `verification_report.json`;
- `multimodel_approval_report.json`;
- `human_qa_report.json`;
- `storyboard.yaml`;
- `asset_manifest.json`;
- `render_manifest.json`;
- `production_qa_report.json`;
- `publish_manifest.json`;
- `analytics_report.json`.

For incomplete templates, allow an explicit `--allow-draft` flag if implemented.

## Acceptance criteria

- Valid examples pass.
- Invalid examples fail with clear messages.
- Complete episode directory validates.
- Missing required artifact fails.
- `--json` output is deterministic.
- Tests cover at least:
  - valid artifact;
  - invalid artifact;
  - unknown artifact;
  - missing episode file;
  - malformed YAML/JSON.

## Validation commands

```bash
pnpm test
pnpm typecheck
pnpm validate -- examples/valid/topic.yaml
```

## Mutation policy

Forbidden:

- weakening schemas to make examples pass;
- treating validation warnings as success for required fields;
- adding network calls;
- adding model calls.

## PR summary requirements

Include:

- CLI commands added;
- validation behavior;
- tests run;
- known limitations;
- draft-mode behavior if implemented.
