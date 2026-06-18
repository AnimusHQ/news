package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

type externalConfig struct {
	Command    string
	InputRoot  string
	OutputRoot string
	Timeout    time.Duration
}

func writeTopicYAML(episodeDir string, manifest EpisodeManifest) error {
	body := fmt.Sprintf(`schema_version: "1.0"
episode_id: "%s"
title_working: "Runtime pilot %s"
status: "generated"
language: "%s"
duration: "%s"
platforms:
%s
prompt: %q
`, manifest.EpisodeID, manifest.EpisodeID, manifest.Language, manifest.Duration, yamlList(manifest.Platforms), manifest.OriginalPrompt)
	return writeText(filepath.Join(episodeDir, "topic.yaml"), body)
}

func writePilotResearchPack(episodeDir string, manifest EpisodeManifest) error {
	promptHash, err := textHash(manifest.OriginalPrompt)
	if err != nil {
		return err
	}
	pack := map[string]any{
		"schema_version":  SchemaVersion,
		"episode_id":      manifest.EpisodeID,
		"artifact_id":     "research-pack-" + manifest.EpisodeID + "-l1-prompt-v1",
		"created_at":      manifest.CreatedAt,
		"created_by":      "system:l1-pilot",
		"status":          StatusGenerated,
		"scope":           "operator_prompt_only",
		"core_question":   manifest.OriginalPrompt,
		"source_boundary": "L1 pilot uses the operator prompt as source state and must not claim full external source-grounded finality.",
		"sources": []map[string]any{{
			"source_id":    "operator-prompt-001",
			"title":        "Operator prompt",
			"type":         "operator_prompt",
			"trust_level":  "operator_intent",
			"content_hash": promptHash,
		}},
		"blocking_limitations": []string{
			"No external source ingestion is performed in L1.",
			"Public publishing still requires future source-grounded release workflow.",
		},
	}
	return writeJSON(filepath.Join(episodeDir, "research_pack.json"), pack)
}

func yamlList(values []string) string {
	if len(values) == 0 {
		return "  []"
	}
	var b strings.Builder
	for _, value := range values {
		fmt.Fprintf(&b, "  - %s\n", value)
	}
	return strings.TrimRight(b.String(), "\n")
}

func buildScript(manifest EpisodeManifest) string {
	prompt := strings.TrimSpace(manifest.OriginalPrompt)
	if prompt == "" {
		return ""
	}
	lang := strings.ToLower(strings.TrimSpace(manifest.Language))
	if lang == "ru" || strings.HasPrefix(lang, "ru-") {
		return fmt.Sprintf(`# Script

## Hook
%s

## Beats
1. Обозначить вопрос или тезис из runtime prompt.
2. Дать короткий контекст без внешних фактов, если источники не приложены.
3. Сформулировать практический вывод, явно связанный с prompt.
4. Завершить без hardcoded CTA, если оператор не передал CTA отдельным входом.

## Voiceover
%s

Это черновой voiceover для runtime prompt. Он должен оставаться нейтральным и не добавлять тему, аудиторию, CTA или факты, которых нет во входном prompt.

Сначала зафиксируй главный вопрос. Затем объясни контекст простыми словами. После этого покажи, что зрителю нужно понять или проверить дальше. Если для темы нужны источники, не выдавай этот L1 pilot за source-grounded финальный материал.

## Visual Notes
- Vertical 9:16 educational scenes derived only from the runtime prompt.
- Use neutral explanatory visuals, abstract context, and readable pacing.
- No fake readable UI text, no real brand logos, no misleading screenshots.

## CTA
No CTA was supplied as runtime input. Do not invent one.

## Estimated Timing
Target: %s.

## AI Disclosure Note
AI-assisted script and media generation may be used. Final publishing requires human QA, source/provenance review, and disclosure review.
`, prompt, prompt, manifest.Duration)
	}
	return fmt.Sprintf(`# Script

## Hook
%s

## Beats
1. State the question or thesis from the runtime prompt.
2. Give brief context without adding external facts when sources are not attached.
3. Name the practical takeaway that follows from the prompt.
4. Close without a hardcoded CTA when the operator did not provide one.

## Voiceover
%s

This is a draft voiceover for the runtime prompt. It must stay neutral and avoid adding a topic, audience, CTA, or factual claim that was not supplied as runtime content input.

Start by naming the central question. Explain the context in simple language. Then show what the viewer should understand or verify next. If the topic needs sources, do not present this L1 pilot as source-grounded final material.

## Visual Notes
- Vertical 9:16 educational scenes derived only from the runtime prompt.
- Use neutral explanatory visuals, abstract context, and readable pacing.
- No fake readable UI text, no real brand logos, no misleading screenshots.

## CTA
No CTA was supplied as runtime input. Do not invent one.

## Estimated Timing
Target: %s.

## AI Disclosure Note
AI-assisted script and media generation may be used. Final publishing requires human QA, source/provenance review, and disclosure review.
`, prompt, prompt, manifest.Duration)
}

