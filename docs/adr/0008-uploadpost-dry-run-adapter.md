# ADR-0008: Upload-Post Dry-Run Adapter

Status: accepted.

## Context

M1 represented Upload-Post publishing with a deterministic mock manifest. M2
needs a real-shaped adapter contract that can validate publishing intent while
keeping live upload and scheduling impossible.

## Decision

Add an Upload-Post dry-run provider under
`internal/shortform/providers/uploadpost`. The provider implements
`providers.PublishingProvider` and is disabled by default.

The provider accepts only `dry_run` mode. Any live, schedule, or publish mode
returns a clear M2 refusal error. Dry-run does not require an API key and does
not perform a network call.

Before constructing an `uploadpost_publish_manifest`, the adapter requires:

- a valid approved render artifact;
- approved render outputs;
- production QA decision `approved`;
- a valid approved release approval;
- human release approval recorded;
- explicit supported platforms;
- explicit visibility and schedule fields where applicable;
- AI disclosure text when disclosure is required;
- production QA and release references.

## Consequences

- The repository still has no live public publishing path in M2.
- Direct generated-output-to-publish remains impossible: render, QA, release, and
  disclosure checks are required before dry-run request construction.
- API keys, if present in local configuration, are not required for tests and are
  redacted from diagnostics.
- Future live Upload-Post support must be a new milestone and must not bypass
  production QA or human release approval.

