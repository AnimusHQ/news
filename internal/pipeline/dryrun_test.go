package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
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
		case "editorial_brief.md", "script.md":
			writeArtifact(t, path, "# Test\n")
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