func buildShotRequests(manifest EpisodeManifest) []VisualShotRequest {
	total := manifest.DurationSec
	if total <= 0 {
		total = 45
	}
	durations := []float64{round1(total * 0.34), round1(total * 0.33), round1(total * 0.33)}
	durations[2] = round1(total - durations[0] - durations[1])
	prefix := "Vertical 9:16 educational video derived from the runtime prompt"
	if strings.EqualFold(manifest.Language, "ru") {
		prefix = "Вертикальное 9:16 образовательное видео по runtime prompt"
	}
	return []VisualShotRequest{
		{
			ShotID: "shot-001", SceneID: "scene-001", DurationSec: durations[0],
			Prompt:            fmt.Sprintf("%s. Opening visual for this prompt only: %s. Use neutral abstract imagery and no readable fake UI text.", prefix, manifest.OriginalPrompt),
			NegativePrompt:    "watermark, distorted text, unreadable UI, broken hands, brand logos, fake screenshots",
			Width:             1080,
			Height:            1920,
			FPS:               30,
			Camera:            "slow push-in",
			Motion:            "subtle cinematic movement",
			SourceScriptLines: []string{"Hook", "Beat 1"},
		},
		{
			ShotID: "shot-002", SceneID: "scene-002", DurationSec: durations[1],
			Prompt:            fmt.Sprintf("Vertical 9:16 educational middle scene expanding the runtime prompt with abstract context, relationships, and motion. Prompt: %s", manifest.OriginalPrompt),
			NegativePrompt:    "watermark, distorted text, readable fake UI, broken hands, brand logos, social media logos",
			Width:             1080,
			Height:            1920,
			FPS:               30,
			Camera:            "smooth lateral move",
			Motion:            "clean explanatory motion",
			SourceScriptLines: []string{"Beat 2", "Beat 3"},
		},
		{
			ShotID: "shot-003", SceneID: "scene-003", DurationSec: durations[2],
			Prompt:            fmt.Sprintf("Vertical 9:16 closing scene for the runtime prompt, focused on a clear generic takeaway without adding a new topic. Prompt: %s", manifest.OriginalPrompt),
			NegativePrompt:    "watermark, distorted text, unreadable UI, broken hands, brand logos, fake terminal text",
			Width:             1080,
			Height:            1920,
			FPS:               30,
			Camera:            "slow pull-back",
			Motion:            "subtle cinematic movement",
			SourceScriptLines: []string{"Takeaway", "CTA"},
		},
	}
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func externalCommandConfig(prefix string) (externalConfig, []string) {
	commandKey := prefix + "_COMMAND"
	inputRootKey := prefix + "_INPUT_ROOT"
	outputRootKey := prefix + "_OUTPUT_ROOT"
	timeoutKey := prefix + "_TIMEOUT"
	cfg := externalConfig{
		Command:    strings.TrimSpace(os.Getenv(commandKey)),
		InputRoot:  strings.TrimSpace(os.Getenv(inputRootKey)),
		OutputRoot: strings.TrimSpace(os.Getenv(outputRootKey)),
		Timeout:    parseEnvTimeout(timeoutKey, 2*time.Minute),
	}
	var missing []string
	if cfg.Command == "" {
		missing = append(missing, commandKey)
	}
	if cfg.InputRoot == "" {
		missing = append(missing, inputRootKey)
	}
	if cfg.OutputRoot == "" {
		missing = append(missing, outputRootKey)
	}
	return cfg, missing
}

func parseEnvTimeout(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if d, err := time.ParseDuration(value); err == nil {
		return d
	}
	return fallback
}

func runExternalJSON(ctx context.Context, cfg externalConfig, input any, output any) error {
	command, err := resolveCommand(cfg.Command)
	if err != nil {
		return err
	}
	if _, err := localexec.ExistingDirUnder(cfg.InputRoot, ".", "external input root"); err != nil {
		return err
	}
	if _, err := localexec.ExistingDirUnder(cfg.OutputRoot, ".", "external output root"); err != nil {
		return err
	}
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, command)
	cmd.Stdin = strings.NewReader(string(data))
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if runCtx.Err() != nil {
		return fmt.Errorf("external command timed out after %s", timeout)
	}
	if err != nil {
		return fmt.Errorf("external command failed: %s", localexec.Redact(stderr.String()))
	}
	if err := json.Unmarshal(stdout, output); err != nil {
		return fmt.Errorf("external command returned invalid JSON: %w", err)
	}
	return nil
}

