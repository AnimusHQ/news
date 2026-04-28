# Codex Task Pack Template

Use this template for every Codex implementation task.

A task pack is the contract between the human operator and Codex. It defines exactly what may change, what must not change, how success is verified, and how review should happen.

```yaml
task_id: "ACC-000"
title: "Short imperative title"
phase: "0 - Foundation"
status: "draft | ready | in_progress | blocked | done"
priority: "P0 | P1 | P2 | P3"
risk: "low | medium | high | critical"
owner: "human | codex | mixed"

objective: |
  Explain the engineering purpose of this task in one or two paragraphs.

context:
  required_reading:
    - AGENTS.md
    - README.md
    - docs/SYSTEM_BLUEPRINT.md
    - docs/MULTIMODEL_STRATEGY.md
    - docs/QUALITY_GATES.md
    - docs/SECURITY_AND_SAFETY.md
    - docs/SCHEMAS.md
    - docs/ARCHITECTURE_DECISIONS.md
    - docs/CODEX_USAGE.md
    - docs/CODEX_MASTER_PLAN.md
  upstream_tasks:
    - "ACC-000"
  downstream_tasks:
    - "ACC-001"

scope:
  allowed_paths:
    - "src/**"
    - "tests/**"
  forbidden_paths:
    - "docs/SYSTEM_BLUEPRINT.md"
    - "docs/ARCHITECTURE_DECISIONS.md"
  allowed_operations:
    - create
    - update
  forbidden_operations:
    - delete
    - rename_canonical_artifacts
    - direct_public_publish_path

non_goals:
  - "Explicitly list what must not be implemented."

requirements:
  functional:
    - "What the code or docs must do."
  security:
    - "Security requirements."
  privacy:
    - "Privacy and data classification requirements."
  reliability:
    - "Reliability/idempotency requirements."
  observability:
    - "Logging/metrics/audit requirements."
  documentation:
    - "Docs to update or preserve."

interfaces:
  inputs:
    - name: "input_name"
      type: "type or schema"
      required: true
  outputs:
    - name: "output_name"
      type: "type or schema"
      required: true
  errors:
    - code: "ERROR_CODE"
      when: "condition"

acceptance_criteria:
  - "Concrete pass/fail criterion."

tests:
  required:
    - name: "test name"
      purpose: "what it proves"
  commands:
    - "pnpm test"
    - "pnpm typecheck"

validation:
  required_commands:
    - "pnpm validate"
  manual_checks:
    - "Human reviewer verifies no unrelated mutations."

mutation_policy:
  forbidden:
    - "No unrelated rewrites."
    - "No provider lock-in."
  allowed_if_justified:
    - "Small docs alignment updates."

codex_prompt: |
  You are implementing one bounded Animus News task pack.

  Read first:
  - AGENTS.md
  - README.md
  - docs/SYSTEM_BLUEPRINT.md
  - docs/MULTIMODEL_STRATEGY.md
  - docs/QUALITY_GATES.md
  - docs/SECURITY_AND_SAFETY.md
  - docs/SCHEMAS.md
  - docs/ARCHITECTURE_DECISIONS.md
  - docs/CODEX_USAGE.md
  - docs/CODEX_MASTER_PLAN.md

  Implement only this task pack:
  <paste task pack>

  Rules:
  - Stay within allowed paths unless absolutely necessary.
  - Do not weaken security, quality gates, source provenance, multimodel independence, or human QA.
  - Do not introduce real secrets or credentials.
  - Add tests.
  - Run relevant checks.
  - Return changed files, commands run, assumptions, risks, and follow-ups.

review_packet:
  codex_must_report:
    - changed_files
    - commands_run
    - assumptions
    - risks
    - follow_ups
    - boundary_exceptions
  human_reviewer_must_check:
    - scope_compliance
    - tests_passed
    - no_security_regression
    - no_architecture_regression
    - no_provider_lock_in
```

## Task pack quality rules

A task pack is not ready for Codex unless:

- scope is explicit;
- forbidden paths are listed;
- non-goals are listed;
- acceptance criteria are testable;
- validation commands are known or intentionally deferred;
- security and privacy requirements are stated;
- review expectations are clear.

## Bad task examples

Bad:

```text
Implement the whole production system.
```

Bad:

```text
Make the AI pipeline better.
```

Bad:

```text
Add model support and publishing and rendering.
```

## Good task example

Good:

```text
Implement ACC-003 artifact validation CLI.
Allowed paths: src/cli/**, src/schemas/**, tests/cli/**, package.json.
Acceptance: valid examples pass, invalid examples fail, validate-episode rejects missing required artifacts.
Non-goals: no workflow engine, no model providers, no rendering.
```
