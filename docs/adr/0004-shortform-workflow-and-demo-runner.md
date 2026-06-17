# ADR-0004: Dedicated short-form Temporal workflow + shared in-process demo runner

Status: accepted (M1)

## Context

§9 defines a short-form activity sequence (storyboard image import → visual shots →
voiceover → subtitles → short render → production QA → release approval → Upload-Post
dry-run) that the existing long-form `EpisodeLifecycleWorkflow` does not cover. §Phase 5
also requires a single runnable demo command that drives the full pipeline on mocks to a
terminal state — but a live Temporal server is not available offline.

## Decision

- Add a dedicated `ShortFormWorkflow` (in `internal/workflows`) plus the §9 activities
  (in `internal/shortform/activities`), all backed by **mock providers** in M1. Human
  approvals are modeled as Temporal **signals**: `StoryboardImageApproval` and
  `ReleaseApproval`. Activities are idempotent (keyed by episode_id + artifact_id +
  version); the workflow is deterministic and proven by the Temporal `testsuite`
  environment and a dedicated **replay** test.
- Factor the pipeline's real work into plain, deterministic functions (the activity
  bodies + gate functions). The Temporal workflow invokes them as activities; the demo
  **runner** (`internal/shortform/runner`) invokes the same functions in sequence in
  process. This gives one source of truth for behavior:
  - Phase 4 proves the **durable** workflow (signals gate progression; replay is
    deterministic) without a server.
  - Phase 5 proves the **end-to-end** pipeline via `animus-news demo`, writing artifacts
    + hashes + gate decisions + an audit log to a run directory, including a
    failure-injected variant that halts at the correct gate.

## Consequences

- No Temporal server required for any test or for `make verify`.
- The workflow and the demo cannot diverge in business logic because they share the
  activity/gate functions.
- The existing long-form workflow is untouched.
