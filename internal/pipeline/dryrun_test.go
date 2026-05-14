package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/productionqa"
)

func TestDryRunPassesForCompleteBundle(t *testing.T) {
	dir := t.TempDir()
	writeCompleteEpisodeFixture(t, dir)

	report, err := DryRun(dir)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if !report.ArtifactsValid {
		t.Fatal("expected artifacts to be valid")
	}
	if len(report.Blockers) != 0 {
		t.Fatalf("expected no blockers, got %v", report.Blockers)
	}
	if report.CouncilConsensus != council.ConsensusRevisionRequired {
		t.Fatalf("expected revision-required council consensus, got %s", report.CouncilConsensus)
	}
	if len(report.CouncilSelected) != 3 {
		t.Fatalf("expected 3 selected council models, got %d", len(report.CouncilSelected))
	}
	if report.CouncilDissent != 1 {
		t.Fatalf("expected one dissenting revision review, got %d", report.CouncilDissent)
	}
	if report.HumanQARecommendation != artifacts.HumanDecisionRequestRevision {
		t.Fatalf("expected human QA request_revision recommendation, got %s", report.HumanQARecommendation)
	}
	if report.HumanQAUnresolved == 0 {
		t.Fatal("expected human QA packet to surface unresolved generated claims")
	}
	if report.StoryboardStatus != "skipped_by_human_qa_gate" {
		t.Fatalf("expected storyboard to stay gated by human QA, got %s", report.StoryboardStatus)
	}
	if report.RenderStatus != "skipped_by_storyboard_gate" {
		t.Fatalf("expected render preview to stay gated by storyboard, got %s", report.RenderStatus)
	}
	if report.ProductionQAStatus != "skipped_by_render_gate" {
		t.Fatalf("expected production QA to stay gated by render preview, got %s", report.ProductionQAStatus)
	}
}

func TestDryRunApprovedFixtureGeneratesDownstreamOutputs(t *testing.T) {
	dir := t.TempDir()
	writeCompleteEpisodeFixture(t, dir)
	writeApprovedDryRunFixtures(t, dir)

	report, err := DryRunWithOptions(dir, DryRunOptions{UseApprovedFixtures: true})
	if err != nil {
		t.Fatalf("approved fixture dry run failed: %v\n%s", err, report.String())
	}
	if report.StoryboardStatus != "generated" {
		t.Fatalf("expected storyboard generation, got %s", report.StoryboardStatus)
	}
	if report.StoryboardSceneCount == 0 {
		t.Fatal("expected generated storyboard scenes")
	}
	if report.RenderStatus != "preview_generated" {
		t.Fatalf("expected render preview, got %s", report.RenderStatus)
	}
	if report.ProductionQADecision != productionqa.DecisionApproved {
		t.Fatalf("expected production QA approval, got %s\n%s", report.ProductionQADecision, report.String())
	}
	if report.AnalyticsWindow != "72h" || report.AnalyticsInsightCount == 0 {
		t.Fatalf("expected fixture analytics insights, got window=%s count=%d", report.AnalyticsWindow, report.AnalyticsInsightCount)
	}
	if len(report.GeneratedOutputPaths) == 0 {
		t.Fatal("expected generated output paths in final summary")
	}
}

func TestDryRunFailsForMissingBundle(t *testing.T) {
	report, err := DryRun(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected dry run to fail for missing directory")
	}
	if len(report.Blockers) == 0 {
		t.Fatal("expected blockers to be populated")
	}
}

func writeApprovedDryRunFixtures(t *testing.T, dir string) {
	t.Helper()
	writeArtifact(t, filepath.Join(dir, "claims.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "claims-test-approved",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "human:test",
  "status": "approved",
  "claims": [
    {
      "claim_id": "claim-001",
      "text": "CI validates the change",
      "type": "technical",
      "risk_level": "medium",
      "source_ids": ["source-test"],
      "evidence_locators": [{"source_id": "source-test", "section": "test", "range": "test"}],
      "verification_status": "supported"
    }
  ]
}`)
	writeArtifact(t, filepath.Join(dir, "human_qa_report.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "human-qa-test-approved",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "human:test",
  "status": "approved",
  "reviewer": "human:test",
  "decision": "approve",
  "notes": "Fixture approval for downstream dry-run coverage.",
  "required_changes": []
}`)
}

