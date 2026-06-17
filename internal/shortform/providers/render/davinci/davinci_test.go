package davinci

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/shortform/providers/mcp"
)

var testNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestDaVinciProviderDisabledByDefault(t *testing.T) {
	p := NewProvider(DaVinciResolveConfig{}, nil)
	_, err := p.RenderShort(context.Background(), davinciRequest(t.TempDir()))
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled fail-closed error, got %v", err)
	}
}

func TestDaVinciProviderMissingMCPURLFailsClosed(t *testing.T) {
	root := t.TempDir()
	makeInputs(t, root)
	cfg := davinciConfig(t, root)
	cfg.MCPURL = ""
	p := NewProvider(cfg, nil)
	_, err := p.RenderShort(context.Background(), davinciRequest(root))
	if err == nil || !strings.Contains(err.Error(), "MCP URL") {
		t.Fatalf("expected missing MCP URL error, got %v", err)
	}
}

func TestDaVinciProviderRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	makeInputs(t, root)
	req := davinciRequest(root)
	req.Subtitles.SRTPath = "../captions.srt"
	p := NewProvider(davinciConfig(t, root), nil)
	_, err := p.RenderShort(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected traversal block, got %v", err)
	}
}

func TestDaVinciProviderDryRunManifestConstruction(t *testing.T) {
	root := t.TempDir()
	makeInputs(t, root)
	p := NewProvider(davinciConfig(t, root), nil)
	manifest, err := p.RenderShort(context.Background(), davinciRequest(root))
	if err != nil {
		t.Fatalf("dry-run render: %v", err)
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		t.Fatalf("DaVinci manifest rejected: %v", issues)
	}
	if manifest.Status != shortform.StatusDraft || manifest.Outputs[0].Status != shortform.StatusDraft {
		t.Fatalf("DaVinci output must remain draft: %+v", manifest)
	}
	if manifest.Renderer.Name != "davinci_resolve_mcp" {
		t.Fatalf("unexpected renderer: %+v", manifest.Renderer)
	}
	if manifest.ProviderMetadata == nil || manifest.ProviderMetadata.Provider != "davinci_resolve_mcp" {
		t.Fatalf("provider metadata missing: %+v", manifest.ProviderMetadata)
	}
	if len(manifest.ProviderMetadata.MCPTools) == 0 {
		t.Fatal("expected dry-run MCP tool plan to be recorded")
	}
}

func TestDaVinciProviderStartRenderRequiresExplicitAllowance(t *testing.T) {
	root := t.TempDir()
	makeInputs(t, root)
	req := davinciRequest(root)
	req.StartRender = true
	p := NewProvider(davinciConfig(t, root), nil)
	_, err := p.RenderShort(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "AllowStartRender") {
		t.Fatalf("expected start-render refusal, got %v", err)
	}

	client := &recordingClient{}
	cfg := davinciConfig(t, root)
	cfg.AllowStartRender = true
	manifest, err := NewProvider(cfg, client).RenderShort(context.Background(), req)
	if err != nil {
		t.Fatalf("allowed start-render dry-run failed: %v", err)
	}
	if manifest.Status != shortform.StatusDraft {
		t.Fatalf("allowed start-render still must produce draft output, got %s", manifest.Status)
	}
	if !client.saw(mcp.ToolResolveStartRenderIfAllowed) {
		t.Fatal("expected start_render_if_allowed MCP tool in plan")
	}
}

func TestDaVinciProviderCannotBypassProductionQA(t *testing.T) {
	root := t.TempDir()
	makeInputs(t, root)
	manifest, err := NewProvider(davinciConfig(t, root), nil).RenderShort(context.Background(), davinciRequest(root))
	if err != nil {
		t.Fatal(err)
	}
	result := gates.RenderGate(gates.RenderInput{Manifest: manifest, ProductionQADecision: "request_revision"})
	if !result.Blocked() {
		t.Fatal("DaVinci draft render must not bypass production QA")
	}
}

