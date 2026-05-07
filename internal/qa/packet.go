package qa

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/verification"
)

const (
	SchemaVersion = "1.0"
	packetStatus  = "draft"
)

// Input contains the deterministic artifacts a human operator needs before
// making a QA decision. It does not contain or imply human approval.
type Input struct {
	EpisodeID         string
	ArtifactID        string
	EpisodePurpose    string
	EpisodeFormat     string
	ScriptPath        string
	ResearchSummary   string
	Claims            []artifacts.Claim
	Verification      verification.Report
	Council           council.Report
	QualityGateStatus string
	OperatorNotes     []string
}

// ClaimRiskSummary captures the shape of the claims under review.
type ClaimRiskSummary struct {
	Low                int `json:"low"`
	Medium             int `json:"medium"`
	High               int `json:"high"`
	Critical           int `json:"critical"`
	Unsupported        int `json:"unsupported"`
	Contradicted       int `json:"contradicted"`
	NeedsReview        int `json:"needs_review"`
	PartiallySupported int `json:"partially_supported"`
}

// ClaimIssue records an unresolved claim that must stay visible to the human.
type ClaimIssue struct {
	ClaimID   string                `json:"claim_id"`
	Text      string                `json:"text,omitempty"`
	Type      string                `json:"type,omitempty"`
	RiskLevel artifacts.ClaimRisk   `json:"risk_level,omitempty"`
	Status    artifacts.ClaimStatus `json:"status"`
	Notes     string                `json:"notes,omitempty"`
}

// Packet is the human-facing QA decision packet. RecommendedDecision is a
// machine recommendation only; it is not a persisted operator approval.
type Packet struct {
	SchemaVersion        string                  `json:"schema_version"`
	EpisodeID            string                  `json:"episode_id"`
	ArtifactID           string                  `json:"artifact_id"`
	Status               string                  `json:"status"`
	Purpose              string                  `json:"purpose"`
	Format               string                  `json:"format"`
	ScriptPath           string                  `json:"script_path,omitempty"`
	ResearchSummary      string                  `json:"research_summary,omitempty"`
	QualityGateStatus    string                  `json:"quality_gate_status,omitempty"`
	ClaimRiskSummary     ClaimRiskSummary        `json:"claim_risk_summary"`
	UnresolvedClaims     []ClaimIssue            `json:"unresolved_claims,omitempty"`
	ModelApprovals       []council.ModelReview   `json:"model_approvals,omitempty"`
	ModelDissent         []council.ModelReview   `json:"model_dissent,omitempty"`
	BlockingIssues       []string                `json:"blocking_issues,omitempty"`
	SafetyPolicyBlockers []string                `json:"safety_policy_blockers,omitempty"`
	UnresolvedRisks      []string                `json:"unresolved_risks,omitempty"`
	RecommendedDecision  artifacts.HumanDecision `json:"recommended_decision"`
	OperatorNotes        []string                `json:"operator_notes,omitempty"`
}

// HumanQAReportDraft is a validation-compatible draft shape. It is suitable for
// an operator to review, but it must not be treated as a recorded human approval.
type HumanQAReportDraft struct {
	SchemaVersion   string   `json:"schema_version"`
	EpisodeID       string   `json:"episode_id"`
	ArtifactID      string   `json:"artifact_id"`
	CreatedBy       string   `json:"created_by"`
	Status          string   `json:"status"`
	Reviewer        string   `json:"reviewer"`
	Decision        string   `json:"decision"`
	Notes           string   `json:"notes"`
	RequiredChanges []string `json:"required_changes,omitempty"`
}

