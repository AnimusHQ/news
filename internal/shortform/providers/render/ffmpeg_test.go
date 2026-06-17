package render

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestFFmpegProviderDisabledByDefault(t *testing.T) {
	p := NewFFmpegProvider(FFmpegConfig{})
	_, err := p.RenderShort(context.Background(), providers.RenderRequest{EpisodeID: "episode-0001"})
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled fail-closed error, got %v", err)
	}
}

func TestFFmpegProviderMissingBinaryFailsClosed(t *testing.T) {
	p := NewFFmpegProvider(FFmpegConfig{Enabled: true, FFmpegPath: filepath.Join(t.TempDir(), "missing-ffmpeg")})
	_, err := p.RenderShort(context.Background(), providers.RenderRequest{EpisodeID: "episode-0001"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing ffmpeg error, got %v", err)
	}
}

func TestFFmpegProviderRejectsInvalidInputPath(t *testing.T) {
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	ffmpeg := writeExecutable(t, "exit 0\n")
	req := ffmpegRequest(inputRoot)
	req.Shots.Shots[0].OutputPath = filepath.Join(t.TempDir(), "outside.mp4")

	p := NewFFmpegProvider(FFmpegConfig{Enabled: true, FFmpegPath: ffmpeg, InputRoot: inputRoot, OutputRoot: outputRoot})
	_, err := p.RenderShort(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected input containment error, got %v", err)
	}
}

func TestFFmpegProviderRejectsPathTraversal(t *testing.T) {
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	makePlaceholderInputs(t, inputRoot)
	ffmpeg := writeExecutable(t, "exit 0\n")
	req := ffmpegRequest(inputRoot)
	req.Subtitles.SRTPath = "../captions.srt"

	p := NewFFmpegProvider(FFmpegConfig{Enabled: true, FFmpegPath: ffmpeg, InputRoot: inputRoot, OutputRoot: outputRoot})
	_, err := p.RenderShort(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected path traversal error, got %v", err)
	}
}

func TestFFmpegProviderTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake executable is POSIX-only")
	}
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	makePlaceholderInputs(t, inputRoot)
	ffmpeg := writeExecutable(t, "sleep 5\n")

	p := NewFFmpegProvider(FFmpegConfig{
		Enabled: true, FFmpegPath: ffmpeg, InputRoot: inputRoot, OutputRoot: outputRoot,
		Timeout: 20 * time.Millisecond,
	})
	_, err := p.RenderShort(context.Background(), ffmpegRequest(inputRoot))
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}

func TestFFmpegCommandDoesNotInvokeShell(t *testing.T) {
	cmd := buildFFmpegCommand(context.Background(), "/tmp/ffmpeg", "/tmp/input.mp4", "/tmp/audio.wav", "/tmp/out.mp4", "/tmp/work")
	if filepath.Base(cmd.Path) == "sh" || filepath.Base(cmd.Path) == "bash" {
		t.Fatalf("ffmpeg command must not invoke a shell: %s", cmd.Path)
	}
	for _, arg := range cmd.Args {
		if arg == "-c" {
			t.Fatalf("ffmpeg command must not use shell-style -c args: %v", cmd.Args)
		}
	}
}

func TestFFmpegProviderNormalizesLocalFixtureWhenAvailable(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg binary not available; skipping local render integration test")
	}
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	makeFFmpegFixture(t, ffmpeg, inputRoot)

	p := NewFFmpegProvider(FFmpegConfig{
		Enabled: true, FFmpegPath: ffmpeg, InputRoot: inputRoot, OutputRoot: outputRoot,
		Timeout: 30 * time.Second, Version: "test",
	})
	manifest, err := p.RenderShort(context.Background(), ffmpegRequest(inputRoot))
	if err != nil {
		t.Fatalf("render local fixture: %v", err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("render manifest rejected: %v", issues)
	}
	if len(manifest.Outputs) != 1 {
		t.Fatalf("expected one output, got %d", len(manifest.Outputs))
	}
	out := manifest.Outputs[0]
	if out.Resolution != shortform.TargetResolution || out.Aspect != shortform.TargetAspect || out.FPS != shortform.TargetFPS {
		t.Fatalf("unexpected output properties: %+v", out)
	}
	if out.VideoCodec != shortform.TargetVideoCodec || out.AudioCodec != shortform.TargetAudioCodec || !out.AudioTrack || !out.SubtitlesBurned {
		t.Fatalf("unexpected codec/audio/subtitle properties: %+v", out)
	}
	outputPath := filepath.Join(outputRoot, "episode-0001", filepath.FromSlash(out.Path))
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("render output missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("render output is empty")
	}
	if out.Hash == "" {
		t.Fatal("render output hash not recorded")
	}
}

