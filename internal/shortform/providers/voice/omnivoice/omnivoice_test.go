package omnivoice

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestOmniVoiceDisabledByDefault(t *testing.T) {
	p := NewProvider(OmniVoiceConfig{})
	_, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled fail-closed error, got %v", err)
	}
}

func TestOmniVoiceMissingBinaryFailsClosed(t *testing.T) {
	cfg := omniConfig(t, "dry_run")
	cfg.BinaryPath = filepath.Join(t.TempDir(), "missing-omnivoice")
	p := NewProvider(cfg)
	_, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing binary error, got %v", err)
	}
}

func TestOmniVoiceMissingModelFailsClosed(t *testing.T) {
	cfg := omniConfig(t, "dry_run")
	cfg.ModelPath = "missing-model"
	p := NewProvider(cfg)
	_, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err == nil || !strings.Contains(err.Error(), "model") {
		t.Fatalf("expected missing model error, got %v", err)
	}
}

func TestOmniVoiceDryRunVoiceManifestValidates(t *testing.T) {
	p := NewProvider(omniConfig(t, "dry_run"))
	manifest, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err != nil {
		t.Fatalf("dry-run voice: %v", err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("voiceover manifest rejected: %v", issues)
	}
	if manifest.Status != shortform.StatusDraft || manifest.OperatorApproval {
		t.Fatalf("OmniVoice output must remain draft and unapproved: %+v", manifest)
	}
	if manifest.Provider.Name != "omnivoice" || manifest.Output.SampleRateHz == 0 {
		t.Fatalf("missing provider/sample-rate metadata: %+v", manifest)
	}
}

func TestOmniVoiceReferenceVoiceRequiresConsentMetadata(t *testing.T) {
	cfg := omniConfig(t, "dry_run")
	req := voiceRequest()
	req.VoicePromptReference = "voice-prompts/editorial-style.json"
	req.VoiceConsentRequired = true
	p := NewProvider(cfg)
	_, err := p.SynthesizeVoiceover(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "consent") {
		t.Fatalf("expected consent error, got %v", err)
	}

	req.VoiceConsentReference = "consent/operator-approved-voice.json"
	manifest, err := p.SynthesizeVoiceover(context.Background(), req)
	if err != nil {
		t.Fatalf("consented dry-run voice failed: %v", err)
	}
	if manifest.VoiceConsentReference == "" || manifest.VoicePromptReference == "" {
		t.Fatalf("consent metadata not recorded: %+v", manifest)
	}
}

func TestOmniVoiceReferenceAudioRequiresAllowedAndHash(t *testing.T) {
	cfg := omniConfig(t, "dry_run")
	makeReferenceAudio(t, cfg.InputRoot)
	req := voiceRequest()
	req.ReferenceAudioPath = "refs/reference.wav"
	req.VoiceConsentReference = "consent/operator-approved-voice.json"
	req.VoiceConsentRequired = true
	p := NewProvider(cfg)
	_, err := p.SynthesizeVoiceover(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "explicitly allowed") {
		t.Fatalf("expected reference audio allowed error, got %v", err)
	}
	req.ReferenceAudioAllowed = true
	_, err = p.SynthesizeVoiceover(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "hash") {
		t.Fatalf("expected reference audio hash error, got %v", err)
	}
	req.ReferenceAudioHash = "sha256:e8c949f1480dd695f839a2b117dbcb7ea5bc8a3612a953d0568f18338137d8fe"
	if _, err := p.SynthesizeVoiceover(context.Background(), req); err != nil {
		t.Fatalf("consented reference audio dry-run failed: %v", err)
	}
}

func TestOmniVoiceSidecarInvalidOutputPathBlocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake sidecar is POSIX-only")
	}
	cfg := omniConfig(t, "local_sidecar")
	cfg.BinaryPath = writeSidecar(t, "../outside.wav")
	p := NewProvider(cfg)
	_, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected invalid output path block, got %v", err)
	}
}

func TestOmniVoiceSidecarOutputManifestValidates(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-script fake sidecar is POSIX-only")
	}
	cfg := omniConfig(t, "local_sidecar")
	cfg.BinaryPath = writeSidecar(t, "voice.wav")
	p := NewProvider(cfg)
	manifest, err := p.SynthesizeVoiceover(context.Background(), voiceRequest())
	if err != nil {
		t.Fatalf("sidecar voice: %v", err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("voiceover manifest rejected: %v", issues)
	}
	if manifest.Output.Hash == "" || manifest.Output.Path != "voice/voice.wav" {
		t.Fatalf("unexpected sidecar output metadata: %+v", manifest.Output)
	}
}

func TestOmniVoiceConfigRedactsLocalDetails(t *testing.T) {
	cfg := OmniVoiceConfig{BinaryPath: "/tmp/private/bin", ModelRoot: "/tmp/private/models", ModelPath: "secret-model", InputRoot: "/tmp/private/in", OutputRoot: "/tmp/private/out"}
	got := cfg.Redacted()
	for _, value := range []string{got.BinaryPath, got.ModelRoot, got.ModelPath, got.InputRoot, got.OutputRoot} {
		if value != "[REDACTED]" {
			t.Fatalf("expected redacted value, got %q", value)
		}
	}
}

func omniConfig(t *testing.T, mode string) OmniVoiceConfig {
	t.Helper()
	modelRoot := t.TempDir()
	inputRoot := t.TempDir()
	outputRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(modelRoot, "base"), 0o755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "omnivoice")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return OmniVoiceConfig{
		Enabled: true, Mode: mode, BinaryPath: binary,
		ModelRoot: modelRoot, ModelPath: "base", InputRoot: inputRoot, OutputRoot: outputRoot,
		Timeout: time.Second, RequireConsent: true, Version: "test",
	}
}

func voiceRequest() providers.VoiceoverRequest {
	return providers.VoiceoverRequest{
		EpisodeID: "episode-0001",
		Now:       testNow,
		ScriptRef: "script.md",
		Text:      "hello from animus news",
		Language:  "en",
	}
}

func makeReferenceAudio(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "refs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "refs", "reference.wav"), []byte("reference audio"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeSidecar(t *testing.T, outputPath string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "omnivoice-sidecar")
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
printf 'audio' > "$out/voice.wav"
printf '{"output_path":"` + outputPath + `","duration_sec":1,"format":"wav","sample_rate_hz":24000}'
`
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}
