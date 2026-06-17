// Package omnivoice contains the optional OmniVoice voice provider boundary.
// It is disabled by default and emits draft voiceover manifests only.
package omnivoice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	ModeDisabled     = "disabled"
	ModeDryRun       = "dry_run"
	ModeLocalSidecar = "local_sidecar"
)

// OmniVoiceConfig controls the optional local/multilingual voice lane.
type OmniVoiceConfig struct {
	Enabled        bool
	Mode           string
	BinaryPath     string
	ModelRoot      string
	ModelPath      string
	InputRoot      string
	OutputRoot     string
	Timeout        time.Duration
	RequireConsent bool
	Version        string
}

// Redacted returns a copy safe for diagnostics.
func (c OmniVoiceConfig) Redacted() OmniVoiceConfig {
	if c.BinaryPath != "" {
		c.BinaryPath = "[REDACTED]"
	}
	if c.ModelRoot != "" {
		c.ModelRoot = "[REDACTED]"
	}
	if c.ModelPath != "" {
		c.ModelPath = "[REDACTED]"
	}
	if c.InputRoot != "" {
		c.InputRoot = "[REDACTED]"
	}
	if c.OutputRoot != "" {
		c.OutputRoot = "[REDACTED]"
	}
	return c
}

// Provider implements providers.VoiceProvider.
type Provider struct {
	cfg OmniVoiceConfig
}

// NewProvider builds a disabled-by-default OmniVoice provider.
func NewProvider(cfg OmniVoiceConfig) *Provider {
	return &Provider{cfg: cfg}
}

// SynthesizeVoiceover implements providers.VoiceProvider.
func (p *Provider) SynthesizeVoiceover(ctx context.Context, req providers.VoiceoverRequest) (*shortform.VoiceoverManifest, error) {
	if !p.cfg.Enabled {
		return nil, fmt.Errorf("OmniVoice adapter is disabled")
	}
	mode := p.mode()
	if mode != ModeDryRun && mode != ModeLocalSidecar {
		return nil, fmt.Errorf("OmniVoice mode %q is not enabled", mode)
	}
	prepared, err := p.prepare(req)
	if err != nil {
		return nil, err
	}
	var manifest *shortform.VoiceoverManifest
	switch mode {
	case ModeDryRun:
		manifest = p.dryRunManifest(req, prepared)
	case ModeLocalSidecar:
		manifest, err = p.sidecarManifest(ctx, req, prepared)
		if err != nil {
			return nil, err
		}
	}
	if err := shortform.Stamp(manifest); err != nil {
		return nil, err
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		return nil, fmt.Errorf("OmniVoice voiceover manifest failed validation: %v", issues)
	}
	return manifest, nil
}

type preparedVoice struct {
	binaryPath string
	modelPath  string
	outputDir  string
	language   string
}

func (p *Provider) prepare(req providers.VoiceoverRequest) (preparedVoice, error) {
	if req.EpisodeID == "" {
		return preparedVoice{}, fmt.Errorf("episode_id is required")
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return preparedVoice{}, err
	}
	if req.ScriptRef == "" {
		return preparedVoice{}, fmt.Errorf("OmniVoice requires a source script reference")
	}
	binaryPath, err := existingFile(p.cfg.BinaryPath, "OmniVoice binary")
	if err != nil {
		return preparedVoice{}, err
	}
	modelPath, err := localexec.ExistingDirUnder(p.cfg.ModelRoot, p.cfg.ModelPath, "OmniVoice model")
	if err != nil {
		return preparedVoice{}, err
	}
	if err := p.validateConsent(req); err != nil {
		return preparedVoice{}, err
	}
	if req.ReferenceAudioPath != "" {
		if _, err := localexec.ExistingFileUnder(p.cfg.InputRoot, req.ReferenceAudioPath, "OmniVoice reference audio"); err != nil {
			return preparedVoice{}, err
		}
	}
	outputDir, err := localexec.EnsureOutputDir(p.cfg.OutputRoot, req.EpisodeID, "voice")
	if err != nil {
		return preparedVoice{}, err
	}
	language := req.Language
	if language == "" {
		language = "en"
	}
	return preparedVoice{binaryPath: binaryPath, modelPath: modelPath, outputDir: outputDir, language: language}, nil
}

func (p *Provider) validateConsent(req providers.VoiceoverRequest) error {
	usesReferenceVoice := req.ReferenceAudioPath != "" || req.VoicePromptReference != ""
	if !usesReferenceVoice {
		return nil
	}
	if p.cfg.RequireConsent || req.VoiceConsentRequired {
		if req.VoiceConsentReference == "" {
			return fmt.Errorf("voice consent reference is required for OmniVoice reference voice use")
		}
		if req.ReferenceAudioPath != "" && !req.ReferenceAudioAllowed {
			return fmt.Errorf("reference audio must be explicitly allowed")
		}
		if req.ReferenceAudioPath != "" && req.ReferenceAudioHash == "" {
			return fmt.Errorf("reference audio hash must be recorded")
		}
	}
	return nil
}

