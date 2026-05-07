package storyboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestGenerateStoryboardFromApprovedScript(t *testing.T) {
	file, err := Generate(Input{
		EpisodeID:             "episode-test",
		ScriptMarkdown:        storyboardScript(),
		HumanQARecommendation: artifacts.HumanDecisionApprove,
		Claims:                storyboardClaims(),
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	if len(file.Scenes) != 4 {
		t.Fatalf("expected 4 scenes, got %d", len(file.Scenes))
	}
	if file.Scenes[0].SceneID != "scene-001" || file.Scenes[1].SceneID != "scene-002" {
		t.Fatalf("expected stable scene ids, got %+v", file.Scenes[:2])
	}
	if file.Scenes[0].TimeTarget != "0:00-0:08" {
		t.Fatalf("expected stable first timing target, got %s", file.Scenes[0].TimeTarget)
	}
	for _, scene := range file.Scenes {
		if scene.Narration == "" {
			t.Fatalf("scene missing narration: %+v", scene)
		}
		if scene.Visual.Type == "" || scene.Visual.Content == "" {
			t.Fatalf("scene missing visual plan: %+v", scene)
		}
	}
}

func TestGenerateRequiresHumanQAGate(t *testing.T) {
	_, err := Generate(Input{
		EpisodeID:             "episode-test",
		ScriptMarkdown:        storyboardScript(),
		HumanQARecommendation: artifacts.HumanDecisionRequestRevision,
		Claims:                storyboardClaims(),
	})
	if err == nil {
		t.Fatal("expected storyboarding to require approving human QA recommendation")
	}
}

func TestGenerateDoesNotDropTechnicalClaims(t *testing.T) {
	_, err := Generate(Input{
		EpisodeID:             "episode-test",
		ScriptMarkdown:        "CI validates the change.",
		HumanQARecommendation: artifacts.HumanDecisionApprove,
		Claims: []artifacts.Claim{
			{
				ID:        "claim-001",
				Text:      "CI validates the change",
				Type:      "technical",
				RiskLevel: artifacts.ClaimRiskMedium,
				SourceIDs: []string{"github-actions-docs-001"},
			},
			{
				ID:        "claim-002",
				Text:      "Rollback exists because production is never theoretical",
				Type:      "technical",
				RiskLevel: artifacts.ClaimRiskMedium,
				SourceIDs: []string{"kubernetes-docs-001"},
			},
		},
	})
	if err == nil {
		t.Fatal("expected generator to reject a missing technical claim reference")
	}
}

func TestGeneratedStoryboardValidatesAsArtifact(t *testing.T) {
	file, err := Generate(Input{
		EpisodeID:             "episode-test",
		ScriptMarkdown:        storyboardScript(),
		HumanQARecommendation: artifacts.HumanDecisionApproveWithMinorEdits,
		Claims:                storyboardClaims(),
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	encoded, err := MarshalYAML(file)
	if err != nil {
		t.Fatalf("marshal storyboard: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "storyboard.yaml")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write storyboard: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected generated storyboard to validate: %+v", report.Issues)
	}
}

func storyboardScript() string {
	return `# Script

You typed git push.
CI validates the change.
Build artifacts or container images may be produced.
Rollback exists because production is never theoretical.
`
}

func storyboardClaims() []artifacts.Claim {
	return []artifacts.Claim{
		{
			ID:        "claim-001",
			Text:      "CI validates the change",
			Type:      "technical",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"github-actions-docs-001"},
		},
		{
			ID:        "claim-002",
			Text:      "Build artifacts or container images may be produced",
			Type:      "technical",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"docker-docs-001"},
		},
		{
			ID:        "claim-003",
			Text:      "Rollback exists because production is never theoretical",
			Type:      "technical",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"kubernetes-docs-001"},
		},
	}
}