// Generate builds a deterministic QA packet from already-generated artifacts.
func Generate(input Input) (Packet, error) {
	if strings.TrimSpace(input.EpisodeID) == "" {
		return Packet{}, fmt.Errorf("episode id is required")
	}
	if len(input.Claims) == 0 {
		return Packet{}, fmt.Errorf("at least one claim is required")
	}

	packet := Packet{
		SchemaVersion:     SchemaVersion,
		EpisodeID:         input.EpisodeID,
		ArtifactID:        artifactID(input),
		Status:            packetStatus,
		Purpose:           defaultText(input.EpisodePurpose, "unspecified episode purpose"),
		Format:            defaultText(input.EpisodeFormat, "unspecified episode format"),
		ScriptPath:        strings.TrimSpace(input.ScriptPath),
		ResearchSummary:   strings.TrimSpace(input.ResearchSummary),
		QualityGateStatus: strings.TrimSpace(input.QualityGateStatus),
		ClaimRiskSummary:  summarizeClaimRisks(input.Claims),
		OperatorNotes:     sortedStrings(input.OperatorNotes),
	}

	claimByID := map[string]artifacts.Claim{}
	for _, claim := range input.Claims {
		claimByID[claim.ID] = claim
	}

	packet.UnresolvedClaims = unresolvedClaims(input.Claims, input.Verification, claimByID)
	updateUnresolvedCounts(&packet.ClaimRiskSummary, packet.UnresolvedClaims)
	packet.ModelApprovals = modelApprovals(input.Council.Reviews)
	packet.ModelDissent = sortedReviews(input.Council.Dissent)
	packet.BlockingIssues = blockingIssues(input)
	packet.SafetyPolicyBlockers = safetyPolicyBlockers(input.Council, packet.UnresolvedClaims, packet.BlockingIssues)
	packet.UnresolvedRisks = unresolvedRisks(packet.UnresolvedClaims, packet.ModelDissent, packet.BlockingIssues, packet.SafetyPolicyBlockers, input.QualityGateStatus)
	packet.RecommendedDecision = recommend(packet, input)

	return packet, nil
}

// ToHumanQAReportDraft returns a validation-compatible draft report. The
// reviewer marker makes the human-in-the-loop requirement explicit.
func (p Packet) ToHumanQAReportDraft() HumanQAReportDraft {
	requiredChanges := append([]string{}, p.BlockingIssues...)
	for _, risk := range p.UnresolvedRisks {
		if !slices.Contains(requiredChanges, risk) {
			requiredChanges = append(requiredChanges, risk)
		}
	}
	sort.Strings(requiredChanges)

	return HumanQAReportDraft{
		SchemaVersion:   p.SchemaVersion,
		EpisodeID:       p.EpisodeID,
		ArtifactID:      "human-qa-" + p.EpisodeID + "-draft-from-packet",
		CreatedBy:       "system:qa-packet-generator",
		Status:          packetStatus,
		Reviewer:        "human-required",
		Decision:        string(p.RecommendedDecision),
		Notes:           "Draft generated for operator review; this is not recorded human approval.",
		RequiredChanges: requiredChanges,
	}
}

func artifactID(input Input) string {
	if strings.TrimSpace(input.ArtifactID) != "" {
		return strings.TrimSpace(input.ArtifactID)
	}
	return "human-qa-packet-" + input.EpisodeID + "-v1"
}

func defaultText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func summarizeClaimRisks(claims []artifacts.Claim) ClaimRiskSummary {
	var summary ClaimRiskSummary
	for _, claim := range claims {
		switch claim.RiskLevel {
		case artifacts.ClaimRiskLow:
			summary.Low++
		case artifacts.ClaimRiskMedium:
			summary.Medium++
		case artifacts.ClaimRiskHigh:
			summary.High++
		case artifacts.ClaimRiskCritical:
			summary.Critical++
		}
	}
	return summary
}

func updateUnresolvedCounts(summary *ClaimRiskSummary, issues []ClaimIssue) {
	for _, issue := range issues {
		switch issue.Status {
		case artifacts.ClaimStatusUnsupported:
			summary.Unsupported++
		case artifacts.ClaimStatusContradicted:
			summary.Contradicted++
		case artifacts.ClaimStatusNeedsHumanReview:
			summary.NeedsReview++
		case artifacts.ClaimStatusPartiallySupported:
			summary.PartiallySupported++
		}
	}
}