func (p *Provider) dryRunManifest(req providers.VoiceoverRequest, prepared preparedVoice) *shortform.VoiceoverManifest {
	path := "voice/omnivoice.wav"
	return p.baseManifest(req, prepared, shortform.MediaOutput{
		Path: path, Hash: deterministicHash(req.EpisodeID, path, prepared.language, "omnivoice-dry-run"),
		DurationSec: 1, Format: "wav", SampleRateHz: 24000,
	})
}

func (p *Provider) sidecarManifest(ctx context.Context, req providers.VoiceoverRequest, prepared preparedVoice) (*shortform.VoiceoverManifest, error) {
	parent, cancel := context.WithTimeout(ctx, timeoutOrDefault(p.cfg.Timeout))
	defer cancel()
	args := []string{
		"--model-path", prepared.modelPath,
		"--output-dir", prepared.outputDir,
		"--language", prepared.language,
		"--script-ref", req.ScriptRef,
		"--require-consent", strconv.FormatBool(p.cfg.RequireConsent || req.VoiceConsentRequired),
	}
	cmd := exec.CommandContext(parent, prepared.binaryPath, args...)
	cmd.Dir = prepared.outputDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if parent.Err() != nil {
		return nil, fmt.Errorf("OmniVoice sidecar timed out after %s: %w", timeoutOrDefault(p.cfg.Timeout), parent.Err())
	}
	if err != nil {
		detail := localexec.Redact(stderr.String()+stdout.String(), prepared.modelPath, prepared.outputDir)
		if detail == "" {
			detail = err.Error()
		}
		return nil, fmt.Errorf("OmniVoice sidecar failed: %s", detail)
	}
	var response sidecarResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("OmniVoice sidecar returned invalid JSON: %w", err)
	}
	if response.OutputPath == "" {
		return nil, fmt.Errorf("OmniVoice sidecar response requires output_path")
	}
	outputPath, err := localexec.ExistingFileUnder(prepared.outputDir, response.OutputPath, "OmniVoice output")
	if err != nil {
		return nil, err
	}
	hash, err := localexec.FileSHA256(outputPath)
	if err != nil {
		return nil, fmt.Errorf("hash OmniVoice output: %w", err)
	}
	format := response.Format
	if format == "" {
		format = filepath.Ext(response.OutputPath)
		if len(format) > 0 {
			format = format[1:]
		}
	}
	if format == "" {
		format = "wav"
	}
	duration := response.DurationSec
	if duration <= 0 {
		duration = 1
	}
	sampleRate := response.SampleRateHz
	if sampleRate <= 0 {
		sampleRate = 24000
	}
	return p.baseManifest(req, prepared, shortform.MediaOutput{
		Path: filepath.ToSlash(filepath.Join("voice", filepath.Base(response.OutputPath))),
		Hash: hash, DurationSec: duration, Format: format, SampleRateHz: sampleRate,
	}), nil
}

func (p *Provider) baseManifest(req providers.VoiceoverRequest, prepared preparedVoice, output shortform.MediaOutput) *shortform.VoiceoverManifest {
	return &shortform.VoiceoverManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       req.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-omnivoice-v1", shortform.KindVoiceoverManifest, req.EpisodeID),
			CreatedAt:       rfc3339(req.Now),
			CreatedBy:       "system:omnivoice",
			SourceArtifacts: []string{req.ScriptRef},
			Status:          shortform.StatusDraft,
		},
		Provider: shortform.ProviderRef{
			Name: "omnivoice", Model: filepath.Base(prepared.modelPath), Version: versionOrDefault(p.cfg.Version),
		},
		SourceScriptRef:       req.ScriptRef,
		Language:              prepared.language,
		VoicePromptReference:  req.VoicePromptReference,
		VoiceConsentReference: req.VoiceConsentReference,
		Output:                output,
		OperatorApproval:      false,
	}
}

type sidecarResponse struct {
	OutputPath   string  `json:"output_path"`
	DurationSec  float64 `json:"duration_sec"`
	Format       string  `json:"format"`
	SampleRateHz int     `json:"sample_rate_hz"`
}

func existingFile(path, label string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s must be configured", label)
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s %q not found: %w", label, path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s %q is a directory", label, path)
	}
	return path, nil
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

func (p *Provider) mode() string {
	if p.cfg.Mode == "" {
		return ModeDisabled
	}
	return p.cfg.Mode
}

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

var _ providers.VoiceProvider = (*Provider)(nil)
