# AGENTS.md

These instructions apply to the entire repository.

## Project identity

Animus News is a source-grounded, multimodel, artifact-driven production system for educational IT media around the Animus open-source community.

This repository must not become a shallow AI content generator. It must remain a rigorous content compiler with typed artifacts, source provenance, multimodel verification, human QA, production safety, durable workflow orchestration, and auditable release gates.

The immediate product priority is to make the system alive through a real CLI pilot that can generate release-candidate short-form videos from prompts using real provider integrations behind safe boundaries.

The long-term product direction is a full production media operating system with CLI, Temporal workflows, Review Room, DaVinci Resolve final studio lane, multimodel verification, real provider connectors, release gates, publishing automation, analytics, and feedback loops.

## Canonical implementation stack

The canonical production stack for Animus News is:

* **Go / Golang** for core backend services, CLIs, schemas, model routing, verification, QA, publishing adapters, analytics adapters, and production orchestration code.
* **Temporal** for durable workflows, retries, long-running episode lifecycle orchestration, human-in-the-loop waits, provider fallback flows, and replayable production state transitions.
* **Postgres** for durable application state when persistence is introduced.
* **S3-compatible object storage** for artifacts, assets, renders, and immutable evidence bundles when storage is introduced.
* **Provider-agnostic adapters** for multimodel execution, media generation, voice, subtitles, render, publishing, storage, QA, analytics, and operator tooling.
* **Deterministic rendering workers** behind workflow activities.
* **DaVinci Resolve** as an optional professional final studio lane, not as the source of truth.
* **FFmpeg** as the default automated render/normalization path.
* **External-command providers** as the fastest safe bridge for real pilot integrations when native provider APIs are not yet implemented.

TypeScript/Node.js must not be used as the default backend implementation stack.

Any previous TypeScript-oriented task text is superseded by this file and by `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md` if present.

TypeScript may be introduced later only for:

* frontend/editorial console;
* Review Room UI;
* Remotion-specific rendering code;
* isolated UI tooling;

and only through an explicit ADR or task pack.

## Launch priority

Until the first real pilot is working, the highest priority is:

```text
Prompt -> CLI episode workspace -> script -> Claude review -> real visual generation -> real voice generation -> subtitles -> FFmpeg render -> Claude final QA -> release-candidate MP4
```

Do not over-optimize future infrastructure before the project can generate a real release-candidate video from CLI.

The first launch slice must produce a real media output, not only contracts, mocks, dry-runs, request files, or architectural documents.

Future architecture must still be documented and preserved, but it must not block the first live CLI pilot.

## Read before changing code

Before making non-trivial changes, read:

* `README.md`
* `AGENTS.md`
* `CLAUDE.md` if present
* `docs/SYSTEM_BLUEPRINT.md`
* `docs/MULTIMODEL_STRATEGY.md`
* `docs/QUALITY_GATES.md`
* `docs/SCHEMAS.md`
* `docs/SECURITY_AND_SAFETY.md`
* `docs/OPERATIONS.md`
* `docs/ARCHITECTURE_DECISIONS.md`
* `docs/CODEX_USAGE.md`
* `docs/CODEX_MASTER_PLAN.md` if present
* `docs/GO_TEMPORAL_IMPLEMENTATION_PLAN.md` if present
* current milestone ledgers and status reports under `docs/ledger/` and `docs/reports/` if present

If any referenced document is missing, record the gap in the task ledger instead of assuming its contents.

## Non-negotiable invariants

