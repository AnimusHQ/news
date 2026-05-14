package models

// Modality describes model input/output capabilities.
type Modality string

const (
	ModalityText   Modality = "text"
	ModalityVision Modality = "vision"
	ModalityAudio  Modality = "audio"
	ModalityVideo  Modality = "video"
	ModalityCode   Modality = "code"
)

// Capability describes a task category a model can handle.
type Capability string

const (
	CapabilityResearchSynthesis     Capability = "research_synthesis"
	CapabilityTechnicalVerification Capability = "technical_verification"
	CapabilityScriptWriting         Capability = "script_writing"
	CapabilityEditorialReview       Capability = "editorial_review"
	CapabilityStoryboardPlanning    Capability = "storyboard_planning"
	CapabilityVisualReasoning       Capability = "visual_reasoning"
	CapabilitySafetyReview          Capability = "safety_review"
	CapabilityAnalytics             Capability = "analytics_interpretation"
)

// PrivacyTier controls which data classes a model may receive.
type PrivacyTier string

const (
	PrivacyTierPublic           PrivacyTier = "public"
	PrivacyTierInternalApproved PrivacyTier = "internal_approved"
	PrivacyTierRestricted       PrivacyTier = "restricted"
	PrivacyTierLocalOnly        PrivacyTier = "local_only"
)

// ModelStatus describes operational availability.
type ModelStatus string

const (
	ModelStatusActive   ModelStatus = "active"
	ModelStatusDegraded ModelStatus = "degraded"
	ModelStatusDisabled ModelStatus = "disabled"
)

// RiskLevel controls whether a task needs a single model or a council.
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// Verdict is a normalized model review outcome.
type Verdict string

const (
	VerdictApprove                Verdict = "approve"
	VerdictApproveWithSuggestions Verdict = "approve_with_suggestions"
	VerdictRequestRevision        Verdict = "request_revision"
	VerdictBlock                  Verdict = "block"
)

// ModelReview is one reviewer model's normalized output.
type ModelReview struct {
	ModelID    string  `json:"model_id"`
	Provider   string  `json:"provider"`
	Task       string  `json:"task"`
	Verdict    Verdict `json:"verdict"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes"`
}

// ModelRecord is a provider-agnostic registry record.
type ModelRecord struct {
	ID           string       `json:"id" yaml:"id"`
	Provider     string       `json:"provider" yaml:"provider"`
	Version      string       `json:"version" yaml:"version"`
	Status       ModelStatus  `json:"status" yaml:"status"`
	PrivacyTier  PrivacyTier  `json:"privacy_tier" yaml:"privacy_tier"`
	Modalities   []Modality   `json:"modalities" yaml:"modalities"`
	Capabilities []Capability `json:"capabilities" yaml:"capabilities"`
	CostScore    float64      `json:"cost_score,omitempty" yaml:"cost_score,omitempty"`
	LatencyScore float64      `json:"latency_score,omitempty" yaml:"latency_score,omitempty"`
	QualityScore float64      `json:"quality_score,omitempty" yaml:"quality_score,omitempty"`
}

// TaskRequest is the normalized input to the model router.
type TaskRequest struct {
	TaskID      string
	Capability  Capability
	RiskLevel   RiskLevel
	Modality    Modality
	PrivacyTier PrivacyTier
	EpisodeID   string
	ArtifactID  string
	Description string
}

// RoutingDecision explains the router's choice.
type RoutingDecision struct {
	Selected        []ModelRecord
	Rejected        []RejectedModel
	Policy          string
	FallbackReasons []string
}

// RejectedModel records why a candidate was not selected.
type RejectedModel struct {
	ModelID string
	Reason  string
}
