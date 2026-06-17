# ADR-0001: Extend the canonical artifact graph in a dedicated short-form package

Status: accepted (M1)

## Context

M1 adds eight short-form artifacts (storyboard_image_manifest, visual_shot_manifest,
voiceover_manifest, subtitle_manifest, short_render_manifest, production_candidate,
release_approval, uploadpost_publish_manifest). The existing `internal/artifacts`
package owns a stable 14-file canonical episode bundle that `episode-0001` and many
tests depend on. Adding the new files to `RequiredEpisodeFiles` would break existing
episode validation (the legacy bundle does not contain them).

## Decision

Implement the new artifacts, their JSON Schemas, validators, and content hashing in a
dedicated `internal/shortform` package tree that **imports** `internal/artifacts` for
the shared envelope (`Metadata`), `Source`, and `Claim` types. The short-form
artifacts **extend** the graph: they reference upstream `source_artifacts` and reuse
the common envelope, but they are not added to the required long-form bundle.

The existing `EpisodeLifecycleWorkflow`, canonical validators, and `RequiredEpisodeFiles`
are left untouched except for two backward-compatible additions (the new statuses in
ADR-0002).

## Consequences

- No regression risk to the existing canonical bundle or its tests.
- Clear ownership boundary: long-form canonical graph in `internal/artifacts`;
  short-form execution contracts in `internal/shortform`.
- The short-form artifacts are still validated by the same envelope rules (schema_version,
  episode_id, artifact_id, status, content_hash prefix, RFC3339 created_at) by reusing
  the shared types.