func unresolvedClaims(claims []artifacts.Claim, report verification.Report, claimByID map[string]artifacts.Claim) []ClaimIssue {
	var issues []ClaimIssue
	seen := map[string]bool{}

	for _, result := range report.ClaimResults {
		if !isUnresolvedStatus(result.Status) {
			continue
		}
		claim := claimByID[result.ClaimID]
		issues = append(issues, ClaimIssue{
			ClaimID:   result.ClaimID,
			Text:      claim.Text,
			Type:      claim.Type,
			RiskLevel: claim.RiskLevel,
			Status:    result.Status,
			Notes:     result.Notes,
		})
		seen[result.ClaimID] = true
	}

	for _, claim := range claims {
		if seen[claim.ID] || !isUnresolvedStatus(claim.Status) {
			continue
		}
		issues = append(issues, ClaimIssue{
			ClaimID:   claim.ID,
			Text:      claim.Text,
			Type:      claim.Type,
			RiskLevel: claim.RiskLevel,
			Status:    claim.Status,
			Notes:     "claim artifact is unresolved and has no overriding verification result",
		})
	}

	sort.SliceStable(issues, func(i, j int) bool {
		return issues[i].ClaimID < issues[j].ClaimID
	})
	return issues
}

func isUnresolvedStatus(status artifacts.ClaimStatus) bool {
	switch status {
	case artifacts.ClaimStatusUnsupported,
		artifacts.ClaimStatusContradicted,
		artifacts.ClaimStatusNeedsHumanReview,
		artifacts.ClaimStatusPartiallySupported:
		return true
	default:
		return false
	}
}

func modelApprovals(reviews []council.ModelReview) []council.ModelReview {
	var approvals []council.ModelReview
	for _, review := range reviews {
		if review.Verdict == council.VerdictApprove || review.Verdict == council.VerdictApproveWithSuggestions {
			approvals = append(approvals, review)
		}
	}
	return sortedReviews(approvals)
}

func sortedReviews(reviews []council.ModelReview) []council.ModelReview {
	out := append([]council.ModelReview(nil), reviews...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Task != out[j].Task {
			return out[i].Task < out[j].Task
		}
		if out[i].ModelID != out[j].ModelID {
			return out[i].ModelID < out[j].ModelID
		}
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Notes < out[j].Notes
	})
	return out
}

func blockingIssues(input Input) []string {
	issues := append([]string{}, input.Verification.BlockingIssues...)
	for _, objection := range input.Council.BlockingObjections {
		issues = append(issues, fmt.Sprintf("%s: %s", reviewLabel(objection), strings.TrimSpace(objection.Notes)))
	}
	if input.Council.Consensus == council.ConsensusBlocked {
		issues = append(issues, "model council consensus is blocked")
	}
	if isFailingQualityStatus(input.QualityGateStatus) {
		issues = append(issues, "quality gate status is "+strings.TrimSpace(input.QualityGateStatus))
	}
	return sortedStrings(compactStrings(issues))
}

func safetyPolicyBlockers(report council.Report, unresolved []ClaimIssue, blocking []string) []string {
	var blockers []string
	for _, review := range append(append([]council.ModelReview{}, report.Dissent...), report.BlockingObjections...) {
		text := strings.ToLower(review.Task + " " + review.Notes)
		if containsAny(text, "safety", "policy", "unsafe", "security", "credential", "secret", "private data") {
			blockers = append(blockers, fmt.Sprintf("%s: %s", reviewLabel(review), strings.TrimSpace(review.Notes)))
		}
	}
	for _, issue := range unresolved {
		text := strings.ToLower(issue.Type + " " + issue.Text + " " + issue.Notes)
		if containsAny(text, "safety", "policy", "unsafe", "security", "credential", "secret", "private data") {
			blockers = append(blockers, fmt.Sprintf("%s: %s", issue.ClaimID, defaultText(issue.Notes, "safety or policy claim remains unresolved")))
		}
	}
	for _, issue := range blocking {
		if containsAny(strings.ToLower(issue), "safety", "policy", "unsafe", "security", "credential", "secret", "private data") {
			blockers = append(blockers, issue)
		}
	}
	return sortedStrings(compactStrings(blockers))
}

