# Release Checklist Runbook

## Purpose

Guide a private or scheduled Animus News release through validation, QA, approval, and safe dry-run publishing gates.

## Severity Guidance

- SEV-1: private data, unsafe content, or unsupported high-risk claim is present.
- SEV-2: release metadata or source list is materially wrong.
- SEV-3: render, captions, or publish pack has a production defect.
- SEV-4: minor copy issue with no public safety impact.

## Detection Signals

- `validate-episode` fails.
- Dry-run reports blockers.
- Production QA decision is not approved.
- Human release approval is missing.
- Publish visibility is public by default.

## Immediate Containment

Stop release scheduling. Keep any uploaded draft private. Do not retry public publication until blockers are resolved.

## Diagnosis Steps

Inspect the episode directory, validation issues, council report, human QA report, production QA report, and publish manifest.

## Resolution Steps

Fix invalid artifacts, rerun verification, obtain explicit human QA and release approval, then rerun dry-run publishing.

## Communication Guidance

Tell the editor, technical verifier, and production engineer what blocked release and which artifact owns the fix.

## Prevention And Follow-Up

Add a regression test or checklist item for the failed gate.

## Artifacts And Logs To Inspect

`claims.json`, `verification_report.json`, `multimodel_approval_report.json`, `human_qa_report.json`, `production_qa_report.json`, `publish_manifest.json`, audit logs, dry-run summary.

## Owner Role

Editor-in-Chief with Production Engineer support.
