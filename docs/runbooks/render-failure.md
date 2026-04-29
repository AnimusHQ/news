# Render Failure Runbook

## Purpose

Recover safely from deterministic preview or render failures without bypassing production QA.

## Severity Guidance

- SEV-1: render contains private data or unsafe misleading visuals.
- SEV-2: final render is wrong, missing assets, or contradicts narration.
- SEV-3: render job fails or output is incomplete.
- SEV-4: cosmetic issue in preview.

## Detection Signals

Missing render output, invalid render manifest, broken asset path, production QA failure, caption or timing mismatch.

## Immediate Containment

Block publish pack generation and release approval until render output and manifest pass validation.

## Diagnosis Steps

Inspect storyboard scene data, asset manifest, render manifest, generated output paths, and renderer logs.

## Resolution Steps

Fix scene or asset metadata, regenerate deterministic preview/render, rerun production QA, then regenerate publish pack if metadata changed.

## Communication Guidance

Tell editor which scenes are affected and whether narration, captions, or assets need revision.

## Prevention And Follow-Up

Add a failing scene fixture or manifest validation test.

## Artifacts And Logs To Inspect

`storyboard.yaml`, `asset_manifest.json`, `render_manifest.json`, renderer logs, production QA report.

## Owner Role

Production Engineer.
