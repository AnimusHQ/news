package pilot

import (
	"fmt"
	"strings"
)

const (
	SchemaVersion = "1.0"

	StatusGenerated        = "generated"
	StatusNeedsReview      = "needs_review"
	StatusVerified         = "verified"
	StatusReleaseBlocked   = "release_blocked"
	StatusReleaseCandidate = "release_candidate"

	StageWorkspace          = "workspace"
	StageScript             = "script"
	StageClaudeScriptReview = "claude_script_review"
	StageVisualRequests     = "visual_shot_requests"
	StageVisualGeneration   = "visual_generation"
	StageVoiceGeneration    = "voice_generation"
	StageSubtitles          = "subtitles"
	StageRender             = "render"
	StageClaudeFinalReview  = "claude_final_review"
	StageProductionQA       = "production_qa"
	StageReleaseCandidate   = "release_candidate"
)

type GenerateRequest struct {
	EpisodeID        string
	Prompt           string
	Language         string
	Duration         string
	Platforms        []string
	VisualProvider   string
	VoiceProvider    string
	SubtitleProvider string
	RenderProvider   string
	ClaudeReview     string
	OutDir           string
}

type ImportClaudeReviewRequest struct {
	EpisodeDir string
	Kind       string
	File       string
}

type ImportAssetRequest struct {
	EpisodeDir string
	ShotID     string
	File       string
}

type Result struct {
	EpisodeID     string
	EpisodeDir    string
	Stage         string
	BlockedGate   string
	NextAction    string
	Ready         bool
	ReleasePath   string
	Artifacts     []string
	Missing       []string
	BlockingIssue string
}

func (r Result) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "episode: %s\n", r.EpisodeID)
	fmt.Fprintf(&b, "episode_dir: %s\n", r.EpisodeDir)
	fmt.Fprintf(&b, "stage: %s\n", r.Stage)
	if r.ReleasePath != "" {
		fmt.Fprintf(&b, "release_candidate: %s\n", r.ReleasePath)
	}
	fmt.Fprintf(&b, "release_candidate_ready: %v\n", r.Ready)
	if r.BlockedGate != "" {
		fmt.Fprintf(&b, "blocking_gate: %s\n", r.BlockedGate)
	}
	if r.BlockingIssue != "" {
		fmt.Fprintf(&b, "blocking_issue: %s\n", r.BlockingIssue)
	}
	if r.NextAction != "" {
		fmt.Fprintf(&b, "next_action: %s\n", r.NextAction)
	}
	if len(r.Artifacts) > 0 {
		fmt.Fprintf(&b, "existing_artifacts: %s\n", strings.Join(r.Artifacts, ", "))
	}
	if len(r.Missing) > 0 {
		fmt.Fprintf(&b, "missing_artifacts: %s\n", strings.Join(r.Missing, ", "))
	}
	return strings.TrimRight(b.String(), "\n")
}

type ValidationReport struct {
	EpisodeID               string   `json:"episode_id"`
	EpisodeDir              string   `json:"episode_dir"`
	Valid                   bool     `json:"valid"`
	CurrentStage            string   `json:"current_stage"`
	ReleaseCandidateReady   bool     `json:"release_candidate_ready"`
	ReleaseCandidatePath    string   `json:"release_candidate_path,omitempty"`
	ExistingArtifacts       []string `json:"existing_artifacts"`
	MissingArtifacts        []string `json:"missing_artifacts"`
	BlockingGate            string   `json:"blocking_gate,omitempty"`
	NextAction              string   `json:"next_action,omitempty"`
	Issues                  []string `json:"issues,omitempty"`
	NoPublicPublishingPath  bool     `json:"no_public_publishing_path"`
	FinalClaudeReviewPassed bool     `json:"final_claude_review_passed"`
}

type ProviderSelections struct {
	Visual       string `json:"visual"`
	Voice        string `json:"voice"`
	Subtitle     string `json:"subtitle"`
	Render       string `json:"render"`
	ClaudeReview string `json:"claude_review"`
}

type EpisodeManifest struct {
	SchemaVersion  string             `json:"schema_version"`
	EpisodeID      string             `json:"episode_id"`
	CreatedAt      string             `json:"created_at"`
	UpdatedAt      string             `json:"updated_at"`
	Status         string             `json:"status"`
	OriginalPrompt string             `json:"original_prompt"`
	Language       string             `json:"language"`
	Duration       string             `json:"duration"`
	DurationSec    float64            `json:"duration_sec"`
	Platforms      []string           `json:"platforms"`
	Providers      ProviderSelections `json:"providers"`
	ContentHash    string             `json:"content_hash,omitempty"`
}

type ScriptManifest struct {
	SchemaVersion        string  `json:"schema_version"`
	EpisodeID            string  `json:"episode_id"`
	CreatedAt            string  `json:"created_at"`
	Status               string  `json:"status"`
	ScriptPath           string  `json:"script_path"`
	ScriptHash           string  `json:"script_hash"`
	EstimatedDurationSec float64 `json:"estimated_duration_sec"`
	SourcePromptHash     string  `json:"source_prompt_hash"`
	ContentHash          string  `json:"content_hash,omitempty"`
}