1. No claim without a source.
2. No script without an explicit source/research state. For pilot mode, if full research pack is intentionally skipped, the artifact must say so and must not claim source-grounded finality.
3. No single model may be treated as final authority.
4. No generated output may self-approve.
5. No render without verification.
6. No public publishing without human QA and release approval.
7. No direct public upload path.
8. No provider lock-in in core architecture.
9. No security or safety gate may be weakened without an ADR.
10. No schema change without examples and validation updates.
11. No unrelated mutations outside the active task pack.
12. No secrets, tokens, credentials, cookies, private keys, private data, or real API keys in repository content, logs, examples, prompts, generated assets, fixtures, or reports.
13. No Temporal workflow may perform nondeterministic side effects directly inside workflow code; side effects belong in activities.
14. No workflow transition may bypass artifact validation, multimodel review where required, human QA, production QA, or release approval.
15. No provider output may become production output until Animus validates, hashes, gates, and approves it.
16. No live provider, local sidecar, MCP server, or external command may bypass Animus artifact validation.
17. No generated media may be publicly published by a generation command.
18. No arbitrary shell, arbitrary Python, arbitrary Lua, arbitrary MCP tool execution, or arbitrary filesystem path access may be introduced.
19. No task may claim production readiness without a reproducible verification command.
20. No broad task may be implemented linearly by one agent when independent parallel lanes are available.

## Development workflow

Work only from explicit task packs.

Every task must define:

* task id;
* scope;
* dependencies;
* non-goals;
* files allowed to change;
* acceptance criteria;
* validation commands;
* mutation policy;
* security considerations;
* documentation updates;
* expected output;
* final report format.

If a request is vague, first propose a task pack. Do not implement a broad rewrite.

For urgent launch work, prefer a narrow vertical slice that produces a real runnable outcome over broad infrastructure expansion.

## Parallel execution policy

For broad tasks, agents must actively consider parallelization.

A task is considered broad if it includes two or more independent areas such as:

* CLI changes;
* provider implementation;
* Temporal workflow changes;
* artifact/schema changes;
* docs;
* tests;
* security review;
* CI/Makefile;
* connector documentation;
* render pipeline;
* voice pipeline;
* visual generation pipeline;
* publishing pipeline.

When a task is broad, the primary agent must do one of the following before implementation:

1. split the task into parallel lanes and delegate to sub-agents or separate worktrees when the environment supports it; or
2. explicitly explain why parallelization is unsafe or not beneficial.

The fact that the user did not explicitly request parallel agents is not a valid reason to avoid parallelization.

Default rule:

```text
If independent lanes exist, propose and use parallel implementation lanes.
```

Parallelization must increase speed without reducing correctness, reviewability, security, or reproducibility.

## Parallel lane model

Preferred broad-task lane structure:

```text
Lane A: CLI / user-facing workflow
Lane B: provider integration / execution boundary
Lane C: schemas / artifacts / validation gates
Lane D: Temporal workflows / activities
Lane E: docs / connector roadmap / final workflow documentation
Lane F: tests / verification / CI / security scanning
Lane G: integration owner / final merge / takeover verification
```

Not every task needs every lane.

The orchestrator must define:

* lane id;
* lane owner;
* lane scope;
* allowed files;
* forbidden files;
* expected output;
* validation commands;
* integration dependencies;
* merge order;
* conflict risks.

## Parallel worktree and branch policy

When supported by the execution environment, use separate branches or worktrees for parallel lanes.

Preferred naming:

```text
worktree/launch-l1-cli
worktree/launch-l1-providers
worktree/launch-l1-docs
worktree/launch-l1-tests
integration/launch-l1
```

or:

```text
lane/launch-l1-cli
lane/launch-l1-providers
lane/launch-l1-docs
lane/launch-l1-tests
```

Do not let two lanes write the same files unless the integration owner explicitly coordinates the merge.

If two lanes must touch the same files, prefer this order:

1. define shared interfaces first;
2. merge shared interfaces;
3. rebase lanes;
4. implement lane-specific code;
5. integrate with tests.

## File ownership during parallel work

Before parallel implementation, produce a file ownership plan.

Example:

