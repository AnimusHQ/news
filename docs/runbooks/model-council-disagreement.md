# Model Council Disagreement Runbook

## Purpose

Resolve multimodel disagreement without suppressing dissent or treating model consensus as human approval.

## Severity Guidance

- SEV-1: safety or technical blocker affects public release.
- SEV-2: central thesis or high-risk claim is disputed.
- SEV-3: non-blocking clarity or structure disagreement.
- SEV-4: stylistic disagreement.

## Detection Signals

Council consensus is `revision_required` or `blocked`, dissent list is non-empty, confidence is low, or reviewer notes conflict.

## Immediate Containment

Pause downstream progression for blockers. Preserve all dissenting notes in the decision packet.

## Diagnosis Steps

Compare reviewer role, claim references, source evidence, and artifact version reviewed.

## Resolution Steps

Revise sources, claims, or script. Rerun council only after upstream artifact changes. Send unresolved disagreement to human QA.

## Communication Guidance

Summarize the disagreement plainly: who objected, what artifact is affected, and what evidence is missing.

## Prevention And Follow-Up

Add a council fixture for the disagreement pattern and update routing or prompt policy if needed.

## Artifacts And Logs To Inspect

`multimodel_approval_report.json`, `verification_report.json`, `claims.json`, model routing decision, audit logs.

## Owner Role

Model Council Arbiter.
