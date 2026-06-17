package shortform

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AnimusHQ/news/internal/shortform/contenthash"
)

func fakeHash(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return "sha256:" + hex.EncodeToString(sum[:])
}

const (
	testEpisode = "episode-0001"
	testTime    = "2026-06-17T12:00:00Z"
)

func env(kind, status, by string) Envelope {
	return Envelope{
		SchemaVersion: SchemaVersion,
		EpisodeID:     testEpisode,
		ArtifactID:    kind + "-" + testEpisode + "-v1",
		CreatedAt:     testTime,
		CreatedBy:     by,
		Status:        status,
	}
}

func validStoryboardImageManifest() Artifact {
	return &StoryboardImageManifest{
		Envelope: env(KindStoryboardImageManifest, "approved", "system"),
		Source:   "chatgpt_manual",
		Operator: "operator:ci",
		Images: []StoryboardImage{{
			SceneID: "scene-001", ImagePath: "storyboards/scene-001/v001.png",
			ImageHash: fakeHash("img-1"), VersionID: "v001", Status: "approved",
			ExpectedStartSec: 0, ExpectedEndSec: 5, VisualReviewPassed: true,
			ApprovedBy: "human:editor", ApprovedAt: testTime,
		}},
	}
}

func validVisualShotManifest() Artifact {
	return &VisualShotManifest{
		Envelope:     env(KindVisualShotManifest, "approved", "model:seedance-mock"),
		Provider:     ProviderRef{Name: "mock", Model: "seedance-mock", Version: "0.1.0"},
		AspectRatio:  TargetAspect,
		RenderTarget: RenderTarget{Resolution: TargetResolution, Aspect: TargetAspect, FPS: TargetFPS, VideoCodec: TargetVideoCodec},
		Shots: []VisualShot{{
			SceneID: "scene-001", Prompt: "establishing shot", NegativePrompt: "no text artifacts",
			ReferenceImageHash: fakeHash("img-1"), OutputPath: "shots/scene-001/v1.mp4",
			OutputHash: fakeHash("shot-1"), DurationSec: 5, Camera: "static", Style: "documentary",
			Status: "approved", OperatorApproval: true,
		}},
	}
}

func validVoiceoverManifest() Artifact {
	return &VoiceoverManifest{
		Envelope:         env(KindVoiceoverManifest, "approved", "model:elevenlabs-mock"),
		Provider:         ProviderRef{Name: "elevenlabs", Model: "mock", Version: "0.1.0"},
		SourceScriptRef:  "script.md",
		Language:         "en",
		Output:           MediaOutput{Path: "voice/vo.mp3", Hash: fakeHash("vo-1"), DurationSec: 30, Format: "mp3"},
		OperatorApproval: true,
	}
}

func validSubtitleManifest() Artifact {
	return &SubtitleManifest{
		Envelope:       env(KindSubtitleManifest, "approved", "model:faster-whisper-mock"),
		Provider:       ProviderRef{Name: "faster_whisper", Model: "base", Version: "0.1.0"},
		Language:       "en",
		TranscriptPath: "subtitles/transcript.json", TranscriptHash: fakeHash("tr-1"),
		SRTPath: "subtitles/captions.srt", SRTHash: fakeHash("srt-1"),
		ASSPath: "subtitles/captions.ass", ASSHash: fakeHash("ass-1"),
		Checks:           SubtitleChecks{WordTimestamps: true, SafeZone: true, Sync: true},
		OperatorApproval: true,
	}
}

func validShortRenderManifest() Artifact {
	return &ShortRenderManifest{
		Envelope: env(KindShortRenderManifest, "approved", "system"),
		Renderer: RendererRef{Name: "ffmpeg", Version: "0.1.0-mock"},
		Inputs:   []string{"visual_shot_manifest.json", "voiceover_manifest.json", "subtitle_manifest.json"},
		Outputs: []RenderOutput{{
			Platform: "master", Path: "renders/master.mp4", Hash: fakeHash("render-1"),
			Resolution: TargetResolution, Aspect: TargetAspect, FPS: TargetFPS,
			VideoCodec: TargetVideoCodec, AudioCodec: TargetAudioCodec, AudioTrack: true,
			SubtitlesBurned: true, DurationSec: 30, Status: "approved",
		}},
	}
}

func validProductionCandidate() Artifact {
	return &ProductionCandidate{
		Envelope:    env(KindProductionCandidate, "locked", "system"),
		CandidateID: "cand-001",
		Immutable:   true,
		Components: []CandidateComponent{
			{ArtifactID: "short_render_manifest-episode-0001-v1", Kind: KindShortRenderManifest, ContentHash: fakeHash("render-1")},
		},
	}
}

func validReleaseApproval() Artifact {
	return &ReleaseApproval{
		Envelope:             env(KindReleaseApproval, "approved", "human:editor"),
		CandidateID:          "cand-001",
		Platforms:            []string{"youtube"},
		Visibility:           "private",
		AIDisclosureRequired: true,
		AIDisclosure:         "AI-generated visuals and synthetic voice.",
		HumanReleaseApproval: true,
		ApprovedBy:           "human:editor",
		ApprovedAt:           testTime,
		ProductionQARef:      "production_qa_report.json",
		RiskAcceptance:       RiskAcceptance{AIGeneratedVisuals: true, AIDisclosurePresent: true, BrandSafetyChecked: true},
	}
}

