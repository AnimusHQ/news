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

func TestValidatePathFailsForMalformedCanonicalArtifacts(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "topic.yaml",
			content: `schema_version: "1.0"
episode_id: "episode-test"
artifact_id: "topic-test"
status: "draft"
title_working: "Test"
format: "how_it_works"
audience:
  primary: "engineers"
scores: {}
operator_decision:
  decision: "approve"
`,
			want: "scores.educational_value",
		},
		{
			name: "research_pack.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "status": "draft",
  "core_question": "How does validation work?",
  "learning_objectives": ["Explain validation."],
  "sources": [
    {"source_id": "source-test", "title": "Test", "uri": "not-a-url", "type": "official_docs", "trust_level": "primary"}
  ]
}`,
			want: "source uri",
		},
		{
			name: "claims.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "claims-test",
  "status": "draft",
  "claims": [
    {"claim_id": "claim-test", "text": "Test", "type": "technical", "risk_level": "danger", "source_ids": ["source-test"], "verification_status": "needs_human_review"}
  ]
}`,
			want: "risk_level",
		},
		{
			name: "verification_report.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "verification-test",
  "status": "draft",
  "summary": "Test",
  "claim_results": [],
  "decision": "approved"
}`,
			want: "claim_results",
		},
		{
			name: "multimodel_approval_report.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "multimodel-test",
  "status": "draft",
  "model_panel": [
    {"model_id": "one", "provider": "local", "task": "review", "verdict": "approve", "confidence": 0.8}
  ],
  "consensus": "approved",
  "dissent": [],
  "operator_summary": "Test"
}`,
			want: "at least two model reviews",
		},
		{
			name: "human_qa_report.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "human-qa-test",
  "status": "draft",
  "reviewer": "human-required",
  "decision": "request_revision",
  "notes": "Test"
}`,
			want: "required_changes",
		},
		{
			name: "storyboard.yaml",
			content: `schema_version: "1.0"
episode_id: "episode-test"
artifact_id: "storyboard-test"
status: "draft"
scenes:
  - scene_id: "scene-001"
    time_target: "0:00-0:08"
    narration: "Test"
    mascot:
      mode: "Explainer"
      emotion: "focused"
      action: "points"
    visual:
      type: "diagram"
    on_screen_text: "Test"
`,
			want: "visual.content",
		},
		{
			name: "asset_manifest.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "asset-test",
  "status": "draft",
  "assets": [
    {"asset_id": "asset-test", "type": "text", "path": "assets/test.txt", "generated_by": "test", "hash": "placeholder"}
  ]
}`,
			want: "license",
		},
		{
			name: "render_manifest.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "render-test",
  "status": "draft",
  "renderer": "local",
  "renderer_version": "0.0.0",
  "inputs": ["storyboard.yaml"],
  "outputs": [
    {"type": "preview", "path": "dist/test.html", "duration_seconds": 0, "resolution": "html", "hash": "placeholder"}
  ]
}`,
			want: "duration_seconds",
		},
		{
			name: "production_qa_report.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "production-qa-test",
  "status": "draft",
  "checks": {"claims": "pass", "asset_provenance": "pass", "policy": "pass"},
  "blocking_issues": ["should not be here"],
  "decision": "approved"
}`,
			want: "approved production QA",
		},
		{
			name: "publish_manifest.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "publish-test",
  "status": "draft",
  "platform": "youtube",
  "visibility": "scheduled",
  "title": "Test",
  "description_path": "dist/description.md",
  "thumbnail_path": "dist/thumbnail.png",
  "human_release_approval": false
}`,
			want: "scheduled_at",
		},
		{
			name: "analytics_report.json",
			content: `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "analytics-test",
  "status": "draft",
  "window": "72h",
  "metrics": {
    "ctr": 2,
    "average_view_duration_seconds": 0,
    "first_30s_retention": 0,
    "subscriber_delta": 0,
    "community_clicks": 0
  },
  "insights": [],
  "recommended_actions": []
}`,
			want: "ctr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, tt.name)
			writeArtifact(t, path, tt.content)

			report := ValidatePath(path)
			if report.Valid {
				t.Fatalf("expected %s to fail validation", tt.name)
			}
			if !validationIssueContains(report, tt.want) {
				t.Fatalf("expected issue containing %q, got %+v", tt.want, report.Issues)
			}
		})
	}
}

func writeCompleteEpisodeFixture(t *testing.T, dir string) {
	t.Helper()
	for _, name := range RequiredEpisodeFiles {
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
		case "editorial_brief.md", "script.md":
			writeArtifact(t, path, "# Test\n")
		case "research_pack.json":
			writeArtifact(t, path, `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "human:test",
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

func validationIssueContains(report ValidationReport, text string) bool {
	for _, issue := range report.Issues {
		if strings.Contains(issue.Field, text) || strings.Contains(issue.Message, text) {
			return true
		}
	}
	return false
}
