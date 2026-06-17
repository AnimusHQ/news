// Package shortform defines the typed, validated, content-hashed artifact
// contracts that extend the News canonical graph with the short-form video
// pipeline (storyboard images → visual shots → voiceover → subtitles → render →
// production candidate → release → guarded publish). See docs/adr/0001.
//
// Every artifact embeds the common envelope, is validated against a committed
// JSON Schema plus semantic Go checks (validate.go), and carries a deterministic
// content hash (contenthash package).
package shortform

import "github.com/AnimusHQ/news/internal/shortform/contenthash"

// SchemaVersion is the envelope schema version for all short-form artifacts.
const SchemaVersion = "1.0"

// Envelope status values (mirrors internal/artifacts.ArtifactStatus).
const (
	StatusDraft      = "draft"
	StatusInReview   = "in_review"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusSuperseded = "superseded"
	StatusLocked     = "locked"
)

// Artifact kinds. These also name the committed schema files
// (internal/shortform/schemas/<kind>.schema.json).
const (
	KindStoryboardImageManifest   = "storyboard_image_manifest"
	KindVisualShotManifest        = "visual_shot_manifest"
	KindVoiceoverManifest         = "voiceover_manifest"
	KindSubtitleManifest          = "subtitle_manifest"
	KindShortRenderManifest       = "short_render_manifest"
	KindProductionCandidate       = "production_candidate"
	KindReleaseApproval           = "release_approval"
	KindUploadPostPublishManifest = "uploadpost_publish_manifest"
)

// Fixed render target for short-form vertical video (§7).
const (
	TargetResolution = "1080x1920"
	TargetAspect     = "9:16"
	TargetFPS        = 30
	TargetVideoCodec = "h264"
	TargetAudioCodec = "aac"
)

// Envelope is the common metadata carried by every short-form artifact.
type Envelope struct {
	SchemaVersion   string   `json:"schema_version"`
	EpisodeID       string   `json:"episode_id"`
	ArtifactID      string   `json:"artifact_id"`
	CreatedAt       string   `json:"created_at"`
	CreatedBy       string   `json:"created_by"`
	SourceArtifacts []string `json:"source_artifacts,omitempty"`
	ContentHash     string   `json:"content_hash,omitempty"`
	Status          string   `json:"status"`
}

// Artifact is implemented by every short-form artifact so shared tooling
// (hashing, validation) can operate generically.
type Artifact interface {
	Kind() string
	EnvelopeRef() *Envelope
}

// Stamp computes and stores the deterministic content hash on the artifact's
// envelope. It excludes the content_hash field itself, so stamping is
// idempotent.
func Stamp(a Artifact) error {
	hash, err := contenthash.Compute(a)
	if err != nil {
		return err
	}
	a.EnvelopeRef().ContentHash = hash
	return nil
}

// ----- Shared value objects -----

// ProviderRef records which provider produced an artifact.
type ProviderRef struct {
	Name    string `json:"name"`
	Model   string `json:"model,omitempty"`
	Version string `json:"version,omitempty"`
}

// RenderTarget pins the vertical short-form output format.
type RenderTarget struct {
	Resolution string `json:"resolution"`
	Aspect     string `json:"aspect"`
	FPS        int    `json:"fps"`
	VideoCodec string `json:"video_codec"`
}

// MediaOutput is a produced media file reference with provenance.
type MediaOutput struct {
	Path        string  `json:"path"`
	Hash        string  `json:"hash"`
	DurationSec float64 `json:"duration_sec"`
	Format      string  `json:"format"`
}

// ----- 1. storyboard_image_manifest.json -----

// StoryboardImageManifest records imported ChatGPT reference images with per
// image hash and approval, preventing unapproved images from being used
// downstream.
type StoryboardImageManifest struct {
	Envelope
	Source   string            `json:"source"`
	Operator string            `json:"operator,omitempty"`
	Images   []StoryboardImage `json:"images"`
}

