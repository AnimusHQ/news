package gates

import (
	"testing"

	"github.com/AnimusHQ/news/internal/shortform"
)

// ----- shared assertions -----

func assertPass(t *testing.T, r Result) {
	t.Helper()
	if r.Blocked() {
		t.Fatalf("%s: expected pass, blocked with %v", r.Gate, r.Reasons)
	}
}

func assertBlocked(t *testing.T, r Result, code string) {
	t.Helper()
	if !r.Blocked() {
		t.Fatalf("%s: expected block with code %q, got pass", r.Gate, code)
	}
	for _, reason := range r.Reasons {
		if reason.Code == code {
			return
		}
	}
	t.Fatalf("%s: expected reason code %q, got %v", r.Gate, code, r.Reasons)
}

const goodHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

// ----- storyboard image gate -----

func validStoryboard() *shortform.StoryboardImageManifest {
	return &shortform.StoryboardImageManifest{
		Source:   "chatgpt_manual",
		Operator: "operator:ci",
		Images: []shortform.StoryboardImage{{
			SceneID: "scene-001", ImagePath: "p.png", ImageHash: goodHash,
			VersionID: "v001", Status: shortform.StatusApproved, VisualReviewPassed: true,
		}},
	}
}

func TestStoryboardImageGatePasses(t *testing.T) {
	assertPass(t, StoryboardImageGate(StoryboardImageInput{Manifest: validStoryboard(), RequiredScenes: []string{"scene-001"}}))
}

func TestStoryboardImageGateBlocks(t *testing.T) {
	cases := []struct {
		name  string
		code  string
		mutfn func(*shortform.StoryboardImageManifest)
		req   []string
	}{
		{"nil_manifest", "manifest_missing", func(m *shortform.StoryboardImageManifest) {}, nil},
		{"bad_source", "source_not_chatgpt_manual", func(m *shortform.StoryboardImageManifest) { m.Source = "midjourney" }, []string{"scene-001"}},
		{"no_operator", "operator_approval_missing", func(m *shortform.StoryboardImageManifest) { m.Operator = "" }, []string{"scene-001"}},
		{"missing_scene", "scene_image_missing", func(m *shortform.StoryboardImageManifest) {}, []string{"scene-001", "scene-002"}},
		{"no_path", "image_path_missing", func(m *shortform.StoryboardImageManifest) { m.Images[0].ImagePath = "" }, []string{"scene-001"}},
		{"no_hash", "image_hash_missing", func(m *shortform.StoryboardImageManifest) { m.Images[0].ImageHash = "" }, []string{"scene-001"}},
		{"not_approved", "image_not_approved", func(m *shortform.StoryboardImageManifest) { m.Images[0].Status = shortform.StatusInReview }, []string{"scene-001"}},
		{"review_failed", "visual_review_failed", func(m *shortform.StoryboardImageManifest) { m.Images[0].VisualReviewPassed = false }, []string{"scene-001"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m *shortform.StoryboardImageManifest
			if tc.name != "nil_manifest" {
				m = validStoryboard()
				tc.mutfn(m)
			}
			assertBlocked(t, StoryboardImageGate(StoryboardImageInput{Manifest: m, RequiredScenes: tc.req}), tc.code)
		})
	}
}

// ----- visual shot gate -----

func validVisualInput() VisualShotInput {
	return VisualShotInput{
		Manifest: &shortform.VisualShotManifest{
			Provider: shortform.ProviderRef{Name: "mock"},
			Shots: []shortform.VisualShot{{
				SceneID: "scene-001", Prompt: "p", NegativePrompt: "n",
				ReferenceImageHash: goodHash, OutputPath: "o.mp4", OutputHash: goodHash,
				DurationSec: 5, Status: shortform.StatusApproved, OperatorApproval: true,
			}},
		},
		ApprovedImageHashes: map[string]bool{goodHash: true},
		KnownScenes:         map[string]bool{"scene-001": true},
		MaxDurationSec:      60,
	}
}

