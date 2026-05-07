package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
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

func TestDryRunFailsForMissingBundle(t *testing.T) {
	report, err := DryRun(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected dry run to fail for missing directory")
	}
	if len(report.Blockers) == 0 {
		t.Fatal("expected blockers to be populated")
	}
}

func writeCompleteEpisodeFixture(t *testing.T, dir string) {
	t.Helper()
	for _, name := range artifacts.RequiredEpisodeFiles {
		path := filepath.Join(dir, name)
		switch name {
		case "topic.yaml", "storyboard.yaml":
			writeArtifact(t, path, fmt.Sprintf("schema_version: \"1.0\"\nepisode_id: \"episode-test\"\nartifact_id: \"%s\"\ncreated_at: \"2026-04-29T00:00:00Z\"\ncreated_by: \"test\"\nstatus: \"draft\"\n", name))
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
		case "publish_manifest.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "publish-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "visibility": "private",
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
  "decision": "request_revision"
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
