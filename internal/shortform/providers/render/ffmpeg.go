// Package render contains local render provider adapters. The FFmpeg adapter is
// disabled by default and must be invoked only from activities or local runners.
package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

const (
	defaultTimeout = 2 * time.Minute
	defaultVersion = "local-ffmpeg"
)

// FFmpegConfig controls the local FFmpeg adapter. Enabled must be true or the
// provider fails closed without inspecting local media.
type FFmpegConfig struct {
	Enabled    bool
	FFmpegPath string
	InputRoot  string
	OutputRoot string
	Timeout    time.Duration
	Version    string
}

// FFmpegProvider normalizes a controlled local media fixture to the short-form
// vertical render target.
type FFmpegProvider struct {
	cfg FFmpegConfig
}

// NewFFmpegProvider builds a disabled-by-default FFmpeg render provider.
func NewFFmpegProvider(cfg FFmpegConfig) *FFmpegProvider {
	return &FFmpegProvider{cfg: cfg}
}

// RenderShort implements providers.RenderProvider with exec.CommandContext.
func (p *FFmpegProvider) RenderShort(ctx context.Context, req providers.RenderRequest) (*shortform.ShortRenderManifest, error) {
	if !p.cfg.Enabled {
		return nil, fmt.Errorf("ffmpeg render adapter is disabled")
	}
	ffmpegPath, err := resolveFFmpeg(p.cfg.FFmpegPath)
	if err != nil {
		return nil, err
	}
	prepared, err := p.prepare(req)
	if err != nil {
		return nil, err
	}

	outputs := make([]shortform.RenderOutput, 0, len(prepared.platforms))
	for _, platform := range prepared.platforms {
		absOut := filepath.Join(prepared.workDir, platform+".mp4")
		cmd := buildFFmpegCommand(ctx, ffmpegPath, prepared.videoPath, prepared.audioPath, absOut, prepared.workDir)
		if err := runFFmpeg(ctx, cmd, timeoutOrDefault(p.cfg.Timeout)); err != nil {
			return nil, err
		}
		hash, err := localexec.FileSHA256(absOut)
		if err != nil {
			return nil, fmt.Errorf("hash ffmpeg output: %w", err)
		}
		outputs = append(outputs, shortform.RenderOutput{
			Platform:        platform,
			Path:            filepath.ToSlash(filepath.Join("renders", platform+".mp4")),
			Hash:            hash,
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

	manifest := &shortform.ShortRenderManifest{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     req.EpisodeID,
			ArtifactID:    fmt.Sprintf("%s-%s-v1", shortform.KindShortRenderManifest, req.EpisodeID),
			CreatedAt:     rfc3339(req.Now),
			CreatedBy:     "system:ffmpeg-local",
			SourceArtifacts: []string{
				req.Shots.ArtifactID,
				req.Voiceover.ArtifactID,
				req.Subtitles.ArtifactID,
			},
			Status: shortform.StatusDraft,
		},
		Renderer: shortform.RendererRef{Name: "ffmpeg", Version: versionOrDefault(p.cfg.Version)},
		Inputs:   []string{prepared.videoRel, prepared.audioRel, prepared.subtitleRel},
		Outputs:  outputs,
	}
	if err := shortform.Stamp(manifest); err != nil {
		return nil, err
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		return nil, fmt.Errorf("ffmpeg render manifest failed validation: %v", issues)
	}
	return manifest, nil
}

type preparedRender struct {
	videoPath    string
	audioPath    string
	subtitleRel  string
	subtitlePath string
	videoRel     string
	audioRel     string
	workDir      string
	platforms    []string
	durationSec  float64
}

func (p *FFmpegProvider) prepare(req providers.RenderRequest) (preparedRender, error) {
	if req.EpisodeID == "" {
		return preparedRender{}, fmt.Errorf("episode_id is required")
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return preparedRender{}, err
	}
	if req.Shots == nil || len(req.Shots.Shots) == 0 {
		return preparedRender{}, fmt.Errorf("ffmpeg render requires at least one visual shot")
	}
	if req.Voiceover == nil {
		return preparedRender{}, fmt.Errorf("ffmpeg render requires a voiceover manifest")
	}
	if req.Subtitles == nil {
		return preparedRender{}, fmt.Errorf("ffmpeg render requires a subtitle manifest")
	}
	videoRel := req.Shots.Shots[0].OutputPath
	audioRel := req.Voiceover.Output.Path
	subtitleRel := req.Subtitles.SRTPath
	if subtitleRel == "" {
		subtitleRel = req.Subtitles.ASSPath
	}
	videoPath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, videoRel, "visual shot")
	if err != nil {
		return preparedRender{}, err
	}
	audioPath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, audioRel, "voiceover")
	if err != nil {
		return preparedRender{}, err
	}
	subtitlePath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, subtitleRel, "subtitles")
	if err != nil {
		return preparedRender{}, err
	}
	workDir, err := localexec.EnsureOutputDir(p.cfg.OutputRoot, req.EpisodeID, "renders")
	if err != nil {
		return preparedRender{}, err
	}
	if err := copyFile(filepath.Join(workDir, "captions.srt"), subtitlePath); err != nil {
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
		videoRel: videoRel, audioRel: audioRel, subtitleRel: subtitleRel,
		workDir: workDir, platforms: platforms, durationSec: duration,
	}, nil
}

func buildFFmpegCommand(ctx context.Context, ffmpegPath, videoPath, audioPath, outputPath, workDir string) *exec.Cmd {
	args := []string{
		"-hide_banner",
		"-loglevel", "warning",
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-map", "0:v:0",
		"-map", "1:a:0",
		"-vf", "subtitles=captions.srt,scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2,fps=30,format=yuv420p",
		"-r", "30",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "veryfast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-shortest",
		"-map_metadata", "-1",
		"-movflags", "+faststart",
		"-threads", "1",
		outputPath,
	}
	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	cmd.Dir = workDir
	return cmd
}

func runFFmpeg(parent context.Context, cmd *exec.Cmd, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	dir := cmd.Dir
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return fmt.Errorf("ffmpeg render timed out after %s: %w", timeout, ctx.Err())
	}
	if err != nil {
		detail := localexec.Redact(stderr.String() + stdout.String())
		if detail == "" {
			detail = err.Error()
		}
		return fmt.Errorf("ffmpeg render failed: %s", detail)
	}
	return nil
}

func resolveFFmpeg(configured string) (string, error) {
	if configured == "" {
		path, err := exec.LookPath("ffmpeg")
		if err != nil {
			return "", fmt.Errorf("ffmpeg binary not found; configure FFmpegPath or install ffmpeg")
		}
		return path, nil
	}
	if filepath.Base(configured) == configured {
		path, err := exec.LookPath(configured)
		if err != nil {
			return "", fmt.Errorf("ffmpeg binary %q not found", configured)
		}
		return path, nil
	}
	info, err := os.Stat(configured)
	if err != nil {
		return "", fmt.Errorf("ffmpeg binary %q not found: %w", configured, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("ffmpeg path %q is a directory", configured)
	}
	return configured, nil
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

func copyFile(dst, src string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read subtitle file: %w", err)
	}
	return os.WriteFile(dst, data, 0o644)
}

func timeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultTimeout
	}
	return timeout
}

func versionOrDefault(version string) string {
	if version == "" {
		return defaultVersion
	}
	return version
}

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

func IsTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

var _ providers.RenderProvider = (*FFmpegProvider)(nil)
