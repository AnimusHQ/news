# ACC-003 — Artifact Validation CLI

## Status

Ready after ACC-002.

## Priority

P0.

## Risk

Medium.

## Objective

Implement a CLI that validates single artifact files and complete episode directories against the canonical schemas. This task turns the schemas from static contracts into an enforceable developer and CI tool.

The CLI is foundational: future workflow state transitions, dry runs, QA reports, and CI checks should depend on it.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/QUALITY_GATES.md`
- `docs/CODEX_MASTER_PLAN.md`
- `docs/codex/tasks/ACC-002-canonical-schemas.md`

## Dependencies

- ACC-000
- ACC-002

## Allowed paths

- `src/cli/**`
- `src/schemas/**`
- `src/artifacts/**`
- `tests/cli/**`
- `tests/fixtures/**`
- `examples/artifacts/**`
- `package.json`
- `docs/codex/tasks/ACC-003-artifact-validation-cli.md` only for status notes if needed

## Forbidden paths

- `docs/SYSTEM_BLUEPRINT.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `src/models/**`
- `src/render/**`
- `src/publishing/**`
- `.github/**` unless only adding an existing validation command to docs is impossible without CI changes

## Non-goals

- Do not implement the episode workflow engine.
- Do not implement model routing.
- Do not implement research generation.
- Do not implement rendering.
- Do not implement publishing.
- Do not call network services.

## Functional requirements

Add CLI commands:

```text
animus-news validate <path>
animus-news validate-episode <episode-dir>
```

Command behavior:

1. `validate <path>`:
   - infer artifact type from filename when possible;
   - validate JSON/YAML/Markdown-backed artifacts where applicable;
   - print readable errors;
   - exit non-zero on failure.
2. `validate-episode <episode-dir>`:
   - check required canonical artifact files;
   - validate machine-readable files;
   - verify basic dependency presence;
   - fail if required artifacts are missing.
3. Support optional JSON output:

```text
animus-news validate <path> --json
```

4. Add package script:

```text
pnpm validate
```

If no episode exists yet, `pnpm validate` may validate examples only, but this must be documented in output or package script comments.

## Required canonical episode files

The complete episode directory should eventually require:

- `topic.yaml`
- `research_pack.json`
- `claims.json`
- `editorial_brief.md`
- `script.md`
- `verification_report.json`
- `multimodel_approval_report.json`
- `human_qa_report.json`
- `storyboard.yaml`
- `asset_manifest.json`
- `render_manifest.json`
- `production_qa_report.json`
- `publish_manifest.json`
- `analytics_report.json`

For ACC-003, Markdown files may be checked for existence only.

## Security requirements

- Do not execute artifact content.
- Treat artifact files as data only.
- Do not follow remote URLs.
- Do not expand untrusted shell commands.
- Avoid unsafe YAML parsing features.

## Acceptance criteria

- CLI validates a valid artifact.
- CLI rejects an invalid artifact.
- CLI validates a directory containing all required files.
- CLI rejects missing required files.
- `--json` output is machine-readable.
- Errors include path and reason.
- Tests cover success and failure cases.

## Required tests

- valid single artifact passes;
- invalid single artifact fails;
- unknown artifact type fails clearly;
- complete episode directory passes with fixtures;
- incomplete episode directory fails;
- JSON output includes status and errors.

## Required validation commands

Codex must run or explain inability to run:

```bash
pnpm test
pnpm typecheck
pnpm validate
```

## Mutation policy

Forbidden:

- evaluating or executing artifact contents;
- adding network access;
- adding workflow transitions;
- adding model provider calls;
- adding publishing behavior.

Allowed:

- CLI dependencies if lightweight and justified;
- YAML parser if safe;
- artifact helper utilities.

## Codex prompt

```text
You are implementing ACC-003 — Artifact Validation CLI for Animus News.

Read first:
- AGENTS.md
- docs/SCHEMAS.md
- docs/QUALITY_GATES.md
- docs/CODEX_MASTER_PLAN.md
- docs/codex/tasks/ACC-002-canonical-schemas.md

Implement only the validation CLI and tests. Do not implement workflow, model routing, rendering, publishing, or network ingestion.

Run:
- pnpm test
- pnpm typecheck
- pnpm validate

Return changed files, commands run, assumptions, risks, and follow-ups.
```
