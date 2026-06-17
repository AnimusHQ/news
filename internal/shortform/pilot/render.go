package pilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

func (r Runner) ensureRender(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	manifestPath := filepath.Join(episodeDir, "short_render_manifest.json")
	outputRel := filepath.ToSlash(filepath.Join("dist", manifest.EpisodeID+"-release-candidate.mp4"))
	outputAbs := filepath.Join(episodeDir, filepath.FromSlash(outputRel))
	if fileExists(manifestPath) && fileExists(outputAbs) {
		return nil
	}
	if fileExists(outputAbs) && !fileExists(manifestPath) {
		return fmt.Errorf("refusing to overwrite existing release candidate without regeneration: %s", outputRel)
	}
	if manifest.Providers.Render != "ffmpeg" {
		return fmt.Errorf("unsupported render provider %q; L1 supports ffmpeg", manifest.Providers.Render)
	}
	ffmpegPath, err := resolveFFmpeg()
	if err != nil {
		return err
	}
	ffprobePath, err := resolveFFprobe()
	if err != nil {
		return err
	}
	var visual shortform.VisualShotManifest
	if err := readJSON(filepath.Join(episodeDir, "visual_shot_manifest.json"), &visual); err != nil {
		return err
	}
	var voice shortform.VoiceoverManifest
	if err := readJSON(filepath.Join(episodeDir, "voiceover_manifest.json"), &voice); err != nil {
		return err
	}
	var subtitles shortform.SubtitleManifest
	if err := readJSON(filepath.Join(episodeDir, "subtitle_manifest.json"), &subtitles); err != nil {
		return err
	}
	if len(visual.Shots) == 0 {
		return fmt.Errorf("render requires at least one visual shot")
	}
	if err := os.MkdirAll(filepath.Dir(outputAbs), 0o755); err != nil {
		return err
	}
	videoPaths := make([]string, 0, len(visual.Shots))
	for _, shot := range visual.Shots {
		path, err := localexec.ExistingFileUnder(episodeDir, shot.OutputPath, "visual shot")
		if err != nil {
			return err
		}
		videoPaths = append(videoPaths, path)
	}
	audioPath, err := localexec.ExistingFileUnder(episodeDir, voice.Output.Path, "voiceover")
	if err != nil {
		return err
	}
	if _, err := localexec.ExistingFileUnder(episodeDir, subtitles.SRTPath, "captions"); err != nil {
		return err
	}
	if err := runFFmpegRender(ctx, ffmpegPath, episodeDir, videoPaths, audioPath, outputAbs); err != nil {
		return err
	}
	props, err := ffprobe(ctx, ffprobePath, outputAbs)
	if err != nil {
		return err
	}
	if props.Width != 1080 || props.Height != 1920 {
		return fmt.Errorf("release candidate has invalid resolution %dx%d", props.Width, props.Height)
	}
	if !props.HasAudio {
		return fmt.Errorf("release candidate has no audio stream")
	}
	hash, err := localexec.FileSHA256(outputAbs)
	if err != nil {
		return err
	}
	renderManifest := &shortform.ShortRenderManifest{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     manifest.EpisodeID,
			ArtifactID:    fmt.Sprintf("%s-%s-v1", shortform.KindShortRenderManifest, manifest.EpisodeID),
			CreatedAt:     r.now().Format(time.RFC3339),
			CreatedBy:     "system:ffmpeg-l1-real-pilot",
			SourceArtifacts: []string{
				"visual_shot_manifest.json",
				"voiceover_manifest.json",
				"subtitle_manifest.json",
			},
			Status: shortform.StatusInReview,
		},
		Renderer: shortform.RendererRef{Name: "ffmpeg", Version: ffmpegPath},
		ProviderMetadata: &shortform.RenderProviderMetadata{
			Provider: "ffmpeg",
			Mode:     "local_exec",
			Timeline: shortform.TimelineConfig{Resolution: shortform.TargetResolution, Aspect: shortform.TargetAspect, FPS: shortform.TargetFPS},
		},
		Inputs: []string{
			"visual_shot_manifest.json",
			"voiceover_manifest.json",
			"subtitle_manifest.json",
		},
		Outputs: []shortform.RenderOutput{{
			Platform:        "master",
			Path:            outputRel,
			Hash:            hash,
			Resolution:      shortform.TargetResolution,
			Aspect:          shortform.TargetAspect,
			FPS:             30,
			VideoCodec:      shortform.TargetVideoCodec,
			AudioCodec:      shortform.TargetAudioCodec,
			AudioTrack:      true,
			SubtitlesBurned: true,
			DurationSec:     props.DurationSec,
			Status:          shortform.StatusInReview,
		}},
	}
	if err := shortform.Stamp(renderManifest); err != nil {
		return err
	}
	if issues := shortform.Validate(renderManifest); len(issues) > 0 {
		return fmt.Errorf("short_render_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(manifestPath, renderManifest); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageRender, "ffmpeg rendered release_candidate MP4 and ffprobe checks passed")
}

func resolveFFmpeg() (string, error) {
	return resolveBinaryEnv("ANIMUS_FFMPEG_BINARY", "ffmpeg")
}

func resolveFFprobe() (string, error) {
	return resolveBinaryEnv("ANIMUS_FFPROBE_BINARY", "ffprobe")
}

func resolveBinaryEnv(envKey, fallback string) (string, error) {
	configured := strings.TrimSpace(os.Getenv(envKey))
	if configured == "" {
		configured = fallback
	}
	if filepath.Base(configured) == configured {
		path, err := exec.LookPath(configured)
		if err != nil {
			return "", fmt.Errorf("%s binary %q not found", fallback, configured)
		}
		return path, nil
	}
	info, err := os.Stat(configured)
	if err != nil {
		return "", fmt.Errorf("%s binary %q not found: %w", fallback, configured, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s binary %q is a directory", fallback, configured)
	}
	return configured, nil
}

func runFFmpegRender(ctx context.Context, ffmpegPath, episodeDir string, videoPaths []string, audioPath, outputPath string) error {
	timeout := parseEnvTimeout("ANIMUS_FFMPEG_TIMEOUT", 3*time.Minute)
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	args := []string{"-hide_banner", "-loglevel", "warning", "-y"}
	for _, path := range videoPaths {
		args = append(args, "-i", path)
	}
	args = append(args, "-i", audioPath)
	filter := buildFilterComplex(len(videoPaths))
	args = append(args,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", fmt.Sprintf("%d:a:0", len(videoPaths)),
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
	)
	cmd := exec.CommandContext(runCtx, ffmpegPath, args...)
	cmd.Dir = episodeDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); runCtx.Err() != nil {
		return fmt.Errorf("ffmpeg render timed out after %s", timeout)
	} else if err != nil {
		return fmt.Errorf("ffmpeg render failed: %s", localexec.Redact(stderr.String()))
	}
	return nil
}

func buildFilterComplex(videoCount int) string {
	var b strings.Builder
	for i := 0; i < videoCount; i++ {
		fmt.Fprintf(&b, "[%d:v]scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2,fps=30,setsar=1[v%d];", i, i)
	}
	for i := 0; i < videoCount; i++ {
		fmt.Fprintf(&b, "[v%d]", i)
	}
	fmt.Fprintf(&b, "concat=n=%d:v=1:a=0[vcat];", videoCount)
	b.WriteString("[vcat]subtitles=subtitles/captions.srt[vout]")
	return b.String()
}

type probeResult struct {
	Width       int
	Height      int
	FPS         float64
	DurationSec float64
	HasAudio    bool
}

func ffprobe(ctx context.Context, ffprobePath, mediaPath string) (probeResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(runCtx, ffprobePath, "-v", "error", "-show_streams", "-show_format", "-of", "json", mediaPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if runCtx.Err() != nil {
		return probeResult{}, fmt.Errorf("ffprobe timed out")
	}
	if err != nil {
		return probeResult{}, fmt.Errorf("ffprobe failed: %s", localexec.Redact(stderr.String()))
	}
	var parsed struct {
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(stdout, &parsed); err != nil {
		return probeResult{}, err
	}
	var res probeResult
	for _, stream := range parsed.Streams {
		switch stream.CodecType {
		case "video":
			res.Width = stream.Width
			res.Height = stream.Height
			res.FPS = parseRate(stream.RFrameRate)
		case "audio":
			res.HasAudio = true
		}
	}
	if parsed.Format.Duration != "" {
		if d, err := strconv.ParseFloat(parsed.Format.Duration, 64); err == nil {
			res.DurationSec = d
		}
	}
	if res.Width == 0 || res.Height == 0 {
		return probeResult{}, fmt.Errorf("ffprobe found no video stream")
	}
	if res.FPS > 0 && (res.FPS < 29.5 || res.FPS > 30.5) {
		return probeResult{}, fmt.Errorf("release candidate has invalid fps %.2f", res.FPS)
	}
	return res, nil
}

func parseRate(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		if den != 0 {
			return num / den
		}
	}
	out, _ := strconv.ParseFloat(rate, 64)
	return out
}