func resolveCommand(command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("external command is not configured")
	}
	if filepath.Base(command) == command {
		resolved, err := exec.LookPath(command)
		if err != nil {
			return "", fmt.Errorf("external command %q not found", command)
		}
		return resolved, nil
	}
	info, err := os.Stat(command)
	if err != nil {
		return "", fmt.Errorf("external command %q not found: %w", command, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("external command %q is a directory", command)
	}
	return command, nil
}

func normalizeVisualResponse(episodeDir string, requests VisualShotRequests, response ExternalVisualResponse) ([]shortform.VisualShot, string, error) {
	if response.SchemaVersion != SchemaVersion || response.EpisodeID != requests.EpisodeID {
		return nil, "", fmt.Errorf("visual provider response has invalid schema_version or episode_id")
	}
	requestByID := map[string]VisualShotRequest{}
	for _, shot := range requests.Shots {
		requestByID[shot.ShotID] = shot
	}
	responseByID := map[string]ExternalVisualOutput{}
	for _, shot := range response.Shots {
		if _, ok := requestByID[shot.ShotID]; !ok {
			return nil, "", fmt.Errorf("visual provider returned unknown shot_id %q", shot.ShotID)
		}
		responseByID[shot.ShotID] = shot
	}
	if len(responseByID) != len(requestByID) {
		return nil, "", fmt.Errorf("visual provider returned %d shot(s), expected %d", len(responseByID), len(requestByID))
	}
	out := make([]shortform.VisualShot, 0, len(requests.Shots))
	for _, req := range requests.Shots {
		got, ok := responseByID[req.ShotID]
		if !ok {
			return nil, "", fmt.Errorf("visual provider missing shot_id %q", req.ShotID)
		}
		if got.Status != StatusGenerated && got.Status != "generated" {
			return nil, "", fmt.Errorf("visual shot %s status must be generated", req.ShotID)
		}
		abs, err := localexec.ExistingFileUnder(episodeDir, got.OutputPath, "visual output")
		if err != nil {
			return nil, "", err
		}
		hash, err := localexec.FileSHA256(abs)
		if err != nil {
			return nil, "", err
		}
		if got.OutputHash != "" && got.OutputHash != hash {
			return nil, "", fmt.Errorf("visual shot %s hash mismatch: provider=%s actual=%s", req.ShotID, got.OutputHash, hash)
		}
		if got.Width != 1080 || got.Height != 1920 || got.FPS != 30 {
			return nil, "", fmt.Errorf("visual shot %s has invalid properties: %dx%d fps=%d", req.ShotID, got.Width, got.Height, got.FPS)
		}
		rel, err := filepath.Rel(episodeDir, abs)
		if err != nil {
			return nil, "", err
		}
		out = append(out, shortform.VisualShot{
			SceneID:            req.SceneID,
			Prompt:             req.Prompt,
			NegativePrompt:     req.NegativePrompt,
			ReferenceImageHash: requests.SourceScriptHash,
			OutputPath:         filepath.ToSlash(rel),
			OutputHash:         hash,
			DurationSec:        got.DurationSec,
			Camera:             req.Camera,
			Style:              req.Motion,
			Status:             shortform.StatusInReview,
			OperatorApproval:   false,
		})
	}
	provider := strings.TrimSpace(response.Provider)
	if provider == "" {
		provider = "external-command"
	}
	return out, provider, nil
}