func TestDaVinciConfigRedactsLocalDetails(t *testing.T) {
	cfg := DaVinciResolveConfig{
		MCPURL: "http://127.0.0.1:8989/mcp", InputRoot: "/tmp/private/input",
		OutputRoot: "/tmp/private/output", ProjectRoot: "/tmp/private/projects",
	}
	got := cfg.Redacted()
	for _, value := range []string{got.MCPURL, got.InputRoot, got.OutputRoot, got.ProjectRoot} {
		if value != "[REDACTED]" {
			t.Fatalf("expected redacted value, got %q", value)
		}
	}
}

type recordingClient struct {
	calls []string
}

func (c *recordingClient) Call(_ context.Context, req mcp.CallRequest) (mcp.CallResponse, error) {
	if err := mcp.ValidateResolveTool(req.Tool); err != nil {
		return mcp.CallResponse{}, err
	}
	c.calls = append(c.calls, req.Tool)
	return mcp.CallResponse{Tool: req.Tool, OK: true, Status: "test"}, nil
}

func (c *recordingClient) saw(tool string) bool {
	for _, call := range c.calls {
		if call == tool {
			return true
		}
	}
	return false
}

func davinciConfig(t *testing.T, root string) DaVinciResolveConfig {
	t.Helper()
	projectRoot := filepath.Join(t.TempDir(), "projects")
	outputRoot := filepath.Join(t.TempDir(), "outputs")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	return DaVinciResolveConfig{
		Enabled: true, Mode: ModeDryRun, MCPURL: "http://127.0.0.1:8989/mcp",
		InputRoot: root, OutputRoot: outputRoot, ProjectRoot: projectRoot,
		Timeout: time.Second, Version: "test",
	}
}

func davinciRequest(_ string) providers.RenderRequest {
	return providers.RenderRequest{
		EpisodeID: "episode-0001",
		Now:       testNow,
		Shots: &shortform.VisualShotManifest{
			Envelope: shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "visual_shot_manifest-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider: shortform.ProviderRef{Name: "fixture"}, AspectRatio: shortform.TargetAspect,
			RenderTarget: shortform.RenderTarget{Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS, VideoCodec: shortform.TargetVideoCodec},
			Shots: []shortform.VisualShot{{
				SceneID: "scene-001", OutputPath: "shots/input.mp4", OutputHash: fixtureHash("shot"),
				DurationSec: 1, Status: shortform.StatusApproved, OperatorApproval: true,
			}},
		},
		Voiceover: &shortform.VoiceoverManifest{
			Envelope: shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "voiceover_manifest-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider: shortform.ProviderRef{Name: "fixture"}, SourceScriptRef: "script.md", Language: "en",
			Output: shortform.MediaOutput{Path: "voice/voice.wav", Hash: fixtureHash("voice"), DurationSec: 1, Format: "wav", SampleRateHz: 48000},
		},
		Subtitles: &shortform.SubtitleManifest{
			Envelope:       shortform.Envelope{SchemaVersion: shortform.SchemaVersion, EpisodeID: "episode-0001", ArtifactID: "subtitle_manifest-episode-0001-v1", CreatedAt: testNow.Format(time.RFC3339), CreatedBy: "model:fixture", Status: shortform.StatusApproved},
			Provider:       shortform.ProviderRef{Name: "fixture"},
			Language:       "en",
			TranscriptPath: "subtitles/transcript.json", TranscriptHash: fixtureHash("transcript"),
			SRTPath: "subtitles/captions.srt", SRTHash: fixtureHash("srt"),
			Checks:           shortform.SubtitleChecks{WordTimestamps: true, SafeZone: true, Sync: true},
			OperatorApproval: true,
		},
		Platforms: []string{"youtube"},
	}
}

func makeInputs(t *testing.T, root string) {
	t.Helper()
	for _, dir := range []string{"shots", "voice", "subtitles"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "shots", "input.mp4"), []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "voice", "voice.wav"), []byte("audio"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "subtitles", "captions.srt"), []byte("1\n00:00:00,000 --> 00:00:00,900\nhello\n"), 0o644); err != nil {
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
