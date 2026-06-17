# Short-form artifact JSON Schemas

The canonical, load-bearing JSON Schema documents for the short-form artifacts
live next to the code that embeds them, at:

    internal/shortform/schemas/*.schema.json

They are embedded into the binary via `go:embed` (see
`internal/shortform/validate.go`) and validated by the dependency-free interpreter
in `internal/shortform/schema` (see ADR-0003). The schemas are kept inside the
package because `go:embed` cannot reference paths above the embedding source file.

| Artifact | Schema file |
| --- | --- |
| storyboard_image_manifest.json | `internal/shortform/schemas/storyboard_image_manifest.schema.json` |
| visual_shot_manifest.json | `internal/shortform/schemas/visual_shot_manifest.schema.json` |
| voiceover_manifest.json | `internal/shortform/schemas/voiceover_manifest.schema.json` |
| subtitle_manifest.json | `internal/shortform/schemas/subtitle_manifest.schema.json` |
| short_render_manifest.json | `internal/shortform/schemas/short_render_manifest.schema.json` |
| production_candidate.json | `internal/shortform/schemas/production_candidate.schema.json` |
| release_approval.json | `internal/shortform/schemas/release_approval.schema.json` |
| uploadpost_publish_manifest.json | `internal/shortform/schemas/uploadpost_publish_manifest.schema.json` |

Validate a file from the CLI:

    go run ./cmd/animus-news validate-shortform <path-to-artifact>.json
