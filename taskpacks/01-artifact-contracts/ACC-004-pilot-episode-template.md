# ACC-004 — Pilot Episode Artifact Template

## Goal

Create a complete pilot episode artifact directory for the first Animus News episode: **What happens after `git push`?**

This task proves that canonical artifacts can represent a real episode from topic to analytics placeholder.

## Required reading

- `AGENTS.md`
- `docs/SCHEMAS.md`
- `docs/EDITORIAL_STANDARD.md`
- `docs/QUALITY_GATES.md`
- `docs/SYSTEM_BLUEPRINT.md`

## Dependencies

- ACC-003

## Scope

Allowed files:

- `episodes/0001-after-git-push/**`
- `examples/**` only if sharing fixtures
- `docs/ROADMAP.md` only if linking pilot status

## Non-goals

Do not add:

- generated video binaries;
- real TTS audio;
- real uploaded YouTube metadata;
- real API keys;
- unverified claims marked as approved.

## Required files

Create:

```text
episodes/0001-after-git-push/
  topic.yaml
  research_pack.json
  claims.json
  editorial_brief.md
  script.md
  verification_report.json
  multimodel_approval_report.json
  human_qa_report.json
  storyboard.yaml
  asset_manifest.json
  render_manifest.json
  production_qa_report.json
  publish_manifest.json
  analytics_report.json
```

## Content requirements

1. Topic should target the `How It Works` format.
2. Include at least 5 sample technical claims.
3. Each claim must have a source ID and evidence locator placeholder.
4. Mark all placeholder fields clearly.
5. `publish_manifest.visibility` must be `private` or `scheduled`, not `public`.
6. `analytics_report` should be marked as placeholder/draft.
7. `script.md` should be a short skeleton, not a final episode script.
8. Add source entries for public documentation such as Git, GitHub Actions, Docker, Kubernetes, or generic CI/CD docs if used.

## Acceptance criteria

- `validate-episode episodes/0001-after-git-push` passes, or passes with explicit draft flag if draft mode exists.
- Every canonical artifact file is present.
- No generated binaries are committed.
- No secrets or private data.
- No unsupported claim is marked as fully approved unless source locator is adequate.

## Validation commands

```bash
pnpm validate:episode -- episodes/0001-after-git-push
pnpm test
```

## Mutation policy

Forbidden:

- weakening schemas for pilot convenience;
- adding real credentials;
- adding generated media binaries;
- marking placeholders as production-ready.

## PR summary requirements

Include:

- artifact files created;
- validation command result;
- placeholder fields remaining;
- follow-up tasks needed to turn pilot into a real episode.
