package productionqa

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/storyboard"
	"github.com/AnimusHQ/news/internal/verification"
)

func TestRunApprovesPassingFixture(t *testing.T) {
	report, err := Run(passingInput(t))
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision != DecisionApproved {
		t.Fatalf("expected approved report, got %s with blockers %+v", report.Decision, report.BlockingIssues)
	}
	if report.Checks["asset_provenance"] != "pass" {
		t.Fatalf("expected asset provenance pass, got %+v", report.Checks)
	}
}

func TestRunFailsForMissingRenderOutput(t *testing.T) {
	input := passingInput(t)
	input.Render.Preview.Content = ""
	report, err := Run(input)
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision == DecisionApproved {
		t.Fatalf("expected missing render output to block approval")
	}
	if report.Checks["render_outputs"] != "fail" {
		t.Fatalf("expected render output check to fail, got %+v", report.Checks)
	}
}

func TestRunFailsForMissingAssetProvenance(t *testing.T) {
	input := passingInput(t)
	input.Render.AssetManifest.Assets[0].Provenance = nil
	report, err := Run(input)
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision == DecisionApproved {
		t.Fatalf("expected missing provenance to block approval")
	}
	if report.Checks["asset_provenance"] != "fail" {
		t.Fatalf("expected asset provenance check to fail, got %+v", report.Checks)
	}
}

func TestRunBlocksDirectPublicPublishIntent(t *testing.T) {
	input := passingInput(t)
	input.PublishVisibility = artifacts.PublishVisibilityPublic
	report, err := Run(input)
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision != DecisionBlock {
		t.Fatalf("expected direct public publish intent to block, got %s", report.Decision)
	}
}

func TestRunFailsForUnresolvedHighRiskClaim(t *testing.T) {
	input := passingInput(t)
	input.Claims[0].RiskLevel = artifacts.ClaimRiskHigh
	input.Verification.Decision = "request_revision"
	input.Verification.ClaimResults[0].Status = artifacts.ClaimStatusNeedsHumanReview
	report, err := Run(input)
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision == DecisionApproved {
		t.Fatalf("expected unresolved high-risk claim to block approval")
	}
	if report.Checks["claims"] != "fail" {
		t.Fatalf("expected claims check to fail, got %+v", report.Checks)
	}
}

func TestRunFailsWithoutHumanQAApproval(t *testing.T) {
	input := passingInput(t)
	input.HumanQARecommendation = artifacts.HumanDecisionRequestRevision
	report, err := Run(input)
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	if report.Decision != DecisionBlock {
		t.Fatalf("expected missing human QA approval to block, got %s", report.Decision)
	}
}

func TestReportValidatesAsProductionQAArtifact(t *testing.T) {
	report, err := Run(passingInput(t))
	if err != nil {
		t.Fatalf("run QA failed: %v", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "production_qa_report.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write report: %v", err)
	}
	validation := artifacts.ValidatePath(path)
	if !validation.Valid {
		t.Fatalf("expected production QA report to validate: %+v", validation.Issues)
	}
}

func passingInput(t *testing.T) Input {
	t.Helper()
	renderResult, err := render.GeneratePreview(render.Input{
		EpisodeID:  "episode-test",
		Storyboard: qaStoryboard(),
	})
	if err != nil {
		t.Fatalf("generate render fixture: %v", err)
	}
	return Input{
		EpisodeID:             "episode-test",
		Render:                renderResult,
		Claims:                []artifacts.Claim{supportedClaim()},
		Verification:          verification.Report{Decision: "approved", ClaimResults: []verification.ClaimResult{{ClaimID: "claim-001", Status: artifacts.ClaimStatusSupported}}},
		HumanQARecommendation: artifacts.HumanDecisionApprove,
		PublishVisibility:     artifacts.PublishVisibilityPrivate,
	}
}

func qaStoryboard() storyboard.File {
	return storyboard.File{
		SchemaVersion: storyboard.SchemaVersion,
		EpisodeID:     "episode-test",
		ArtifactID:    "storyboard-test",
		Status:        "draft",
		Scenes: []storyboard.Scene{
			{
				SceneID:      "scene-001",
				TimeTarget:   "0:00-0:08",
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

func supportedClaim() artifacts.Claim {
	return artifacts.Claim{
		ID:        "claim-001",
		Text:      "CI validates the change",
		Type:      "technical",
		RiskLevel: artifacts.ClaimRiskMedium,
		SourceIDs: []string{"github-actions-docs-001"},
		EvidenceLocators: []artifacts.EvidenceLocator{
			{SourceID: "github-actions-docs-001", Section: "events"},
		},
		Status: artifacts.ClaimStatusSupported,
	}
}