func TestVisualShotGatePasses(t *testing.T) {
	assertPass(t, VisualShotGate(validVisualInput()))
}

func TestVisualShotGateBlocks(t *testing.T) {
	cases := []struct {
		name string
		code string
		mut  func(*VisualShotInput)
	}{
		{"nil", "manifest_missing", func(in *VisualShotInput) { in.Manifest = nil }},
		{"no_provider", "provider_metadata_missing", func(in *VisualShotInput) { in.Manifest.Provider.Name = "" }},
		{"no_prompt", "prompt_missing", func(in *VisualShotInput) { in.Manifest.Shots[0].Prompt = "" }},
		{"no_neg_prompt", "negative_prompt_missing", func(in *VisualShotInput) { in.Manifest.Shots[0].NegativePrompt = "" }},
		{"no_output", "output_missing", func(in *VisualShotInput) { in.Manifest.Shots[0].OutputHash = "" }},
		{"bad_duration", "duration_out_of_tolerance", func(in *VisualShotInput) { in.Manifest.Shots[0].DurationSec = 999 }},
		{"ref_not_approved", "reference_image_not_approved", func(in *VisualShotInput) { in.ApprovedImageHashes = map[string]bool{} }},
		{"bad_scene", "scene_mapping_invalid", func(in *VisualShotInput) { in.KnownScenes = map[string]bool{"scene-999": true} }},
		{"no_operator", "operator_approval_missing", func(in *VisualShotInput) { in.Manifest.Shots[0].OperatorApproval = false }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validVisualInput()
			tc.mut(&in)
			assertBlocked(t, VisualShotGate(in), tc.code)
		})
	}
}

// ----- subtitle gate -----

func validSubtitle() *shortform.SubtitleManifest {
	return &shortform.SubtitleManifest{
		TranscriptPath: "t.json", TranscriptHash: goodHash, SRTPath: "c.srt", SRTHash: goodHash,
		Checks:           shortform.SubtitleChecks{WordTimestamps: true, SafeZone: true, Sync: true},
		OperatorApproval: true,
	}
}

func TestSubtitleGatePasses(t *testing.T) {
	assertPass(t, SubtitleGate(SubtitleInput{Manifest: validSubtitle()}))
}

func TestSubtitleGateBlocks(t *testing.T) {
	cases := []struct {
		name string
		code string
		mut  func(*shortform.SubtitleManifest)
		nil_ bool
	}{
		{"nil", "manifest_missing", nil, true},
		{"no_transcript", "transcript_missing", func(m *shortform.SubtitleManifest) { m.TranscriptHash = "" }, false},
		{"no_captions", "captions_missing", func(m *shortform.SubtitleManifest) { m.SRTPath = ""; m.ASSPath = "" }, false},
		{"no_word_ts", "word_timestamps_missing", func(m *shortform.SubtitleManifest) { m.Checks.WordTimestamps = false }, false},
		{"no_safe_zone", "safe_zone_failed", func(m *shortform.SubtitleManifest) { m.Checks.SafeZone = false }, false},
		{"no_sync", "sync_failed", func(m *shortform.SubtitleManifest) { m.Checks.Sync = false }, false},
		{"no_operator", "operator_approval_missing", func(m *shortform.SubtitleManifest) { m.OperatorApproval = false }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m *shortform.SubtitleManifest
			if !tc.nil_ {
				m = validSubtitle()
				tc.mut(m)
			}
			assertBlocked(t, SubtitleGate(SubtitleInput{Manifest: m}), tc.code)
		})
	}
}

// ----- render gate -----