func normalizeVoiceResponse(episodeDir string, manifest EpisodeManifest, response ExternalVoiceResponse) (*shortform.VoiceoverManifest, error) {
	if response.SchemaVersion != SchemaVersion || response.EpisodeID != manifest.EpisodeID {
		return nil, fmt.Errorf("voice provider response has invalid schema_version or episode_id")
	}
	abs, err := localexec.ExistingFileUnder(episodeDir, response.OutputPath, "voice output")
	if err != nil {
		return nil, err
	}
	hash, err := localexec.FileSHA256(abs)
	if err != nil {
		return nil, err
	}
	if response.OutputHash != "" && response.OutputHash != hash {
		return nil, fmt.Errorf("voice output hash mismatch: provider=%s actual=%s", response.OutputHash, hash)
	}
	rel, err := filepath.Rel(episodeDir, abs)
	if err != nil {
		return nil, err
	}
	provider := strings.TrimSpace(response.Provider)
	if provider == "" {
		provider = "external-command"
	}
	return &shortform.VoiceoverManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       manifest.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", shortform.KindVoiceoverManifest, manifest.EpisodeID),
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			CreatedBy:       "system:external-command-voice",
			SourceArtifacts: []string{"script.md", "claude_script_review_response.json"},
			Status:          shortform.StatusInReview,
		},
		Provider:              shortform.ProviderRef{Name: provider},
		SourceScriptRef:       "script.md",
		Language:              manifest.Language,
		VoiceConsentReference: response.VoiceConsentReference,
		Output: shortform.MediaOutput{
			Path:         filepath.ToSlash(rel),
			Hash:         hash,
			DurationSec:  response.DurationSec,
			Format:       strings.TrimPrefix(strings.ToLower(filepath.Ext(abs)), "."),
			SampleRateHz: response.SampleRate,
		},
		OperatorApproval: false,
	}, nil
}

func normalizeSubtitleResponse(episodeDir string, manifest EpisodeManifest, response SubtitleSidecarResponse) (*shortform.SubtitleManifest, error) {
	if response.SchemaVersion != SchemaVersion || response.EpisodeID != manifest.EpisodeID {
		return nil, fmt.Errorf("subtitle provider response has invalid schema_version or episode_id")
	}
	if response.TranscriptPath == "" || response.SRTPath == "" {
		return nil, fmt.Errorf("subtitle response requires transcript_path and srt_path")
	}
	return manifestForSubtitleFiles(episodeDir, manifest, defaultText(response.Provider, "faster-whisper"))
}