type VisualShotRequests struct {
	SchemaVersion    string              `json:"schema_version"`
	EpisodeID        string              `json:"episode_id"`
	CreatedAt        string              `json:"created_at"`
	Status           string              `json:"status"`
	SourceScriptHash string              `json:"source_script_hash"`
	Shots            []VisualShotRequest `json:"shots"`
	ContentHash      string              `json:"content_hash,omitempty"`
}

type VisualShotRequest struct {
	ShotID            string   `json:"shot_id"`
	SceneID           string   `json:"scene_id"`
	DurationSec       float64  `json:"duration_sec"`
	Prompt            string   `json:"prompt"`
	NegativePrompt    string   `json:"negative_prompt"`
	Width             int      `json:"width"`
	Height            int      `json:"height"`
	FPS               int      `json:"fps"`
	Camera            string   `json:"camera"`
	Motion            string   `json:"motion"`
	SourceScriptLines []string `json:"source_script_lines"`
}

type ClaudeReviewResponse struct {
	SchemaVersion                 string   `json:"schema_version"`
	EpisodeID                     string   `json:"episode_id"`
	Verdict                       string   `json:"verdict"`
	ProductionReadiness           int      `json:"production_readiness"`
	BlockingIssues                []string `json:"blocking_issues"`
	SuggestedRevisions            []string `json:"suggested_revisions"`
	ApprovedScriptHash            string   `json:"approved_script_hash,omitempty"`
	CanContinueToVisualGeneration bool     `json:"can_continue_to_visual_generation,omitempty"`
	CanReleaseCandidate           bool     `json:"can_release_candidate,omitempty"`
	OperatorOverride              bool     `json:"operator_override,omitempty"`
	OperatorOverrideReason        string   `json:"operator_override_reason,omitempty"`
}

type ExternalVisualInput struct {
	SchemaVersion string              `json:"schema_version"`
	EpisodeID     string              `json:"episode_id"`
	Provider      string              `json:"provider"`
	Shots         []VisualShotRequest `json:"shots"`
	OutputDir     string              `json:"output_dir"`
}

type ExternalVisualResponse struct {
	SchemaVersion string                 `json:"schema_version"`
	EpisodeID     string                 `json:"episode_id"`
	Provider      string                 `json:"provider"`
	Shots         []ExternalVisualOutput `json:"shots"`
}

type ExternalVisualOutput struct {
	ShotID      string  `json:"shot_id"`
	Status      string  `json:"status"`
	OutputPath  string  `json:"output_path"`
	OutputHash  string  `json:"output_hash,omitempty"`
	DurationSec float64 `json:"duration_sec"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FPS         int     `json:"fps"`
}

type ExternalVoiceInput struct {
	SchemaVersion string `json:"schema_version"`
	EpisodeID     string `json:"episode_id"`
	Language      string `json:"language"`
	Text          string `json:"text"`
	OutputDir     string `json:"output_dir"`
}

type ExternalVoiceResponse struct {
	SchemaVersion         string  `json:"schema_version"`
	EpisodeID             string  `json:"episode_id"`
	Provider              string  `json:"provider"`
	OutputPath            string  `json:"output_path"`
	OutputHash            string  `json:"output_hash,omitempty"`
	DurationSec           float64 `json:"duration_sec"`
	SampleRate            int     `json:"sample_rate"`
	VoiceConsentReference string  `json:"voice_consent_reference,omitempty"`
}

type SubtitleSidecarInput struct {
	SchemaVersion string `json:"schema_version"`
	EpisodeID     string `json:"episode_id"`
	Provider      string `json:"provider"`
	Language      string `json:"language"`
	AudioPath     string `json:"audio_path"`
	OutputDir     string `json:"output_dir"`
}

type SubtitleSidecarResponse struct {
	SchemaVersion  string `json:"schema_version"`
	EpisodeID      string `json:"episode_id"`
	Provider       string `json:"provider"`
	TranscriptPath string `json:"transcript_path"`
	SRTPath        string `json:"srt_path"`
	ASSPath        string `json:"ass_path,omitempty"`
	WordTimestamps bool   `json:"word_timestamps"`
	SafeZone       bool   `json:"safe_zone"`
	Sync           bool   `json:"sync"`
}

type ProductionQAReport struct {
	SchemaVersion  string            `json:"schema_version"`
	EpisodeID      string            `json:"episode_id"`
	ArtifactID     string            `json:"artifact_id"`
	CreatedAt      string            `json:"created_at"`
	Status         string            `json:"status"`
	Checks         map[string]string `json:"checks"`
	BlockingIssues []string          `json:"blocking_issues"`
	Decision       string            `json:"decision"`
	ContentHash    string            `json:"content_hash,omitempty"`
}

type PublishManifest struct {
	SchemaVersion         string   `json:"schema_version"`
	EpisodeID             string   `json:"episode_id"`
	ArtifactID            string   `json:"artifact_id"`
	CreatedAt             string   `json:"created_at"`
	Status                string   `json:"status"`
	Mode                  string   `json:"mode"`
	LivePublishingEnabled bool     `json:"live_publishing_enabled"`
	Platforms             []string `json:"platforms"`
	Visibility            string   `json:"visibility"`
	ReleaseCandidatePath  string   `json:"release_candidate_path"`
	RenderManifestRef     string   `json:"render_manifest_ref"`
	ProductionQARef       string   `json:"production_qa_ref"`
	FinalReviewRef        string   `json:"final_review_ref"`
	ContentHash           string   `json:"content_hash,omitempty"`
}
