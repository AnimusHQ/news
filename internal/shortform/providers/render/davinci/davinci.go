// Package davinci contains the DaVinci Resolve MCP render provider boundary.
// It is disabled by default and does not expose arbitrary MCP execution.
package davinci

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
	"github.com/AnimusHQ/news/internal/shortform/providers/mcp"
)

const (
	ModeDisabled = "disabled"
	ModeDryRun   = "dry_run"
	ModeLocalMCP = "local_mcp"
)

// DaVinciResolveConfig controls the optional Resolve MCP finishing lane.
type DaVinciResolveConfig struct {
	Enabled          bool
	MCPURL           string
	InputRoot        string
	OutputRoot       string
	ProjectRoot      string
	Timeout          time.Duration
	AllowStartRender bool
	Mode             string
	Version          string
}

// Redacted returns a copy safe for diagnostics.
func (c DaVinciResolveConfig) Redacted() DaVinciResolveConfig {
	if c.MCPURL != "" {
		c.MCPURL = "[REDACTED]"
	}
	if c.InputRoot != "" {
		c.InputRoot = "[REDACTED]"
	}
	if c.OutputRoot != "" {
		c.OutputRoot = "[REDACTED]"
	}
	if c.ProjectRoot != "" {
		c.ProjectRoot = "[REDACTED]"
	}
	return c
}

// Provider implements providers.RenderProvider through a constrained MCP client.
type Provider struct {
	cfg    DaVinciResolveConfig
	client mcp.Client
}

// NewProvider builds a DaVinci Resolve MCP provider. Pass nil client for dry-run
// mode to use an in-memory validating client.
func NewProvider(cfg DaVinciResolveConfig, client mcp.Client) *Provider {
	return &Provider{cfg: cfg, client: client}
}

// RenderShort implements providers.RenderProvider.
func (p *Provider) RenderShort(ctx context.Context, req providers.RenderRequest) (*shortform.ShortRenderManifest, error) {
	if !p.cfg.Enabled {
		return nil, fmt.Errorf("DaVinci Resolve MCP render adapter is disabled")
	}
	mode := p.mode()
	if mode != ModeDryRun && mode != ModeLocalMCP {
		return nil, fmt.Errorf("DaVinci Resolve MCP mode %q is not enabled", mode)
	}
	if err := mcp.ValidateEndpoint(p.cfg.MCPURL); err != nil {
		return nil, err
	}
	if req.StartRender && !p.cfg.AllowStartRender {
		return nil, fmt.Errorf("DaVinci Resolve start render refused; AllowStartRender must be true")
	}
	prepared, err := p.prepare(req)
	if err != nil {
		return nil, err
	}
	client, err := p.clientForMode(mode)
	if err != nil {
		return nil, err
	}
	tools, err := p.callPlan(ctx, client, prepared, req.StartRender)
	if err != nil {
		return nil, err
	}
	manifest := p.manifest(req, prepared, mode, tools)
	if err := shortform.Stamp(manifest); err != nil {
		return nil, err
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		return nil, fmt.Errorf("DaVinci render manifest failed validation: %v", issues)
	}
	return manifest, nil
}

type preparedRender struct {
	videoPath    string
	audioPath    string
	subtitlePath string
	projectDir   string
	outputDir    string
	platforms    []string
	durationSec  float64
}

func (p *Provider) prepare(req providers.RenderRequest) (preparedRender, error) {
	if req.EpisodeID == "" {
		return preparedRender{}, fmt.Errorf("episode_id is required")
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return preparedRender{}, err
	}
	if req.Shots == nil || len(req.Shots.Shots) == 0 {
		return preparedRender{}, fmt.Errorf("DaVinci render requires at least one visual shot")
	}
	if req.Voiceover == nil {
		return preparedRender{}, fmt.Errorf("DaVinci render requires a voiceover manifest")
	}
	if req.Subtitles == nil {
		return preparedRender{}, fmt.Errorf("DaVinci render requires a subtitle manifest")
	}
	videoPath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, req.Shots.Shots[0].OutputPath, "visual shot")
	if err != nil {
		return preparedRender{}, err
	}
	audioPath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, req.Voiceover.Output.Path, "voiceover")
	if err != nil {
		return preparedRender{}, err
	}
	subtitleRel := req.Subtitles.SRTPath
	if subtitleRel == "" {
		subtitleRel = req.Subtitles.ASSPath
	}
	subtitlePath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, subtitleRel, "subtitles")
	if err != nil {
		return preparedRender{}, err
	}
	projectDir, err := localexec.EnsureOutputDir(p.cfg.ProjectRoot, req.EpisodeID)
	if err != nil {
		return preparedRender{}, err
	}
	outputDir, err := localexec.EnsureOutputDir(p.cfg.OutputRoot, req.EpisodeID, "renders")
	if err != nil {
		return preparedRender{}, err
	}
	platforms, err := normalizePlatforms(req.Platforms)
	if err != nil {
		return preparedRender{}, err
	}
	duration := req.Voiceover.Output.DurationSec
	if duration <= 0 {
		duration = req.Shots.Shots[0].DurationSec
	}
	return preparedRender{
		videoPath: videoPath, audioPath: audioPath, subtitlePath: subtitlePath,
		projectDir: projectDir, outputDir: outputDir, platforms: platforms,
		durationSec: duration,
	}, nil
}