func manifestForSubtitleFiles(episodeDir string, manifest EpisodeManifest, provider string) (*shortform.SubtitleManifest, error) {
	transcript := filepath.Join("subtitles", "transcript.json")
	srt := filepath.Join("subtitles", "captions.srt")
	ass := filepath.Join("subtitles", "captions.ass")
	transcriptAbs, err := localexec.ExistingFileUnder(episodeDir, transcript, "transcript")
	if err != nil {
		return nil, err
	}
	if data, err := os.ReadFile(transcriptAbs); err != nil || !json.Valid(data) {
		return nil, fmt.Errorf("transcript.json must be valid JSON")
	}
	srtAbs, err := localexec.ExistingFileUnder(episodeDir, srt, "srt")
	if err != nil {
		return nil, err
	}
	assAbs, err := localexec.ExistingFileUnder(episodeDir, ass, "ass")
	if err != nil {
		return nil, err
	}
	transcriptHash, _ := localexec.FileSHA256(transcriptAbs)
	srtHash, _ := localexec.FileSHA256(srtAbs)
	assHash, _ := localexec.FileSHA256(assAbs)
	return &shortform.SubtitleManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       manifest.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", shortform.KindSubtitleManifest, manifest.EpisodeID),
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			CreatedBy:       "system:" + provider,
			SourceArtifacts: []string{"voiceover_manifest.json"},
			Status:          shortform.StatusInReview,
		},
		Provider:       shortform.ProviderRef{Name: provider},
		Language:       manifest.Language,
		TranscriptPath: filepath.ToSlash(transcript),
		TranscriptHash: transcriptHash,
		SRTPath:        filepath.ToSlash(srt),
		SRTHash:        srtHash,
		ASSPath:        filepath.ToSlash(ass),
		ASSHash:        assHash,
		Checks: shortform.SubtitleChecks{
			WordTimestamps: true,
			SafeZone:       true,
			Sync:           true,
		},
		OperatorApproval: false,
	}, nil
}

func voiceoverText(scriptPath string) (string, error) {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}
	text := string(data)
	marker := "## Voiceover"
	idx := strings.Index(text, marker)
	if idx == -1 {
		return strings.TrimSpace(text), nil
	}
	rest := text[idx+len(marker):]
	next := strings.Index(rest, "\n## ")
	if next >= 0 {
		rest = rest[:next]
	}
	lines := strings.Split(rest, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n\n")), nil
}

func srtTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	totalMillis := int(seconds * 1000)
	ms := totalMillis % 1000
	totalSeconds := totalMillis / 1000
	s := totalSeconds % 60
	totalMinutes := totalSeconds / 60
	m := totalMinutes % 60
	h := totalMinutes / 60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func assSubtitle(text string, duration float64) string {
	return fmt.Sprintf(`[Script Info]
ScriptType: v4.00+
PlayResX: 1080
PlayResY: 1920

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,58,&H00FFFFFF,&H000000FF,&H64000000,&H96000000,1,0,0,0,100,100,0,0,1,4,1,2,80,80,180,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:00.00,%s,Default,,0,0,0,,%s
`, assTimestamp(duration), escapeASS(text))
}

func assTimestamp(seconds float64) string {
	totalCentis := int(seconds * 100)
	cs := totalCentis % 100
	totalSeconds := totalCentis / 100
	s := totalSeconds % 60
	totalMinutes := totalSeconds / 60
	m := totalMinutes % 60
	h := totalMinutes / 60
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs)
}

func escapeASS(text string) string {
	text = strings.ReplaceAll(text, "\n", `\N`)
	text = strings.ReplaceAll(text, "{", "")
	text = strings.ReplaceAll(text, "}", "")
	return text
}

