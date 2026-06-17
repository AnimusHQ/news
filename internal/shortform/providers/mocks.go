package providers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
)

// fakeHash returns a deterministic, schema-valid sha256 hash for mock outputs.
// No real bytes are produced (M1 spends nothing, generates nothing).
func fakeHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

func draftEnvelope(kind, episodeID, createdBy string, now time.Time, sources []string) shortform.Envelope {
	return shortform.Envelope{
		SchemaVersion:   shortform.SchemaVersion,
		EpisodeID:       episodeID,
		ArtifactID:      fmt.Sprintf("%s-%s-v1", kind, episodeID),
		CreatedAt:       rfc3339(now),
		CreatedBy:       createdBy,
		SourceArtifacts: sources,
		Status:          shortform.StatusDraft,
	}
}

// ----- StoryboardImageProvider -----

// MockStoryboardImageProvider deterministically records imported ChatGPT
// reference images as a draft manifest (no approval — approval is a later gate).
type MockStoryboardImageProvider struct{ Defect Defect }

func (p MockStoryboardImageProvider) ImportStoryboardImages(_ context.Context, req StoryboardImageRequest) (*shortform.StoryboardImageManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock storyboard image provider: injected error")
	}
	if len(req.Scenes) == 0 {
		return nil, fmt.Errorf("storyboard import requires at least one scene")
	}
	images := make([]shortform.StoryboardImage, 0, len(req.Scenes))
	for _, scene := range req.Scenes {
		images = append(images, shortform.StoryboardImage{
			SceneID:            scene.SceneID,
			ImagePath:          fmt.Sprintf("storyboards/%s/v001.png", scene.SceneID),
			ImageHash:          fakeHash(req.EpisodeID, scene.SceneID, "storyboard-image"),
			VersionID:          "v001",
			Status:             shortform.StatusDraft,
			ExpectedStartSec:   scene.StartSec,
			ExpectedEndSec:     scene.EndSec,
			VisualReviewPassed: false,
		})
	}
	manifest := &shortform.StoryboardImageManifest{
		Envelope: draftEnvelope(shortform.KindStoryboardImageManifest, req.EpisodeID, "system", req.Now, nil),
		Source:   "chatgpt_manual",
		Operator: req.Operator,
		Images:   images,
	}
	return manifest, shortform.Stamp(manifest)
}

// ----- VisualVideoProvider -----

// MockVisualVideoProvider produces deterministic fake image-to-video shots from
// the (approved) storyboard images.
type MockVisualVideoProvider struct{ Defect Defect }

func (p MockVisualVideoProvider) GenerateShots(_ context.Context, req VisualShotRequest) (*shortform.VisualShotManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock visual video provider: injected error")
	}
	if req.StoryboardImage == nil || len(req.StoryboardImage.Images) == 0 {
		return nil, fmt.Errorf("visual shot generation requires storyboard images")
	}
	shots := make([]shortform.VisualShot, 0, len(req.StoryboardImage.Images))
	for _, img := range req.StoryboardImage.Images {
		duration := img.ExpectedEndSec - img.ExpectedStartSec
		if duration <= 0 {
			duration = 5
		}
		shots = append(shots, shortform.VisualShot{
			SceneID:            img.SceneID,
			Prompt:             fmt.Sprintf("animate scene %s with subtle parallax motion", img.SceneID),
			NegativePrompt:     "no on-screen text, no watermark, no distortion",
			ReferenceImageHash: img.ImageHash,
			OutputPath:         fmt.Sprintf("shots/%s/v001.mp4", img.SceneID),
			OutputHash:         fakeHash(req.EpisodeID, img.SceneID, "visual-shot"),
			DurationSec:        duration,
			Camera:             "slow_push_in",
			Style:              "documentary",
			Status:             shortform.StatusDraft,
			OperatorApproval:   false,
		})
	}
	manifest := &shortform.VisualShotManifest{
		Envelope:    draftEnvelope(shortform.KindVisualShotManifest, req.EpisodeID, "model:seedance-mock", req.Now, []string{req.StoryboardImage.ArtifactID}),
		Provider:    shortform.ProviderRef{Name: "mock", Model: "seedance-2.0-mock", Version: "0.1.0"},
		AspectRatio: shortform.TargetAspect,
		RenderTarget: shortform.RenderTarget{
			Resolution: shortform.TargetResolution,
			Aspect:     shortform.TargetAspect,
			FPS:        shortform.TargetFPS,
			VideoCodec: shortform.TargetVideoCodec,
		},
		Shots: shots,
	}
	return manifest, shortform.Stamp(manifest)
}

