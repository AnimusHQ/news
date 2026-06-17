package uploadpost

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestUploadPostDryRunDisabledByDefault(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{})
	_, err := p.UploadPostDryRun(context.Background(), publishRequest())
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled fail-closed error, got %v", err)
	}
}

func TestUploadPostDryRunSuccess(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	manifest, err := p.UploadPostDryRun(context.Background(), publishRequest())
	if err != nil {
		t.Fatalf("dry-run success: %v", err)
	}
	if manifest.Mode != "dry_run" || !manifest.DryRun {
		t.Fatalf("manifest must be dry-run only: %+v", manifest)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("publish manifest rejected: %v", issues)
	}
	if manifest.ContentHash == "" {
		t.Fatal("publish manifest content hash missing")
	}
}

func TestUploadPostDryRunBlocksMissingProductionQA(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	req := publishRequest()
	req.ProductionQADecision = "request_revision"
	_, err := p.UploadPostDryRun(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "production QA") {
		t.Fatalf("expected production QA block, got %v", err)
	}
}

func TestUploadPostDryRunBlocksMissingReleaseApproval(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	req := publishRequest()
	req.Release.HumanReleaseApproval = false
	_, err := p.UploadPostDryRun(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "release") {
		t.Fatalf("expected release approval block, got %v", err)
	}
}

func TestUploadPostDryRunBlocksMissingDisclosureWhenRequired(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	req := publishRequest()
	req.Release.AIDisclosure = ""
	req.Release.RiskAcceptance.AIDisclosurePresent = false
	_, err := p.UploadPostDryRun(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "disclosure") {
		t.Fatalf("expected disclosure block, got %v", err)
	}
}

func TestUploadPostLiveModeRefusesInM2AndRedactsAPIKey(t *testing.T) {
	key := "up_live_secret_123456789"
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "publish", APIKey: key})
	_, err := p.UploadPostDryRun(context.Background(), publishRequest())
	if err == nil || !strings.Contains(err.Error(), "refused in M2") {
		t.Fatalf("expected live mode refusal, got %v", err)
	}
	if strings.Contains(err.Error(), key) {
		t.Fatalf("error leaked API key: %v", err)
	}
	if got := (DryRunConfig{APIKey: key}).Redacted().APIKey; got != "[REDACTED]" {
		t.Fatalf("redacted config leaked key: %q", got)
	}
}

func TestUploadPostDryRunDoesNotRequireAPIKey(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	if _, err := p.UploadPostDryRun(context.Background(), publishRequest()); err != nil {
		t.Fatalf("dry-run must not require API key: %v", err)
	}
}

func TestUploadPostDryRunValidatesPlatforms(t *testing.T) {
	p := NewDryRunProvider(DryRunConfig{Enabled: true, Mode: "dry_run"})
	req := publishRequest()
	req.Release.Platforms = []string{"master"}
	_, err := p.UploadPostDryRun(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "platform") {
		t.Fatalf("expected platform validation error, got %v", err)
	}
}

func publishRequest() providers.PublishRequest {
	return providers.PublishRequest{
		EpisodeID:            "episode-0001",
		Now:                  testNow,
		Render:               approvedRender(),
		Release:              approvedRelease(),
		ProductionQADecision: gates.ProductionQAApproved,
		ProductionQARef:      "production_qa_report.json",
	}
}

func approvedRender() *shortform.ShortRenderManifest {
	return &shortform.ShortRenderManifest{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     "episode-0001",
			ArtifactID:    "short_render_manifest-episode-0001-v1",
			CreatedAt:     testNow.Format(time.RFC3339),
			CreatedBy:     "system:ffmpeg-local",
			Status:        shortform.StatusApproved,
		},
		Renderer: shortform.RendererRef{Name: "ffmpeg", Version: "test"},
		Inputs:   []string{"visual_shot_manifest.json", "voiceover_manifest.json", "subtitle_manifest.json"},
		Outputs: []shortform.RenderOutput{{
			Platform: "youtube", Path: "renders/youtube.mp4",
			Hash:       "sha256:3e37adf74d5fa927e21aa8225d0ccb1fb3a4daf8b3f4ad6c69b440fb81d03dc0",
			Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS,
			VideoCodec: shortform.TargetVideoCodec, AudioCodec: shortform.TargetAudioCodec,
			AudioTrack: true, SubtitlesBurned: true, DurationSec: 1, Status: shortform.StatusApproved,
		}},
	}
}

func approvedRelease() *shortform.ReleaseApproval {
	return &shortform.ReleaseApproval{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     "episode-0001",
			ArtifactID:    "release_approval-episode-0001-v1",
			CreatedAt:     testNow.Format(time.RFC3339),
			CreatedBy:     "human:editor",
			Status:        shortform.StatusApproved,
		},
		CandidateID:          "candidate-001",
		Platforms:            []string{"youtube"},
		Visibility:           "private",
		AIDisclosureRequired: true,
		AIDisclosure:         "AI-generated visuals and synthetic voice.",
		HumanReleaseApproval: true,
		ApprovedBy:           "human:editor",
		ApprovedAt:           testNow.Format(time.RFC3339),
		ProductionQARef:      "production_qa_report.json",
		RiskAcceptance:       shortform.RiskAcceptance{AIGeneratedVisuals: true, AIDisclosurePresent: true, BrandSafetyChecked: true},
	}
}
