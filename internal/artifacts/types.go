package artifacts

import "time"

// ArtifactStatus describes lifecycle status for persisted pipeline artifacts.
type ArtifactStatus string

const (
	ArtifactStatusDraft      ArtifactStatus = "draft"
	ArtifactStatusInReview   ArtifactStatus = "in_review"
	ArtifactStatusApproved   ArtifactStatus = "approved"
	ArtifactStatusRejected   ArtifactStatus = "rejected"
	ArtifactStatusSuperseded ArtifactStatus = "superseded"
	ArtifactStatusLocked     ArtifactStatus = "locked"
)

// IsTerminalImmutable reports whether a status marks an artifact as frozen
// against further mutation. Approved and locked artifacts must never be mutated
// in place; a new versioned artifact must be produced instead.
func (s ArtifactStatus) IsTerminalImmutable() bool {
	return s == ArtifactStatusApproved || s == ArtifactStatusLocked
}

// Metadata is embedded into machine-readable artifacts.
type Metadata struct {
	SchemaVersion   string         `json:"schema_version" yaml:"schema_version"`
	EpisodeID       string         `json:"episode_id" yaml:"episode_id"`
	ArtifactID      string         `json:"artifact_id" yaml:"artifact_id"`
	CreatedAt       time.Time      `json:"created_at" yaml:"created_at"`
	CreatedBy       string         `json:"created_by" yaml:"created_by"`
	SourceArtifacts []string       `json:"source_artifacts,omitempty" yaml:"source_artifacts,omitempty"`
	ContentHash     string         `json:"content_hash,omitempty" yaml:"content_hash,omitempty"`
	Status          ArtifactStatus `json:"status" yaml:"status"`
}

// ClaimRisk captures verification and release risk for factual claims.
type ClaimRisk string

const (
	ClaimRiskLow      ClaimRisk = "low"
	ClaimRiskMedium   ClaimRisk = "medium"
	ClaimRiskHigh     ClaimRisk = "high"
	ClaimRiskCritical ClaimRisk = "critical"
)

// ClaimStatus describes verification status for a factual claim.
type ClaimStatus string

const (
	ClaimStatusSupported          ClaimStatus = "supported"
	ClaimStatusPartiallySupported ClaimStatus = "partially_supported"
	ClaimStatusUnsupported        ClaimStatus = "unsupported"
	ClaimStatusContradicted       ClaimStatus = "contradicted"
	ClaimStatusNeedsHumanReview   ClaimStatus = "needs_human_review"
	ClaimStatusRemoved            ClaimStatus = "removed"
)

// Source describes a source used as evidence.
type Source struct {
	ID           string `json:"source_id" yaml:"source_id"`
	Title        string `json:"title" yaml:"title"`
	URI          string `json:"uri" yaml:"uri"`
	Type         string `json:"type" yaml:"type"`
	TrustLevel   string `json:"trust_level" yaml:"trust_level"`
	ContentHash  string `json:"content_hash,omitempty" yaml:"content_hash,omitempty"`
	LicenseNotes string `json:"license_notes,omitempty" yaml:"license_notes,omitempty"`
}

// EvidenceLocator points to the specific source location supporting a claim.
type EvidenceLocator struct {
	SourceID  string `json:"source_id" yaml:"source_id"`
	Section   string `json:"section,omitempty" yaml:"section,omitempty"`
	Range     string `json:"range,omitempty" yaml:"range,omitempty"`
	QuoteHash string `json:"quote_hash,omitempty" yaml:"quote_hash,omitempty"`
}

// Claim is a factual statement extracted from script/research.
type Claim struct {
	ID               string            `json:"claim_id" yaml:"claim_id"`
	Text             string            `json:"text" yaml:"text"`
	Type             string            `json:"type" yaml:"type"`
	RiskLevel        ClaimRisk         `json:"risk_level" yaml:"risk_level"`
	SourceIDs        []string          `json:"source_ids" yaml:"source_ids"`
	EvidenceLocators []EvidenceLocator `json:"evidence_locators" yaml:"evidence_locators"`
	Status           ClaimStatus       `json:"verification_status" yaml:"verification_status"`
}

// HumanDecision is a human gate decision.
type HumanDecision string

const (
	HumanDecisionApprove               HumanDecision = "approve"
	HumanDecisionApproveWithMinorEdits HumanDecision = "approve_with_minor_edits"
	HumanDecisionRequestRevision       HumanDecision = "request_revision"
	HumanDecisionBlock                 HumanDecision = "block"
)

// PublishVisibility is intentionally safe-by-default.
type PublishVisibility string

const (
	PublishVisibilityPrivate   PublishVisibility = "private"
	PublishVisibilityScheduled PublishVisibility = "scheduled"
	PublishVisibilityPublic    PublishVisibility = "public"
)
