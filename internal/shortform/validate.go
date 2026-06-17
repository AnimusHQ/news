package shortform

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AnimusHQ/news/internal/shortform/contenthash"
	"github.com/AnimusHQ/news/internal/shortform/schema"
)

//go:embed schemas/*.schema.json
var schemaFS embed.FS

var (
	schemasOnce     sync.Once
	compiledSchemas map[string]*schema.Schema
	schemaLoadErr   error
)

func loadSchemas() (map[string]*schema.Schema, error) {
	schemasOnce.Do(func() {
		compiledSchemas = map[string]*schema.Schema{}
		entries, err := schemaFS.ReadDir("schemas")
		if err != nil {
			schemaLoadErr = err
			return
		}
		for _, entry := range entries {
			name := entry.Name()
			data, err := schemaFS.ReadFile("schemas/" + name)
			if err != nil {
				schemaLoadErr = err
				return
			}
			compiled, err := schema.Compile(data)
			if err != nil {
				schemaLoadErr = fmt.Errorf("compile %s: %w", name, err)
				return
			}
			compiledSchemas[strings.TrimSuffix(name, ".schema.json")] = compiled
		}
	})
	return compiledSchemas, schemaLoadErr
}

// SchemaFor returns the compiled schema for an artifact kind.
func SchemaFor(kind string) (*schema.Schema, error) {
	schemas, err := loadSchemas()
	if err != nil {
		return nil, err
	}
	s, ok := schemas[kind]
	if !ok {
		return nil, fmt.Errorf("no schema registered for kind %q", kind)
	}
	return s, nil
}

// KnownKinds returns the registered short-form artifact kinds.
func KnownKinds() []string {
	schemas, err := loadSchemas()
	if err != nil {
		return nil
	}
	kinds := make([]string, 0, len(schemas))
	for kind := range schemas {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	return kinds
}

// Validate checks an artifact against its JSON Schema plus envelope and content
// hash integrity, returning a sorted list of human-readable issues. An empty
// slice means valid.
func Validate(a Artifact) []string {
	var issues []string
	s, err := SchemaFor(a.Kind())
	if err != nil {
		return []string{err.Error()}
	}
	issues = append(issues, schema.ValidateValue(s, a)...)
	issues = append(issues, validateEnvelopeSemantics(a.EnvelopeRef())...)
	issues = append(issues, validateKindSemantics(a)...)
	sort.Strings(issues)
	return issues
}

// ValidateBytes validates raw JSON for a given artifact kind against its schema
// only (no Go-typed semantic checks). Used by the CLI for single-file checks.
func ValidateBytes(kind string, data []byte) []string {
	s, err := SchemaFor(kind)
	if err != nil {
		return []string{err.Error()}
	}
	return schema.ValidateBytes(s, data)
}

// KindFromFilename maps an artifact filename (e.g. visual_shot_manifest.json) to
// its kind. Returns "" if it is not a short-form artifact.
func KindFromFilename(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), ".json")
	for _, kind := range KnownKinds() {
		if kind == base {
			return kind
		}
	}
	return ""
}

// ValidateFile reads a short-form artifact file and validates it against its
// schema. Unknown filenames return an issue rather than silently passing.
func ValidateFile(path string) []string {
	kind := KindFromFilename(path)
	if kind == "" {
		return []string{fmt.Sprintf("%s: not a recognized short-form artifact", filepath.Base(path))}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read: %v", filepath.Base(path), err)}
	}
	issues := ValidateBytes(kind, data)
	// Also verify content hash integrity for files that carry one.
	var env Envelope
	if json.Unmarshal(data, &env) == nil && env.ContentHash != "" {
		var generic map[string]any
		if json.Unmarshal(data, &generic) == nil {
			if got, err := contenthash.Compute(generic); err == nil && got != env.ContentHash {
				issues = append(issues, fmt.Sprintf("%s: content_hash mismatch (recomputed %s)", filepath.Base(path), got))
			}
		}
	}
	sort.Strings(issues)
	return issues
}

var createdByPattern = regexp.MustCompile(`^(system|human)$|^(system|human|model):.+`)

func validateEnvelopeSemantics(env *Envelope) []string {
	var issues []string
	if env.SchemaVersion != SchemaVersion {
		issues = append(issues, fmt.Sprintf("schema_version must be %q", SchemaVersion))
	}
	if env.CreatedAt != "" {
		if _, err := time.Parse(time.RFC3339, env.CreatedAt); err != nil {
			issues = append(issues, "created_at must be RFC3339")
		}
	}
	if env.CreatedBy != "" && !createdByPattern.MatchString(env.CreatedBy) {
		issues = append(issues, "created_by must be system, human, or <human|system|model>:<id>")
	}
	if env.ContentHash != "" && !strings.HasPrefix(env.ContentHash, contenthash.Prefix) {
		// Full-artifact hash integrity is verified by ValidateFile (which has the
		// serialized bytes) and by Stamp round-trip tests; here we only flag an
		// obviously malformed prefix.
		issues = append(issues, "content_hash must use sha256: prefix")
	}
	return issues
}

func validateKindSemantics(a Artifact) []string {
	switch v := a.(type) {
	case *StoryboardImageManifest:
		return uniqueSceneIDs(sceneIDsFromImages(v.Images))
	case *VisualShotManifest:
		issues := uniqueSceneIDs(sceneIDsFromShots(v.Shots))
		if v.AspectRatio != v.RenderTarget.Aspect {
			issues = append(issues, "aspect_ratio must match render_target.aspect")
		}
		return issues
	case *ProductionCandidate:
		var issues []string
		if v.Status != string(statusLocked) {
			issues = append(issues, "production_candidate must have status locked")
		}
		if !v.Immutable {
			issues = append(issues, "production_candidate must be immutable")
		}
		return issues
	default:
		return nil
	}
}

// statusLocked mirrors artifacts.ArtifactStatusLocked without importing it here.
const statusLocked = "locked"

func sceneIDsFromImages(images []StoryboardImage) []string {
	out := make([]string, 0, len(images))
	for _, img := range images {
		out = append(out, img.SceneID)
	}
	return out
}

func sceneIDsFromShots(shots []VisualShot) []string {
	out := make([]string, 0, len(shots))
	for _, shot := range shots {
		out = append(out, shot.SceneID)
	}
	return out
}

func uniqueSceneIDs(ids []string) []string {
	seen := map[string]bool{}
	var issues []string
	for _, id := range ids {
		if seen[id] {
			issues = append(issues, fmt.Sprintf("duplicate scene_id %q", id))
		}
		seen[id] = true
	}
	return issues
}
