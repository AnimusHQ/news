// Package subtitles contains local subtitle-generation adapter boundaries. The
// faster-whisper adapter is a sidecar contract, not Python in the core backend.
package subtitles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

const (
	defaultTimeout = 2 * time.Minute
	defaultVersion = "sidecar-contract"
)

// FasterWhisperConfig controls the local sidecar boundary. Enabled must be true
// and all paths must be configured, otherwise the provider fails closed.
type FasterWhisperConfig struct {
	Enabled         bool
	BinaryPath      string
	ScriptPath      string
	InputRoot       string
	OutputRoot      string
	ModelRoot       string
	ModelPath       string
	Timeout         time.Duration
	ProviderVersion string
}

// FasterWhisperProvider executes a configured sidecar command and normalizes its
// JSON response into a SubtitleManifest.
type FasterWhisperProvider struct {
	cfg FasterWhisperConfig
}

// NewFasterWhisperProvider builds a disabled-by-default sidecar provider.
func NewFasterWhisperProvider(cfg FasterWhisperConfig) *FasterWhisperProvider {
	return &FasterWhisperProvider{cfg: cfg}
}

// GenerateSubtitles implements providers.SubtitleProvider.
func (p *FasterWhisperProvider) GenerateSubtitles(ctx context.Context, req providers.SubtitleRequest) (*shortform.SubtitleManifest, error) {
	if !p.cfg.Enabled {
		return nil, fmt.Errorf("faster-whisper subtitle adapter is disabled")
	}
	prepared, err := p.prepare(req)
	if err != nil {
		return nil, err
	}
	response, err := p.runSidecar(ctx, prepared, req.WordTimestampsRequired)
	if err != nil {
		return nil, err
	}
	manifest, err := p.manifestFromResponse(req, prepared, response)
	if err != nil {
		return nil, err
	}
	if req.WordTimestampsRequired && !manifest.Checks.WordTimestamps {
		return nil, fmt.Errorf("faster-whisper sidecar did not provide required word timestamps")
	}
	if err := shortform.Stamp(manifest); err != nil {
		return nil, err
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		return nil, fmt.Errorf("subtitle manifest failed validation: %v", issues)
	}
	return manifest, nil
}

type preparedSubtitle struct {
	binaryPath string
	scriptPath string
	audioPath  string
	outputDir  string
	modelPath  string
	language   string
}

func (p *FasterWhisperProvider) prepare(req providers.SubtitleRequest) (preparedSubtitle, error) {
	if req.EpisodeID == "" {
		return preparedSubtitle{}, fmt.Errorf("episode_id is required")
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return preparedSubtitle{}, err
	}
	if req.Voiceover == nil {
		return preparedSubtitle{}, fmt.Errorf("subtitles require a voiceover manifest")
	}
	if p.cfg.BinaryPath == "" {
		return preparedSubtitle{}, fmt.Errorf("faster-whisper sidecar binary must be configured")
	}
	binaryPath, err := resolveExecutable(p.cfg.BinaryPath, "faster-whisper sidecar binary")
	if err != nil {
		return preparedSubtitle{}, err
	}
	scriptPath := ""
	if p.cfg.ScriptPath != "" {
		scriptPath, err = existingFile(p.cfg.ScriptPath, "faster-whisper sidecar script")
		if err != nil {
			return preparedSubtitle{}, err
		}
	}
	modelPath, err := localexec.ExistingDirUnder(p.cfg.ModelRoot, p.cfg.ModelPath, "faster-whisper model")
	if err != nil {
		return preparedSubtitle{}, err
	}
	audioPath, err := localexec.ExistingFileUnder(p.cfg.InputRoot, req.Voiceover.Output.Path, "voiceover audio")
	if err != nil {
		return preparedSubtitle{}, err
	}
	outputDir, err := localexec.EnsureOutputDir(p.cfg.OutputRoot, req.EpisodeID, "subtitles")
	if err != nil {
		return preparedSubtitle{}, err
	}
	language := req.Language
	if language == "" {
		language = req.Voiceover.Language
	}
	if language == "" {
		language = "en"
	}
	return preparedSubtitle{binaryPath: binaryPath, scriptPath: scriptPath, audioPath: audioPath, outputDir: outputDir, modelPath: modelPath, language: language}, nil
}

