package research

import (
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestAuditPackPassesWithPrimarySourceForHighRiskClaim(t *testing.T) {
	result, err := AuditPack(Pack{
		CoreQuestion: "How does deployment work?",
		Sources: []artifacts.Source{{
			ID: "official", Title: "Official Docs", URI: "https://example.com/docs", Type: "official_docs", TrustLevel: "primary",
		}},
		LearningObjectives:       []string{"Understand deployment."},
		ForbiddenSimplifications: []string{"Do not imply every system uses Kubernetes."},
		VisualOpportunities:      []string{"Pipeline diagram"},
	}, []artifacts.Claim{{
		ID:        "claim-1",
		Text:      "High risk claim.",
		RiskLevel: artifacts.ClaimRiskHigh,
		SourceIDs: []string{"official"},
	}})
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected audit to pass, blockers: %v", result.Blockers)
	}
}

func TestAuditPackBlocksHighRiskClaimWithoutPrimarySource(t *testing.T) {
	result, err := AuditPack(Pack{
		CoreQuestion: "How does deployment work?",
		Sources: []artifacts.Source{{
			ID: "community", Title: "Forum", URI: "https://example.com/forum", Type: "community_discussion", TrustLevel: "community",
		}},
	}, []artifacts.Claim{{
		ID:        "claim-1",
		Text:      "High risk claim.",
		RiskLevel: artifacts.ClaimRiskHigh,
		SourceIDs: []string{"community"},
	}})
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected high-risk community-only evidence to block audit")
	}
	if len(result.Blockers) == 0 {
		t.Fatal("expected blockers")
	}
}

func TestAuditPackWarnsOnMissingEditorialFields(t *testing.T) {
	result, err := AuditPack(Pack{
		CoreQuestion: "How does deployment work?",
		Sources: []artifacts.Source{{
			ID: "official", Title: "Official Docs", URI: "https://example.com/docs", Type: "official_docs", TrustLevel: "primary",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid audit with warnings, got blockers: %v", result.Blockers)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warnings for missing editorial fields")
	}
}

func TestAuditPackBlocksMissingSources(t *testing.T) {
	result, err := AuditPack(Pack{CoreQuestion: "Question"}, nil)
	if err != nil {
		t.Fatalf("audit failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected missing sources to block")
	}
}