func defaultText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func loadVoiceoverManifest(episodeDir string) (*shortform.VoiceoverManifest, error) {
	var manifest shortform.VoiceoverManifest
	if err := readJSON(filepath.Join(episodeDir, "voiceover_manifest.json"), &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func loadRenderManifest(episodeDir string) (*shortform.ShortRenderManifest, error) {
	var manifest shortform.ShortRenderManifest
	if err := readJSON(filepath.Join(episodeDir, "short_render_manifest.json"), &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (r Runner) appendAudit(episodeDir, event, message string) error {
	entry := map[string]string{
		"ts":      r.now().Format(time.RFC3339),
		"event":   event,
		"message": message,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(episodeDir, "audit.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func validateEpisodeManifest(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	if manifest.SchemaVersion != SchemaVersion {
		issues = append(issues, "episode_manifest.json schema_version must be 1.0")
	}
	if err := localexec.SafeSegment(manifest.EpisodeID, "episode_id"); err != nil {
		issues = append(issues, err.Error())
	}
	if manifest.OriginalPrompt == "" {
		issues = append(issues, "episode_manifest.json original_prompt is required")
	}
	if manifest.ContentHash != "" {
		copy := manifest
		hash := copy.ContentHash
		copy.ContentHash = ""
		if err := stampPilotArtifact(&copy); err != nil || copy.ContentHash != hash {
			issues = append(issues, "episode_manifest.json content_hash mismatch")
		}
	}
	if !fileExists(filepath.Join(episodeDir, "audit.log")) {
		issues = append(issues, "audit.log is missing")
	}
	return issues
}

func validateScriptArtifacts(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var sm ScriptManifest
	if err := readJSON(filepath.Join(episodeDir, "script_manifest.json"), &sm); err != nil {
		return []string{err.Error()}
	}
	hash, err := localexec.FileSHA256(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return []string{err.Error()}
	}
	if sm.EpisodeID != manifest.EpisodeID || sm.ScriptHash != hash {
		issues = append(issues, "script_manifest.json does not match current script.md")
	}
	return issues
}

func validateScriptReview(episodeDir string, manifest EpisodeManifest) []string {
	ok, issue, err := Runner{}.scriptReviewPassed(episodeDir, manifest)
	if err != nil {
		return []string{err.Error()}
	}
	if !ok {
		return []string{issue}
	}
	return nil
}

func validateVisualArtifacts(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var m shortform.VisualShotManifest
	if err := readJSON(filepath.Join(episodeDir, "visual_shot_manifest.json"), &m); err != nil {
		return []string{err.Error()}
	}
	if got := shortform.Validate(&m); len(got) > 0 {
		issues = append(issues, got...)
	}
	var requests VisualShotRequests
	if err := readJSON(filepath.Join(episodeDir, "visual_shot_requests.json"), &requests); err != nil {
		issues = append(issues, err.Error())
		return issues
	}
	if len(m.Shots) != len(requests.Shots) {
		issues = append(issues, "visual_shot_manifest.json shot count does not match visual_shot_requests.json")
	}
	for _, shot := range m.Shots {
		abs, err := localexec.ExistingFileUnder(episodeDir, shot.OutputPath, "visual output")
		if err != nil {
			issues = append(issues, err.Error())
			continue
		}
		hash, err := localexec.FileSHA256(abs)
		if err != nil || hash != shot.OutputHash {
			issues = append(issues, shot.OutputPath+": visual output hash mismatch")
		}
	}
	return issues
}

func validateVoiceArtifacts(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var m shortform.VoiceoverManifest
	if err := readJSON(filepath.Join(episodeDir, "voiceover_manifest.json"), &m); err != nil {
		return []string{err.Error()}
	}
	if got := shortform.Validate(&m); len(got) > 0 {
		issues = append(issues, got...)
	}
	abs, err := localexec.ExistingFileUnder(episodeDir, m.Output.Path, "voiceover")
	if err != nil {
		return append(issues, err.Error())
	}
	hash, err := localexec.FileSHA256(abs)
	if err != nil || hash != m.Output.Hash {
		issues = append(issues, "voiceover output hash mismatch")
	}
	return issues
}

func validateSubtitleArtifacts(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var m shortform.SubtitleManifest
	if err := readJSON(filepath.Join(episodeDir, "subtitle_manifest.json"), &m); err != nil {
		return []string{err.Error()}
	}
	if got := shortform.Validate(&m); len(got) > 0 {
		issues = append(issues, got...)
	}
	for _, pair := range []struct {
		path string
		hash string
		name string
	}{{m.TranscriptPath, m.TranscriptHash, "transcript"}, {m.SRTPath, m.SRTHash, "srt"}, {m.ASSPath, m.ASSHash, "ass"}} {
		if pair.path == "" {
			continue
		}
		abs, err := localexec.ExistingFileUnder(episodeDir, pair.path, pair.name)
		if err != nil {
			issues = append(issues, err.Error())
			continue
		}
		hash, err := localexec.FileSHA256(abs)
		if err != nil || hash != pair.hash {
			issues = append(issues, pair.name+" hash mismatch")
		}
	}
	return issues
}

func validateRenderArtifacts(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var m shortform.ShortRenderManifest
	if err := readJSON(filepath.Join(episodeDir, "short_render_manifest.json"), &m); err != nil {
		return []string{err.Error()}
	}
	if got := shortform.Validate(&m); len(got) > 0 {
		issues = append(issues, got...)
	}
	if len(m.Outputs) == 0 {
		return append(issues, "short_render_manifest.json has no outputs")
	}
	out := m.Outputs[0]
	abs, err := localexec.ExistingFileUnder(episodeDir, out.Path, "release candidate")
	if err != nil {
		return append(issues, err.Error())
	}
	hash, err := localexec.FileSHA256(abs)
	if err != nil || hash != out.Hash {
		issues = append(issues, "release candidate hash mismatch")
	}
	if out.Resolution != shortform.TargetResolution || out.Aspect != shortform.TargetAspect || out.FPS != shortform.TargetFPS {
		issues = append(issues, "release candidate render properties are not 1080x1920 9:16 30fps")
	}
	if !out.AudioTrack || !out.SubtitlesBurned {
		issues = append(issues, "release candidate requires audio and burned captions")
	}
	return issues
}

func validateFinalReview(episodeDir string, manifest EpisodeManifest) []string {
	ok, issue, err := Runner{}.finalReviewPassed(episodeDir, manifest)
	if err != nil {
		return []string{err.Error()}
	}
	if !ok {
		return []string{issue}
	}
	return nil
}

func validateProductionQAAndPublish(episodeDir string, manifest EpisodeManifest) []string {
	var issues []string
	var qa ProductionQAReport
	if err := readJSON(filepath.Join(episodeDir, "production_qa_report.json"), &qa); err != nil {
		return []string{err.Error()}
	}
	if qa.Decision != "approved" || len(qa.BlockingIssues) > 0 {
		issues = append(issues, "production_qa_report.json must be approved with no blocking issues")
	}
	var pm PublishManifest
	if err := readJSON(filepath.Join(episodeDir, "publish_manifest.json"), &pm); err != nil {
		return append(issues, err.Error())
	}
	if pm.LivePublishingEnabled || pm.Visibility == "public" || pm.Mode != "release_candidate_only" {
		issues = append(issues, "publish_manifest.json must not enable live or public publishing")
	}
	return issues
}

func publishManifestDisablesLive(episodeDir string) bool {
	var pm PublishManifest
	if err := readJSON(filepath.Join(episodeDir, "publish_manifest.json"), &pm); err != nil {
		return true
	}
	return !pm.LivePublishingEnabled && pm.Visibility != "public" && pm.Mode == "release_candidate_only"
}

func finalReviewFilePasses(episodeDir, episodeID string) bool {
	var review ClaudeReviewResponse
	if err := readJSON(filepath.Join(episodeDir, "final_review_response.json"), &review); err != nil {
		return false
	}
	return review.SchemaVersion == SchemaVersion && review.EpisodeID == episodeID &&
		((review.Verdict == "pass" && review.CanReleaseCandidate && len(review.BlockingIssues) == 0) || review.OperatorOverride)
}

func (r Runner) ImportClaudeReview(req ImportClaudeReviewRequest) error {
	episodeDir, err := filepath.Abs(req.EpisodeDir)
	if err != nil {
		return err
	}
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return err
	}
	var review ClaudeReviewResponse
	if err := readJSON(req.File, &review); err != nil {
		return err
	}
	if review.SchemaVersion != SchemaVersion || review.EpisodeID != manifest.EpisodeID {
		return fmt.Errorf("review response must have schema_version=1.0 and episode_id=%s", manifest.EpisodeID)
	}
	switch req.Kind {
	case "script":
		if review.ApprovedScriptHash == "" {
			return fmt.Errorf("script review requires approved_script_hash")
		}
		dst := filepath.Join(episodeDir, "claude_script_review_response.json")
		if err := writeJSON(dst, review); err != nil {
			return err
		}
		ok, issue, err := r.scriptReviewPassed(episodeDir, manifest)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("script review rejected: %s", issue)
		}
	case "final":
		dst := filepath.Join(episodeDir, "final_review_response.json")
		if err := writeJSON(dst, review); err != nil {
			return err
		}
		ok, issue, err := r.finalReviewPassed(episodeDir, manifest)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("final review rejected: %s", issue)
		}
	default:
		return fmt.Errorf("--kind must be script or final")
	}
	return r.appendAudit(episodeDir, "claude_"+req.Kind+"_review_imported", "Claude review response imported and validated")
}

func (r Runner) ImportVisualShot(req ImportAssetRequest) error {
	episodeDir, err := filepath.Abs(req.EpisodeDir)
	if err != nil {
		return err
	}
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return err
	}
	var requests VisualShotRequests
	if err := readJSON(filepath.Join(episodeDir, "visual_shot_requests.json"), &requests); err != nil {
		return err
	}
	if req.ShotID == "" {
		return fmt.Errorf("--shot-id is required")
	}
	var shotReq *VisualShotRequest
	for i := range requests.Shots {
		if requests.Shots[i].ShotID == req.ShotID {
			shotReq = &requests.Shots[i]
			break
		}
	}
	if shotReq == nil {
		return fmt.Errorf("unknown shot_id %q", req.ShotID)
	}
	if err := os.MkdirAll(filepath.Join(episodeDir, "visual"), 0o755); err != nil {
		return err
	}
	dstRel := filepath.ToSlash(filepath.Join("visual", req.ShotID+strings.ToLower(filepath.Ext(req.File))))
	dst := filepath.Join(episodeDir, filepath.FromSlash(dstRel))
	if fileExists(dst) {
		return fmt.Errorf("refusing to overwrite existing %s without regeneration", dstRel)
	}
	if err := copyFile(dst, req.File); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, "visual_shot_imported", fmt.Sprintf("manual visual shot imported for %s in %s", req.ShotID, manifest.EpisodeID))
}