func validRenderInput() RenderInput {
	return RenderInput{
		Manifest: &shortform.ShortRenderManifest{
			Outputs: []shortform.RenderOutput{{
				Platform: "master", Path: "r.mp4", Hash: goodHash,
				Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS,
				VideoCodec: shortform.TargetVideoCodec, AudioCodec: shortform.TargetAudioCodec,
				AudioTrack: true, SubtitlesBurned: true, DurationSec: 30, Status: shortform.StatusApproved,
			}},
		},
		ProductionQADecision: ProductionQAApproved,
	}
}

func TestRenderGatePasses(t *testing.T) {
	assertPass(t, RenderGate(validRenderInput()))
}

func TestRenderGateBlocks(t *testing.T) {
	cases := []struct {
		name string
		code string
		mut  func(*RenderInput)
	}{
		{"nil", "manifest_missing", func(in *RenderInput) { in.Manifest = nil }},
		{"no_outputs", "outputs_missing", func(in *RenderInput) { in.Manifest.Outputs = nil }},
		{"no_hash", "output_hash_missing", func(in *RenderInput) { in.Manifest.Outputs[0].Hash = "" }},
		{"bad_resolution", "resolution_invalid", func(in *RenderInput) { in.Manifest.Outputs[0].Resolution = "720x1280" }},
		{"bad_aspect", "aspect_invalid", func(in *RenderInput) { in.Manifest.Outputs[0].Aspect = "16:9" }},
		{"bad_fps", "fps_invalid", func(in *RenderInput) { in.Manifest.Outputs[0].FPS = 24 }},
		{"no_audio", "audio_track_missing", func(in *RenderInput) { in.Manifest.Outputs[0].AudioTrack = false }},
		{"no_subs", "subtitles_not_readable", func(in *RenderInput) { in.Manifest.Outputs[0].SubtitlesBurned = false }},
		{"qa_not_approved", "production_qa_not_approved", func(in *RenderInput) { in.ProductionQADecision = "request_revision" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validRenderInput()
			tc.mut(&in)
			assertBlocked(t, RenderGate(in), tc.code)
		})
	}
}

// ----- release gate -----

func validReleaseInput() ReleaseInput {
	return ReleaseInput{
		PublishManifest: &shortform.UploadPostPublishManifest{
			Provider: "upload_post", Mode: "dry_run", DryRun: true,
			Platforms: []string{"youtube"}, Visibility: "private",
			AIDisclosureRequired: true, AIDisclosure: "AI-generated visuals and synthetic voice.",
			HumanReleaseApproval: true,
		},
		ProductionQADecision: ProductionQAApproved,
		DryRunSucceeded:      true,
	}
}

func TestReleaseGatePasses(t *testing.T) {
	assertPass(t, ReleaseGate(validReleaseInput()))
}

func TestReleaseGateBlocks(t *testing.T) {
	cases := []struct {
		name string
		code string
		mut  func(*ReleaseInput)
	}{
		{"nil", "publish_manifest_missing", func(in *ReleaseInput) { in.PublishManifest = nil }},
		{"qa_not_approved", "production_qa_not_approved", func(in *ReleaseInput) { in.ProductionQADecision = "block" }},
		{"no_human", "human_release_approval_missing", func(in *ReleaseInput) { in.PublishManifest.HumanReleaseApproval = false }},
		{"no_platforms", "platforms_not_explicit", func(in *ReleaseInput) { in.PublishManifest.Platforms = nil }},
		{"no_visibility", "visibility_not_set", func(in *ReleaseInput) { in.PublishManifest.Visibility = "" }},
		{"scheduled_no_time", "scheduled_at_missing", func(in *ReleaseInput) { in.PublishManifest.Visibility = "scheduled" }},
		{"disclosure_missing", "disclosure_text_missing", func(in *ReleaseInput) { in.PublishManifest.AIDisclosure = "" }},
		{"dry_run_failed", "dry_run_failed", func(in *ReleaseInput) { in.DryRunSucceeded = false }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validReleaseInput()
			tc.mut(&in)
			assertBlocked(t, ReleaseGate(in), tc.code)
		})
	}
}
