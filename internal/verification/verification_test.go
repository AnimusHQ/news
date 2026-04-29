package verification

import (
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
)

func TestVerifyClaimsApprovesSupportedClaims(t *testing.T) {
	report, err := VerifyClaims([]artifacts.Claim{
		{
			ID:        "claim-1",
			Text:      "Supported claim.",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"source-1"},
			Status:    artifacts.ClaimStatusSupported,
		},
	}, council.Report{Consensus: council.ConsensusApproved})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if report.Decision != "approved" {
		t.Fatalf("expected approved decision, got %s", report.Decision)
	}
	if len(report.BlockingIssues) != 0 {
		t.Fatalf("expected no blockers, got %v", report.BlockingIssues)
	}
}

func TestVerifyClaimsBlocksHighRiskNeedsReview(t *testing.T) {
	report, err := VerifyClaims([]artifacts.Claim{
		{
			ID:        "claim-1",
			Text:      "High-risk claim.",
			RiskLevel: artifacts.ClaimRiskHigh,
			SourceIDs: []string{"source-1"},
			Status:    artifacts.ClaimStatusNeedsHumanReview,
		},
	}, council.Report{Consensus: council.ConsensusApproved})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if report.Decision != "request_revision" {
		t.Fatalf("expected request_revision, got %s", report.Decision)
	}
	if len(report.BlockingIssues) == 0 {
		t.Fatal("expected blocking issues")
	}
}

func TestVerifyClaimsBlocksCouncilRevision(t *testing.T) {
	report, err := VerifyClaims([]artifacts.Claim{
		{
			ID:        "claim-1",
			Text:      "Supported claim.",
			RiskLevel: artifacts.ClaimRiskMedium,
			SourceIDs: []string{"source-1"},
			Status:    artifacts.ClaimStatusSupported,
		},
	}, council.Report{Consensus: council.ConsensusRevisionRequired})
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if report.Decision != "request_revision" {
		t.Fatalf("expected request_revision, got %s", report.Decision)
	}
	if len(report.BlockingIssues) == 0 {
		t.Fatal("expected council blocker")
	}
}

func TestVerifyClaimsRequiresClaim(t *testing.T) {
	_, err := VerifyClaims(nil, council.Report{Consensus: council.ConsensusApproved})
	if err == nil {
		t.Fatal("expected empty claim set to fail")
	}
}
