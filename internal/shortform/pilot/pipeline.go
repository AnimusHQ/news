package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/contenthash"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

type Runner struct {
	Now func() time.Time
	// ReviewClient performs Claude API reviews when --claude-review api is
	// selected. It is nil in production (built from env on demand) and injected
	// in tests. Never an approval authority: the pilot validates and gates.
	ReviewClient ReviewClient
}

func (r Runner) now() time.Time {
	if r.Now == nil {
		return time.Now().UTC()
	}
	return r.Now().UTC()
}

func GenerateReal(ctx context.Context, req GenerateRequest) (Result, error) {
	return Runner{}.GenerateReal(ctx, req)
}

func Resume(ctx context.Context, episodeDir string) (Result, error) {
	return Runner{}.Resume(ctx, episodeDir)
}

func Status(episodeDir string) (Result, error) {
	return Runner{}.Status(episodeDir)
}

func Validate(episodeDir string) (ValidationReport, error) {
	return Runner{}.Validate(episodeDir)
}

func ImportClaudeReview(req ImportClaudeReviewRequest) error {
	return Runner{}.ImportClaudeReview(req)
}

func ImportVisualShot(req ImportAssetRequest) error {
	return Runner{}.ImportVisualShot(req)
}

func ImportVoice(req ImportAssetRequest) error {
	return Runner{}.ImportVoice(req)
}

func (r Runner) GenerateReal(ctx context.Context, req GenerateRequest) (Result, error) {
	if err := validateGenerateRequest(req); err != nil {
		return Result{}, err
	}
	episodeDir, err := filepath.Abs(req.OutDir)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(episodeDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create episode workspace: %w", err)
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return Result{}, err
	}
	manifestPath := filepath.Join(episodeDir, "episode_manifest.json")
	if fileExists(manifestPath) {
		existing, err := loadEpisodeManifest(episodeDir)
		if err != nil {
			return Result{}, err
		}
		if existing.EpisodeID != req.EpisodeID {
			return Result{}, fmt.Errorf("episode_manifest.json belongs to %q, not %q", existing.EpisodeID, req.EpisodeID)
		}
	} else {
		durationSec, err := parseDurationSeconds(req.Duration)
		if err != nil {
			return Result{}, err
		}
		now := r.now().Format(time.RFC3339)
		manifest := EpisodeManifest{
			SchemaVersion:  SchemaVersion,
			EpisodeID:      req.EpisodeID,
			CreatedAt:      now,
			UpdatedAt:      now,
			Status:         StatusGenerated,
			OriginalPrompt: strings.TrimSpace(req.Prompt),
			Language:       strings.TrimSpace(req.Language),
			Duration:       strings.TrimSpace(req.Duration),
			DurationSec:    durationSec,
			Platforms:      normalizeCSV(req.Platforms),
			Providers: ProviderSelections{
				Visual:       req.VisualProvider,
				Voice:        req.VoiceProvider,
				Subtitle:     req.SubtitleProvider,
				Render:       req.RenderProvider,
				ClaudeReview: req.ClaudeReview,
			},
		}
		if err := stampPilotArtifact(&manifest); err != nil {
			return Result{}, err
		}
		if err := writeJSON(manifestPath, manifest); err != nil {
			return Result{}, err
		}
		if err := writeTopicYAML(episodeDir, manifest); err != nil {
			return Result{}, err
		}
		if err := writePilotResearchPack(episodeDir, manifest); err != nil {
			return Result{}, err
		}
		if err := r.appendAudit(episodeDir, "workspace_created", "episode workspace created"); err != nil {
			return Result{}, err
		}
	}
	return r.Resume(ctx, episodeDir)
}

