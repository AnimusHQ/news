package publishing

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/productionqa"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/storyboard"
)

func TestGeneratePackDefaultsToPrivate(t *testing.T) {
	pack, err := GeneratePack(PackInput{
		EpisodeID: "episode-1",
		Title:     "What Happens After git push?",
		Summary:   "A source-grounded explanation.",
		Sources: []artifacts.Source{{
			ID:    "git-docs",
			Title: "Git documentation",
			URI:   "https://git-scm.com/doc",
		}},
	})
	if err != nil {
		t.Fatalf("generate pack failed: %v", err)
	}
	if pack.Visibility != artifacts.PublishVisibilityPrivate {
		t.Fatalf("expected private visibility, got %s", pack.Visibility)
	}
	if !strings.Contains(pack.Description, "Sources:") {
		t.Fatalf("expected sources in description: %s", pack.Description)
	}
}

func TestGeneratePackRejectsPublicWithoutApproval(t *testing.T) {
	_, err := GeneratePack(PackInput{
		EpisodeID:     "episode-1",
		Title:         "What Happens After git push?",
		Summary:       "A source-grounded explanation.",
		Visibility:    artifacts.PublishVisibilityPublic,
		HumanApproved: false,
	})
	if err == nil {
		t.Fatal("expected public visibility without approval to fail")
	}
}

func TestGeneratePackAllowsPublicWithApproval(t *testing.T) {
	pack, err := GeneratePack(PackInput{
		EpisodeID:     "episode-1",
		Title:         "What Happens After git push?",
		Summary:       "A source-grounded explanation.",
		Visibility:    artifacts.PublishVisibilityPublic,
		HumanApproved: true,
	})
	if err != nil {
		t.Fatalf("expected approved public pack to be generated: %v", err)
	}
	if pack.Visibility != artifacts.PublishVisibilityPublic {
		t.Fatalf("expected public visibility, got %s", pack.Visibility)
	}
}

func TestGeneratePackRequiresTitle(t *testing.T) {
	_, err := GeneratePack(PackInput{EpisodeID: "episode-1"})
	if err == nil {
		t.Fatal("expected missing title to fail")
	}
}

func TestGenerateReleasePackBuildsManifestAndChapters(t *testing.T) {
	release, err := GenerateReleasePack(releaseInput())
	if err != nil {
		t.Fatalf("generate release pack failed: %v", err)
	}
	if release.Manifest.Visibility != artifacts.PublishVisibilityPrivate {
		t.Fatalf("expected private manifest visibility, got %s", release.Manifest.Visibility)
	}
	if !strings.Contains(release.Pack.Description, "0:00 From commit to production") {
		t.Fatalf("expected storyboard chapter in description: %s", release.Pack.Description)
	}
	if !strings.Contains(release.Pack.Description, "Sources:") {
		t.Fatalf("expected sources in release description: %s", release.Pack.Description)
	}
}

func TestGenerateReleasePackManifestValidates(t *testing.T) {
	release, err := GenerateReleasePack(releaseInput())
	if err != nil {
		t.Fatalf("generate release pack failed: %v", err)
	}
	encoded, err := json.MarshalIndent(release.Manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "publish_manifest.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected publish manifest to validate: %+v", report.Issues)
	}
}

func TestGenerateReleasePackRejectsUnapprovedProductionQA(t *testing.T) {
	input := releaseInput()
	input.ProductionQA.Decision = productionqa.DecisionRequestRevision
	_, err := GenerateReleasePack(input)
	if err == nil {
		t.Fatal("expected unapproved production QA to fail")
	}
}

func TestGenerateReleasePackRejectsClaimsWithoutSources(t *testing.T) {
	input := releaseInput()
	input.Sources = nil
	_, err := GenerateReleasePack(input)
	if err == nil {
		t.Fatal("expected claims without sources to fail")
	}
}

func TestGenerateReleasePackRejectsPublicWithoutReleaseApproval(t *testing.T) {
	input := releaseInput()
	input.Visibility = artifacts.PublishVisibilityPublic
	input.HumanApproved = false
	_, err := GenerateReleasePack(input)
	if err == nil {
		t.Fatal("expected public visibility without approval to fail")
	}
}

func TestGenerateReleasePackRequiresDisclosureWhenFlagged(t *testing.T) {
	input := releaseInput()
	input.SyntheticDisclosureRequired = true
	input.SyntheticDisclosure = ""
	_, err := GenerateReleasePack(input)
	if err == nil {
		t.Fatal("expected missing synthetic disclosure to fail")
	}
	input.SyntheticDisclosure = "Contains deterministic placeholder visuals."
	release, err := GenerateReleasePack(input)
	if err != nil {
		t.Fatalf("expected disclosure to satisfy release pack: %v", err)
	}
	if release.Manifest.SyntheticDisclosure == "" {
		t.Fatal("expected disclosure to be present in manifest")
	}
}

func releaseInput() ReleasePackInput {
	return ReleasePackInput{
		EpisodeID:  "episode-1",
		Title:      "What Happens After git push?",
		Summary:    "A source-grounded explanation.",
		Visibility: artifacts.PublishVisibilityPrivate,
		Sources: []artifacts.Source{{
			ID:    "git-docs",
			Title: "Git documentation",
			URI:   "https://git-scm.com/doc",
		}},
		Claims: []artifacts.Claim{{
			ID:        "claim-001",
			Text:      "CI validates the change",
			Type:      "technical",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"git-docs"},
		}},
		Storyboard: storyboard.File{
			SchemaVersion: storyboard.SchemaVersion,
			EpisodeID:     "episode-1",
			ArtifactID:    "storyboard-episode-1",
			Status:        "draft",
			Scenes: []storyboard.Scene{{
				SceneID:      "scene-001",
				TimeTarget:   "0:00-0:08",
				Narration:    "From commit to production.",
				Visual:       storyboard.VisualPlan{Type: "terminal_animation", Content: "git push origin main"},
				OnScreenText: "From commit to production",
			}},
		},
		RenderManifest: render.RenderManifest{
			SchemaVersion: render.SchemaVersion,
			EpisodeID:     "episode-1",
			ArtifactID:    "render-episode-1",
			Status:        "draft",
			Outputs: []render.RenderOutput{{
				Type: "html_preview",
				Path: "dist/episode-1-preview.html",
				Hash: "sha256:test",
			}},
		},
		ProductionQA: productionqa.Report{
			SchemaVersion: productionqa.SchemaVersion,
			EpisodeID:     "episode-1",
			ArtifactID:    "production-qa-episode-1",
			Status:        "draft",
			Decision:      productionqa.DecisionApproved,
		},
	}
}
