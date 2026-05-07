package render

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"path"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/storyboard"
)

const (
	SchemaVersion   = "1.0"
	fileStatus      = "draft"
	rendererName    = "local-html-preview"
	rendererVersion = "0.1.0"
)

// Input contains the deterministic data needed to generate a local preview.
type Input struct {
	EpisodeID  string
	Storyboard storyboard.File
	OutputDir  string
}

// Result contains generated preview content plus manifest artifacts.
type Result struct {
	Preview        Preview        `json:"preview"`
	AssetManifest  AssetManifest  `json:"asset_manifest"`
	RenderManifest RenderManifest `json:"render_manifest"`
}

// Preview is a deterministic local preview output. Callers decide whether to
// persist Content to Path.
type Preview struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Hash    string `json:"hash"`
}

// AssetManifest is the generated placeholder asset provenance artifact.
type AssetManifest struct {
	SchemaVersion string  `json:"schema_version"`
	EpisodeID     string  `json:"episode_id"`
	ArtifactID    string  `json:"artifact_id"`
	Status        string  `json:"status"`
	Assets        []Asset `json:"assets"`
}

// Asset records deterministic placeholder asset provenance.
type Asset struct {
	AssetID     string   `json:"asset_id"`
	Type        string   `json:"type"`
	Path        string   `json:"path"`
	GeneratedBy string   `json:"generated_by"`
	License     string   `json:"license"`
	Hash        string   `json:"hash"`
	Provenance  []string `json:"provenance,omitempty"`
}

// RenderManifest is the generated render_manifest.json shape.
type RenderManifest struct {
	SchemaVersion   string         `json:"schema_version"`
	EpisodeID       string         `json:"episode_id"`
	ArtifactID      string         `json:"artifact_id"`
	Status          string         `json:"status"`
	Renderer        string         `json:"renderer"`
	RendererVersion string         `json:"renderer_version"`
	Inputs          []string       `json:"inputs"`
	Outputs         []RenderOutput `json:"outputs"`
}

// RenderOutput describes one deterministic preview output.
type RenderOutput struct {
	Type            string `json:"type"`
	Path            string `json:"path"`
	DurationSeconds int    `json:"duration_seconds"`
	Resolution      string `json:"resolution"`
	Hash            string `json:"hash"`
}

// GeneratePreview produces a deterministic local HTML preview and manifests.
func GeneratePreview(input Input) (Result, error) {
	episodeID := strings.TrimSpace(input.EpisodeID)
	if episodeID == "" {
		episodeID = strings.TrimSpace(input.Storyboard.EpisodeID)
	}
	if episodeID == "" {
		return Result{}, fmt.Errorf("episode id is required")
	}
	if err := validateStoryboard(input.Storyboard); err != nil {
		return Result{}, err
	}
	if input.Storyboard.EpisodeID != "" && input.Storyboard.EpisodeID != episodeID {
		return Result{}, fmt.Errorf("storyboard episode id %s does not match render episode id %s", input.Storyboard.EpisodeID, episodeID)
	}

	outputDir := strings.Trim(strings.ReplaceAll(input.OutputDir, "\\", "/"), "/")
	if outputDir == "" {
		outputDir = "dist"
	}
	previewPath := path.Join(outputDir, episodeID+"-preview.html")
	content := renderHTMLPreview(episodeID, input.Storyboard.Scenes)
	previewHash := contentHash(content)
	assetManifest := generateAssetManifest(episodeID, input.Storyboard.Scenes)

	return Result{
		Preview: Preview{
			Path:    previewPath,
			Content: content,
			Hash:    previewHash,
		},
		AssetManifest: assetManifest,
		RenderManifest: RenderManifest{
			SchemaVersion:   SchemaVersion,
			EpisodeID:       episodeID,
			ArtifactID:      "render-manifest-" + episodeID + "-preview-v1",
			Status:          fileStatus,
			Renderer:        rendererName,
			RendererVersion: rendererVersion,
			Inputs:          []string{"storyboard.yaml", "asset_manifest.json"},
			Outputs: []RenderOutput{
				{
					Type:            "html_preview",
					Path:            previewPath,
					DurationSeconds: durationSeconds(input.Storyboard.Scenes),
					Resolution:      "responsive-html",
					Hash:            previewHash,
				},
			},
		},
	}, nil
}

func validateStoryboard(file storyboard.File) error {
	if len(file.Scenes) == 0 {
		return fmt.Errorf("storyboard must contain at least one scene")
	}
	seen := map[string]bool{}
	for _, scene := range file.Scenes {
		if strings.TrimSpace(scene.SceneID) == "" {
			return fmt.Errorf("scene_id is required")
		}
		if seen[scene.SceneID] {
			return fmt.Errorf("duplicate scene_id: %s", scene.SceneID)
		}
		seen[scene.SceneID] = true
		if strings.TrimSpace(scene.TimeTarget) == "" {
			return fmt.Errorf("%s: time_target is required", scene.SceneID)
		}
		if strings.TrimSpace(scene.Narration) == "" {
			return fmt.Errorf("%s: narration is required", scene.SceneID)
		}
		if strings.TrimSpace(scene.Visual.Type) == "" {
			return fmt.Errorf("%s: visual.type is required", scene.SceneID)
		}
		if strings.TrimSpace(scene.Visual.Content) == "" {
			return fmt.Errorf("%s: visual.content is required", scene.SceneID)
		}
	}
	return nil
}