// ----- VoiceProvider -----

// MockVoiceProvider produces a deterministic fake voiceover manifest.
type MockVoiceProvider struct{ Defect Defect }

func (p MockVoiceProvider) SynthesizeVoiceover(_ context.Context, req VoiceoverRequest) (*shortform.VoiceoverManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock voice provider: injected error")
	}
	if req.ScriptRef == "" {
		return nil, fmt.Errorf("voiceover requires a source script reference")
	}
	language := req.Language
	if language == "" {
		language = "en"
	}
	manifest := &shortform.VoiceoverManifest{
		Envelope:        draftEnvelope(shortform.KindVoiceoverManifest, req.EpisodeID, "model:elevenlabs-mock", req.Now, []string{req.ScriptRef}),
		Provider:        shortform.ProviderRef{Name: "elevenlabs", Model: "eleven_multilingual_v2-mock", Version: "0.1.0"},
		SourceScriptRef: req.ScriptRef,
		Language:        language,
		Output: shortform.MediaOutput{
			Path:        "voice/voiceover.mp3",
			Hash:        fakeHash(req.EpisodeID, "voiceover"),
			DurationSec: 42,
			Format:      "mp3",
		},
		OperatorApproval: false,
	}
	return manifest, shortform.Stamp(manifest)
}

// ----- SubtitleProvider -----

// MockSubtitleProvider produces a deterministic fake subtitle manifest. With
// DefectSubtitleSyncFailed it emits a schema-valid manifest whose sync check
// failed.
type MockSubtitleProvider struct{ Defect Defect }

func (p MockSubtitleProvider) GenerateSubtitles(_ context.Context, req SubtitleRequest) (*shortform.SubtitleManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock subtitle provider: injected error")
	}
	if req.Voiceover == nil {
		return nil, fmt.Errorf("subtitles require a voiceover input")
	}
	language := req.Language
	if language == "" {
		language = req.Voiceover.Language
	}
	checks := shortform.SubtitleChecks{WordTimestamps: true, SafeZone: true, Sync: true}
	if p.Defect == DefectSubtitleSyncFailed {
		checks.Sync = false
	}
	manifest := &shortform.SubtitleManifest{
		Envelope:         draftEnvelope(shortform.KindSubtitleManifest, req.EpisodeID, "model:faster-whisper-mock", req.Now, []string{req.Voiceover.ArtifactID}),
		Provider:         shortform.ProviderRef{Name: "faster_whisper", Model: "base", Version: "0.1.0"},
		Language:         language,
		TranscriptPath:   "subtitles/transcript.json",
		TranscriptHash:   fakeHash(req.EpisodeID, "transcript"),
		SRTPath:          "subtitles/captions.srt",
		SRTHash:          fakeHash(req.EpisodeID, "srt"),
		ASSPath:          "subtitles/captions.ass",
		ASSHash:          fakeHash(req.EpisodeID, "ass"),
		Checks:           checks,
		OperatorApproval: false,
	}
	return manifest, shortform.Stamp(manifest)
}

// ----- RenderProvider -----

// MockRenderProvider composites a deterministic fake render manifest. With
// DefectRenderNoAudio it emits an output with no audio track.
type MockRenderProvider struct{ Defect Defect }