func (p *Provider) callPlan(ctx context.Context, client mcp.Client, prepared preparedRender, start bool) ([]string, error) {
	tools := []string{
		mcp.ToolResolveHealthcheck,
		mcp.ToolResolveCreateProjectFromManifest,
		mcp.ToolResolveImportAssetsFromManifest,
		mcp.ToolResolveCreateVerticalTimeline,
		mcp.ToolResolveQueueRenderFromManifest,
	}
	if start {
		tools = append(tools, mcp.ToolResolveStartRenderIfAllowed)
	}
	tools = append(tools,
		mcp.ToolResolveGetRenderStatus,
		mcp.ToolResolveCollectRenderResult,
		mcp.ToolResolveExportProjectArchive,
	)
	for _, tool := range tools {
		if err := mcp.ValidateResolveTool(tool); err != nil {
			return nil, err
		}
		if _, err := client.Call(ctx, mcp.CallRequest{Tool: tool, Timeout: timeoutOrDefault(p.cfg.Timeout), Payload: map[string]any{
			"video_path":    prepared.videoPath,
			"audio_path":    prepared.audioPath,
			"subtitle_path": prepared.subtitlePath,
			"project_dir":   prepared.projectDir,
			"output_dir":    prepared.outputDir,
		}}); err != nil {
			return nil, err
		}
	}
	return tools, nil
}

func (p *Provider) manifest(req providers.RenderRequest, prepared preparedRender, mode string, tools []string) *shortform.ShortRenderManifest {
	outputs := make([]shortform.RenderOutput, 0, len(prepared.platforms))
	for _, platform := range prepared.platforms {
		path := filepath.ToSlash(filepath.Join("renders", platform+".mp4"))
		outputs = append(outputs, shortform.RenderOutput{
			Platform:        platform,
			Path:            path,
			Hash:            deterministicHash(req.EpisodeID, platform, path, "davinci-dry-run"),
			Resolution:      shortform.TargetResolution,
			Aspect:          shortform.TargetAspect,
			FPS:             shortform.TargetFPS,
			VideoCodec:      shortform.TargetVideoCodec,
			AudioCodec:      shortform.TargetAudioCodec,
			AudioTrack:      true,
			SubtitlesBurned: true,
			DurationSec:     prepared.durationSec,
			Status:          shortform.StatusDraft,
		})
	}
	return &shortform.ShortRenderManifest{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     req.EpisodeID,
			ArtifactID:    fmt.Sprintf("%s-%s-davinci-v1", shortform.KindShortRenderManifest, req.EpisodeID),
			CreatedAt:     rfc3339(req.Now),
			CreatedBy:     "system:davinci-resolve-mcp",
			SourceArtifacts: []string{
				req.Shots.ArtifactID,
				req.Voiceover.ArtifactID,
				req.Subtitles.ArtifactID,
			},
			Status: shortform.StatusDraft,
		},
		Renderer: shortform.RendererRef{Name: "davinci_resolve_mcp", Version: versionOrDefault(p.cfg.Version)},
		ProviderMetadata: &shortform.RenderProviderMetadata{
			Provider: "davinci_resolve_mcp",
			Mode:     mode,
			Timeline: shortform.TimelineConfig{Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS},
			MCPTools: tools,
		},
		Inputs:  []string{"visual_shot_manifest.json", "voiceover_manifest.json", "subtitle_manifest.json"},
		Outputs: outputs,
	}
}

func (p *Provider) clientForMode(mode string) (mcp.Client, error) {
	if p.client != nil {
		return p.client, nil
	}
	if mode == ModeDryRun {
		return &mcp.DryRunClient{}, nil
	}
	return nil, fmt.Errorf("DaVinci Resolve local_mcp mode requires an explicit MCP client")
}

func (p *Provider) mode() string {
	if p.cfg.Mode == "" {
		return ModeDisabled
	}
	return p.cfg.Mode
}

func normalizePlatforms(in []string) ([]string, error) {
	if len(in) == 0 {
		in = []string{"master"}
	}
	allowed := map[string]bool{"master": true, "youtube": true, "instagram": true, "tiktok": true}
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, platform := range in {
		if !allowed[platform] {
			return nil, fmt.Errorf("unsupported render platform %q", platform)
		}
		if err := localexec.SafeSegment(platform, "platform"); err != nil {
			return nil, err
		}
		if !seen[platform] {
			out = append(out, platform)
			seen[platform] = true
		}
	}
	return out, nil
}

func deterministicHash(parts ...string) string {
	data, _ := json.Marshal(parts)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return time.Minute
	}
	return timeout
}

func versionOrDefault(version string) string {
	if version == "" {
		return "m3-dry-run"
	}
	return version
}

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

var _ providers.RenderProvider = (*Provider)(nil)