func (p *FasterWhisperProvider) runSidecar(parent context.Context, prepared preparedSubtitle, wordTimestampsRequired bool) (sidecarResponse, error) {
	ctx, cancel := context.WithTimeout(parent, timeoutOrDefault(p.cfg.Timeout))
	defer cancel()
	args := []string{}
	if prepared.scriptPath != "" {
		args = append(args, prepared.scriptPath)
	}
	args = append(args,
		"--audio", prepared.audioPath,
		"--output-dir", prepared.outputDir,
		"--language", prepared.language,
		"--model-path", prepared.modelPath,
		"--word-timestamps", strconv.FormatBool(wordTimestampsRequired),
	)
	cmd := exec.CommandContext(ctx, prepared.binaryPath, args...)
	cmd.Dir = prepared.outputDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		return sidecarResponse{}, fmt.Errorf("faster-whisper sidecar timed out after %s: %w", timeoutOrDefault(p.cfg.Timeout), ctx.Err())
	}
	if err != nil {
		detail := localexec.Redact(stderr.String()+stdout.String(), prepared.modelPath, prepared.audioPath)
		if detail == "" {
			detail = err.Error()
		}
		return sidecarResponse{}, fmt.Errorf("faster-whisper sidecar failed: %s", detail)
	}
	var response sidecarResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return sidecarResponse{}, fmt.Errorf("faster-whisper sidecar returned invalid JSON: %w", err)
	}
	return response, nil
}

func (p *FasterWhisperProvider) manifestFromResponse(req providers.SubtitleRequest, prepared preparedSubtitle, response sidecarResponse) (*shortform.SubtitleManifest, error) {
	if response.TranscriptPath == "" || response.SRTPath == "" {
		return nil, fmt.Errorf("faster-whisper sidecar response requires transcript_path and srt_path")
	}
	transcriptPath, err := localexec.ExistingFileUnder(prepared.outputDir, response.TranscriptPath, "transcript output")
	if err != nil {
		return nil, err
	}
	transcriptData, err := os.ReadFile(transcriptPath)
	if err != nil {
		return nil, fmt.Errorf("read transcript output: %w", err)
	}
	if !json.Valid(transcriptData) {
		return nil, fmt.Errorf("transcript output must be JSON")
	}
	transcriptHash, err := localexec.FileSHA256(transcriptPath)
	if err != nil {
		return nil, err
	}
	srtPath, err := localexec.ExistingFileUnder(prepared.outputDir, response.SRTPath, "srt output")
	if err != nil {
		return nil, err
	}
	srtHash, err := localexec.FileSHA256(srtPath)
	if err != nil {
		return nil, err
	}
	assRel := response.ASSPath
	assHash := ""
	if assRel != "" {
		assPath, err := localexec.ExistingFileUnder(prepared.outputDir, assRel, "ass output")
		if err != nil {
			return nil, err
		}
		assHash, err = localexec.FileSHA256(assPath)
		if err != nil {
			return nil, err
		}
	}
	provider := response.Provider
	if provider.Name == "" {
		provider.Name = "faster_whisper"
	}
	if provider.Version == "" {
		provider.Version = versionOrDefault(p.cfg.ProviderVersion)
	}
	language := response.Language
	if language == "" {
		language = prepared.language
	}
	assManifestPath := ""
	if assRel != "" {
		assManifestPath = filepath.ToSlash(filepath.Join("subtitles", filepath.Base(assRel)))
	}
	return &shortform.SubtitleManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       req.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", shortform.KindSubtitleManifest, req.EpisodeID),
			CreatedAt:       rfc3339(req.Now),
			CreatedBy:       "system:faster-whisper-sidecar",
			SourceArtifacts: []string{req.Voiceover.ArtifactID},
			Status:          shortform.StatusDraft,
		},
		Provider:         provider,
		Language:         language,
		TranscriptPath:   filepath.ToSlash(filepath.Join("subtitles", filepath.Base(response.TranscriptPath))),
		TranscriptHash:   transcriptHash,
		SRTPath:          filepath.ToSlash(filepath.Join("subtitles", filepath.Base(response.SRTPath))),
		SRTHash:          srtHash,
		ASSPath:          assManifestPath,
		ASSHash:          assHash,
		Checks:           response.Checks,
		OperatorApproval: false,
	}, nil
}

type sidecarResponse struct {
	Provider       shortform.ProviderRef    `json:"provider"`
	Language       string                   `json:"language"`
	TranscriptPath string                   `json:"transcript_path"`
	SRTPath        string                   `json:"srt_path"`
	ASSPath        string                   `json:"ass_path,omitempty"`
	Checks         shortform.SubtitleChecks `json:"checks"`
}

func resolveExecutable(path, label string) (string, error) {
	if filepath.Base(path) == path {
		resolved, err := exec.LookPath(path)
		if err != nil {
			return "", fmt.Errorf("%s %q not found", label, path)
		}
		return resolved, nil
	}
	return existingFile(path, label)
}

func existingFile(path, label string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s %q not found: %w", label, path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s %q is a directory", label, path)
	}
	return path, nil
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

var _ providers.SubtitleProvider = (*FasterWhisperProvider)(nil)
