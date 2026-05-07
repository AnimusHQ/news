package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/storyboard"
)

func TestGeneratePreviewFromStoryboard(t *testing.T) {
	result, err := GeneratePreview(Input{
		EpisodeID:  "episode-test",
		Storyboard: testStoryboard(),
		OutputDir:  "dist",
	})
	if err != nil {
		t.Fatalf("generate preview failed: %v", err)
	}

	if result.Preview.Path != "dist/episode-test-preview.html" {
		t.Fatalf("expected deterministic preview path, got %s", result.Preview.Path)
	}
	if result.Preview.Content == "" || result.Preview.Hash == "" {
		t.Fatalf("expected preview content and hash: %+v", result.Preview)
	}
	if len(result.AssetManifest.Assets) < 2 {
		t.Fatalf("expected placeholder asset provenance, got %+v", result.AssetManifest.Assets)
	}
	if got := result.RenderManifest.Outputs[0].DurationSeconds; got != 16 {
		t.Fatalf("expected duration from storyboard timing, got %d", got)
	}
}

func TestGeneratePreviewIsDeterministic(t *testing.T) {
	first, err := GeneratePreview(Input{EpisodeID: "episode-test", Storyboard: testStoryboard()})
	if err != nil {
		t.Fatalf("generate first preview failed: %v", err)
	}
	second, err := GeneratePreview(Input{EpisodeID: "episode-test", Storyboard: testStoryboard()})
	if err != nil {
		t.Fatalf("generate second preview failed: %v", err)
	}
	if first.Preview.Hash != second.Preview.Hash {
		t.Fatalf("expected deterministic preview hash, got %s and %s", first.Preview.Hash, second.Preview.Hash)
	}
	if first.RenderManifest.Outputs[0].Path != second.RenderManifest.Outputs[0].Path {
		t.Fatalf("expected deterministic output path")
	}
}

func TestGeneratePreviewFailsForMissingRequiredSceneData(t *testing.T) {
	file := testStoryboard()
	file.Scenes[0].Visual.Content = ""
	_, err := GeneratePreview(Input{EpisodeID: "episode-test", Storyboard: file})
	if err == nil {
		t.Fatal("expected missing visual content to fail")
	}
}

func TestRenderManifestValidatesAsArtifact(t *testing.T) {
	result, err := GeneratePreview(Input{
		EpisodeID:  "episode-test",
		Storyboard: testStoryboard(),
	})
	if err != nil {
		t.Fatalf("generate preview failed: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "render_manifest.json")
	encoded, err := json.MarshalIndent(result.RenderManifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal render manifest: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write render manifest: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected render manifest to validate: %+v", report.Issues)
	}
}

func testStoryboard() storyboard.File {
	return storyboard.File{
		SchemaVersion: storyboard.SchemaVersion,
		EpisodeID:     "episode-test",
		ArtifactID:    "storyboard-test",
		Status:        "draft",
		Scenes: []storyboard.Scene{
			{
				SceneID:      "scene-001",
				TimeTarget:   "0:00-0:08",
				Narration:    "You typed git push",
				Mascot:       storyboard.MascotPlan{Mode: "Production Mode", Emotion: "curious", Action: "opens terminal"},
				Visual:       storyboard.VisualPlan{Type: "terminal_animation", Content: "git push origin main"},
				OnScreenText: "You typed git push",
				CaptionPlan:  "captions_from_narration",
			},
			{
				SceneID:      "scene-002",
				TimeTarget:   "0:08-0:16",
				Narration:    "CI validates the change",
				Mascot:       storyboard.MascotPlan{Mode: "Explainer", Emotion: "focused", Action: "points at pipeline diagram"},
				Visual:       storyboard.VisualPlan{Type: "pipeline_diagram", Content: "repository event -> CI checks"},
				OnScreenText: "CI validates the change",
				CaptionPlan:  "captions_from_narration",
				ClaimIDs:     []string{"claim-001"},
				SourceIDs:    []string{"github-actions-docs-001"},
			},
		},
	}
}
