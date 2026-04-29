# Provider Outage Runbook

## Purpose

Respond safely when model, rendering, publishing, analytics, storage, or workflow providers are unavailable or degraded.

## Severity Guidance

- SEV-1: release gate cannot verify or publish safely.
- SEV-2: critical provider degraded with no approved fallback.
- SEV-3: retryable local worker or sandbox provider issue.
- SEV-4: non-blocking analytics delay.

## Detection Signals

Provider unavailable error, router no-candidate error, degraded health status, retry exhaustion, timeout trend.

## Immediate Containment

Fail closed for verification, QA, publishing, and privacy-sensitive tasks. Do not silently route restricted data to a lower-trust provider.

## Diagnosis Steps

Inspect router decision, provider health, privacy tier, fallback policy, retries, and cost budget.

## Resolution Steps

Use an approved fallback if privacy and policy allow it. Otherwise pause the workflow and request human decision.

## Communication Guidance

State impacted stage, blocked episodes, available fallback, privacy constraints, and estimated delay.

## Prevention And Follow-Up

Add provider health fixture, fallback test, and budget guard if outage triggered repeated retries.

## Artifacts And Logs To Inspect

Model registry, router decisions, audit events, cost reports, workflow state, provider error logs.

## Owner Role

Production Engineer.
