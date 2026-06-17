# ADR-0002: Status model additions and deterministic content hashing

Status: accepted (M1)

## Context

1. The M1 envelope (§7) requires statuses `draft | in_review | approved | rejected |
   superseded | locked`. The repo (`internal/artifacts/types.go`) defines only
   `draft | approved | rejected | superseded`.
2. M1 §7 / §4.10 require a deterministic `content_hash` computed as "sha256 over
   canonicalized content, excluding the hash field itself." The repo only has
   raw-bytes sha256 helpers (`storage.contentHash`, `render.contentHash`,
   `artifacts.FileContentHash`), none of which canonicalize or exclude the hash field.

## Decision

1. **Statuses.** Add `ArtifactStatusInReview = "in_review"` and `ArtifactStatusLocked
   = "locked"` to `internal/artifacts/types.go` and to `validArtifactStatus`. This is
   backward compatible: existing tests only assert `draft`/`approved` and the rejection
   of unknown statuses; `in_review`/`locked` were previously unknown and are now valid.
   `locked` and `approved` are treated as terminal/immutable for mutation gates.

2. **Content hashing.** Add `internal/shortform/contenthash`:
   - `Canonicalize(v any) ([]byte, error)`: marshal to JSON, decode to a generic value,
     and re-encode with object keys sorted recursively (RFC 8785-style canonical key
     ordering; numbers/strings kept as produced by `encoding/json`).
   - `Compute(v any) (string, error)`: canonicalize a copy with any `content_hash`
     field removed, then `"sha256:" + hex(sha256(canonical))`.
   - `Verify(v any, expected string) error`.
   Excluding the `content_hash` field makes the hash stable whether or not the artifact
   already carries its own hash, and immune to object key ordering.

## Consequences

- Deterministic, order-independent hashing provable by round-trip and stability tests.
- A single hashing definition for all short-form artifacts; legacy raw-bytes hashing in
  `storage` is unchanged (it hashes opaque stored bytes, a different concern).
- Mutation/immutability gates can treat `approved`/`locked` as frozen.
