package subtitles

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestFasterWhisperProviderDisabledByDefault(t *testing.T) {
	p := NewFasterWhisperProvider(FasterWhisperConfig{})
	_, err := p.GenerateSubtitles(context.Background(), providers.SubtitleRequest{EpisodeID: "episode-0001"})
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled fail-closed error, got %v", err)
	}
}

func TestFasterWhisperProviderMissingBinaryFailsClosed(t *testing.T) {
	cfg := subtitleConfig(t, nil)
	cfg.BinaryPath = filepath.Join(t.TempDir(), "missing-sidecar")
	p := NewFasterWhisperProvider(cfg)
	_, err := p.GenerateSubtitles(context.Background(), subtitleRequest())
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing binary error, got %v", err)
	}
}

func TestFasterWhisperProviderMissingModelFailsClosed(t *testing.T) {
	cfg := subtitleConfig(t, nil)
	cfg.ModelPath = "missing-model"
	p := NewFasterWhisperProvider(cfg)
	_, err := p.GenerateSubtitles(context.Background(), subtitleRequest())
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestFasterWhisperProviderRejectsInvalidAudioPath(t *testing.T) {
	cfg := subtitleConfig(t, nil)
	req := subtitleRequest()
	req.Voiceover.Output.Path = "../voice.wav"
	p := NewFasterWhisperProvider(cfg)
	_, err := p.GenerateSubtitles(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected path containment error, got %v", err)
	}
}

func TestFasterWhisperProviderValidatesSidecarManifest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake executable is POSIX-only")
	}
	cfg := subtitleConfig(t, &sidecarOptions{WordTimestamps: true, SafeZone: true, Sync: true})
	p := NewFasterWhisperProvider(cfg)
	manifest, err := p.GenerateSubtitles(context.Background(), subtitleRequest())
	if err != nil {
		t.Fatalf("generate subtitles: %v", err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("subtitle manifest rejected: %v", issues)
	}
	if manifest.ContentHash == "" || manifest.TranscriptHash == "" || manifest.SRTHash == "" || manifest.ASSHash == "" {
		t.Fatalf("expected content and output hashes, got %+v", manifest)
	}
	if manifest.OperatorApproval {
		t.Fatal("sidecar output must remain a draft artifact requiring approval")
	}
}

func TestFasterWhisperProviderRequiresWordTimestampsWhenRequested(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake executable is POSIX-only")
	}
	cfg := subtitleConfig(t, &sidecarOptions{WordTimestamps: false, SafeZone: true, Sync: true})
	p := NewFasterWhisperProvider(cfg)
	_, err := p.GenerateSubtitles(context.Background(), subtitleRequest())
	if err == nil || !strings.Contains(err.Error(), "word timestamps") {
		t.Fatalf("expected word timestamp requirement error, got %v", err)
	}
}

func TestFasterWhisperProviderPreservesSafeZoneCheckForGate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake executable is POSIX-only")
	}
	cfg := subtitleConfig(t, &sidecarOptions{WordTimestamps: true, SafeZone: false, Sync: true})
	p := NewFasterWhisperProvider(cfg)
	req := subtitleRequest()
	req.WordTimestampsRequired = true
	manifest, err := p.GenerateSubtitles(context.Background(), req)
	if err != nil {
		t.Fatalf("safe-zone failure should remain a gate decision, got provider error: %v", err)
	}
	result := gates.SubtitleGate(gates.SubtitleInput{Manifest: manifest})
	if !result.Blocked() {
		t.Fatal("safe-zone failure must still block at the subtitle gate")
	}
}

func TestMockSubtitlePathStillPasses(t *testing.T) {
	ctx := context.Background()
	vo, err := providers.MockVoiceProvider{}.SynthesizeVoiceover(ctx, providers.VoiceoverRequest{EpisodeID: "episode-0001", Now: testNow, ScriptRef: "script.md", Language: "en"})
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := providers.MockSubtitleProvider{}.GenerateSubtitles(ctx, providers.SubtitleRequest{
		EpisodeID: "episode-0001", Now: testNow, Voiceover: vo, Language: "en", WordTimestampsRequired: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("mock subtitle manifest rejected: %v", issues)
	}
	if !manifest.Checks.WordTimestamps {
		t.Fatal("mock subtitle provider must continue representing word timestamps")
	}
}

type sidecarOptions struct {
	WordTimestamps bool
	SafeZone       bool
	Sync           bool
}

func subtitleConfig(t *testing.T, opts *sidecarOptions) FasterWhisperConfig {
	t.Helper()
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	modelRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(inputRoot, "voice"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(modelRoot, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputRoot, "voice", "voice.wav"), []byte("audio"), 0o644); err != nil {
		t.Fatal(err)
	}
	return FasterWhisperConfig{
		Enabled: true, BinaryPath: writeSidecar(t, opts),
		InputRoot: inputRoot, OutputRoot: outputRoot, ModelRoot: modelRoot, ModelPath: "base",
		Timeout: time.Second, ProviderVersion: "test",
	}
}

func subtitleRequest() providers.SubtitleRequest {
	return providers.SubtitleRequest{
		EpisodeID: "episode-0001",
		Now:       testNow,
		Voiceover: &shortform.VoiceoverManifest{
			Envelope: shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "voiceover-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider: shortform.ProviderRef{Name: "fixture"}, SourceScriptRef: "script.md", Language: "en",
			Output: shortform.MediaOutput{Path: "voice/voice.wav", Hash: "sha256:e8c949f1480dd695f839a2b117dbcb7ea5bc8a3612a953d0568f18338137d8fe", DurationSec: 1, Format: "wav"},
		},
		Language:               "en",
		WordTimestampsRequired: true,
	}
}

func writeSidecar(t *testing.T, opts *sidecarOptions) string {
	t.Helper()
	if opts == nil {
		opts = &sidecarOptions{WordTimestamps: true, SafeZone: true, Sync: true}
	}
	path := filepath.Join(t.TempDir(), "fw-sidecar")
	body := `#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --output-dir) out="$2"; shift 2 ;;
    *) shift ;;
  esac
done
mkdir -p "$out"
printf '{"language":"en","segments":[{"start":0,"end":1,"text":"hello","words":[{"word":"hello","start":0,"end":1}]}]}' > "$out/transcript.json"
printf '1\n00:00:00,000 --> 00:00:00,900\nhello\n' > "$out/captions.srt"
printf '[Script Info]\nTitle: fixture\n[V4+ Styles]\n[Events]\n' > "$out/captions.ass"
printf '{"provider":{"name":"faster_whisper","model":"base","version":"test"},"language":"en","transcript_path":"transcript.json","srt_path":"captions.srt","ass_path":"captions.ass","checks":{"word_timestamps":` + boolLiteral(opts.WordTimestamps) + `,"safe_zone":` + boolLiteral(opts.SafeZone) + `,"sync":` + boolLiteral(opts.Sync) + `}}'
`
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func boolLiteral(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
