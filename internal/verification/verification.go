package verification

import (
	"fmt"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
)

// ClaimResult is the verification result for a single claim.
type ClaimResult struct {
	ClaimID string
	Status  artifacts.ClaimStatus
	Notes   string
}

// Report is the in-memory claim verification report used before artifact persistence.
type Report struct {
	Summary        string
	ClaimResults   []ClaimResult
	BlockingIssues []string
	Decision       string
}

// VerifyClaims checks deterministic source/evidence coverage and combines it
// with a model council report. It does not call external providers.
func VerifyClaims(claims []artifacts.Claim, councilReport council.Report) (Report, error) {
	if len(claims) == 0 {
		return Report{}, fmt.Errorf("at least one claim is required")
	}

	report := Report{
		ClaimResults: make([]ClaimResult, 0, len(claims)),
	}

	for _, claim := range claims {
		result := verifyClaimCoverage(claim)
		report.ClaimResults = append(report.ClaimResults, result)
		if result.Status == artifacts.ClaimStatusUnsupported || result.Status == artifacts.ClaimStatusContradicted || result.Status == artifacts.ClaimStatusNeedsHumanReview {
			if claim.RiskLevel == artifacts.ClaimRiskHigh || claim.RiskLevel == artifacts.ClaimRiskCritical {
				report.BlockingIssues = append(report.BlockingIssues, fmt.Sprintf("%s: %s", claim.ID, result.Notes))
			}
		}
	}

	if councilReport.Consensus == council.ConsensusBlocked {
		report.BlockingIssues = append(report.BlockingIssues, "model council blocked verification")
	}
	if councilReport.Consensus == council.ConsensusRevisionRequired {
		report.BlockingIssues = append(report.BlockingIssues, "model council requires revision")
	}

	if len(report.BlockingIssues) > 0 {
		report.Decision = "request_revision"
		report.Summary = "Verification found blockers that must be remediated before production release."
		return report, nil
	}

	report.Decision = "approved"
	report.Summary = "All claims have sufficient deterministic coverage for this verification pass."
	return report, nil
}

func verifyClaimCoverage(claim artifacts.Claim) ClaimResult {
	if claim.ID == "" {
		return ClaimResult{Status: artifacts.ClaimStatusUnsupported, Notes: "claim id is required"}
	}
	if claim.Text == "" {
		return ClaimResult{ClaimID: claim.ID, Status: artifacts.ClaimStatusUnsupported, Notes: "claim text is required"}
	}
	if len(claim.SourceIDs) == 0 {
		return ClaimResult{ClaimID: claim.ID, Status: artifacts.ClaimStatusUnsupported, Notes: "claim has no source references"}
	}
	if (claim.RiskLevel == artifacts.ClaimRiskHigh || claim.RiskLevel == artifacts.ClaimRiskCritical) && len(claim.EvidenceLocators) == 0 {
		return ClaimResult{ClaimID: claim.ID, Status: artifacts.ClaimStatusNeedsHumanReview, Notes: "high/critical risk claim lacks evidence locator"}
	}
	if claim.Status == artifacts.ClaimStatusContradicted || claim.Status == artifacts.ClaimStatusUnsupported {
		return ClaimResult{ClaimID: claim.ID, Status: claim.Status, Notes: "claim already marked as unresolved by upstream artifact"}
	}
	if claim.Status == artifacts.ClaimStatusNeedsHumanReview {
		return ClaimResult{ClaimID: claim.ID, Status: artifacts.ClaimStatusNeedsHumanReview, Notes: "claim requires human review before production release"}
	}
	return ClaimResult{ClaimID: claim.ID, Status: artifacts.ClaimStatusSupported, Notes: "claim has source references and required locator coverage"}
}
