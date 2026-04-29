package council

import (
	"fmt"

	"github.com/AnimusHQ/news/internal/models"
)

// Verdict is a model review outcome.
type Verdict = models.Verdict

const (
	VerdictApprove                = models.VerdictApprove
	VerdictApproveWithSuggestions = models.VerdictApproveWithSuggestions
	VerdictRequestRevision        = models.VerdictRequestRevision
	VerdictBlock                  = models.VerdictBlock
)

// Consensus is the normalized council decision.
type Consensus string

const (
	ConsensusApproved                Consensus = "approved"
	ConsensusApprovedWithSuggestions Consensus = "approved_with_suggestions"
	ConsensusRevisionRequired        Consensus = "revision_required"
	ConsensusBlocked                 Consensus = "blocked"
)

// ModelReview is one reviewer model's output.
type ModelReview = models.ModelReview

// Report is the canonical in-memory council aggregation result.
type Report struct {
	Reviews            []ModelReview `json:"reviews"`
	Consensus          Consensus     `json:"consensus"`
	Dissent            []ModelReview `json:"dissent,omitempty"`
	BlockingObjections []ModelReview `json:"blocking_objections,omitempty"`
	OperatorSummary    string        `json:"operator_summary"`
}

// Aggregate turns independent model reviews into a council report. It preserves
// dissent and treats any block verdict as a hard blocker.
func Aggregate(reviews []ModelReview) (Report, error) {
	if len(reviews) == 0 {
		return Report{}, fmt.Errorf("at least one model review is required")
	}

	report := Report{Reviews: append([]ModelReview(nil), reviews...)}
	var suggestions, revisions, blockers []ModelReview
	for _, review := range reviews {
		switch review.Verdict {
		case VerdictApprove:
			// no-op
		case VerdictApproveWithSuggestions:
			suggestions = append(suggestions, review)
		case VerdictRequestRevision:
			revisions = append(revisions, review)
			report.Dissent = append(report.Dissent, review)
		case VerdictBlock:
			blockers = append(blockers, review)
			report.Dissent = append(report.Dissent, review)
		default:
			return Report{}, fmt.Errorf("unknown verdict from %s: %q", review.ModelID, review.Verdict)
		}
	}

	report.BlockingObjections = blockers
	switch {
	case len(blockers) > 0:
		report.Consensus = ConsensusBlocked
		report.OperatorSummary = "Council blocked the artifact; human review should not approve without remediation."
	case len(revisions) > 0:
		report.Consensus = ConsensusRevisionRequired
		report.OperatorSummary = "Council requires revision before human QA approval."
	case len(suggestions) > 0:
		report.Consensus = ConsensusApprovedWithSuggestions
		report.OperatorSummary = "Council approves with suggestions preserved for human QA."
	default:
		report.Consensus = ConsensusApproved
		report.OperatorSummary = "Council approves. Human QA remains final authority."
	}

	return report, nil
}
