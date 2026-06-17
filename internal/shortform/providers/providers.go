// Package providers defines the short-form execution provider interfaces and
// their deterministic mock implementations (see docs/adr/0005). M1 ships mocks
// only; real adapters (Seedance, ElevenLabs, faster-whisper, FFmpeg, Upload-Post)
// are deferred to M2/M3 behind flags.
//
// Every provider emits a schema-valid DRAFT artifact and never self-approves.
// Mocks support deterministic failure injection (Defect) so gate negative tests
// are exercisable.
package providers

import (
	"context"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
)

// Defect selects a deterministic failure injection for a mock provider.
type Defect string

const (
	// DefectNone is the default healthy behavior.
	DefectNone Defect = ""
	// DefectError makes the provider return an error (exercises retry/fallback
	// and workflow block paths).
	DefectError Defect = "error"
	// DefectSubtitleSyncFailed emits a schema-valid subtitle manifest whose sync
	// check failed (exercises the subtitle gate negative path).
	DefectSubtitleSyncFailed Defect = "subtitle_sync_failed"
	// DefectRenderNoAudio emits a schema-valid render whose output has no audio
	// track (exercises the render gate negative path).
	DefectRenderNoAudio Defect = "render_no_audio"
	// DefectMissingDisclosure emits a publish manifest that requires AI
	// disclosure but omits the disclosure text (exercises the release gate).
	DefectMissingDisclosure Defect = "missing_disclosure"
)

// SceneSpec is the per-scene timing input for storyboard image import.
type SceneSpec struct {
	SceneID  string
	StartSec float64
	EndSec   float64
	Prompt   string
}

// StoryboardImageRequest imports ChatGPT reference images for an episode.
type StoryboardImageRequest struct {
	EpisodeID string
	Now       time.Time
	Operator  string
	Scenes    []SceneSpec
}

// VisualShotRequest generates image-to-video shots from approved storyboard
// images.
type VisualShotRequest struct {
	EpisodeID       string
	Now             time.Time
	StoryboardImage *shortform.StoryboardImageManifest
}

// VoiceoverRequest synthesizes voiceover from an approved script reference.
type VoiceoverRequest struct {
	EpisodeID string
	Now       time.Time
	ScriptRef string
	Language  string
}

// SubtitleRequest generates transcript and captions from a voiceover.
type SubtitleRequest struct {
	EpisodeID              string
	Now                    time.Time
	Voiceover              *shortform.VoiceoverManifest
	Language               string
	WordTimestampsRequired bool
}

// RenderRequest composites shots + voiceover + subtitles into final renders.
type RenderRequest struct {
	EpisodeID string
	Now       time.Time
	Shots     *shortform.VisualShotManifest
	Voiceover *shortform.VoiceoverManifest
	Subtitles *shortform.SubtitleManifest
	Platforms []string
}

// PublishRequest builds a guarded Upload-Post dry-run publish manifest.
type PublishRequest struct {
	EpisodeID            string
	Now                  time.Time
	Release              *shortform.ReleaseApproval
	Render               *shortform.ShortRenderManifest
	ProductionQADecision string
	ProductionQARef      string
}

// StoryboardImageProvider imports and records ChatGPT reference images with
// provenance. It never generates images itself (chatgpt_manual import only).
type StoryboardImageProvider interface {
	ImportStoryboardImages(ctx context.Context, req StoryboardImageRequest) (*shortform.StoryboardImageManifest, error)
}

// VisualVideoProvider produces image-to-video shots (Seedance in M3).
type VisualVideoProvider interface {
	GenerateShots(ctx context.Context, req VisualShotRequest) (*shortform.VisualShotManifest, error)
}

// VoiceProvider synthesizes voiceover audio (ElevenLabs in M3).
type VoiceProvider interface {
	SynthesizeVoiceover(ctx context.Context, req VoiceoverRequest) (*shortform.VoiceoverManifest, error)
}

// SubtitleProvider transcribes audio and emits captions (faster-whisper in M2).
// Word timestamps only; no pyannote diarization (ADR-0005).
type SubtitleProvider interface {
	GenerateSubtitles(ctx context.Context, req SubtitleRequest) (*shortform.SubtitleManifest, error)
}

// RenderProvider composites the final vertical renders (FFmpeg in M2).
type RenderProvider interface {
	RenderShort(ctx context.Context, req RenderRequest) (*shortform.ShortRenderManifest, error)
}

// PublishingProvider performs a guarded Upload-Post dry-run (M2) — never a real
// publish in M1.
type PublishingProvider interface {
	UploadPostDryRun(ctx context.Context, req PublishRequest) (*shortform.UploadPostPublishManifest, error)
}
