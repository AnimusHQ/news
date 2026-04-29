# Private Data Exposure Runbook

## Purpose

Contain and remediate any exposure of secrets, credentials, private user data, embargoed content, or internal-only material.

## Severity Guidance

- SEV-1: credential or private data is public or sent to an unrestricted provider.
- SEV-2: sensitive material is present in an internal artifact or log.
- SEV-3: low-confidence scanner finding needs review.
- SEV-4: documentation cleanup only.

## Detection Signals

Secret scanner finding, human report, provider policy alert, unexpected private text in render or publish pack.

## Immediate Containment

Stop publication. Remove exposed artifact from release flow. Rotate affected credentials if any real secret may have leaked.

## Diagnosis Steps

Identify source artifact, logs, model prompts, render outputs, and publish metadata containing the sensitive value.

## Resolution Steps

Redact values, regenerate affected artifacts, rerun scan, rerun QA, and record an audit event.

## Communication Guidance

Notify security owner and affected integration owner. Do not paste the secret into chat, issue text, logs, or model prompts.

## Prevention And Follow-Up

Add scanner pattern coverage and update artifact handling to avoid recurrence.

## Artifacts And Logs To Inspect

Secret scan output, audit logs, generated descriptions, prompt-like files, render manifests, publish packs.

## Owner Role

Safety Reviewer with Production Engineer support.