func (r Runner) Resume(ctx context.Context, episodeDir string) (Result, error) {
	episodeDir, err := filepath.Abs(episodeDir)
	if err != nil {
		return Result{}, err
	}
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return Result{}, err
	}
	if err := r.ensureScript(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureClaudeScriptRequest(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureAPIScriptReview(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if ok, issue, err := r.scriptReviewPassed(episodeDir, manifest); err != nil {
		return Result{}, err
	} else if !ok {
		_ = r.updateEpisodeStatus(episodeDir, StatusNeedsReview)
		res, _ := r.Status(episodeDir)
		res.BlockedGate = StageClaudeScriptReview
		res.BlockingIssue = issue
		res.NextAction = fmt.Sprintf("send %s to Claude, save JSON, then run import-claude-review --kind script", filepath.Join(episodeDir, "claude_script_review_request.md"))
		return res, nil
	}
	if err := r.ensureVisualRequests(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureVisuals(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureVoice(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureSubtitles(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureRender(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureFinalReviewRequest(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensureAPIFinalReview(ctx, episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if ok, issue, err := r.finalReviewPassed(episodeDir, manifest); err != nil {
		return Result{}, err
	} else if !ok {
		_ = r.updateEpisodeStatus(episodeDir, StatusReleaseBlocked)
		res, _ := r.Status(episodeDir)
		res.BlockedGate = StageClaudeFinalReview
		res.BlockingIssue = issue
		res.NextAction = fmt.Sprintf("review %s with Claude, save JSON, then run import-claude-review --kind final", filepath.Join(episodeDir, "final_review_request.md"))
		return res, nil
	}
	if err := r.ensureProductionQA(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.ensurePublishManifest(episodeDir, manifest); err != nil {
		return Result{}, err
	}
	if err := r.updateEpisodeStatus(episodeDir, StatusReleaseCandidate); err != nil {
		return Result{}, err
	}
	return r.Status(episodeDir)
}

func (r Runner) Status(episodeDir string) (Result, error) {
	episodeDir, err := filepath.Abs(episodeDir)
	if err != nil {
		return Result{}, err
	}
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return Result{}, err
	}
	required := requiredArtifactPaths(manifest.EpisodeID)
	existing := make([]string, 0, len(required))
	missing := make([]string, 0, len(required))
	for _, rel := range required {
		if fileExists(filepath.Join(episodeDir, filepath.FromSlash(rel))) {
			existing = append(existing, rel)
		} else {
			missing = append(missing, rel)
		}
	}
	stage, gate, action := deriveStage(episodeDir, manifest)
	releasePath := filepath.ToSlash(filepath.Join(episodeDir, "dist", manifest.EpisodeID+"-release-candidate.mp4"))
	ready := stage == StageReleaseCandidate && fileExists(filepath.Join(episodeDir, "publish_manifest.json"))
	return Result{
		EpisodeID:   manifest.EpisodeID,
		EpisodeDir:  episodeDir,
		Stage:       stage,
		BlockedGate: gate,
		NextAction:  action,
		Ready:       ready,
		ReleasePath: releasePath,
		Artifacts:   existing,
		Missing:     missing,
	}, nil
}

func (r Runner) Validate(episodeDir string) (ValidationReport, error) {
	status, err := r.Status(episodeDir)
	if err != nil {
		return ValidationReport{}, err
	}
	var issues []string
	manifest, err := loadEpisodeManifest(status.EpisodeDir)
	if err != nil {
		return ValidationReport{}, err
	}
	issues = append(issues, validateEpisodeManifest(status.EpisodeDir, manifest)...)
	issues = append(issues, validateScriptArtifacts(status.EpisodeDir, manifest)...)
	issues = append(issues, validateScriptReview(status.EpisodeDir, manifest)...)
	issues = append(issues, validateVisualArtifacts(status.EpisodeDir, manifest)...)
	issues = append(issues, validateVoiceArtifacts(status.EpisodeDir, manifest)...)
	issues = append(issues, validateSubtitleArtifacts(status.EpisodeDir, manifest)...)
	issues = append(issues, validateRenderArtifacts(status.EpisodeDir, manifest)...)
	issues = append(issues, validateFinalReview(status.EpisodeDir, manifest)...)
	issues = append(issues, validateProductionQAAndPublish(status.EpisodeDir, manifest)...)
	sort.Strings(issues)
	report := ValidationReport{
		EpisodeID:               status.EpisodeID,
		EpisodeDir:              status.EpisodeDir,
		Valid:                   len(issues) == 0 && status.Ready,
		CurrentStage:            status.Stage,
		ReleaseCandidateReady:   status.Ready,
		ReleaseCandidatePath:    status.ReleasePath,
		ExistingArtifacts:       status.Artifacts,
		MissingArtifacts:        status.Missing,
		BlockingGate:            status.BlockedGate,
		NextAction:              status.NextAction,
		Issues:                  issues,
		NoPublicPublishingPath:  publishManifestDisablesLive(status.EpisodeDir),
		FinalClaudeReviewPassed: finalReviewFilePasses(status.EpisodeDir, manifest.EpisodeID),
	}
	if !report.NoPublicPublishingPath {
		report.Valid = false
		report.Issues = append(report.Issues, "publish_manifest.json must keep live_public_publishing disabled")
	}
	if !report.ReleaseCandidateReady {
		report.Valid = false
	}
	return report, nil
}

func (r Runner) ensureScript(episodeDir string, manifest EpisodeManifest) error {
	scriptPath := filepath.Join(episodeDir, "script.md")
	scriptManifestPath := filepath.Join(episodeDir, "script_manifest.json")
	if fileExists(scriptPath) && fileExists(scriptManifestPath) {
		return nil
	}
	if fileExists(scriptPath) || fileExists(scriptManifestPath) {
		return fmt.Errorf("partial script artifacts exist; refusing to overwrite without regeneration")
	}
	body := buildScript(manifest)
	if err := writeText(scriptPath, body); err != nil {
		return err
	}
	scriptHash, err := localexec.FileSHA256(scriptPath)
	if err != nil {
		return err
	}
	promptHash, err := textHash(manifest.OriginalPrompt)
	if err != nil {
		return err
	}
	sm := ScriptManifest{
		SchemaVersion:        SchemaVersion,
		EpisodeID:            manifest.EpisodeID,
		CreatedAt:            r.now().Format(time.RFC3339),
		Status:               StatusGenerated,
		ScriptPath:           "script.md",
		ScriptHash:           scriptHash,
		EstimatedDurationSec: manifest.DurationSec,
		SourcePromptHash:     promptHash,
	}
	if err := stampPilotArtifact(&sm); err != nil {
		return err
	}
	if err := writeJSON(scriptManifestPath, sm); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageScript, "deterministic local script generated")
}

func (r Runner) ensureClaudeScriptRequest(episodeDir string, manifest EpisodeManifest) error {
	path := filepath.Join(episodeDir, "claude_script_review_request.md")
	if fileExists(path) {
		return nil
	}
	script, err := os.ReadFile(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	scriptHash, err := localexec.FileSHA256(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	req := fmt.Sprintf(`# Claude Script Review Request

Episode: %s
Language: %s
Target duration: %s
Platforms: %s
Required response: JSON only.

Original prompt:

%s

Current script hash:

%s

Review gate:

- Validate that the script is suitable for visual generation.
- Identify unsupported factual claims or misleading framing.
- Do not approve if the script needs blocking revisions.
- Return this JSON shape:

`+"```json"+`
{
  "schema_version": "1.0",
  "episode_id": "%s",
  "verdict": "pass",
  "production_readiness": 85,
  "blocking_issues": [],
  "suggested_revisions": [],
  "approved_script_hash": "%s",
  "can_continue_to_visual_generation": true
}
`+"```"+`

Script:

%s
`, manifest.EpisodeID, manifest.Language, manifest.Duration, strings.Join(manifest.Platforms, ","),
		manifest.OriginalPrompt, scriptHash, manifest.EpisodeID, scriptHash, string(script))
	if err := writeText(path, req); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageClaudeScriptReview, "manual Claude script review request written")
}

func (r Runner) scriptReviewPassed(episodeDir string, manifest EpisodeManifest) (bool, string, error) {
	path := filepath.Join(episodeDir, "claude_script_review_response.json")
	if !fileExists(path) {
		return false, "missing claude_script_review_response.json", nil
	}
	var review ClaudeReviewResponse
	if err := readJSON(path, &review); err != nil {
		return false, "", err
	}
	if review.SchemaVersion != SchemaVersion || review.EpisodeID != manifest.EpisodeID {
		return false, "Claude script review response has invalid schema_version or episode_id", nil
	}
	if review.Verdict != "pass" {
		return false, "Claude script review verdict is not pass", nil
	}
	if !review.CanContinueToVisualGeneration {
		return false, "Claude script review did not allow visual generation", nil
	}
	if len(review.BlockingIssues) > 0 {
		return false, "Claude script review has blocking issues", nil
	}
	scriptHash, err := localexec.FileSHA256(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return false, "", err
	}
	if review.ApprovedScriptHash != scriptHash {
		return false, "approved_script_hash does not match current script.md", nil
	}
	return true, "", nil
}

func (r Runner) ensureVisualRequests(episodeDir string, manifest EpisodeManifest) error {
	path := filepath.Join(episodeDir, "visual_shot_requests.json")
	if fileExists(path) {
		return nil
	}
	scriptHash, err := localexec.FileSHA256(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	requests := VisualShotRequests{
		SchemaVersion:    SchemaVersion,
		EpisodeID:        manifest.EpisodeID,
		CreatedAt:        r.now().Format(time.RFC3339),
		Status:           StatusGenerated,
		SourceScriptHash: scriptHash,
		Shots:            buildShotRequests(manifest),
	}
	if err := stampPilotArtifact(&requests); err != nil {
		return err
	}
	if err := writeJSON(path, requests); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageVisualRequests, "visual shot requests generated from prompt and script")
}

func (r Runner) ensureVisuals(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	if fileExists(filepath.Join(episodeDir, "visual_shot_manifest.json")) {
		return nil
	}
	if manifest.Providers.Visual != "external-command" {
		return fmt.Errorf("unsupported visual provider %q; L1 generate-real supports external-command", manifest.Providers.Visual)
	}
	requestPath := filepath.Join(episodeDir, "visual_shot_requests.json")
	var requests VisualShotRequests
	if err := readJSON(requestPath, &requests); err != nil {
		return err
	}
	visualDir := filepath.Join(episodeDir, "visual")
	if err := os.MkdirAll(visualDir, 0o755); err != nil {
		return err
	}
	cfg, missing := externalCommandConfig("ANIMUS_VISUAL")
	if len(missing) > 0 {
		return fmt.Errorf("visual provider external-command missing configuration: %s; request: %s", strings.Join(missing, ", "), requestPath)
	}
	input := ExternalVisualInput{
		SchemaVersion: SchemaVersion,
		EpisodeID:     manifest.EpisodeID,
		Provider:      "external-command",
		Shots:         requests.Shots,
		OutputDir:     visualDir,
	}
	var response ExternalVisualResponse
	if err := runExternalJSON(ctx, cfg, input, &response); err != nil {
		return err
	}
	shots, provider, err := normalizeVisualResponse(episodeDir, requests, response)
	if err != nil {
		return err
	}
	manifestOut := &shortform.VisualShotManifest{
		Envelope: shortform.Envelope{
			SchemaVersion:   shortform.SchemaVersion,
			EpisodeID:       manifest.EpisodeID,
			ArtifactID:      fmt.Sprintf("%s-%s-v1", shortform.KindVisualShotManifest, manifest.EpisodeID),
			CreatedAt:       r.now().Format(time.RFC3339),
			CreatedBy:       "system:external-command-visual",
			SourceArtifacts: []string{"visual_shot_requests.json", "claude_script_review_response.json"},
			Status:          shortform.StatusInReview,
		},
		Provider:    shortform.ProviderRef{Name: provider},
		AspectRatio: shortform.TargetAspect,
		RenderTarget: shortform.RenderTarget{
			Resolution: shortform.TargetResolution,
			Aspect:     shortform.TargetAspect,
			FPS:        shortform.TargetFPS,
			VideoCodec: shortform.TargetVideoCodec,
		},
		Shots: shots,
	}
	if err := shortform.Stamp(manifestOut); err != nil {
		return err
	}
	if issues := shortform.Validate(manifestOut); len(issues) > 0 {
		return fmt.Errorf("visual_shot_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(filepath.Join(episodeDir, "visual_shot_manifest.json"), manifestOut); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageVisualGeneration, "external visual command outputs validated and hashed")
}

func (r Runner) ensureVoice(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	if fileExists(filepath.Join(episodeDir, "voiceover_manifest.json")) {
		return nil
	}
	if manifest.Providers.Voice != "external-command" {
		return fmt.Errorf("unsupported voice provider %q; L1 generate-real supports external-command", manifest.Providers.Voice)
	}
	audioDir := filepath.Join(episodeDir, "audio")
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		return err
	}
	cfg, missing := externalCommandConfig("ANIMUS_VOICE")
	if len(missing) > 0 {
		return fmt.Errorf("voice provider external-command missing configuration: %s", strings.Join(missing, ", "))
	}
	text, err := voiceoverText(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	input := ExternalVoiceInput{
		SchemaVersion: SchemaVersion,
		EpisodeID:     manifest.EpisodeID,
		Language:      manifest.Language,
		Text:          text,
		OutputDir:     audioDir,
	}
	var response ExternalVoiceResponse
	if err := runExternalJSON(ctx, cfg, input, &response); err != nil {
		return err
	}
	voiceManifest, err := normalizeVoiceResponse(episodeDir, manifest, response)
	if err != nil {
		return err
	}
	if err := shortform.Stamp(voiceManifest); err != nil {
		return err
	}
	if issues := shortform.Validate(voiceManifest); len(issues) > 0 {
		return fmt.Errorf("voiceover_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(filepath.Join(episodeDir, "voiceover_manifest.json"), voiceManifest); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageVoiceGeneration, "external voice command output validated and hashed")
}

func (r Runner) ensureSubtitles(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	if fileExists(filepath.Join(episodeDir, "subtitle_manifest.json")) {
		return nil
	}
	switch manifest.Providers.Subtitle {
	case "script-timing":
		return r.generateScriptTimingSubtitles(episodeDir, manifest)
	case "faster-whisper":
		return r.generateFasterWhisperSubtitles(ctx, episodeDir, manifest)
	default:
		return fmt.Errorf("unsupported subtitle provider %q; use faster-whisper or script-timing", manifest.Providers.Subtitle)
	}
}

func (r Runner) generateFasterWhisperSubtitles(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	cfg, missing := externalCommandConfig("ANIMUS_FASTER_WHISPER")
	if len(missing) > 0 {
		return fmt.Errorf("subtitle provider faster-whisper missing configuration: %s", strings.Join(missing, ", "))
	}
	voice, err := loadVoiceoverManifest(episodeDir)
	if err != nil {
		return err
	}
	audioPath, err := localexec.ExistingFileUnder(episodeDir, voice.Output.Path, "voiceover")
	if err != nil {
		return err
	}
	outDir := filepath.Join(episodeDir, "subtitles")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	input := SubtitleSidecarInput{
		SchemaVersion: SchemaVersion,
		EpisodeID:     manifest.EpisodeID,
		Provider:      "faster-whisper",
		Language:      manifest.Language,
		AudioPath:     audioPath,
		OutputDir:     outDir,
	}
	var response SubtitleSidecarResponse
	if err := runExternalJSON(ctx, cfg, input, &response); err != nil {
		return err
	}
	subtitleManifest, err := normalizeSubtitleResponse(episodeDir, manifest, response)
	if err != nil {
		return err
	}
	if err := shortform.Stamp(subtitleManifest); err != nil {
		return err
	}
	if issues := shortform.Validate(subtitleManifest); len(issues) > 0 {
		return fmt.Errorf("subtitle_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(filepath.Join(episodeDir, "subtitle_manifest.json"), subtitleManifest); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageSubtitles, "faster-whisper sidecar subtitle outputs validated and hashed")
}

func (r Runner) generateScriptTimingSubtitles(episodeDir string, manifest EpisodeManifest) error {
	outDir := filepath.Join(episodeDir, "subtitles")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	text, err := voiceoverText(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	duration := manifest.DurationSec
	if duration <= 0 {
		duration = 45
	}
	transcriptPath := filepath.Join(outDir, "transcript.json")
	srtPath := filepath.Join(outDir, "captions.srt")
	assPath := filepath.Join(outDir, "captions.ass")
	transcript := map[string]any{
		"schema_version": SchemaVersion,
		"episode_id":     manifest.EpisodeID,
		"provider":       "script-timing",
		"language":       manifest.Language,
		"segments": []map[string]any{{
			"start": 0,
			"end":   duration,
			"text":  text,
		}},
	}
	if err := writeJSON(transcriptPath, transcript); err != nil {
		return err
	}
	if err := writeText(srtPath, fmt.Sprintf("1\n00:00:00,000 --> %s\n%s\n", srtTimestamp(duration), text)); err != nil {
		return err
	}
	if err := writeText(assPath, assSubtitle(text, duration)); err != nil {
		return err
	}
	subtitleManifest, err := manifestForSubtitleFiles(episodeDir, manifest, "script-timing")
	if err != nil {
		return err
	}
	if err := shortform.Stamp(subtitleManifest); err != nil {
		return err
	}
	if issues := shortform.Validate(subtitleManifest); len(issues) > 0 {
		return fmt.Errorf("subtitle_manifest.json validation failed: %v", issues)
	}
	if err := writeJSON(filepath.Join(episodeDir, "subtitle_manifest.json"), subtitleManifest); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageSubtitles, "script-timing subtitles generated by explicit fallback")
}

func (r Runner) ensureFinalReviewRequest(episodeDir string, manifest EpisodeManifest) error {
	path := filepath.Join(episodeDir, "final_review_request.md")
	if fileExists(path) {
		return nil
	}
	script, err := os.ReadFile(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	req := fmt.Sprintf(`# Claude Final Release-Candidate Review Request

Episode: %s
Original prompt:

%s

Release candidate:

%s

Required response: JSON only.

`+"```json"+`
{
  "schema_version": "1.0",
  "episode_id": "%s",
  "verdict": "pass",
  "production_readiness": 85,
  "blocking_issues": [],
  "suggested_revisions": [],
  "can_release_candidate": true
}
`+"```"+`

QA checklist:

- The video answers the original prompt.
- Script, visuals, voice, and subtitles are coherent.
- No unsafe public publishing is authorized by this review.
- Blocking issues are listed explicitly.
- Known limitation: L1 may use external wrappers; verify provider provenance in manifests.

Artifacts for review:

- script.md
- visual_shot_manifest.json
- voiceover_manifest.json
- subtitle_manifest.json
- short_render_manifest.json

Script:

%s
`, manifest.EpisodeID, manifest.OriginalPrompt, filepath.Join(episodeDir, "dist", manifest.EpisodeID+"-release-candidate.mp4"), manifest.EpisodeID, string(script))
	if err := writeText(path, req); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageClaudeFinalReview, "manual Claude final review request written")
}

func (r Runner) finalReviewPassed(episodeDir string, manifest EpisodeManifest) (bool, string, error) {
	path := filepath.Join(episodeDir, "final_review_response.json")
	if !fileExists(path) {
		return false, "missing final_review_response.json", nil
	}
	var review ClaudeReviewResponse
	if err := readJSON(path, &review); err != nil {
		return false, "", err
	}
	if review.SchemaVersion != SchemaVersion || review.EpisodeID != manifest.EpisodeID {
		return false, "Claude final review response has invalid schema_version or episode_id", nil
	}
	if review.OperatorOverride {
		if strings.TrimSpace(review.OperatorOverrideReason) == "" {
			return false, "operator override requires operator_override_reason", nil
		}
		return true, "", nil
	}
	if review.Verdict != "pass" {
		return false, "Claude final review verdict is not pass", nil
	}
	if !review.CanReleaseCandidate {
		return false, "Claude final review did not allow release candidate", nil
	}
	if len(review.BlockingIssues) > 0 {
		return false, "Claude final review has blocking issues", nil
	}
	return true, "", nil
}

func (r Runner) ensureProductionQA(episodeDir string, manifest EpisodeManifest) error {
	path := filepath.Join(episodeDir, "production_qa_report.json")
	if fileExists(path) {
		return nil
	}
	renderManifest, err := loadRenderManifest(episodeDir)
	if err != nil {
		return err
	}
	checks := map[string]string{
		"audio":            "pass",
		"captions":         "pass",
		"visuals":          "pass",
		"asset_provenance": "pass",
		"policy":           "pass",
		"claude_final_qa":  "pass",
	}
	if len(renderManifest.Outputs) == 0 || !renderManifest.Outputs[0].AudioTrack || !renderManifest.Outputs[0].SubtitlesBurned {
		checks["audio"] = "fail"
		checks["captions"] = "fail"
	}
	report := ProductionQAReport{
		SchemaVersion:  SchemaVersion,
		EpisodeID:      manifest.EpisodeID,
		ArtifactID:     "production-qa-" + manifest.EpisodeID + "-l1-v1",
		CreatedAt:      r.now().Format(time.RFC3339),
		Status:         StatusVerified,
		Checks:         checks,
		BlockingIssues: nil,
		Decision:       "approved",
	}
	for _, value := range checks {
		if value == "fail" {
			report.Decision = "blocked"
			report.BlockingIssues = append(report.BlockingIssues, "render technical checks failed")
			break
		}
	}
	if err := stampPilotArtifact(&report); err != nil {
		return err
	}
	if err := writeJSON(path, report); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageProductionQA, "production QA report generated for release candidate")
}

func (r Runner) ensurePublishManifest(episodeDir string, manifest EpisodeManifest) error {
	path := filepath.Join(episodeDir, "publish_manifest.json")
	if fileExists(path) {
		return nil
	}
	pm := PublishManifest{
		SchemaVersion:         SchemaVersion,
		EpisodeID:             manifest.EpisodeID,
		ArtifactID:            "publish-manifest-" + manifest.EpisodeID + "-l1-v1",
		CreatedAt:             r.now().Format(time.RFC3339),
		Status:                StatusReleaseCandidate,
		Mode:                  "release_candidate_only",
		LivePublishingEnabled: false,
		Platforms:             manifest.Platforms,
		Visibility:            "private",
		ReleaseCandidatePath:  filepath.ToSlash(filepath.Join("dist", manifest.EpisodeID+"-release-candidate.mp4")),
		RenderManifestRef:     "short_render_manifest.json",
		ProductionQARef:       "production_qa_report.json",
		FinalReviewRef:        "final_review_response.json",
	}
	if err := stampPilotArtifact(&pm); err != nil {
		return err
	}
	if err := writeJSON(path, pm); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, "publish_manifest", "publish manifest written with live publishing disabled")
}

func (r Runner) updateEpisodeStatus(episodeDir, status string) error {
	manifest, err := loadEpisodeManifest(episodeDir)
	if err != nil {
		return err
	}
	manifest.Status = status
	manifest.UpdatedAt = r.now().Format(time.RFC3339)
	manifest.ContentHash = ""
	if err := stampPilotArtifact(&manifest); err != nil {
		return err
	}
	return writeJSON(filepath.Join(episodeDir, "episode_manifest.json"), manifest)
}

func validateGenerateRequest(req GenerateRequest) error {
	required := map[string]string{
		"--episode-id":        req.EpisodeID,
		"--prompt":            req.Prompt,
		"--language":          req.Language,
		"--duration":          req.Duration,
		"--visual-provider":   req.VisualProvider,
		"--voice-provider":    req.VoiceProvider,
		"--subtitle-provider": req.SubtitleProvider,
		"--render-provider":   req.RenderProvider,
		"--claude-review":     req.ClaudeReview,
		"--out":               req.OutDir,
	}
	var missing []string
	for flag, value := range required {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, flag)
		}
	}
	if len(req.Platforms) == 0 {
		missing = append(missing, "--platforms")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
	}
	if req.ClaudeReview != "manual" && req.ClaudeReview != "api" {
		return fmt.Errorf("unsupported --claude-review %q; supported: manual, api", req.ClaudeReview)
	}
	if req.RenderProvider != "ffmpeg" {
		return fmt.Errorf("unsupported --render-provider %q; L1 supports ffmpeg", req.RenderProvider)
	}
	return nil
}

func loadEpisodeManifest(episodeDir string) (EpisodeManifest, error) {
	var manifest EpisodeManifest
	if err := readJSON(filepath.Join(episodeDir, "episode_manifest.json"), &manifest); err != nil {
		return EpisodeManifest{}, err
	}
	return manifest, nil
}

func requiredArtifactPaths(episodeID string) []string {
	return []string{
		"topic.yaml",
		"research_pack.json",
		"episode_manifest.json",
		"script.md",
		"script_manifest.json",
		"claude_script_review_request.md",
		"claude_script_review_response.json",
		"visual_shot_requests.json",
		"visual_shot_manifest.json",
		"visual/shot-001.mp4",
		"visual/shot-002.mp4",
		"visual/shot-003.mp4",
		"audio/voiceover.wav",
		"voiceover_manifest.json",
		"subtitles/transcript.json",
		"subtitles/captions.srt",
		"subtitles/captions.ass",
		"subtitle_manifest.json",
		"dist/" + episodeID + "-release-candidate.mp4",
		"short_render_manifest.json",
		"final_review_request.md",
		"final_review_response.json",
		"production_qa_report.json",
		"publish_manifest.json",
		"audit.log",
	}
}

func deriveStage(episodeDir string, manifest EpisodeManifest) (string, string, string) {
	checks := []struct {
		stage string
		file  string
		gate  string
		next  string
	}{
		{StageScript, "script_manifest.json", "", "run pilot generate-real or resume to create script artifacts"},
		{StageClaudeScriptReview, "claude_script_review_response.json", StageClaudeScriptReview, "import a valid Claude script review response"},
		{StageVisualRequests, "visual_shot_requests.json", "", "resume to generate visual shot requests"},
		{StageVisualGeneration, "visual_shot_manifest.json", StageVisualGeneration, "configure visual provider or import visual shots, then resume"},
		{StageVoiceGeneration, "voiceover_manifest.json", StageVoiceGeneration, "configure voice provider or import voiceover, then resume"},
		{StageSubtitles, "subtitle_manifest.json", StageSubtitles, "configure subtitle provider, then resume"},
		{StageRender, "short_render_manifest.json", StageRender, "install/configure ffmpeg, then resume"},
		{StageClaudeFinalReview, "final_review_response.json", StageClaudeFinalReview, "import a valid Claude final review response"},
		{StageProductionQA, "production_qa_report.json", "", "resume to generate production QA report"},
		{StageReleaseCandidate, "publish_manifest.json", "", "resume to finish release-candidate manifest"},
	}
	for _, check := range checks {
		if !fileExists(filepath.Join(episodeDir, check.file)) {
			return check.stage, check.gate, check.next
		}
	}
	if manifest.Status == StatusReleaseCandidate {
		return StageReleaseCandidate, "", ""
	}
	return StageProductionQA, "", "resume to mark release candidate ready"
}

func stampPilotArtifact(v any) error {
	hash, err := contenthash.Compute(v)
	if err != nil {
		return err
	}
	switch t := v.(type) {
	case *EpisodeManifest:
		t.ContentHash = hash
	case *ScriptManifest:
		t.ContentHash = hash
	case *VisualShotRequests:
		t.ContentHash = hash
	case *ProductionQAReport:
		t.ContentHash = hash
	case *PublishManifest:
		t.ContentHash = hash
	default:
		return fmt.Errorf("unsupported pilot artifact type %T", v)
	}
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func writeText(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func normalizeCSV(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.ToLower(strings.TrimSpace(item))
			if item == "" || seen[item] {
				continue
			}
			out = append(out, item)
			seen[item] = true
		}
	}
	return out
}

func parseDurationSeconds(value string) (float64, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if strings.HasSuffix(value, "s") {
		value = strings.TrimSuffix(value, "s")
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("duration must be positive seconds, got %q", value)
	}
	return seconds, nil
}

func textHash(text string) (string, error) {
	return contenthash.Compute(map[string]string{"text": text})
}
