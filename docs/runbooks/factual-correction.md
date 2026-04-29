# Factual Correction Runbook

## Purpose

Handle factual corrections before or after release while preserving source provenance and public trust.

## Severity Guidance

- SEV-1: harmful or materially misleading claim is public.
- SEV-2: central technical claim is wrong.
- SEV-3: local explanation detail is imprecise.
- SEV-4: typo or wording issue.

## Detection Signals

Viewer report, model council dissent, human QA objection, failed verification, or post-publication analytics/comments indicating confusion.

## Immediate Containment

Pause release if unpublished. If public and harmful, unlist or remove according to severity.

## Diagnosis Steps

Trace the claim to `claims.json`, evidence locators, source records, verification notes, and script lines.

## Resolution Steps

Update or remove the claim, add corrected evidence, rerun verification and human QA, then update description, pinned correction, or replacement release as needed.

## Communication Guidance

Be direct. State what was wrong, what changed, and which source supports the correction.

## Prevention And Follow-Up

Add a regression fixture for the claim pattern and update forbidden simplifications if needed.

## Artifacts And Logs To Inspect

`research_pack.json`, `claims.json`, `script.md`, `verification_report.json`, council report, comments, audit events.

## Owner Role

Technical Verifier.
