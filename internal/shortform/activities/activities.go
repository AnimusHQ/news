// Package activities implements the short-form pipeline activities (§9). In M1
// every activity is backed by a deterministic mock provider; real adapters are
// deferred to M2/M3 behind flags (the *NotEnabledInM1 activities prove they never
// run). Activities own side effects; gates (pure) run in the orchestrator.
//
// Activities are idempotent: a given input yields a byte-identical artifact
// (deterministic mocks + deterministic content hash), so re-execution keyed by
// episode_id + artifact_id + version is safe.
package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

// Activities bundles the provider set used by the short-form pipeline. In M1 the
// providers are mocks; defects can be injected for failure-path testing.
type Activities struct {
	Storyboard providers.StoryboardImageProvider
	Visual     providers.VisualVideoProvider
	Voice      providers.VoiceProvider
	Subtitle   providers.SubtitleProvider
	Render     providers.RenderProvider
	Publishing providers.PublishingProvider
}

// MockDefects selects per-provider failure injection for the mock activity set.
type MockDefects struct {
	Storyboard providers.Defect
	Visual     providers.Defect
	Voice      providers.Defect
	Subtitle   providers.Defect
	Render     providers.Defect
	Publishing providers.Defect
}

// NewMockActivities builds the healthy M1 mock activity set.
func NewMockActivities() *Activities { return NewMockActivitiesWithDefects(MockDefects{}) }

// NewMockActivitiesWithDefects builds a mock activity set with injected defects.
func NewMockActivitiesWithDefects(d MockDefects) *Activities {
	return &Activities{
		Storyboard: providers.MockStoryboardImageProvider{Defect: d.Storyboard},
		Visual:     providers.MockVisualVideoProvider{Defect: d.Visual},
		Voice:      providers.MockVoiceProvider{Defect: d.Voice},
		Subtitle:   providers.MockSubtitleProvider{Defect: d.Subtitle},
		Render:     providers.MockRenderProvider{Defect: d.Render},
		Publishing: providers.MockPublishingProvider{Defect: d.Publishing},
	}
}

// ValidationResult is the structural validation outcome of an artifact.
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Issues []string `json:"issues,omitempty"`
}

func validate(a shortform.Artifact) ValidationResult {
	issues := shortform.Validate(a)
	return ValidationResult{Valid: len(issues) == 0, Issues: issues}
}

// ----- prompt pack (BuildChatGPTStoryboardPromptPackActivity) -----

// PromptPackInput requests a deterministic ChatGPT storyboard prompt pack.
type PromptPackInput struct {
	EpisodeID string
	Scenes    []providers.SceneSpec
}

// ScenePrompt is one deterministic prompt for an operator to paste into ChatGPT.
type ScenePrompt struct {
	SceneID  string  `json:"scene_id"`
	StartSec float64 `json:"start_sec"`
	EndSec   float64 `json:"end_sec"`
	Prompt   string  `json:"prompt"`
}

// StoryboardPromptPack is the deterministic prompt pack handed to the operator.
type StoryboardPromptPack struct {
	EpisodeID string        `json:"episode_id"`
	Source    string        `json:"source"`
	Scenes    []ScenePrompt `json:"scenes"`
}

func (a *Activities) BuildStoryboardPromptPack(_ context.Context, in PromptPackInput) (StoryboardPromptPack, error) {
	if len(in.Scenes) == 0 {
		return StoryboardPromptPack{}, fmt.Errorf("prompt pack requires at least one scene")
	}
	scenes := make([]ScenePrompt, 0, len(in.Scenes))
	for _, s := range in.Scenes {
		prompt := s.Prompt
		if prompt == "" {
			prompt = fmt.Sprintf("vertical 9:16 reference image for scene %s", s.SceneID)
		}
		scenes = append(scenes, ScenePrompt{SceneID: s.SceneID, StartSec: s.StartSec, EndSec: s.EndSec, Prompt: prompt})
	}
	return StoryboardPromptPack{EpisodeID: in.EpisodeID, Source: "chatgpt_manual", Scenes: scenes}, nil
}

// ----- storyboard images -----

// ImportStoryboardInput imports ChatGPT reference images (mock import in M1).
type ImportStoryboardInput struct {
	EpisodeID string
	Now       time.Time
	Operator  string
	Scenes    []providers.SceneSpec
}

func (a *Activities) ImportStoryboardImages(ctx context.Context, in ImportStoryboardInput) (*shortform.StoryboardImageManifest, error) {
	return a.Storyboard.ImportStoryboardImages(ctx, providers.StoryboardImageRequest{
		EpisodeID: in.EpisodeID, Now: in.Now, Operator: in.Operator, Scenes: in.Scenes,
	})
}