func (r Runner) ImportVoice(req ImportAssetRequest) error {
	episodeDir, err := filepath.Abs(req.EpisodeDir)
	if err != nil {
		return err
	}
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(episodeDir, "audio"), 0o755); err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(req.File))
	if ext == "" {
		ext = ".wav"
	}
	dstRel := filepath.ToSlash(filepath.Join("audio", "voiceover"+ext))
	dst := filepath.Join(episodeDir, filepath.FromSlash(dstRel))
	if fileExists(dst) {
		return fmt.Errorf("refusing to overwrite existing %s without regeneration", dstRel)
	}
	if err := copyFile(dst, req.File); err != nil {
		return err
	}
	hash, err := localexec.FileSHA256(dst)
	if err != nil {
		return err
	}
	vm := &shortform.VoiceoverManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       manifest.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", shortform.KindVoiceoverManifest, manifest.EpisodeID),
			CreatedAt:       r.now().Format(time.RFC3339),
			CreatedBy:       "human:manual-import",
			SourceArtifacts: []string{"script.md", "claude_script_review_response.json"},
			Status:          shortform.StatusInReview,
		},
		Provider:        shortform.ProviderRef{Name: "manual-import"},
		SourceScriptRef: "script.md",
		Language:        manifest.Language,
		Output: shortform.MediaOutput{
			Path:        dstRel,
			Hash:        hash,
			DurationSec: manifest.DurationSec,
			Format:      strings.TrimPrefix(ext, "."),
		},
		OperatorApproval: false,
	}
	if err := shortform.Stamp(vm); err != nil {
		return err
	}
	if issues := shortform.Validate(vm); len(issues) > 0 {
		return fmt.Errorf("voiceover_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(filepath.Join(episodeDir, "voiceover_manifest.json"), vm); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, "voice_imported", "manual voiceover imported and hashed")
}

func copyFile(dst, src string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

var whitespace = regexp.MustCompile(`\s+`)

func compactText(text string) string {
	return strings.TrimSpace(whitespace.ReplaceAllString(text, " "))
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
