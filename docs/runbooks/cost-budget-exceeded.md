# Cost Budget Exceeded Runbook

## Purpose

Control model, render, publishing, and analytics costs without weakening quality gates.

## Severity Guidance

- SEV-1: runaway cost or repeated paid provider retry.
- SEV-2: episode exceeds approved budget before verification or render.
- SEV-3: warning threshold exceeded.
- SEV-4: reporting discrepancy.

## Detection Signals

Cost budget decision blocks automation, provider cost spike, repeated retries, or unusually high per-stage cost.

## Immediate Containment

Pause non-critical automation and paid retries. Do not skip verification, QA, or safety review to save cost.

## Diagnosis Steps

Aggregate cost by episode, stage, provider, model, and operation type.

## Resolution Steps

Use approved cheaper models for low-risk tasks, batch work, cache unchanged artifacts, or request human budget approval.

## Communication Guidance

Report total cost, stage drivers, recommendation, and whether quality gates are affected.

## Prevention And Follow-Up

Add budget thresholds and regression tests for costly retry paths.

## Artifacts And Logs To Inspect

Cost events, router decisions, audit logs, workflow retry history, provider health status.

## Owner Role

Production Engineer with Editor-in-Chief approval for budget exceptions.