func renderHTMLPreview(episodeID string, scenes []storyboard.Scene) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString("<title>")
	b.WriteString(html.EscapeString(episodeID))
	b.WriteString(" preview</title>\n")
	b.WriteString("<style>body{font-family:system-ui,sans-serif;margin:0;background:#101418;color:#f4f7fb}main{max-width:960px;margin:auto;padding:24px}.scene{border:1px solid #334155;padding:16px;margin:12px 0}.meta{color:#9fb0c7}.visual{background:#172033;padding:12px}</style>\n")
	b.WriteString("</head>\n<body>\n<main>\n")
	b.WriteString("<h1>")
	b.WriteString(html.EscapeString(episodeID))
	b.WriteString(" local preview</h1>\n")
	for _, scene := range scenes {
		b.WriteString("<section class=\"scene\" data-scene=\"")
		b.WriteString(html.EscapeString(scene.SceneID))
		b.WriteString("\">\n")
		b.WriteString("<p class=\"meta\">")
		b.WriteString(html.EscapeString(scene.SceneID + " | " + scene.TimeTarget))
		b.WriteString("</p>\n<h2>")
		b.WriteString(html.EscapeString(scene.OnScreenText))
		b.WriteString("</h2>\n<p>")
		b.WriteString(html.EscapeString(scene.Narration))
		b.WriteString("</p>\n<div class=\"visual\">")
		b.WriteString(html.EscapeString(scene.Visual.Type + ": " + scene.Visual.Content))
		b.WriteString("</div>\n<p class=\"meta\">Mascot: ")
		b.WriteString(html.EscapeString(scene.Mascot.Mode + " / " + scene.Mascot.Emotion + " / " + scene.Mascot.Action))
		b.WriteString("</p>\n<p class=\"meta\">Captions: ")
		b.WriteString(html.EscapeString(scene.CaptionPlan))
		b.WriteString("</p>\n</section>\n")
	}
	b.WriteString("</main>\n</body>\n</html>\n")
	return b.String()
}

func generateAssetManifest(episodeID string, scenes []storyboard.Scene) AssetManifest {
	assets := []Asset{
		placeholderAsset("mascot-placeholder", "vector_placeholder", "assets/placeholders/mascot.svg", []string{"storyboard.yaml"}),
	}
	visualTypes := map[string][]string{}
	for _, scene := range scenes {
		visualTypes[scene.Visual.Type] = append(visualTypes[scene.Visual.Type], scene.SceneID)
	}
	var keys []string
	for key := range visualTypes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		sort.Strings(visualTypes[key])
		assets = append(assets, placeholderAsset(
			"visual-"+slug(key),
			"visual_placeholder",
			path.Join("assets/placeholders", slug(key)+".svg"),
			visualTypes[key],
		))
	}

	return AssetManifest{
		SchemaVersion: SchemaVersion,
		EpisodeID:     episodeID,
		ArtifactID:    "asset-manifest-" + episodeID + "-preview-v1",
		Status:        fileStatus,
		Assets:        assets,
	}
}

func placeholderAsset(id, assetType, assetPath string, provenance []string) Asset {
	content := id + "|" + assetType + "|" + assetPath + "|" + strings.Join(provenance, ",")
	return Asset{
		AssetID:     id,
		Type:        assetType,
		Path:        assetPath,
		GeneratedBy: rendererName,
		License:     "owned/generated-placeholder",
		Hash:        contentHash(content),
		Provenance:  append([]string(nil), provenance...),
	}
}

func durationSeconds(scenes []storyboard.Scene) int {
	total := 0
	for _, scene := range scenes {
		total += sceneDuration(scene.TimeTarget)
	}
	if total == 0 {
		return len(scenes) * 8
	}
	return total
}

func sceneDuration(target string) int {
	parts := strings.Split(target, "-")
	if len(parts) != 2 {
		return 0
	}
	start, okStart := parseClock(parts[0])
	end, okEnd := parseClock(parts[1])
	if !okStart || !okEnd || end <= start {
		return 0
	}
	return end - start
}

func parseClock(value string) (int, bool) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, false
	}
	var minutes, seconds int
	if _, err := fmt.Sscanf(value, "%d:%d", &minutes, &seconds); err != nil {
		return 0, false
	}
	return minutes*60 + seconds, true
}

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		isWord := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if isWord {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
