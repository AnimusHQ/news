# Security Finding Runbook

## Purpose

Respond to security scanner findings, unsafe content signals, malicious source injection, or policy risks.

## Severity Guidance

- SEV-1: real secret, private data, or unsafe public content detected.
- SEV-2: high-confidence finding in an internal artifact.
- SEV-3: low-confidence scanner warning needing triage.
- SEV-4: false positive requiring fixture or scanner tuning.

## Detection Signals

`scan-secrets` finding, suspicious source text, policy review blocker, audit event, or human report.

## Immediate Containment

Block release and remove the affected artifact from downstream use until triage completes.

## Diagnosis Steps

Identify finding type, file, line, artifact stage, whether the value is real, and whether it reached model context or render output.

## Resolution Steps

Redact, rotate if needed, regenerate affected artifacts, rerun scanning, and record the resolution.

## Communication Guidance

Use redacted values only. Notify Security/Safety Reviewer and artifact owner.

## Prevention And Follow-Up

Add scanner pattern, fixture, or source normalization rule to prevent recurrence.

## Artifacts And Logs To Inspect

Secret scan output, source ingestion notes, audit logs, prompts, generated artifacts, publish pack.

## Owner Role

Safety Reviewer.
