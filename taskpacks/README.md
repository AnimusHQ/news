# Codex Task Packs

This directory contains executable task packs for building Animus News into a production-grade, source-grounded, multimodel content compiler.

Each task pack is designed to be copied into Codex as a bounded implementation request.

## Required Codex procedure

Before executing any task, Codex must read:

- `AGENTS.md`
- `README.md`
- `docs/SYSTEM_BLUEPRINT.md`
- `docs/MULTIMODEL_STRATEGY.md`
- `docs/QUALITY_GATES.md`
- `docs/SECURITY_AND_SAFETY.md`
- `docs/SCHEMAS.md`
- `docs/ARCHITECTURE_DECISIONS.md`
- `docs/CODEX_USAGE.md`
- `docs/CODEX_MASTER_PLAN.md`

## Task execution rule

Codex must execute one task pack per branch/PR unless explicitly instructed otherwise.

## Task pack groups

```mermaid
flowchart TD
  A[00-foundation] --> B[01-artifact-contracts]
  B --> C[02-model-layer]
  C --> D[03-council-verification]
  B --> E[04-workflow]
  D --> F[05-research-claims-qa]
  E --> F
  F --> G[06-production-rendering]
  G --> H[07-publishing]
  H --> I[08-analytics]
  I --> J[09-hardening]
```

## Status tracking

Task status should be tracked in issues or project board later. For now, task packs are the canonical execution backlog.