func (a *Activities) ValidateImportedStoryboardImages(_ context.Context, m *shortform.StoryboardImageManifest) (ValidationResult, error) {
	if m == nil {
		return ValidationResult{Valid: false, Issues: []string{"storyboard image manifest is nil"}}, nil
	}
	return validate(m), nil
}

// ApproveStoryboardInput applies human approval to imported images.
type ApproveStoryboardInput struct {
	Manifest *shortform.StoryboardImageManifest
	Approver string
	Now      time.Time
}

func (a *Activities) ApproveStoryboardImages(_ context.Context, in ApproveStoryboardInput) (*shortform.StoryboardImageManifest, error) {
	if err := shortform.ApproveStoryboardImages(in.Manifest, in.Approver, in.Now); err != nil {
		return nil, err
	}
	return in.Manifest, nil
}

// ----- visual shots -----

// VisualShotsInput generates image-to-video shots from approved storyboard images.
type VisualShotsInput struct {
	EpisodeID  string
	Now        time.Time
	Storyboard *shortform.StoryboardImageManifest
}

// GenerateMockVisualShots is the M1 default (Seedance mock).
func (a *Activities) GenerateMockVisualShots(ctx context.Context, in VisualShotsInput) (*shortform.VisualShotManifest, error) {
	return a.Visual.GenerateShots(ctx, providers.VisualShotRequest{EpisodeID: in.EpisodeID, Now: in.Now, StoryboardImage: in.Storyboard})
}

// GenerateSeedanceShots is the real provider path; never runs in M1.
func (a *Activities) GenerateSeedanceShots(_ context.Context, _ VisualShotsInput) (*shortform.VisualShotManifest, error) {
	return nil, errNotEnabledInM1("GenerateSeedanceShots")
}

func (a *Activities) ApproveVisualShots(_ context.Context, m *shortform.VisualShotManifest, now time.Time) (*shortform.VisualShotManifest, error) {
	if err := shortform.ApproveVisualShots(m, now); err != nil {
		return nil, err
	}
	return m, nil
}

// ----- voiceover -----

// VoiceoverInput synthesizes voiceover (ElevenLabs mock in M1).
type VoiceoverInput struct {
	EpisodeID string
	Now       time.Time
	ScriptRef string
	Language  string
}

func (a *Activities) GenerateElevenLabsVoiceover(ctx context.Context, in VoiceoverInput) (*shortform.VoiceoverManifest, error) {
	return a.Voice.SynthesizeVoiceover(ctx, providers.VoiceoverRequest{EpisodeID: in.EpisodeID, Now: in.Now, ScriptRef: in.ScriptRef, Language: in.Language})
}

func (a *Activities) ApproveVoiceover(_ context.Context, m *shortform.VoiceoverManifest, now time.Time) (*shortform.VoiceoverManifest, error) {
	if err := shortform.ApproveVoiceover(m, now); err != nil {
		return nil, err
	}
	return m, nil
}

// ----- subtitles -----

// SubtitlesInput generates subtitles from an approved voiceover.
type SubtitlesInput struct {
	EpisodeID string
	Now       time.Time
	Voiceover *shortform.VoiceoverManifest
	Language  string
}

func (a *Activities) GenerateSubtitles(ctx context.Context, in SubtitlesInput) (*shortform.SubtitleManifest, error) {
	return a.Subtitle.GenerateSubtitles(ctx, providers.SubtitleRequest{EpisodeID: in.EpisodeID, Now: in.Now, Voiceover: in.Voiceover, Language: in.Language})
}

func (a *Activities) ValidateSubtitles(_ context.Context, m *shortform.SubtitleManifest) (ValidationResult, error) {
	if m == nil {
		return ValidationResult{Valid: false, Issues: []string{"subtitle manifest is nil"}}, nil
	}
	return validate(m), nil
}

func (a *Activities) ApproveSubtitles(_ context.Context, m *shortform.SubtitleManifest, now time.Time) (*shortform.SubtitleManifest, error) {
	if err := shortform.ApproveSubtitles(m, now); err != nil {
		return nil, err
	}
	return m, nil
}

// ----- render -----

// RenderInput composites the final render (FFmpeg mock in M1).
type RenderInput struct {
	EpisodeID string
	Now       time.Time
	Shots     *shortform.VisualShotManifest
	Voiceover *shortform.VoiceoverManifest
	Subtitles *shortform.SubtitleManifest
	Platforms []string
}

// RenderShortPreview and RenderShortFinal share the mock render in M1.
func (a *Activities) RenderShortFinal(ctx context.Context, in RenderInput) (*shortform.ShortRenderManifest, error) {
	return a.Render.RenderShort(ctx, providers.RenderRequest{
		EpisodeID: in.EpisodeID, Now: in.Now, Shots: in.Shots, Voiceover: in.Voiceover, Subtitles: in.Subtitles, Platforms: in.Platforms,
	})
}

// ProductionQAResult is the deterministic production QA decision.
type ProductionQAResult struct {
	Decision       string   `json:"decision"`
	BlockingIssues []string `json:"blocking_issues,omitempty"`
}

