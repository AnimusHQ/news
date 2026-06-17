package activities

import (
	"context"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform/providers"
)

var aNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func scenes() []providers.SceneSpec {
	return []providers.SceneSpec{{SceneID: "scene-001", StartSec: 0, EndSec: 5}}
}

// TestActivitiesAreIdempotent asserts that re-running an activity with the same
// input yields a byte-identical artifact (same content hash), so retries keyed by
// episode_id + artifact_id + version are safe.
func TestActivitiesAreIdempotent(t *testing.T) {
	ctx := context.Background()
	a := NewMockActivities()

	in := ImportStoryboardInput{EpisodeID: "episode-0001", Now: aNow, Operator: "operator:ci", Scenes: scenes()}
	first, err := a.ImportStoryboardImages(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	second, err := a.ImportStoryboardImages(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if first.ContentHash == "" || first.ContentHash != second.ContentHash {
		t.Fatalf("activity not idempotent: %q vs %q", first.ContentHash, second.ContentHash)
	}
	if first.ArtifactID != second.ArtifactID {
		t.Fatalf("artifact id not stable: %q vs %q", first.ArtifactID, second.ArtifactID)
	}
}

func TestProductionQADependsOnRenderQuality(t *testing.T) {
	ctx := context.Background()
	// Healthy render -> approved.
	good := NewMockActivities()
	sb, _ := good.ImportStoryboardImages(ctx, ImportStoryboardInput{EpisodeID: "e", Now: aNow, Operator: "op", Scenes: scenes()})
	shots, _ := good.GenerateMockVisualShots(ctx, VisualShotsInput{EpisodeID: "e", Now: aNow, Storyboard: sb})
	vo, _ := good.GenerateElevenLabsVoiceover(ctx, VoiceoverInput{EpisodeID: "e", Now: aNow, ScriptRef: "s"})
	subs, _ := good.GenerateSubtitles(ctx, SubtitlesInput{EpisodeID: "e", Now: aNow, Voiceover: vo})
	render, _ := good.RenderShortFinal(ctx, RenderInput{EpisodeID: "e", Now: aNow, Shots: shots, Voiceover: vo, Subtitles: subs})
	qa, _ := good.RunProductionQA(ctx, render)
	if qa.Decision != "approved" {
		t.Fatalf("healthy render should pass QA, got %s", qa.Decision)
	}

	// No-audio render -> request_revision.
	bad := NewMockActivitiesWithDefects(MockDefects{Render: providers.DefectRenderNoAudio})
	badRender, _ := bad.RenderShortFinal(ctx, RenderInput{EpisodeID: "e", Now: aNow, Shots: shots, Voiceover: vo, Subtitles: subs})
	badQA, _ := bad.RunProductionQA(ctx, badRender)
	if badQA.Decision == "approved" {
		t.Fatal("no-audio render must fail QA")
	}
}

func TestDeferredActivitiesNeverRunInM1(t *testing.T) {
	ctx := context.Background()
	a := NewMockActivities()
	if _, err := a.GenerateSeedanceShots(ctx, VisualShotsInput{}); err == nil {
		t.Fatal("GenerateSeedanceShots must be disabled in M1")
	}
	if err := a.UploadPostSchedulePublish(ctx, nil); err == nil {
		t.Fatal("UploadPostSchedulePublish must be disabled in M1")
	}
}

func TestUploadPostDryRunRefusesNonDryRunMode(t *testing.T) {
	ctx := context.Background()
	a := NewMockActivities()
	res, err := a.UploadPostDryRun(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.OK {
		t.Fatal("nil manifest must not pass dry-run")
	}
}