func validUploadPostManifest() Artifact {
	return &UploadPostPublishManifest{
		Envelope:             env(KindUploadPostPublishManifest, "draft", "system"),
		Provider:             "upload_post",
		Mode:                 "dry_run",
		DryRun:               true,
		Platforms:            []string{"youtube"},
		Visibility:           "private",
		AIDisclosureRequired: true,
		AIDisclosure:         "AI-generated visuals and synthetic voice.",
		HumanReleaseApproval: true,
		ProductionQARef:      "production_qa_report.json",
		ReleaseApprovalRef:   "release_approval.json",
	}
}

type artifactCase struct {
	name     string
	build    func() Artifact
	empty    func() Artifact
	breakOne func(map[string]any)
}

func artifactCases() []artifactCase {
	return []artifactCase{
		{"storyboard_image_manifest", validStoryboardImageManifest, func() Artifact { return &StoryboardImageManifest{} },
			func(m map[string]any) { m["source"] = "midjourney" }},
		{"visual_shot_manifest", validVisualShotManifest, func() Artifact { return &VisualShotManifest{} },
			func(m map[string]any) { m["render_target"].(map[string]any)["fps"] = 24.0 }},
		{"voiceover_manifest", validVoiceoverManifest, func() Artifact { return &VoiceoverManifest{} },
			func(m map[string]any) { m["output"].(map[string]any)["hash"] = "not-a-hash" }},
		{"subtitle_manifest", validSubtitleManifest, func() Artifact { return &SubtitleManifest{} },
			func(m map[string]any) { delete(m["checks"].(map[string]any), "sync") }},
		{"short_render_manifest", validShortRenderManifest, func() Artifact { return &ShortRenderManifest{} },
			func(m map[string]any) { m["outputs"].([]any)[0].(map[string]any)["aspect"] = "16:9" }},
		{"production_candidate", validProductionCandidate, func() Artifact { return &ProductionCandidate{} },
			func(m map[string]any) { m["immutable"] = false }},
		{"release_approval", validReleaseApproval, func() Artifact { return &ReleaseApproval{} },
			func(m map[string]any) { m["visibility"] = "viral" }},
		{"uploadpost_publish_manifest", validUploadPostManifest, func() Artifact { return &UploadPostPublishManifest{} },
			func(m map[string]any) { m["provider"] = "youtube" }},
	}
}

func TestAllSchemasCompile(t *testing.T) {
	if _, err := loadSchemas(); err != nil {
		t.Fatalf("schemas must compile: %v", err)
	}
	if len(KnownKinds()) != 8 {
		t.Fatalf("expected 8 short-form schemas, got %d: %v", len(KnownKinds()), KnownKinds())
	}
}

func TestArtifactsValidateStampAndRoundTrip(t *testing.T) {
	for _, tc := range artifactCases() {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.build()
			if a.Kind() != tc.name {
				t.Fatalf("kind mismatch: %s != %s", a.Kind(), tc.name)
			}
			if err := Stamp(a); err != nil {
				t.Fatalf("stamp: %v", err)
			}
			if a.EnvelopeRef().ContentHash == "" {
				t.Fatal("stamp must set content hash")
			}
			if issues := Validate(a); len(issues) != 0 {
				t.Fatalf("valid artifact rejected: %v", issues)
			}

			// Round-trip: marshal -> unmarshal -> deep equal.
			data, err := json.Marshal(a)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got := tc.empty()
			if err := json.Unmarshal(data, got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(a, got) {
				t.Fatalf("round-trip mismatch:\n got=%+v\nwant=%+v", got, a)
			}

			// ValidateFile exercises schema + content-hash integrity from bytes.
			path := filepath.Join(t.TempDir(), tc.name+".json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatalf("write: %v", err)
			}
			if issues := ValidateFile(path); len(issues) != 0 {
				t.Fatalf("ValidateFile rejected valid artifact: %v", issues)
			}
		})
	}
}

func TestArtifactHashIsDeterministicAndExcludesHashField(t *testing.T) {
	for _, tc := range artifactCases() {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.build()
			h1, err := contentHashOf(a)
			if err != nil {
				t.Fatal(err)
			}
			// Independently built instance must hash identically.
			b := tc.build()
			h2, err := contentHashOf(b)
			if err != nil {
				t.Fatal(err)
			}
			if h1 != h2 {
				t.Fatalf("hash not deterministic across instances: %s != %s", h1, h2)
			}
			// Stamping must not change the hash (hash excludes content_hash).
			if err := Stamp(a); err != nil {
				t.Fatal(err)
			}
			h3, err := contentHashOf(a)
			if err != nil {
				t.Fatal(err)
			}
			if h3 != h1 {
				t.Fatalf("stamping changed the hash: %s != %s", h3, h1)
			}
			if a.EnvelopeRef().ContentHash != h1 {
				t.Fatalf("stamped hash %s != computed %s", a.EnvelopeRef().ContentHash, h1)
			}
		})
	}
}

func TestArtifactsRejectInvalidAgainstSchema(t *testing.T) {
	for _, tc := range artifactCases() {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.build()
			data, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			var generic map[string]any
			if err := json.Unmarshal(data, &generic); err != nil {
				t.Fatal(err)
			}
			tc.breakOne(generic)
			broken, err := json.Marshal(generic)
			if err != nil {
				t.Fatal(err)
			}
			if issues := ValidateBytes(tc.name, broken); len(issues) == 0 {
				t.Fatalf("expected schema to reject broken %s", tc.name)
			}
		})
	}
}

func contentHashOf(a Artifact) (string, error) {
	return contenthash.Compute(a)
}