func writeCompleteEpisodeFixture(t *testing.T, dir string) {
	t.Helper()
	for _, name := range artifacts.RequiredEpisodeFiles {
		path := filepath.Join(dir, name)
		switch name {
		case "topic.yaml":
			writeArtifact(t, path, `schema_version: "1.0"
episode_id: "episode-test"
artifact_id: "topic-test"
created_at: "2026-04-29T00:00:00Z"
created_by: "human:test"
status: "draft"
title_working: "Test topic"
format: "how_it_works"
audience:
  primary: "engineers"
scores:
  educational_value: 8
  evergreen_value: 8
  community_fit: 8
  visual_potential: 8
  production_cost: 4
  factual_risk: 3
  funnel_value: 5
operator_decision:
  decision: "approved_for_fixture"
`)
		case "editorial_brief.md":
			writeArtifact(t, path, "# Test\n")
		case "script.md":
			writeArtifact(t, path, `# Test

CI validates the change.
Build artifacts may be produced.
Deployment strategy moves the change toward production.
`)
		case "claims.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "claims-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "claims": [
    {
      "claim_id": "claim-test",
      "text": "Test claim.",
      "type": "technical",
      "risk_level": "medium",
      "source_ids": ["source-test"],
      "evidence_locators": [{"source_id": "source-test", "section": "test", "range": "test"}],
      "verification_status": "needs_human_review"
    }
  ]
}`)
		case "research_pack.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "core_question": "How does the test pipeline preserve source grounding?",
  "learning_objectives": ["Explain source-backed dry-run validation."],
  "sources": [
    {
      "source_id": "source-test",
      "title": "Test primary source",
      "uri": "https://example.com/test-source",
      "type": "official_docs",
      "trust_level": "primary",
      "license_notes": "test fixture"
    }
  ],
  "forbidden_simplifications": ["Do not treat mock review as human approval."],
  "visual_opportunities": ["pipeline diagram"]
}`)
		case "verification_report.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "verification-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "summary": "Fixture verification requires revision.",
  "claim_results": [
    {"claim_id": "claim-test", "status": "needs_human_review", "notes": "Fixture."}
  ],
  "blocking_issues": ["Fixture requires review."],
  "decision": "request_revision"
}`)
		case "multimodel_approval_report.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "multimodel-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "model_panel": [
    {"model_id": "technical", "provider": "local", "task": "technical", "verdict": "request_revision", "confidence": 0.5, "notes": "Fixture."},
    {"model_id": "editorial", "provider": "local", "task": "editorial", "verdict": "approve_with_suggestions", "confidence": 0.7, "notes": "Fixture."}
  ],
  "consensus": "revision_required",
  "dissent": [{"model_id": "technical"}],
  "operator_summary": "Fixture requires human review."
}`)
		case "publish_manifest.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "publish-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "platform": "youtube",
  "visibility": "private",
  "title": "Fixture title",
  "description_path": "dist/description.md",
  "thumbnail_path": "dist/thumbnail.png",
  "scheduled_at": null,
  "human_release_approval": false
}`)
		case "human_qa_report.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "human-qa-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "reviewer": "human-required",
  "decision": "request_revision",
  "notes": "Fixture requires revision.",
  "required_changes": ["Review fixture."]
}`)
		case "storyboard.yaml":
			writeArtifact(t, path, `schema_version: "1.0"
episode_id: "episode-test"
artifact_id: "storyboard-test"
created_at: "2026-04-29T00:00:00Z"
created_by: "system:test"
status: "draft"
scenes:
  - scene_id: "scene-001"
    time_target: "0:00-0:08"
    narration: "Fixture narration"
    mascot:
      mode: "Explainer"
      emotion: "focused"
      action: "points"
    visual:
      type: "diagram"
      content: "fixture"
    on_screen_text: "Fixture"
`)
		case "asset_manifest.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "asset-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "assets": [
    {"asset_id": "asset-test", "type": "text", "path": "assets/test.txt", "generated_by": "test", "license": "owned/generated", "hash": "placeholder"}
  ]
}`)
		case "render_manifest.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "render-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "renderer": "local-test",
  "renderer_version": "0.0.0",
  "inputs": ["storyboard.yaml", "asset_manifest.json"],
  "outputs": [
    {"type": "preview", "path": "dist/test.html", "duration_seconds": 8, "resolution": "responsive-html", "hash": "placeholder"}
  ]
}`)
		case "production_qa_report.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "production-qa-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "checks": {
    "claims": "fail",
    "asset_provenance": "pass",
    "policy": "pass"
  },
  "blocking_issues": ["Fixture requires review."],
  "decision": "request_revision"
}`)
		case "analytics_report.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "analytics-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "window": "dry_run",
  "metrics": {
    "ctr": 0,
    "average_view_duration_seconds": 0,
    "first_30s_retention": 0,
    "subscriber_delta": 0,
    "community_clicks": 0
  },
  "insights": [],
  "recommended_actions": []
}`)
		default:
			writeArtifact(t, path, fmt.Sprintf(`{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "%s",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft"
}`, name))
		}
	}
}

func writeArtifact(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