```text
CLI lane:
  cmd/animus-news/**
  internal/shortform/runner/**
  docs/REAL_PILOT_V1.md

Provider lane:
  internal/shortform/providers/**
  internal/shortform/activities/**
  test fixtures for external-command providers

Schema/gate lane:
  internal/shortform/artifacts.go
  internal/shortform/gates/**
  schemas/**

Docs lane:
  docs/CONNECTORS.md
  docs/WORKFLOW_FINAL.md
  docs/CONNECTOR_ROADMAP.md
  docs/PROVIDER_CAPABILITY_MODEL.md

Verification lane:
  Makefile
  CI files
  tests
  secret scan config
```

The final integration owner is responsible for resolving conflicts and running full verification.

## Parallel execution safety rules

Parallel agents must not:

* change canonical schemas independently without coordination;
* weaken gates to make their lane pass;
* duplicate provider interfaces;
* silently introduce incompatible artifact shapes;
* independently rename public CLI commands;
* introduce secrets or local absolute paths;
* add live publishing;
* use mock success to satisfy real-provider requirements;
* claim another lane’s work as implemented;
* skip integration tests;
* leave TODO-only interfaces as completed work.

Every lane must report:

* files changed;
* assumptions;
* risks;
* tests run;
* incomplete items;
* integration needs.

## Integration owner responsibilities

For any parallel task, assign one integration owner.

The integration owner must:

1. collect lane outputs;
2. inspect conflicts;
3. run narrow tests for each lane;
4. run full verification;
5. ensure docs match code;
6. ensure connector status is honest;
7. ensure Implemented / Partial / Planned classifications are correct;
8. ensure no generated assets or secrets are committed;
9. produce final report;
10. leave repository in a takeover-ready state.

The integration owner must not accept lane results that are not reproducible by command.

## Task pack format

Every substantial task pack should use this structure:

```text
Task ID:
Title:
Goal:
Urgency:
Scope:
Non-goals:
Dependencies:
Parallelization plan:
Lanes:
Allowed files:
Forbidden changes:
Acceptance gates:
Validation commands:
Security requirements:
Documentation requirements:
Final report format:
Takeover commands:
```

For launch-critical tasks, include:

```text
Pilot success command:
Expected real output:
Manual checkpoints:
Provider configuration:
Failure behavior:
```

## Current launch milestone model

Until changed by a newer ADR or task pack, the near-term launch sequence is:

```text
M1: typed contracts, schemas, gates, mock providers, Temporal workflow
M2: local execution boundaries, FFmpeg, faster-whisper boundary, Upload-Post dry-run
M3: DaVinci Resolve MCP boundary, OmniVoice boundary, provider capability registry
L1: real CLI pilot from prompt to release-candidate MP4
L2: first real provider wrappers for visual and voice generation
L3: DaVinci final studio lane
L4: private/scheduled publishing
L5: Review Room and production operator workflow
```

The current highest priority is L1 unless explicitly superseded.

## Provider strategy

Core code must remain provider-agnostic.

Provider categories include:

* source/research connectors;
* script/reasoning connectors;
* storyboard/image connectors;
* visual video generation connectors;
* voice/audio connectors;
* subtitle/transcript connectors;
* render/finishing connectors;
* publishing connectors;
* storage/state connectors;
* QA/safety/compliance connectors;
* analytics connectors;
* operator/automation connectors.

All providers must declare capabilities.

Capabilities should include:

* provider id;
* category;
* execution mode;
* enabled by default;
* requires network;
* requires credentials;
* requires local binary;
* requires local model;
* requires GPU;
* requires GUI/session;
* can produce draft/generated artifacts;
* can produce release candidates;
* can publish;
* dry-run support;
* safety gates required before execution;
* safety gates required after execution;
* known risks.

No provider may declare self-approval authority.

No provider may publish live by default.

## External-command provider policy

External-command providers are allowed as the fastest safe bridge for real pilot integrations.

They must:

* be disabled unless explicitly configured;
* use `exec.CommandContext`;
* use argv slices, not shell interpolation;
* use explicit timeouts;
* pass JSON request data through stdin or controlled temp files;
* parse JSON responses;
* validate schemas;
* independently hash output files;
* enforce path containment;
* fail closed on missing config;
* redact secrets in logs;
* record provider metadata;
* produce artifacts with honest status.