// RunProductionQA deterministically evaluates a render: approved only when every
// output meets the vertical target with audio + burned subtitles.
func (a *Activities) RunProductionQA(_ context.Context, m *shortform.ShortRenderManifest) (ProductionQAResult, error) {
	if m == nil || len(m.Outputs) == 0 {
		return ProductionQAResult{Decision: "request_revision", BlockingIssues: []string{"no render outputs"}}, nil
	}
	var issues []string
	for _, out := range m.Outputs {
		if !out.AudioTrack {
			issues = append(issues, out.Platform+": missing audio track")
		}
		if !out.SubtitlesBurned {
			issues = append(issues, out.Platform+": subtitles not burned")
		}
		if out.Resolution != shortform.TargetResolution || out.Aspect != shortform.TargetAspect || out.FPS != shortform.TargetFPS {
			issues = append(issues, out.Platform+": render target mismatch")
		}
	}
	if len(issues) > 0 {
		return ProductionQAResult{Decision: "request_revision", BlockingIssues: issues}, nil
	}
	return ProductionQAResult{Decision: "approved"}, nil
}

func (a *Activities) ApproveRenderOutputs(_ context.Context, m *shortform.ShortRenderManifest, now time.Time) (*shortform.ShortRenderManifest, error) {
	if err := shortform.ApproveRenderOutputs(m, now); err != nil {
		return nil, err
	}
	return m, nil
}

func (a *Activities) ValidateShortRender(_ context.Context, m *shortform.ShortRenderManifest) (ValidationResult, error) {
	if m == nil {
		return ValidationResult{Valid: false, Issues: []string{"short render manifest is nil"}}, nil
	}
	return validate(m), nil
}

// ----- production candidate + release -----

// AssembleCandidateInput pins approved artifacts into an immutable candidate.
// Components are concrete descriptors so the input is serializable across the
// activity boundary.
type AssembleCandidateInput struct {
	EpisodeID   string
	CandidateID string
	Now         time.Time
	Components  []shortform.ComponentRef
}

func (a *Activities) AssembleProductionCandidate(_ context.Context, in AssembleCandidateInput) (*shortform.ProductionCandidate, error) {
	return shortform.AssembleProductionCandidate(in.EpisodeID, in.CandidateID, in.Now, in.Components)
}

func (a *Activities) BuildReleaseApproval(_ context.Context, in shortform.BuildReleaseApprovalInput) (*shortform.ReleaseApproval, error) {
	return shortform.BuildReleaseApproval(in)
}

// ----- publishing -----

// PublishManifestInput builds a guarded Upload-Post dry-run manifest.
type PublishManifestInput struct {
	EpisodeID       string
	Now             time.Time
	Release         *shortform.ReleaseApproval
	ProductionQARef string
}

func (a *Activities) GenerateUploadPostPublishManifest(ctx context.Context, in PublishManifestInput) (*shortform.UploadPostPublishManifest, error) {
	return a.Publishing.UploadPostDryRun(ctx, providers.PublishRequest{EpisodeID: in.EpisodeID, Now: in.Now, Release: in.Release, ProductionQARef: in.ProductionQARef})
}

func (a *Activities) ValidateUploadPostPublishManifest(_ context.Context, m *shortform.UploadPostPublishManifest) (ValidationResult, error) {
	if m == nil {
		return ValidationResult{Valid: false, Issues: []string{"publish manifest is nil"}}, nil
	}
	return validate(m), nil
}

// DryRunResult is the outcome of a guarded Upload-Post dry-run.
type DryRunResult struct {
	OK     bool   `json:"ok"`
	Mode   string `json:"mode"`
	Detail string `json:"detail"`
}

// UploadPostDryRun simulates the guarded dry-run. It never performs a real
// upload (M1) and requires dry_run mode.
func (a *Activities) UploadPostDryRun(_ context.Context, m *shortform.UploadPostPublishManifest) (DryRunResult, error) {
	if m == nil {
		return DryRunResult{OK: false, Detail: "publish manifest is nil"}, nil
	}
	if m.Mode != "dry_run" || !m.DryRun {
		return DryRunResult{OK: false, Mode: m.Mode, Detail: "M1 only permits dry_run mode"}, nil
	}
	return DryRunResult{OK: true, Mode: "dry_run", Detail: "dry-run validated; no upload performed"}, nil
}

// UploadPostSchedulePublish is the real scheduled/public publish path; never runs
// in M1.
func (a *Activities) UploadPostSchedulePublish(_ context.Context, _ *shortform.UploadPostPublishManifest) error {
	return errNotEnabledInM1("UploadPostSchedulePublish")
}

func errNotEnabledInM1(name string) error {
	return fmt.Errorf("%s is deferred to M2/M3 and must not run in M1", name)
}