func (p MockRenderProvider) RenderShort(_ context.Context, req RenderRequest) (*shortform.ShortRenderManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock render provider: injected error")
	}
	if req.Shots == nil || req.Voiceover == nil || req.Subtitles == nil {
		return nil, fmt.Errorf("render requires shots, voiceover, and subtitles")
	}
	platforms := req.Platforms
	if len(platforms) == 0 {
		platforms = []string{"master"}
	}
	audioTrack := p.Defect != DefectRenderNoAudio
	outputs := make([]shortform.RenderOutput, 0, len(platforms))
	for _, platform := range platforms {
		outputs = append(outputs, shortform.RenderOutput{
			Platform:        platform,
			Path:            fmt.Sprintf("renders/%s.mp4", platform),
			Hash:            fakeHash(req.EpisodeID, platform, "render"),
			Resolution:      shortform.TargetResolution,
			Aspect:          shortform.TargetAspect,
			FPS:             shortform.TargetFPS,
			VideoCodec:      shortform.TargetVideoCodec,
			AudioCodec:      shortform.TargetAudioCodec,
			AudioTrack:      audioTrack,
			SubtitlesBurned: true,
			DurationSec:     req.Voiceover.Output.DurationSec,
			Status:          shortform.StatusDraft,
		})
	}
	manifest := &shortform.ShortRenderManifest{
		Envelope: draftEnvelope(shortform.KindShortRenderManifest, req.EpisodeID, "system", req.Now,
			[]string{req.Shots.ArtifactID, req.Voiceover.ArtifactID, req.Subtitles.ArtifactID}),
		Renderer: shortform.RendererRef{Name: "ffmpeg", Version: "0.1.0-mock"},
		Inputs:   []string{"visual_shot_manifest.json", "voiceover_manifest.json", "subtitle_manifest.json"},
		Outputs:  outputs,
	}
	return manifest, shortform.Stamp(manifest)
}

// ----- PublishingProvider -----

// MockPublishingProvider builds a guarded Upload-Post dry-run manifest. It never
// performs a real upload. With DefectMissingDisclosure it omits the disclosure
// text while marking it required.
type MockPublishingProvider struct{ Defect Defect }

func (p MockPublishingProvider) UploadPostDryRun(_ context.Context, req PublishRequest) (*shortform.UploadPostPublishManifest, error) {
	if p.Defect == DefectError {
		return nil, fmt.Errorf("mock publishing provider: injected error")
	}
	if req.Release == nil {
		return nil, fmt.Errorf("publish manifest requires a release approval")
	}
	disclosure := req.Release.AIDisclosure
	if p.Defect == DefectMissingDisclosure {
		disclosure = ""
	}
	manifest := &shortform.UploadPostPublishManifest{
		Envelope: draftEnvelope(shortform.KindUploadPostPublishManifest, req.EpisodeID, "system", req.Now,
			[]string{req.Release.ArtifactID}),
		Provider:             "upload_post",
		Mode:                 "dry_run",
		DryRun:               true,
		Platforms:            req.Release.Platforms,
		Visibility:           req.Release.Visibility,
		ScheduledAt:          req.Release.ScheduledAt,
		AIDisclosureRequired: req.Release.AIDisclosureRequired,
		AIDisclosure:         disclosure,
		HumanReleaseApproval: req.Release.HumanReleaseApproval,
		ProductionQARef:      req.ProductionQARef,
		ReleaseApprovalRef:   req.Release.ArtifactID,
	}
	return manifest, shortform.Stamp(manifest)
}

// Compile-time interface checks.
var (
	_ StoryboardImageProvider = MockStoryboardImageProvider{}
	_ VisualVideoProvider     = MockVisualVideoProvider{}
	_ VoiceProvider           = MockVoiceProvider{}
	_ SubtitleProvider        = MockSubtitleProvider{}
	_ RenderProvider          = MockRenderProvider{}
	_ PublishingProvider      = MockPublishingProvider{}
)
