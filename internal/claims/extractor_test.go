package claims

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestExtractProducesExpectedPilotClaims(t *testing.T) {
	result, err := Extract(Input{
		EpisodeID:      "episode-test",
		ArtifactID:     "claims-test",
		ScriptMarkdown: pilotScript(),
		ResearchPack:   testResearchPack(),
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	if len(result.ClaimsFile.Claims) != 6 {
		t.Fatalf("expected 6 factual claims, got %d: %+v", len(result.ClaimsFile.Claims), result.ClaimsFile.Claims)
	}
	assertClaim(t, result.ClaimsFile.Claims[0], "claim-001", TypeTechnical, artifacts.ClaimRiskMedium, []string{"git-docs-001"})
	assertClaim(t, result.ClaimsFile.Claims[3], "claim-004", TypeTechnical, artifacts.ClaimRiskMedium, []string{"docker-docs-001"})
	if result.ClaimsFile.Claims[0].Status != artifacts.ClaimStatusNeedsHumanReview {
		t.Fatalf("extracted claims must require review, got %s", result.ClaimsFile.Claims[0].Status)
	}
	if len(result.ClaimsFile.Claims[0].EvidenceLocators) != 0 {
		t.Fatal("extractor must not fabricate evidence locators")
	}
}

func TestExtractFlagsUnlinkedHighRiskTechnicalClaim(t *testing.T) {
	result, err := Extract(Input{
		EpisodeID:      "episode-test",
		ArtifactID:     "claims-test",
		ScriptMarkdown: "Credential exposure can leak private data during an incident.",
		ResearchPack:   testResearchPack(),
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if len(result.UnlinkedClaimIDs) != 1 {
		t.Fatalf("expected one unlinked high-risk claim, got %+v", result.UnlinkedClaimIDs)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected extraction warning")
	}
	if got := result.ClaimsFile.Claims[0].RiskLevel; got != artifacts.ClaimRiskCritical {
		t.Fatalf("expected critical risk, got %s", got)
	}
}

func TestExtractSkipsNonFactualOpinion(t *testing.T) {
	result, err := Extract(Input{
		EpisodeID:      "episode-test",
		ArtifactID:     "claims-test",
		ScriptMarkdown: "I think this should feel exciting.\nCI validates the change.",
		ResearchPack:   testResearchPack(),
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}
	if len(result.ClaimsFile.Claims) != 1 {
		t.Fatalf("expected only one factual claim, got %+v", result.ClaimsFile.Claims)
	}
	for _, candidate := range result.Candidates {
		if candidate.Type == TypeEditorialOpinion && candidate.Included {
			t.Fatalf("opinion candidate should not be included: %+v", candidate)
		}
	}
}

func TestExtractedClaimsFileValidatesForLinkedMediumRiskClaims(t *testing.T) {
	result, err := Extract(Input{
		EpisodeID:      "episode-test",
		ArtifactID:     "claims-test",
		ScriptMarkdown: "CI validates the change. Build artifacts may be produced.",
		ResearchPack:   testResearchPack(),
	})
	if err != nil {
		t.Fatalf("extract failed: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "claims.json")
	encoded, err := json.MarshalIndent(result.ClaimsFile, "", "  ")
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write claims: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected extracted claims to validate: %+v", report.Issues)
	}
}

func assertClaim(t *testing.T, claim artifacts.Claim, id string, claimType string, risk artifacts.ClaimRisk, sources []string) {
	t.Helper()
	if claim.ID != id {
		t.Fatalf("expected claim id %s, got %s", id, claim.ID)
	}
	if claim.Type != claimType {
		t.Fatalf("expected claim type %s, got %s", claimType, claim.Type)
	}
	if claim.RiskLevel != risk {
		t.Fatalf("expected claim risk %s, got %s", risk, claim.RiskLevel)
	}
	if len(claim.SourceIDs) != len(sources) {
		t.Fatalf("expected sources %+v, got %+v", sources, claim.SourceIDs)
	}
	for i := range sources {
		if claim.SourceIDs[i] != sources[i] {
			t.Fatalf("expected sources %+v, got %+v", sources, claim.SourceIDs)
		}
	}
}

func pilotScript() string {
	return `# Script

You typed git push. A few minutes later, your code may be running in production.

1. Local commits and remote repository update.
2. Repository event triggers automation.
3. CI validates the change.
4. Build artifacts or container images may be produced.
5. Deployment strategy moves the change toward production.
6. Observability tells the team whether reality matches expectations.
7. Rollback exists because production is never theoretical.

If you want to learn more, follow the Voyager path.`
}

func testResearchPack() artifacts.ResearchPackFile {
	return artifacts.ResearchPackFile{
		SchemaVersion: "1.0",
		EpisodeID:     "episode-test",
		ArtifactID:    "research-test",
		Status:        "draft",
		CoreQuestion:  "What happens after git push?",
		Sources: []artifacts.Source{
			{ID: "git-docs-001", Title: "Git documentation", URI: "https://git-scm.com/doc", Type: "official_docs", TrustLevel: "primary"},
			{ID: "github-actions-docs-001", Title: "GitHub Actions documentation", URI: "https://docs.github.com/actions", Type: "official_docs", TrustLevel: "primary"},
			{ID: "docker-docs-001", Title: "Docker documentation", URI: "https://docs.docker.com/", Type: "official_docs", TrustLevel: "primary"},
			{ID: "kubernetes-docs-001", Title: "Kubernetes documentation", URI: "https://kubernetes.io/docs/", Type: "official_docs", TrustLevel: "primary"},
		},
	}
}
