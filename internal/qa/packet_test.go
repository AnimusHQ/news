package qa

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/verification"
)

func TestGenerateApprovePacket(t *testing.T) {
	packet, err := Generate(Input{
		EpisodeID:       "episode-test",
		EpisodePurpose:  "Explain a source-backed deployment flow.",
		EpisodeFormat:   "short educational explainer",
		ResearchSummary: "Official docs support the deployment flow.",
		Claims: []artifacts.Claim{
			supportedClaim("claim-001", artifacts.ClaimRiskMedium),
		},
		Verification: verification.Report{
			Decision: "approved",
			ClaimResults: []verification.ClaimResult{
				{ClaimID: "claim-001", Status: artifacts.ClaimStatusSupported, Notes: "covered"},
			},
		},
		Council: council.Report{
			Consensus: council.ConsensusApproved,
			Reviews: []council.ModelReview{
				{ModelID: "reviewer-a", Task: "technical_verification", Verdict: council.VerdictApprove, Notes: "covered"},
				{ModelID: "reviewer-b", Task: "editorial_review", Verdict: council.VerdictApprove, Notes: "clear"},
			},
		},
		QualityGateStatus: "passed",
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if packet.RecommendedDecision != artifacts.HumanDecisionApprove {
		t.Fatalf("expected approve recommendation, got %s", packet.RecommendedDecision)
	}
	if len(packet.UnresolvedClaims) != 0 {
		t.Fatalf("expected no unresolved claims, got %+v", packet.UnresolvedClaims)
	}
	if len(packet.ModelApprovals) != 2 {
		t.Fatalf("expected approvals to be visible, got %+v", packet.ModelApprovals)
	}
}

func TestGenerateRevisionPacketPreservesDissentAndUnsupportedClaims(t *testing.T) {
	packet, err := Generate(Input{
		EpisodeID: "episode-test",
		Claims: []artifacts.Claim{
			{
				ID:        "claim-001",
				Text:      "Deployment strategy moves the change toward production.",
				Type:      "technical",
				RiskLevel: artifacts.ClaimRiskHigh,
				SourceIDs: []string{"source-001"},
				Status:    artifacts.ClaimStatusNeedsHumanReview,
			},
		},
		Verification: verification.Report{
			Decision: "request_revision",
			ClaimResults: []verification.ClaimResult{
				{ClaimID: "claim-001", Status: artifacts.ClaimStatusNeedsHumanReview, Notes: "high risk claim lacks locator"},
			},
			BlockingIssues: []string{"claim-001: high risk claim lacks locator"},
		},
		Council: council.Report{
			Consensus: council.ConsensusRevisionRequired,
			Reviews: []council.ModelReview{
				{ModelID: "reviewer-a", Task: "technical_verification", Verdict: council.VerdictApprove, Notes: "mostly covered"},
				{ModelID: "reviewer-b", Task: "editorial_review", Verdict: council.VerdictRequestRevision, Notes: "source locator is missing"},
			},
			Dissent: []council.ModelReview{
				{ModelID: "reviewer-b", Task: "editorial_review", Verdict: council.VerdictRequestRevision, Notes: "source locator is missing"},
			},
		},
		QualityGateStatus: "passed",
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if packet.RecommendedDecision != artifacts.HumanDecisionRequestRevision {
		t.Fatalf("expected request_revision recommendation, got %s", packet.RecommendedDecision)
	}
	if len(packet.ModelDissent) != 1 || packet.ModelDissent[0].Notes != "source locator is missing" {
		t.Fatalf("expected dissenting notes to be preserved, got %+v", packet.ModelDissent)
	}
	if len(packet.UnresolvedClaims) != 1 || packet.UnresolvedClaims[0].ClaimID != "claim-001" {
		t.Fatalf("expected unsupported claim to stay visible, got %+v", packet.UnresolvedClaims)
	}
	if len(packet.BlockingIssues) == 0 {
		t.Fatal("expected blocking issue to remain visible")
	}
}

func TestGenerateBlockPacketForSafetyBlocker(t *testing.T) {
	packet, err := Generate(Input{
		EpisodeID: "episode-test",
		Claims: []artifacts.Claim{
			{
				ID:        "claim-001",
				Text:      "Credential exposure can leak private data.",
				Type:      "safety",
				RiskLevel: artifacts.ClaimRiskCritical,
				SourceIDs: []string{"source-001"},
				Status:    artifacts.ClaimStatusUnsupported,
			},
		},
		Verification: verification.Report{
			Decision: "request_revision",
			ClaimResults: []verification.ClaimResult{
				{ClaimID: "claim-001", Status: artifacts.ClaimStatusUnsupported, Notes: "unsupported safety claim"},
			},
			BlockingIssues: []string{"claim-001: unsupported safety claim"},
		},
		Council: council.Report{
			Consensus: council.ConsensusBlocked,
			Reviews: []council.ModelReview{
				{ModelID: "safety-reviewer", Task: "safety_review", Verdict: council.VerdictBlock, Notes: "policy blocker: unsupported credential claim"},
			},
			Dissent: []council.ModelReview{
				{ModelID: "safety-reviewer", Task: "safety_review", Verdict: council.VerdictBlock, Notes: "policy blocker: unsupported credential claim"},
			},
			BlockingObjections: []council.ModelReview{
				{ModelID: "safety-reviewer", Task: "safety_review", Verdict: council.VerdictBlock, Notes: "policy blocker: unsupported credential claim"},
			},
		},
		QualityGateStatus: "blocked",
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if packet.RecommendedDecision != artifacts.HumanDecisionBlock {
		t.Fatalf("expected block recommendation, got %s", packet.RecommendedDecision)
	}
	if len(packet.SafetyPolicyBlockers) == 0 {
		t.Fatal("expected safety/policy blocker to stay visible")
	}
	if len(packet.BlockingIssues) == 0 {
		t.Fatal("expected blocking issues to stay visible")
	}
}

func TestHumanQAReportDraftValidatesForRevisionPacket(t *testing.T) {
	packet, err := Generate(Input{
		EpisodeID: "episode-test",
		Claims: []artifacts.Claim{
			{
				ID:        "claim-001",
				Text:      "A high risk claim requires evidence.",
				Type:      "technical",
				RiskLevel: artifacts.ClaimRiskHigh,
				SourceIDs: []string{"source-001"},
				Status:    artifacts.ClaimStatusNeedsHumanReview,
			},
		},
		Verification: verification.Report{
			Decision: "request_revision",
			ClaimResults: []verification.ClaimResult{
				{ClaimID: "claim-001", Status: artifacts.ClaimStatusNeedsHumanReview, Notes: "missing locator"},
			},
			BlockingIssues: []string{"claim-001: missing locator"},
		},
		Council: council.Report{Consensus: council.ConsensusRevisionRequired},
	})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	draft := packet.ToHumanQAReportDraft()
	if draft.Reviewer != "human-required" {
		t.Fatalf("draft must keep human review explicit, got %s", draft.Reviewer)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "human_qa_report.json")
	encoded, err := json.MarshalIndent(draft, "", "  ")
	if err != nil {
		t.Fatalf("marshal draft: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write draft: %v", err)
	}
	report := artifacts.ValidatePath(path)
	if !report.Valid {
		t.Fatalf("expected draft to validate as human QA artifact: %+v", report.Issues)
	}
}

func supportedClaim(id string, risk artifacts.ClaimRisk) artifacts.Claim {
	return artifacts.Claim{
		ID:        id,
		Text:      "CI validates the change.",
		Type:      "technical",
		RiskLevel: risk,
		SourceIDs: []string{"source-001"},
		EvidenceLocators: []artifacts.EvidenceLocator{
			{SourceID: "source-001", Section: "docs"},
		},
		Status: artifacts.ClaimStatusSupported,
	}
}