// StoryboardImage is one imported reference image for a scene.
type StoryboardImage struct {
	SceneID            string  `json:"scene_id"`
	ImagePath          string  `json:"image_path"`
	ImageHash          string  `json:"image_hash"`
	VersionID          string  `json:"version_id"`
	Status             string  `json:"status"`
	ExpectedStartSec   float64 `json:"expected_start_sec"`
	ExpectedEndSec     float64 `json:"expected_end_sec"`
	VisualReviewPassed bool    `json:"visual_review_passed"`
	ApprovedBy         string  `json:"approved_by,omitempty"`
	ApprovedAt         string  `json:"approved_at,omitempty"`
}

func (m *StoryboardImageManifest) Kind() string           { return KindStoryboardImageManifest }
func (m *StoryboardImageManifest) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 2. visual_shot_manifest.json -----

// VisualShotManifest records Seedance image-to-video shots generated from
// approved storyboard images.
type VisualShotManifest struct {
	Envelope
	Provider     ProviderRef  `json:"provider"`
	AspectRatio  string       `json:"aspect_ratio"`
	RenderTarget RenderTarget `json:"render_target"`
	Shots        []VisualShot `json:"shots"`
}

// VisualShot is one generated image-to-video shot.
type VisualShot struct {
	SceneID            string  `json:"scene_id"`
	Prompt             string  `json:"prompt"`
	NegativePrompt     string  `json:"negative_prompt"`
	ReferenceImageHash string  `json:"reference_image_hash"`
	OutputPath         string  `json:"output_path"`
	OutputHash         string  `json:"output_hash"`
	DurationSec        float64 `json:"duration_sec"`
	Camera             string  `json:"camera,omitempty"`
	Style              string  `json:"style,omitempty"`
	Status             string  `json:"status"`
	OperatorApproval   bool    `json:"operator_approval"`
}

func (m *VisualShotManifest) Kind() string           { return KindVisualShotManifest }
func (m *VisualShotManifest) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 3. voiceover_manifest.json -----

// VoiceoverManifest records synthesized voiceover with provider metadata and
// output provenance.
type VoiceoverManifest struct {
	Envelope
	Provider         ProviderRef `json:"provider"`
	SourceScriptRef  string      `json:"source_script_ref"`
	Language         string      `json:"language"`
	Output           MediaOutput `json:"output"`
	OperatorApproval bool        `json:"operator_approval"`
}

func (m *VoiceoverManifest) Kind() string           { return KindVoiceoverManifest }
func (m *VoiceoverManifest) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 4. subtitle_manifest.json -----

// SubtitleManifest records transcript + caption outputs and the subtitle quality
// checks (word timestamps, safe zone, sync).
type SubtitleManifest struct {
	Envelope
	Provider         ProviderRef    `json:"provider"`
	Language         string         `json:"language"`
	TranscriptPath   string         `json:"transcript_path"`
	TranscriptHash   string         `json:"transcript_hash"`
	SRTPath          string         `json:"srt_path"`
	SRTHash          string         `json:"srt_hash"`
	ASSPath          string         `json:"ass_path,omitempty"`
	ASSHash          string         `json:"ass_hash,omitempty"`
	Checks           SubtitleChecks `json:"checks"`
	OperatorApproval bool           `json:"operator_approval"`
}

// SubtitleChecks are the deterministic subtitle gate checks.
type SubtitleChecks struct {
	WordTimestamps bool `json:"word_timestamps"`
	SafeZone       bool `json:"safe_zone"`
	Sync           bool `json:"sync"`
}

func (m *SubtitleManifest) Kind() string           { return KindSubtitleManifest }
func (m *SubtitleManifest) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 5. short_render_manifest.json -----

// ShortRenderManifest records final per-platform renders.
type ShortRenderManifest struct {
	Envelope
	Renderer RendererRef    `json:"renderer"`
	Inputs   []string       `json:"inputs"`
	Outputs  []RenderOutput `json:"outputs"`
}

