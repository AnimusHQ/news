package artifacts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateEpisodeDirectoryPassesForCompleteBundle(t *testing.T) {
	dir := t.TempDir()
	writeCompleteEpisodeFixture(t, dir)

	if err := ValidateEpisodeDirectory(dir); err != nil {
		t.Fatalf("expected complete bundle to validate: %v", err)
	}
}

func TestValidateEpisodeDirectoryFailsForMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	writeCompleteEpisodeFixture(t, dir)
	if err := os.Remove(filepath.Join(dir, "topic.yaml")); err != nil {
		t.Fatalf("remove topic: %v", err)
	}

	if err := ValidateEpisodeDirectory(dir); err == nil {
		t.Fatal("expected missing artifact to fail validation")
	}
}

func TestValidateEpisodeDirectoryFailsForFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(path, []byte("placeholder"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := ValidateEpisodeDirectory(path); err == nil {
		t.Fatal("expected file path to fail validation")
	}
}

func TestValidateEpisodeDirectoryBlocksPublicWithoutHumanReleaseApproval(t *testing.T) {
	dir := t.TempDir()
	writeCompleteEpisodeFixture(t, dir)
	writeArtifact(t, filepath.Join(dir, "publish_manifest.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "publish-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "visibility": "public",
  "human_release_approval": false
}`)

	err := ValidateEpisodeDirectory(dir)
	if err == nil {
		t.Fatal("expected public publish without approval to fail")
	}
	if !strings.Contains(err.Error(), "public visibility requires explicit human release approval") {
		t.Fatalf("expected publish safety error, got: %v", err)
	}
}

func TestValidatePathPassesForSingleArtifact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "research_pack.json")
	writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "core_question": "How does validation work?",
  "learning_objectives": ["Explain validation."],
  "sources": [
    {
      "source_id": "source-test",
      "title": "Test source",
      "uri": "https://example.com/source",
      "type": "official_docs",
      "trust_level": "primary"
    }
  ]
}`)

	report := ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected artifact to validate: %+v", report.Issues)
	}
}

func TestValidatePathFailsForUnknownArtifact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown.json")
	writeArtifact(t, path, `{"schema_version":"1.0"}`)

	report := ValidatePath(path)
	if report.Valid {
		t.Fatal("expected unknown artifact to fail")
	}
	if len(report.Issues) == 0 || !strings.Contains(report.Issues[0].Message, "unknown canonical artifact") {
		t.Fatalf("expected unknown artifact issue, got %+v", report.Issues)
	}
}

func writeCompleteEpisodeFixture(t *testing.T, dir string) {
	t.Helper()
	for _, name := range RequiredEpisodeFiles {
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