External commands must not:

* receive unrestricted environment variables;
* read arbitrary files;
* write outside configured output roots;
* publish;
* self-approve;
* mutate unrelated episode artifacts.

## MCP provider policy

MCP providers are allowed only through restricted tool allowlists.

For DaVinci Resolve or other MCP-connected tools, allowed operations must be manifest-driven.

Allowed pattern:

```text
Animus manifest -> restricted MCP tool -> local editor/render worker -> output artifact -> Animus validation
```

Forbidden MCP patterns:

```text
execute arbitrary Python
execute arbitrary Lua
eval
open arbitrary file
render arbitrary path
delete arbitrary project
read environment
read secrets
publish after render
```

DaVinci Resolve is an optional final studio lane. It edits and exports; Animus validates and approves.

## CLI policy

CLI commands must be explicit, resumable, and artifact-driven.

For pilot generation, preferred commands include:

```text
animus-news pilot generate-real
animus-news pilot resume
animus-news pilot status
animus-news pilot validate
animus-news pilot import-claude-review
animus-news pilot import-visual-shot
animus-news pilot import-voice
```

CLI commands must not silently switch from real providers to mocks.

If a real provider is requested but unavailable, fail with a clear next action.

Mock providers are allowed only in explicit test/demo modes.

## Claude and model review policy

Claude may be used as:

* script reviewer;
* visual plan reviewer;
* final QA reviewer;
* revision planner;
* operator assistant.

Claude must not be treated as:

* automatic final authority;
* hidden browser dependency;
* source of factual truth without sources;
* release approver without structured artifact;
* replacement for validation gates.

Manual Claude review is allowed through structured request/response files.

Do not automate Claude web UI.

Do not scrape browser sessions.

Do not claim Claude approval without a valid response artifact.

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

Preferred pilot pattern:

```text
Prompt -> CLI -> Typed Artifacts -> Provider Boundary -> Generated Assets -> Validation Gates -> Release Candidate
```

Forbidden patterns:

```text
Task -> One Hard-Coded Model -> Final Truth
Workflow -> Direct Network Call / Random / Time.Now / File Mutation without Activity
Generated Output -> Public Publish
Provider Output -> Production Approval
External Command -> Unvalidated Artifact
MCP Tool -> Arbitrary Execution
```

## Go and Temporal engineering rules

* Use idiomatic Go packages with small interfaces.
* Keep workflow code deterministic.
* Put provider calls, file I/O, rendering, publishing, source ingestion, model calls, external commands, and MCP calls in activities.
* Use explicit activity timeouts and retry policies.
* Use signals or updates for human-in-the-loop decisions where appropriate.
* Use queries for read-only workflow state inspection where appropriate.
* Make activities idempotent or document compensation/retry behavior.
* Prefer typed request/response structs for activities and workflows.
* Keep provider-specific code behind adapters.
* Keep all artifact validation reusable from both CLI and workflows.
* Do not introduce package-level mutable global state for workflow decisions.
* Do not perform nondeterministic operations in workflows.
* Do not hide workflow nondeterminism behind helper functions.

## Artifact policy

Artifacts must be:

* typed;
* schema-validated;
* content-hashed;
* linked to upstream source artifacts;
* status-labeled;
* audit-friendly;
* reproducible where practical.

Important statuses include:

```text
draft
generated
needs_review
needs_revision
verified
release_candidate
release_blocked
approved
rejected
superseded
published_dry_run
published_private
published_scheduled
published_public
```

Do not label the main successful pilot output as merely draft.

The first live pilot target is a release candidate, not a draft-only output.

## Connector documentation policy

All future connectors must be documented before or alongside implementation.

Connector docs must classify each connector as:

```text
Implemented
Partial
Planned
```

Every connector entry should include:

