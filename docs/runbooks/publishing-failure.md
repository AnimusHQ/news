# Publishing Failure Runbook

## Purpose

Handle private draft upload, scheduled release, metadata validation, or dry-run publishing failures.

## Severity Guidance

- SEV-1: unintended public publication or unsafe metadata exposure.
- SEV-2: scheduled release cannot be completed safely.
- SEV-3: private draft or dry-run adapter failed.
- SEV-4: non-blocking metadata warning.

## Detection Signals

Publishing adapter error, visibility not allowed, missing human approval, invalid metadata, platform processing failure.

## Immediate Containment

Keep draft private. Disable or pause schedule. Do not retry with public visibility to bypass the issue.

## Diagnosis Steps

Inspect publish pack, publish manifest, human release approval, adapter response, and dry-run result.

## Resolution Steps

Fix metadata, confirm visibility is private or scheduled, obtain release approval, rerun dry-run or private upload.

## Communication Guidance

Notify Editor-in-Chief and Production Engineer with draft status, visibility, and required approval state.

## Prevention And Follow-Up

Add adapter test for the failed visibility or metadata condition.

## Artifacts And Logs To Inspect

`publish_manifest.json`, publish pack, adapter response, audit events, production QA report.

## Owner Role

Production Engineer.
