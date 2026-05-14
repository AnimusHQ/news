package artifacts

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateTransitionBlocksMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	report := ValidateTransition(dir, StateResearchReady)
	if report.Valid {
		t.Fatal("expected missing research pack to block transition")
	}
}

func TestValidateTransitionBlocksInvalidArtifact(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, filepath.Join(dir, "research_pack.json"), `{"schema_version":"1.0"}`)
	report := ValidateTransition(dir, StateResearchReady)
	if report.Valid {
		t.Fatal("expected invalid research pack to block transition")
	}
}

func TestValidateTransitionBlocksRejectedArtifact(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, filepath.Join(dir, "research_pack.json"), validResearchPackWithStatus("rejected"))
	report := ValidateTransition(dir, StateResearchReady)
	if report.Valid {
		t.Fatal("expected rejected research pack to block transition")
	}
	if !dependencyIssueContains(report, "required artifact status is rejected") {
		t.Fatalf("expected rejected status issue, got %+v", report.Issues)
	}
}

func TestValidateTransitionBlocksStaleDependencyHash(t *testing.T) {
	dir := t.TempDir()
	researchPath := filepath.Join(dir, "research_pack.json")
	writeArtifact(t, researchPath, validResearchPackWithStatus("draft"))
	writeArtifact(t, filepath.Join(dir, "script.md"), "# Script\n")
	writeArtifact(t, filepath.Join(dir, "claims.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "claims-test",
  "status": "draft",
  "source_artifacts": ["research_pack.json@sha256:wrong"],
  "claims": [
    {
      "claim_id": "claim-test",
      "text": "Test claim.",
      "type": "technical",
      "risk_level": "medium",
      "source_ids": ["source-test"],
      "verification_status": "needs_human_review"
    }
  ]
}`)
	report := ValidateTransition(dir, StateVerifying)
	if report.Valid {
		t.Fatal("expected stale dependency hash to block transition")
	}
	if !dependencyIssueContains(report, "hash mismatch") {
		t.Fatalf("expected hash mismatch issue, got %+v", report.Issues)
	}
}

func TestValidateTransitionBlocksStoryboardingWithoutHumanApproval(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, filepath.Join(dir, "human_qa_report.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "human-qa-test",
  "status": "draft",
  "reviewer": "human-required",
  "decision": "request_revision"
}`)
	report := ValidateTransition(dir, StateStoryboarding)
	if report.Valid {
		t.Fatal("expected non-approving human QA to block storyboarding")
	}
}

func TestValidateTransitionAllowsApprovedStoryboarding(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, filepath.Join(dir, "human_qa_report.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "human-qa-test",
  "status": "draft",
  "reviewer": "human-required",
  "decision": "approve"
}`)
	report := ValidateTransition(dir, StateStoryboarding)
	if !report.Valid {
		t.Fatalf("expected approved human QA to allow storyboarding: %+v", report.Issues)
	}
}

func TestValidateTransitionBlocksScheduledWithoutProductionQAApproval(t *testing.T) {
	dir := t.TempDir()
	writeArtifact(t, filepath.Join(dir, "production_qa_report.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "production-qa-test",
  "status": "draft",
  "decision": "request_revision"
}`)
	writeArtifact(t, filepath.Join(dir, "publish_manifest.json"), `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "publish-test",
  "status": "draft",
  "visibility": "private",
  "human_release_approval": false
}`)
	report := ValidateTransition(dir, StateScheduled)
	if report.Valid {
		t.Fatal("expected production QA revision to block scheduling")
	}
}

func validResearchPackWithStatus(status string) string {
	return `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "status": "` + status + `",
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
}`
}

func dependencyIssueContains(report DependencyReport, text string) bool {
	for _, issue := range report.Issues {
		if strings.Contains(issue.Reason, text) {
			return true
		}
	}
	return false
}
