package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestBuildPackProducesValidResearchArtifact(t *testing.T) {
	result, err := BuildPack(builderInput(false))
	if err != nil {
		t.Fatalf("build pack failed: %v", err)
	}
	if len(result.Blockers) != 0 {
		t.Fatalf("expected no blockers, got %+v", result.Blockers)
	}
	if result.Pack.Sources[0].ID != "official" {
		t.Fatalf("expected primary source to be ranked first, got %+v", result.Pack.Sources)
	}
	if len(result.Pack.ClaimCandidates) != 1 {
		t.Fatalf("expected claim candidate from snippet, got %+v", result.Pack.ClaimCandidates)
	}

	encoded, err := json.MarshalIndent(result.Pack, "", "  ")
	if err != nil {
		t.Fatalf("marshal research pack: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "research_pack.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write research pack: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected research pack to validate: %+v", report.Issues)
	}
}

func TestBuildPackFlagsMissingPrimarySourceForHighRiskTopic(t *testing.T) {
	input := builderInput(true)
	input.Sources = []artifacts.Source{{ID: "community", Title: "Forum", URI: "https://example.com/forum", Type: "community_discussion", TrustLevel: "community"}}
	input.Snippets[0].SourceID = "community"
	input.Snippets[0].Locator.SourceID = "community"
	result, err := BuildPack(input)
	if err != nil {
		t.Fatalf("build pack failed: %v", err)
	}
	if len(result.Blockers) == 0 {
		t.Fatal("expected high-risk missing primary source blocker")
	}
}

func TestBuildPackPreservesSourceIDsAndLocators(t *testing.T) {
	result, err := BuildPack(builderInput(false))
	if err != nil {
		t.Fatalf("build pack failed: %v", err)
	}
	if result.Pack.SourceSnippets[0].SourceID != "official" {
		t.Fatalf("expected snippet source id preserved, got %+v", result.Pack.SourceSnippets)
	}
	if result.Pack.SourceSnippets[0].Locator.Section != "events" {
		t.Fatalf("expected locator section preserved, got %+v", result.Pack.SourceSnippets[0].Locator)
	}
}

func TestBuildPackIncludesForbiddenSimplifications(t *testing.T) {
	result, err := BuildPack(builderInput(false))
	if err != nil {
		t.Fatalf("build pack failed: %v", err)
	}
	if len(result.Pack.ForbiddenSimplifications) == 0 {
		t.Fatal("expected forbidden simplifications")
	}
}

func TestBuildPackBlocksUnknownSnippetSource(t *testing.T) {
	input := builderInput(false)
	input.Snippets[0].SourceID = "missing"
	result, err := BuildPack(input)
	if err != nil {
		t.Fatalf("build pack failed: %v", err)
	}
	if len(result.Blockers) == 0 {
		t.Fatal("expected unknown snippet source blocker")
	}
}

func builderInput(highRisk bool) BuilderInput {
	return BuilderInput{
		EpisodeID:    "episode-test",
		CoreQuestion: "How does CI validate a repository event?",
		Audience: map[string]string{
			"primary": "beginner engineers",
		},
		Sources: []artifacts.Source{
			{ID: "community", Title: "Forum", URI: "https://example.com/forum", Type: "community_discussion", TrustLevel: "community"},
			{ID: "official", Title: "Official Docs", URI: "https://example.com/docs", Type: "official_docs", TrustLevel: "primary"},
		},
		Snippets: []artifacts.SourceSnippet{{
			SourceID: "official",
			Locator:  artifacts.EvidenceLocator{SourceID: "official", Section: "events", Range: "p1"},
			Text:     "Repository events can trigger CI validation.",
		}},
		LearningObjectives:       []string{"Explain repository event validation."},
		RequiredTerms:            []string{"CI", "repository event"},
		ForbiddenSimplifications: []string{"Do not imply every CI system deploys automatically."},
		VisualOpportunities:      []string{"pipeline diagram"},
		RecommendedCTA:           "Review the source-backed production path.",
		HighRiskTopic:            highRisk,
	}
}