func ffmpegRequest(_ string) providers.RenderRequest {
	return providers.RenderRequest{
		EpisodeID: "episode-0001",
		Now:       testNow,
		Shots: &shortform.VisualShotManifest{
			Envelope: shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "visual-shot-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Shots: []shortform.VisualShot{{
				SceneID: "scene-001", OutputPath: "shots/input.mp4", OutputHash: fixtureHash("shot"),
				DurationSec: 1, Status: shortform.StatusApproved, OperatorApproval: true,
			}},
			Provider:     shortform.ProviderRef{Name: "fixture"},
			AspectRatio:  shortform.TargetAspect,
			RenderTarget: shortform.RenderTarget{Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS, VideoCodec: shortform.TargetVideoCodec},
		},
		Voiceover: &shortform.VoiceoverManifest{
			Envelope: shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "voiceover-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider: shortform.ProviderRef{Name: "fixture"}, SourceScriptRef: "script.md", Language: "en",
			Output: shortform.MediaOutput{Path: "voice/voice.wav", Hash: fixtureHash("voice"), DurationSec: 1, Format: "wav"},
		},
		Subtitles: &shortform.SubtitleManifest{
			Envelope:       shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "subtitle-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider:       shortform.ProviderRef{Name: "fixture"},
			Language:       "en",
			TranscriptPath: "subtitles/transcript.json", TranscriptHash: fixtureHash("transcript"),
			SRTPath: "subtitles/captions.srt", SRTHash: fixtureHash("srt"),
			Checks:           shortform.SubtitleChecks{WordTimestamps: true, SafeZone: true, Sync: true},
			OperatorApproval: true,
		},
		Platforms: []string{"master"},
	}
}

func makePlaceholderInputs(t *testing.T, root string) {
	t.Helper()
	for _, dir := range []string{"shots", "voice", "subtitles"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(root, "shots", "input.mp4"), "video")
	writeFile(t, filepath.Join(root, "voice", "voice.wav"), "audio")
	writeFile(t, filepath.Join(root, "subtitles", "captions.srt"), "1\n00:00:00,000 --> 00:00:00,900\nhello\n")
}

func makeFFmpegFixture(t *testing.T, ffmpeg, root string) {
	t.Helper()
	makePlaceholderInputs(t, root)
	video := filepath.Join(root, "shots", "input.mp4")
	audio := filepath.Join(root, "voice", "voice.wav")
	if out, err := exec.Command(ffmpeg,
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=size=320x240:rate=30:duration=1",
		"-an", "-c:v", "libx264", "-pix_fmt", "yuv420p", video,
	).CombinedOutput(); err != nil {
		t.Skipf("ffmpeg fixture video generation failed: %v: %s", err, string(out))
	}
	if out, err := exec.Command(ffmpeg,
		"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=1000:duration=1",
		"-c:a", "pcm_s16le", audio,
	).CombinedOutput(); err != nil {
		t.Skipf("ffmpeg fixture audio generation failed: %v: %s", err, string(out))
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ffmpeg")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixtureHash(seed string) string {
	switch seed {
	case "shot":
		return "sha256:2eb4e86b6167fa60ed1078d307e1ee7514a7f106f6b5ab4f5ebccd0bcd3bcb71"
	case "voice":
		return "sha256:e8c949f1480dd695f839a2b117dbcb7ea5bc8a3612a953d0568f18338137d8fe"
	case "transcript":
		return "sha256:8513e0973a7cd8bb2ef693d72fbb5778322557da68d51b97bc8eba9321df86a7"
	default:
		return "sha256:d9f7760a2dd5dd161fb3edf1c0474389196a6c864b8820e84ad2a9a859db44d3"
	}
}
