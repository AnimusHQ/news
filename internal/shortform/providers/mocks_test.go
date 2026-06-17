package providers

import (
	"context"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func sampleScenes() []SceneSpec {
	return []SceneSpec{
		{SceneID: "scene-001", StartSec: 0, EndSec: 5, Prompt: "hook"},
		{SceneID: "scene-002", StartSec: 5, EndSec: 12, Prompt: "explainer"},
	}
}

// driveChain runs the full mock provider chain and returns every emitted
// artifact. A release approval (human-authored, not a provider) is supplied for
// the publishing step.
func driveChain(t *testing.T) (
	*shortform.StoryboardImageManifest,
	*shortform.VisualShotManifest,
	*shortform.VoiceoverManifest,
	*shortform.SubtitleManifest,
	*shortform.ShortRenderManifest,
	*shortform.UploadPostPublishManifest,
) {
	t.Helper()
	ctx := context.Background()
	ep := "episode-0001"

	sb, err := MockStoryboardImageProvider{}.ImportStoryboardImages(ctx, StoryboardImageRequest{
		EpisodeID: ep, Now: testNow, Operator: "operator:ci", Scenes: sampleScenes(),
	})
	if err != nil {
		t.Fatalf("storyboard: %v", err)
	}
	shots, err := MockVisualVideoProvider{}.GenerateShots(ctx, VisualShotRequest{EpisodeID: ep, Now: testNow, StoryboardImage: sb})
	if err != nil {
		t.Fatalf("visual: %v", err)
	}
	vo, err := MockVoiceProvider{}.SynthesizeVoiceover(ctx, VoiceoverRequest{EpisodeID: ep, Now: testNow, ScriptRef: "script.md", Language: "en"})
	if err != nil {
		t.Fatalf("voice: %v", err)
	}
	subs, err := MockSubtitleProvider{}.GenerateSubtitles(ctx, SubtitleRequest{EpisodeID: ep, Now: testNow, Voiceover: vo})
	if err != nil {
		t.Fatalf("subtitle: %v", err)
	}
	render, err := MockRenderProvider{}.RenderShort(ctx, RenderRequest{
		EpisodeID: ep, Now: testNow, Shots: shots, Voiceover: vo, Subtitles: subs, Platforms: []string{"master", "youtube"},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	release := &shortform.ReleaseApproval{
		Envelope:             shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: ep, ArtifactID: "release_approval-episode-0001-v1", CreatedAt: rfc3339(testNow), CreatedBy: "human:editor", Status: shortform.StatusApproved},
		CandidateID:          "cand-001",
		Platforms:            []string{"youtube"},
		Visibility:           "private",
		AIDisclosureRequired: true,
		AIDisclosure:         "AI-generated visuals and synthetic voice.",
		HumanReleaseApproval: true,
		ApprovedBy:           "human:editor",
		ApprovedAt:           rfc3339(testNow),
		ProductionQARef:      "production_qa_report.json",
		RiskAcceptance:       shortform.RiskAcceptance{AIGeneratedVisuals: true, AIDisclosurePresent: true, BrandSafetyChecked: true},
	}
	publish, err := MockPublishingProvider{}.UploadPostDryRun(ctx, PublishRequest{EpisodeID: ep, Now: testNow, Release: release, ProductionQARef: "production_qa_report.json"})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	return sb, shots, vo, subs, render, publish
}

func TestMocksEmitSchemaValidDraftArtifacts(t *testing.T) {
	sb, shots, vo, subs, render, publish := driveChain(t)
	artifacts := []shortform.Artifact{sb, shots, vo, subs, render, publish}
	for _, a := range artifacts {
		if issues := shortform.Validate(a); len(issues) != 0 {
			t.Fatalf("%s emitted invalid artifact: %v", a.Kind(), issues)
		}
		if a.EnvelopeRef().ContentHash == "" {
			t.Fatalf("%s did not stamp a content hash", a.Kind())
		}
		if a.EnvelopeRef().Status != shortform.StatusDraft {
			t.Fatalf("%s must emit a draft (got %q)", a.Kind(), a.EnvelopeRef().Status)
		}
	}
	// Provider must never self-approve: created_by must be system or model:*.
	for _, a := range []shortform.Artifact{sb, shots, vo, subs, render, publish} {
		if by := a.EnvelopeRef().CreatedBy; by == "human:editor" {
			t.Fatalf("%s must not be authored by a human approver", a.Kind())
		}
	}
}

func TestMocksAreDeterministic(t *testing.T) {
	a1, _, _, _, _, _ := driveChain(t)
	a2, _, _, _, _, _ := driveChain(t)
	if a1.ContentHash != a2.ContentHash {
		t.Fatalf("mock storyboard output not deterministic: %s != %s", a1.ContentHash, a2.ContentHash)
	}
}

func TestVisualShotsReferenceStoryboardImageHashes(t *testing.T) {
	sb, shots, _, _, _, _ := driveChain(t)
	byScene := map[string]string{}
	for _, img := range sb.Images {
		byScene[img.SceneID] = img.ImageHash
	}
	for _, shot := range shots.Shots {
		if byScene[shot.SceneID] != shot.ReferenceImageHash {
			t.Fatalf("shot %s reference hash %s does not match storyboard image", shot.SceneID, shot.ReferenceImageHash)
		}
	}
}

func TestDefectErrorIsInjectable(t *testing.T) {
	ctx := context.Background()
	if _, err := (MockStoryboardImageProvider{Defect: DefectError}).ImportStoryboardImages(ctx, StoryboardImageRequest{EpisodeID: "e", Now: testNow, Scenes: sampleScenes()}); err == nil {
		t.Fatal("expected injected error")
	}
	if _, err := (MockVoiceProvider{Defect: DefectError}).SynthesizeVoiceover(ctx, VoiceoverRequest{EpisodeID: "e", Now: testNow, ScriptRef: "s"}); err == nil {
		t.Fatal("expected injected error")
	}
}

func TestDomainDefectsStaySchemaValid(t *testing.T) {
	ctx := context.Background()
	ep := "episode-0001"
	vo, _ := MockVoiceProvider{}.SynthesizeVoiceover(ctx, VoiceoverRequest{EpisodeID: ep, Now: testNow, ScriptRef: "script.md"})

	subs, err := MockSubtitleProvider{Defect: DefectSubtitleSyncFailed}.GenerateSubtitles(ctx, SubtitleRequest{EpisodeID: ep, Now: testNow, Voiceover: vo})
	if err != nil {
		t.Fatalf("subtitle defect: %v", err)
	}
	if issues := shortform.Validate(subs); len(issues) != 0 {
		t.Fatalf("defective subtitle must still be schema-valid: %v", issues)
	}
	if subs.Checks.Sync {
		t.Fatal("DefectSubtitleSyncFailed must set sync=false")
	}

	shots, _ := MockVisualVideoProvider{}.GenerateShots(ctx, VisualShotRequest{EpisodeID: ep, Now: testNow, StoryboardImage: mustStoryboard(t)})
	render, err := MockRenderProvider{Defect: DefectRenderNoAudio}.RenderShort(ctx, RenderRequest{EpisodeID: ep, Now: testNow, Shots: shots, Voiceover: vo, Subtitles: subs})
	if err != nil {
		t.Fatalf("render defect: %v", err)
	}
	if issues := shortform.Validate(render); len(issues) != 0 {
		t.Fatalf("defective render must still be schema-valid: %v", issues)
	}
	if render.Outputs[0].AudioTrack {
		t.Fatal("DefectRenderNoAudio must set audio_track=false")
	}
}

func mustStoryboard(t *testing.T) *shortform.StoryboardImageManifest {
	t.Helper()
	sb, err := MockStoryboardImageProvider{}.ImportStoryboardImages(context.Background(), StoryboardImageRequest{EpisodeID: "episode-0001", Now: testNow, Scenes: sampleScenes()})
	if err != nil {
		t.Fatalf("storyboard: %v", err)
	}
	return sb
}