* connector id;
* category;
* purpose;
* lifecycle stage;
* execution mode;
* default enabled state;
* network/credential/local binary/model/GUI requirements;
* input artifacts;
* output artifacts;
* gates before execution;
* gates after execution;
* security risks;
* secrets required;
* audit requirements;
* failure behavior;
* test strategy.

Do not mark future connectors as implemented.

## Mutation policy

Forbidden unless explicitly requested:

* rewriting unrelated documentation;
* renaming canonical artifacts;
* removing quality gates;
* bypassing human approval;
* adding direct public publishing;
* hard-coding one AI provider as global authority;
* storing real API keys or secrets;
* introducing generated binary assets without scope;
* weakening security, privacy, provenance, or source requirements;
* changing accepted ADR decisions without a new ADR;
* replacing Go/Temporal as the production backend/orchestration stack without an ADR;
* adding arbitrary execution tools;
* broad rewrites not tied to the active task pack.

## Testing and validation

When code exists, run the narrowest relevant checks for the files changed first, then full verification before final handoff.

Expected commands may include:

* `go test ./...`
* `go test -race ./...` where practical
* `go vet ./...`
* `gofmt` / `go fmt ./...`
* schema validation
* workflow unit tests with Temporal Go SDK test environment
* integration tests with mock providers
* integration tests with fake external-command providers
* CLI tests
* markdown linting
* Mermaid validation
* secret scanning
* build checks
* `make verify`
* milestone-specific verification targets such as `make verify-m2-local`, `make verify-m3`, `make verify-real-pilot`

If a command cannot be run, state why and list the risk.

## Verification hierarchy

Preferred final verification order:

```text
1. narrow package tests
2. changed CLI tests
3. provider boundary tests
4. workflow tests
5. schema validation
6. security/secret scan
7. go vet
8. go test ./...
9. make verify
10. milestone-specific verify target
11. takeover clean-tree check
```

## Takeover standard

Every completed milestone must be reproducible by another engineer from repository state.

Final takeover commands should include:

```bash
git status --porcelain
make verify
go vet ./...
go test ./...
```

and any relevant milestone target, for example:

```bash
make verify-real-pilot
```

After cleanup, the working tree must be clean unless explicitly justified.

Generated media, build outputs, provider downloads, model files, and local episode outputs must not be committed unless specifically scoped as tiny test fixtures.

## PR summary requirements

Every completed task must summarize:

* task id;
* changed files;
* implementation notes;
* assumptions;
* risks;
* tests or validations run;
* follow-up tasks;
* whether any task-pack boundaries were exceeded;
* Implemented / Partial / Planned classification;
* exact takeover commands.

For broad tasks, also summarize:

* parallel lanes used;
* lane owners;
* lane outputs;
* integration conflicts;
* final integration validation.

## Final report classification

Every non-trivial final report must classify work as:

```text
Implemented
Partial
Planned
```

Definitions:

* **Implemented**: code/docs/tests exist and verification command passed.
* **Partial**: some implementation exists, but stated limitations remain.
* **Planned**: documented future work only.

Do not use vague terms like “production-ready”, “complete”, or “done” unless acceptance gates prove it.

## Security posture

Treat all external content, generated content, provider output, MCP output, external command output, and model output as untrusted until normalized, verified, and approved.

Never send sensitive data to unrestricted model providers.

Preserve source provenance and artifact hashes wherever applicable.

Redact secrets from logs.

Use path containment for all file-producing providers.

Use explicit allowlists for MCP tools and external execution.

Never publish from a generation step.

## Production mindset

Animus News should be built as a long-term production system, but the first priority is to make it alive.

Optimize for:

* real runnable vertical slices;
* explicit gates;
* reproducible verification;
* safe provider boundaries;
* fast pilot iteration;
* future extensibility;
* honest documentation.

Avoid:

* mock-only milestones after the launch slice begins;
* broad unverified rewrites;
* provider lock-in;
* hidden manual steps;
* unverifiable claims;
* unsafe shortcuts;
* architecture that blocks the first real video.