func unresolvedRisks(unresolved []ClaimIssue, dissent []council.ModelReview, blocking []string, safety []string, qualityStatus string) []string {
	var risks []string
	for _, issue := range unresolved {
		risks = append(risks, fmt.Sprintf("%s %s claim remains %s", issue.RiskLevel, issue.ClaimID, issue.Status))
	}
	for _, review := range dissent {
		risks = append(risks, fmt.Sprintf("%s dissent: %s", reviewLabel(review), strings.TrimSpace(review.Notes)))
	}
	risks = append(risks, blocking...)
	risks = append(risks, safety...)
	if qualityStatus != "" && !isPassingQualityStatus(qualityStatus) {
		risks = append(risks, "quality gate status requires operator attention: "+strings.TrimSpace(qualityStatus))
	}
	return sortedStrings(compactStrings(risks))
}

func recommend(packet Packet, input Input) artifacts.HumanDecision {
	if input.Council.Consensus == council.ConsensusBlocked || len(packet.SafetyPolicyBlockers) > 0 || hasHardClaimBlocker(packet.UnresolvedClaims) {
		return artifacts.HumanDecisionBlock
	}
	if input.Council.Consensus == council.ConsensusRevisionRequired ||
		len(packet.BlockingIssues) > 0 ||
		hasHighRiskUnresolved(packet.UnresolvedClaims) ||
		isRevisionVerificationDecision(input.Verification.Decision) ||
		isFailingQualityStatus(input.QualityGateStatus) {
		return artifacts.HumanDecisionRequestRevision
	}
	if input.Council.Consensus == council.ConsensusApprovedWithSuggestions ||
		len(packet.UnresolvedClaims) > 0 ||
		len(packet.ModelDissent) > 0 {
		return artifacts.HumanDecisionApproveWithMinorEdits
	}
	return artifacts.HumanDecisionApprove
}

func hasHardClaimBlocker(issues []ClaimIssue) bool {
	for _, issue := range issues {
		if issue.RiskLevel != artifacts.ClaimRiskCritical {
			continue
		}
		if issue.Status == artifacts.ClaimStatusContradicted || issue.Status == artifacts.ClaimStatusUnsupported {
			return true
		}
	}
	return false
}

func hasHighRiskUnresolved(issues []ClaimIssue) bool {
	for _, issue := range issues {
		if issue.RiskLevel == artifacts.ClaimRiskHigh || issue.RiskLevel == artifacts.ClaimRiskCritical {
			return true
		}
	}
	return false
}

func isRevisionVerificationDecision(decision string) bool {
	decision = strings.ToLower(strings.TrimSpace(decision))
	return decision != "" && decision != "approved" && decision != "approve"
}

func isPassingQualityStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return status == "" || status == "passed" || status == "approved" || status == "green" || status == "dry_run"
}

func isFailingQualityStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return strings.Contains(status, "fail") || strings.Contains(status, "block") || strings.Contains(status, "rejected")
}

func reviewLabel(review council.ModelReview) string {
	if review.Task == "" {
		return review.ModelID
	}
	return review.ModelID + "/" + review.Task
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		out = append(out, value)
	}
	return out
}

func sortedStrings(values []string) []string {
	out := append([]string{}, compactStrings(values)...)
	sort.Strings(out)
	return slices.Compact(out)
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