// RendererRef identifies the renderer and version.
type RendererRef struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RenderOutput describes one final render output for a platform.
type RenderOutput struct {
	Platform        string  `json:"platform"`
	Path            string  `json:"path"`
	Hash            string  `json:"hash"`
	Resolution      string  `json:"resolution"`
	Aspect          string  `json:"aspect"`
	FPS             int     `json:"fps"`
	VideoCodec      string  `json:"video_codec"`
	AudioCodec      string  `json:"audio_codec"`
	AudioTrack      bool    `json:"audio_track"`
	SubtitlesBurned bool    `json:"subtitles_burned"`
	DurationSec     float64 `json:"duration_sec"`
	Status          string  `json:"status"`
}

func (m *ShortRenderManifest) Kind() string           { return KindShortRenderManifest }
func (m *ShortRenderManifest) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 6. production_candidate.json -----

// ProductionCandidate is an immutable bundle pointer assembled once all upstream
// artifacts are approved. Its status is locked and it must never be mutated in
// place.
type ProductionCandidate struct {
	Envelope
	CandidateID string               `json:"candidate_id"`
	Immutable   bool                 `json:"immutable"`
	Components  []CandidateComponent `json:"components"`
}

// CandidateComponent pins one upstream artifact into the bundle by hash.
type CandidateComponent struct {
	ArtifactID  string `json:"artifact_id"`
	Kind        string `json:"kind"`
	ContentHash string `json:"content_hash"`
}

func (m *ProductionCandidate) Kind() string           { return KindProductionCandidate }
func (m *ProductionCandidate) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 7. release_approval.json -----

// ReleaseApproval is the explicit, human-recorded release decision.
type ReleaseApproval struct {
	Envelope
	CandidateID          string         `json:"candidate_id"`
	Platforms            []string       `json:"platforms"`
	Visibility           string         `json:"visibility"`
	ScheduledAt          string         `json:"scheduled_at,omitempty"`
	AIDisclosureRequired bool           `json:"ai_disclosure_required"`
	AIDisclosure         string         `json:"ai_disclosure,omitempty"`
	HumanReleaseApproval bool           `json:"human_release_approval"`
	ApprovedBy           string         `json:"approved_by,omitempty"`
	ApprovedAt           string         `json:"approved_at,omitempty"`
	ProductionQARef      string         `json:"production_qa_ref"`
	RiskAcceptance       RiskAcceptance `json:"risk_acceptance"`
}

// RiskAcceptance records explicit acknowledgement of generation risks.
type RiskAcceptance struct {
	AIGeneratedVisuals  bool `json:"ai_generated_visuals"`
	AIDisclosurePresent bool `json:"ai_disclosure_present"`
	BrandSafetyChecked  bool `json:"brand_safety_checked"`
}

func (m *ReleaseApproval) Kind() string           { return KindReleaseApproval }
func (m *ReleaseApproval) EnvelopeRef() *Envelope { return &m.Envelope }

// ----- 8. uploadpost_publish_manifest.json -----

// UploadPostPublishManifest is the guarded publishing intent. It references QA
// and release approval and is never a direct publish path.
type UploadPostPublishManifest struct {
	Envelope
	Provider             string   `json:"provider"`
	Mode                 string   `json:"mode"`
	DryRun               bool     `json:"dry_run"`
	Platforms            []string `json:"platforms"`
	Visibility           string   `json:"visibility"`
	ScheduledAt          string   `json:"scheduled_at,omitempty"`
	AIDisclosureRequired bool     `json:"ai_disclosure_required"`
	AIDisclosure         string   `json:"ai_disclosure,omitempty"`
	HumanReleaseApproval bool     `json:"human_release_approval"`
	ProductionQARef      string   `json:"production_qa_ref"`
	ReleaseApprovalRef   string   `json:"release_approval_ref"`
}

func (m *UploadPostPublishManifest) Kind() string           { return KindUploadPostPublishManifest }
func (m *UploadPostPublishManifest) EnvelopeRef() *Envelope { return &m.Envelope }
