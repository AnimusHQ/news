# AGENTS.md

These instructions apply to the entire repository.

## Project identity

Animus News is a source-grounded, multimodel, artifact-driven production system for educational IT media around the Animus open-source community.

This repository must not become a shallow AI content generator. It must remain a rigorous content compiler with typed artifacts, source provenance, multimodel verification, human QA, production safety, durable workflow orchestration, and auditable release gates.

## Canonical implementation stack

The canonical production stack for Animus News is:

- **Go / Golang** for core backend services, CLIs, schemas, model routing, verification, QA, publishing adapters, analytics adapters, and production orchestration code.
- **Temporal** for durable workflows, retries, long-running episode lifecycle orchestration, human-in-the-loop waits, provider fallback flows, and replayable production state transitions.
- **Postgres** for durable application state when persistence is introduced.
- **S3-compatible object storage** for artifacts, assets, renders, and immutable evidence bundles when storage is introduced.
- **Provider-agnostic model adapters** for multimodel execution.
- **Deterministic rendering workers** behind workflow activities.

TypeScript/Node.js must not be used as the default backend implementation stack. Any previous TypeScript-oriented task text is superseded by this file and by `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md` if present.

TypeScript may be introduced later only for a frontend/editorial console, Remotion-specific rendering code, or isolated UI tooling through an explicit ADR or task pack.

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
- `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md` if present

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
13. No Temporal workflow may perform nondeterministic side effects directly inside workflow code; side effects belong in activities.
14. No workflow transition may bypass artifact validation, multimodel review where required, human QA, or production QA.

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
- changing accepted ADR decisions without a new ADR;
- replacing Go/Temporal as the production backend/orchestration stack without an ADR.

## Architecture policy

Core code must remain provider-agnostic and workflow-driven.

Preferred model pattern:

```text
Task -> Model Router -> Provider Adapter -> Normalized Output -> Council Report -> Human QA
```

Preferred orchestration pattern:

```text
Temporal Workflow -> Activity -> Typed Artifact -> Validation Gate -> Next Workflow State
```

Forbidden patterns:

```text
Task -> One Hard-Coded Model -> Final Truth
Workflow -> Direct Network Call / Random / Time.Now / File Mutation without Activity
Generated Output -> Public Publish
```

## Go and Temporal engineering rules

- Use idiomatic Go packages with small interfaces.
- Keep workflow code deterministic.
- Put provider calls, file I/O, rendering, publishing, source ingestion, and model calls in activities.
- Use explicit activity timeouts and retry policies.
- Use signals or updates for human-in-the-loop decisions where appropriate.
- Use queries for read-only workflow state inspection where appropriate.
- Make activities idempotent or document compensation/retry behavior.
- Prefer typed request/response structs for activities and workflows.
- Keep provider-specific code behind adapters.
- Keep all artifact validation reusable from both CLI and workflows.

## Testing and validation

When code exists, run the narrowest relevant checks for the files changed.

Expected future commands may include:

- `go test ./...`
- `go test -race ./...` where practical
- `go vet ./...`
- `gofmt` / `go fmt ./...`
- schema validation;
- workflow unit tests with Temporal Go SDK test environment;
- integration tests with mock providers;
- markdown linting;
- Mermaid validation;
- secret scanning;
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
