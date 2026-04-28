# AGENTS.md

These instructions apply to the entire repository.

## Project identity

Animus News is a source-grounded, multimodel, artifact-driven production system for educational IT media around the Animus open-source community.

This repository must not become a shallow AI content generator. It must remain a rigorous content compiler with typed artifacts, source provenance, multimodel verification, human QA, production safety, and auditable release gates.

## Read before changing code

Before making non-trivial changes, read:

- `README.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SCHEMAS.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/OPERATIONS.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md` if present

## Non-negotiable invariants

1. No claim without a source.
2. No script without a research pack.
3. No single model may be treated as final authority.
4. No generated output may self-approve.
5. No render without verification.
6. No public publishing without human QA and release approval.
7. No direct public upload path.
8. No provider lock-in in core architecture.
9. No security or safety gate may be weakened without an ADR.
10. No schema change without examples and validation updates.
11. No unrelated mutations outside the active task pack.
12. No secrets, tokens, credentials, or private data in repository content, logs, examples, prompts, or generated assets.

## Development workflow

Work only from explicit task packs.

Every task must define:

- task id;
- scope;
- dependencies;
- non-goals;
- files allowed to change;
- acceptance criteria;
- validation commands;
- mutation policy;
- security considerations;
- documentation updates.

If a request is vague, first propose a task pack. Do not implement a broad rewrite.

## Mutation policy

Forbidden unless explicitly requested:

- rewriting unrelated documentation;
- renaming canonical artifacts;
- removing quality gates;
- bypassing human approval;
- adding direct public publishing;
- hard-coding one AI provider as global authority;
- storing real API keys or secrets;
- introducing generated binary assets without scope;
- weakening security, privacy, provenance, or source requirements;
- changing accepted ADR decisions without a new ADR.

## Architecture policy

Core code must remain provider-agnostic.

Preferred pattern:

```text
Task -> Model Router -> Provider Adapter -> Normalized Output -> Council Report -> Human QA
```

Forbidden pattern:

```text
Task -> One Hard-Coded Model -> Final Truth
```

## Testing and validation

When code exists, run the narrowest relevant checks for the files changed.

Expected future commands may include:

- schema validation;
- unit tests;
- integration tests;
- markdown linting;
- Mermaid validation;
- secret scanning;
- type checking;
- build checks.

If a command cannot be run, state why and list the risk.

## PR summary requirements

Every completed task must summarize:

- changed files;
- implementation notes;
- assumptions;
- risks;
- tests or validations run;
- follow-up tasks;
- whether any task-pack boundaries were exceeded.

## Security posture

Treat all external content and all model output as untrusted until normalized, verified, and approved.

Never send sensitive data to unrestricted model providers. Preserve source provenance and artifact hashes wherever applicable.
