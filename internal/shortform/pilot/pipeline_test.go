package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

var fixedNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func TestGenerateRealCreatesWorkspaceAndStopsAtClaudeScriptReview(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "episode")
	res, err := testRunner().GenerateReal(context.Background(), testGenerateRequest(dir))
	if err != nil {
		t.Fatalf("generate-real: %v", err)
	}
	if res.Stage != StageClaudeScriptReview || res.BlockedGate != StageClaudeScriptReview {
		t.Fatalf("expected Claude script checkpoint, got %+v", res)
	}
	for _, rel := range []string{"topic.yaml", "research_pack.json", "episode_manifest.json", "script.md", "script_manifest.json", "claude_script_review_request.md", "audit.log"} {
		if !fileExists(filepath.Join(dir, rel)) {
			t.Fatalf("expected %s", rel)
		}
	}
	if fileExists(filepath.Join(dir, "visual_shot_manifest.json")) {
		t.Fatal("visual generation must be blocked before Claude approval")
	}
}

func TestImportScriptReviewRejectsHashMismatch(t *testing.T) {
	dir := createScriptCheckpoint(t)
	review := scriptReview(t, dir)
	review.ApprovedScriptHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	path := writeReview(t, dir, "bad-script-review.json", review)
	err := testRunner().ImportClaudeReview(ImportClaudeReviewRequest{EpisodeDir: dir, Kind: "script", File: path})
	if err == nil || !strings.Contains(err.Error(), "approved_script_hash") {
		t.Fatalf("expected script hash mismatch, got %v", err)
	}
}

func TestResumeFailsClosedWhenVisualConfigMissing(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	_, err := testRunner().Resume(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "ANIMUS_VISUAL_COMMAND") {
		t.Fatalf("expected missing visual config error, got %v", err)
	}
	if !fileExists(filepath.Join(dir, "visual_shot_requests.json")) {
		t.Fatal("visual request file should exist for provider handoff")
	}
}

func TestExternalVisualHashMismatchRejected(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	providers := buildFakeProviders(t)
	configureVisual(t, dir, providers.Visual)
	t.Setenv("ANIMUS_FAKE_HASH_MISMATCH", "visual")
	_, err := testRunner().Resume(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("expected visual hash mismatch rejection, got %v", err)
	}
}

func TestExternalVisualMissingShotRejected(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	providers := buildFakeProviders(t)
	configureVisual(t, dir, providers.Visual)
	t.Setenv("ANIMUS_FAKE_MISSING_SHOT", "1")
	_, err := testRunner().Resume(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "expected 3") {
		t.Fatalf("expected missing visual shot rejection, got %v", err)
	}
}

func TestVoiceProviderMissingConfigFailsClosed(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	providers := buildFakeProviders(t)
	configureVisual(t, dir, providers.Visual)
	_, err := testRunner().Resume(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "ANIMUS_VOICE_COMMAND") {
		t.Fatalf("expected missing voice config error, got %v", err)
	}
	if !fileExists(filepath.Join(dir, "visual_shot_manifest.json")) {
		t.Fatal("visual manifest should be complete before voice config block")
	}
}

func TestSubtitleProviderMissingConfigFailsClosedWhenRequested(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	providers := buildFakeProviders(t)
	configureVisual(t, dir, providers.Visual)
	configureVoice(t, dir, providers.Voice)
	_, err := testRunner().Resume(context.Background(), dir)
	if err == nil || !strings.Contains(err.Error(), "ANIMUS_FASTER_WHISPER_COMMAND") {
		t.Fatalf("expected missing faster-whisper config error, got %v", err)
	}
}

func TestFullFakeExternalPilotProducesReleaseCandidateAfterFinalClaudeReview(t *testing.T) {
	dir := createApprovedScriptCheckpoint(t)
	providers := buildFakeProviders(t)
	configureVisual(t, dir, providers.Visual)
	configureVoice(t, dir, providers.Voice)
	configureSubtitle(t, dir, providers.Subtitle)

	res, err := testRunner().Resume(context.Background(), dir)
	if err != nil {
		t.Fatalf("resume to final review: %v", err)
	}
	if res.Stage != StageClaudeFinalReview || res.Ready {
		t.Fatalf("expected final Claude checkpoint before readiness, got %+v", res)
	}
	releasePath := filepath.Join(dir, "dist", "animus-test-001-release-candidate.mp4")
	info, err := os.Stat(releasePath)
	if err != nil {
		t.Fatalf("release candidate missing: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("release candidate is empty")
	}
	report, _ := testRunner().Validate(dir)
	if report.Valid {
		t.Fatal("validate must not pass before final Claude review")
	}

	final := ClaudeReviewResponse{
		SchemaVersion:       SchemaVersion,
		EpisodeID:           "animus-test-001",
		Verdict:             "pass",
		ProductionReadiness: 88,
		BlockingIssues:      []string{},
		SuggestedRevisions:  []string{},
		CanReleaseCandidate: true,
	}
	if err := testRunner().ImportClaudeReview(ImportClaudeReviewRequest{EpisodeDir: dir, Kind: "final", File: writeReview(t, dir, "final-review.json", final)}); err != nil {
		t.Fatalf("import final review: %v", err)
	}
	res, err = testRunner().Resume(context.Background(), dir)
	if err != nil {
		t.Fatalf("resume after final review: %v", err)
	}
	if !res.Ready || res.Stage != StageReleaseCandidate {
		t.Fatalf("expected release candidate ready, got %+v", res)
	}
	report, err = testRunner().Validate(dir)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if !report.Valid || !report.NoPublicPublishingPath || !report.FinalClaudeReviewPassed {
		t.Fatalf("unexpected validation report: %+v", report)
	}
	var publish PublishManifest
	if err := readJSON(filepath.Join(dir, "publish_manifest.json"), &publish); err != nil {
		t.Fatal(err)
	}
	if publish.LivePublishingEnabled || publish.Visibility == "public" {
		t.Fatalf("public publishing path opened: %+v", publish)
	}
}

func testRunner() Runner {
	return Runner{Now: func() time.Time { return fixedNow }}
}

// TestExternalVisualPathTraversalRejected proves the visual external-command
// path (used by the Seedance wrapper) refuses an output that escapes the
// episode root, so an untrusted provider cannot write or reference files
// outside the workspace.
func TestExternalVisualPathTraversalRejected(t *testing.T) {
	episodeDir := t.TempDir()
	requests := VisualShotRequests{
		SchemaVersion: SchemaVersion,
		EpisodeID:     "animus-test-001",
		Shots:         []VisualShotRequest{{ShotID: "shot-001"}},
	}
	response := ExternalVisualResponse{
		SchemaVersion: SchemaVersion,
		EpisodeID:     "animus-test-001",
		Provider:      "seedance-wrapper",
		Shots: []ExternalVisualOutput{{
			ShotID:     "shot-001",
			Status:     "generated",
			OutputPath: "../../../etc/passwd",
			Width:      1080, Height: 1920, FPS: 30,
		}},
	}
	if _, _, err := normalizeVisualResponse(episodeDir, requests, response); err == nil || !strings.Contains(err.Error(), "escapes configured root") {
		t.Fatalf("expected path-traversal rejection, got %v", err)
	}
}

// fakeReviewClient is an injected ReviewClient that returns canned JSON without
// any network call, exercising the --claude-review api wiring offline.
type fakeReviewClient struct {
	responses map[string]json.RawMessage
	err       error
	calls     []string
}

func (f *fakeReviewClient) Review(ctx context.Context, kind, episodeID, prompt string) (json.RawMessage, error) {
	f.calls = append(f.calls, kind)
	if f.err != nil {
		return nil, f.err
	}
	resp, ok := f.responses[kind]
	if !ok {
		return nil, fmt.Errorf("fake review client has no %s response", kind)
	}
	return resp, nil
}

func apiGenerateRequest(dir string) GenerateRequest {
	req := testGenerateRequest(dir)
	req.ClaudeReview = "api"
	return req
}

func reviewJSON(t *testing.T, fields map[string]any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestClaudeReviewAPIRejectedByValidation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "episode")
	req := testGenerateRequest(dir)
	req.ClaudeReview = "council"
	_, err := testRunner().GenerateReal(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "claude-review") {
		t.Fatalf("expected unsupported claude-review rejection, got %v", err)
	}
}

func TestClaudeAPIReviewMissingKeyFailsClosed(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	dir := filepath.Join(t.TempDir(), "episode")
	// No injected ReviewClient -> builds from env -> must fail closed.
	_, err := testRunner().GenerateReal(context.Background(), apiGenerateRequest(dir))
	if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("expected fail-closed on missing key, got %v", err)
	}
	if fileExists(filepath.Join(dir, "claude_script_review_response.json")) {
		t.Fatal("no review response should be written without a key")
	}
}

func TestClaudeAPIScriptReviewPassesAndBindsScriptHash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "episode")
	fake := &fakeReviewClient{responses: map[string]json.RawMessage{
		// approved_script_hash deliberately wrong: the pilot must rebind it.
		"script": reviewJSON(t, map[string]any{
			"schema_version":                    "1.0",
			"episode_id":                        "animus-test-001",
			"verdict":                           "pass",
			"production_readiness":              88,
			"blocking_issues":                   []string{},
			"suggested_revisions":               []string{},
			"approved_script_hash":              "sha256:wrong",
			"can_continue_to_visual_generation": true,
		}),
	}}
	runner := Runner{Now: func() time.Time { return fixedNow }, ReviewClient: fake}
	// External visual provider is unconfigured, so resume fails closed at the
	// visual stage — which proves the API script review already passed.
	_, err := runner.GenerateReal(context.Background(), apiGenerateRequest(dir))
	if err == nil || !strings.Contains(err.Error(), "ANIMUS_VISUAL_COMMAND") {
		t.Fatalf("expected to advance past api script review to visual config, got %v", err)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "script" {
		t.Fatalf("expected one script review call, got %v", fake.calls)
	}
	manifest, err := loadEpisodeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	ok, issue, err := runner.scriptReviewPassed(dir, manifest)
	if err != nil || !ok {
		t.Fatalf("api script review should pass with bound hash: ok=%v issue=%q err=%v", ok, issue, err)
	}
}

func TestClaudeAPIScriptReviewFailVerdictBlocksAtGate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "episode")
	fake := &fakeReviewClient{responses: map[string]json.RawMessage{
		"script": reviewJSON(t, map[string]any{
			"schema_version":                    "1.0",
			"episode_id":                        "animus-test-001",
			"verdict":                           "fail",
			"production_readiness":              40,
			"blocking_issues":                   []string{"unsupported factual claim"},
			"suggested_revisions":               []string{"cite a source"},
			"can_continue_to_visual_generation": false,
		}),
	}}
	runner := Runner{Now: func() time.Time { return fixedNow }, ReviewClient: fake}
	res, err := runner.GenerateReal(context.Background(), apiGenerateRequest(dir))
	if err != nil {
		t.Fatalf("fail verdict should block, not error: %v", err)
	}
	if res.BlockedGate != StageClaudeScriptReview {
		t.Fatalf("expected block at claude script review, got %+v", res)
	}
}

func TestClaudeAPIFinalReviewWritesValidatedResponse(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "episode")
	fake := &fakeReviewClient{responses: map[string]json.RawMessage{
		"script": reviewJSON(t, map[string]any{
			"schema_version": "1.0", "episode_id": "animus-test-001", "verdict": "pass",
			"production_readiness": 88, "blocking_issues": []string{}, "suggested_revisions": []string{},
			"can_continue_to_visual_generation": true,
		}),
		"final": reviewJSON(t, map[string]any{
			"schema_version": "1.0", "episode_id": "animus-test-001", "verdict": "pass",
			"production_readiness": 90, "blocking_issues": []string{}, "suggested_revisions": []string{},
			"can_release_candidate": true,
		}),
	}}
	runner := Runner{Now: func() time.Time { return fixedNow }, ReviewClient: fake}
	_, _ = runner.GenerateReal(context.Background(), apiGenerateRequest(dir)) // stops at visual config

	// Exercise the final-review step directly (reaching it organically needs
	// ffmpeg + media providers). The request file is what the pipeline writes.
	if err := writeText(filepath.Join(dir, "final_review_request.md"), "# Claude Final Review Request\nepisode animus-test-001"); err != nil {
		t.Fatal(err)
	}
	manifest, err := loadEpisodeManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.ensureAPIFinalReview(context.Background(), dir, manifest); err != nil {
		t.Fatalf("final review: %v", err)
	}
	if !finalReviewFilePasses(dir, manifest.EpisodeID) {
		t.Fatal("api final review should pass the gate")
	}
}

func testGenerateRequest(dir string) GenerateRequest {
	return GenerateRequest{
		EpisodeID:        "animus-test-001",
		Prompt:           "Объясни, почему open-source разработчикам нужна устойчивая экосистема",
		Language:         "ru",
		Duration:         "3s",
		Platforms:        []string{"tiktok", "instagram", "youtube"},
		VisualProvider:   "external-command",
		VoiceProvider:    "external-command",
		SubtitleProvider: "faster-whisper",
		RenderProvider:   "ffmpeg",
		ClaudeReview:     "manual",
		OutDir:           dir,
	}
}

func createScriptCheckpoint(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "episode")
	_, err := testRunner().GenerateReal(context.Background(), testGenerateRequest(dir))
	if err != nil {
		t.Fatalf("create script checkpoint: %v", err)
	}
	return dir
}

func createApprovedScriptCheckpoint(t *testing.T) string {
	t.Helper()
	dir := createScriptCheckpoint(t)
	review := scriptReview(t, dir)
	if err := testRunner().ImportClaudeReview(ImportClaudeReviewRequest{EpisodeDir: dir, Kind: "script", File: writeReview(t, dir, "script-review.json", review)}); err != nil {
		t.Fatalf("import script review: %v", err)
	}
	return dir
}

func scriptReview(t *testing.T, dir string) ClaudeReviewResponse {
	t.Helper()
	hash, err := localFileHash(filepath.Join(dir, "script.md"))
	if err != nil {
		t.Fatal(err)
	}
	return ClaudeReviewResponse{
		SchemaVersion:                 SchemaVersion,
		EpisodeID:                     "animus-test-001",
		Verdict:                       "pass",
		ProductionReadiness:           86,
		BlockingIssues:                []string{},
		SuggestedRevisions:            []string{},
		ApprovedScriptHash:            hash,
		CanContinueToVisualGeneration: true,
	}
}

func localFileHash(path string) (string, error) {
	return localexec.FileSHA256(path)
}

func writeReview(t *testing.T, dir, name string, review ClaudeReviewResponse) string {
	t.Helper()
	path := filepath.Join(dir, name)
	data, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

type fakeProviderPaths struct {
	Visual   string
	Voice    string
	Subtitle string
}

func buildFakeProviders(t *testing.T) fakeProviderPaths {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg required for real pilot fake-provider integration test")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "fake_provider.go")
	if err := os.WriteFile(source, []byte(fakeProviderSource), 0o644); err != nil {
		t.Fatal(err)
	}
	build := func(name string) string {
		t.Helper()
		out := filepath.Join(dir, name)
		cmd := exec.Command("go", "build", "-o", out, source)
		if data, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build fake provider %s: %v: %s", name, err, string(data))
		}
		return out
	}
	return fakeProviderPaths{
		Visual:   build("fake-visual"),
		Voice:    build("fake-voice"),
		Subtitle: build("fake-subtitle"),
	}
}

func configureVisual(t *testing.T, episodeDir, command string) {
	t.Helper()
	t.Setenv("ANIMUS_VISUAL_COMMAND", command)
	t.Setenv("ANIMUS_VISUAL_INPUT_ROOT", episodeDir)
	t.Setenv("ANIMUS_VISUAL_OUTPUT_ROOT", episodeDir)
}

func configureVoice(t *testing.T, episodeDir, command string) {
	t.Helper()
	t.Setenv("ANIMUS_VOICE_COMMAND", command)
	t.Setenv("ANIMUS_VOICE_INPUT_ROOT", episodeDir)
	t.Setenv("ANIMUS_VOICE_OUTPUT_ROOT", episodeDir)
}

func configureSubtitle(t *testing.T, episodeDir, command string) {
	t.Helper()
	t.Setenv("ANIMUS_FASTER_WHISPER_COMMAND", command)
	t.Setenv("ANIMUS_FASTER_WHISPER_INPUT_ROOT", episodeDir)
	t.Setenv("ANIMUS_FASTER_WHISPER_OUTPUT_ROOT", episodeDir)
}

const fakeProviderSource = `
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	mode := strings.TrimPrefix(filepath.Base(os.Args[0]), "fake-")
	data, err := io.ReadAll(os.Stdin)
	must(err)
	switch mode {
	case "visual":
		runVisual(data)
	case "voice":
		runVoice(data)
	case "subtitle":
		runSubtitle(data)
	default:
		fail("unknown mode " + mode)
	}
}

func runVisual(data []byte) {
	var req struct {
		SchemaVersion string ` + "`json:\"schema_version\"`" + `
		EpisodeID string ` + "`json:\"episode_id\"`" + `
		OutputDir string ` + "`json:\"output_dir\"`" + `
		Shots []struct {
			ShotID string ` + "`json:\"shot_id\"`" + `
			DurationSec float64 ` + "`json:\"duration_sec\"`" + `
		} ` + "`json:\"shots\"`" + `
	}
	must(json.Unmarshal(data, &req))
	must(os.MkdirAll(req.OutputDir, 0755))
	var shots []map[string]any
	for i, shot := range req.Shots {
		if os.Getenv("ANIMUS_FAKE_MISSING_SHOT") == "1" && i == len(req.Shots)-1 {
			continue
		}
		if shot.DurationSec <= 0 {
			shot.DurationSec = 1
		}
		out := filepath.Join(req.OutputDir, shot.ShotID+".mp4")
		args := []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", fmt.Sprintf("testsrc=size=1080x1920:rate=30:duration=%.2f", shot.DurationSec), "-an", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-threads", "1", out}
		run("ffmpeg", args...)
		hash := fileHash(out)
		if os.Getenv("ANIMUS_FAKE_HASH_MISMATCH") == "visual" {
			hash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
		}
		shots = append(shots, map[string]any{"shot_id": shot.ShotID, "status": "generated", "output_path": out, "output_hash": hash, "duration_sec": shot.DurationSec, "width": 1080, "height": 1920, "fps": 30})
	}
	write(map[string]any{"schema_version": "1.0", "episode_id": req.EpisodeID, "provider": "fake-visual", "shots": shots})
}

func runVoice(data []byte) {
	var req struct {
		EpisodeID string ` + "`json:\"episode_id\"`" + `
		OutputDir string ` + "`json:\"output_dir\"`" + `
	}
	must(json.Unmarshal(data, &req))
	must(os.MkdirAll(req.OutputDir, 0755))
	out := filepath.Join(req.OutputDir, "voiceover.wav")
	run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=660:sample_rate=48000:duration=3", "-c:a", "pcm_s16le", out)
	write(map[string]any{"schema_version": "1.0", "episode_id": req.EpisodeID, "provider": "fake-voice", "output_path": out, "output_hash": fileHash(out), "duration_sec": 3.0, "sample_rate": 48000})
}

func runSubtitle(data []byte) {
	var req struct {
		EpisodeID string ` + "`json:\"episode_id\"`" + `
		OutputDir string ` + "`json:\"output_dir\"`" + `
	}
	must(json.Unmarshal(data, &req))
	must(os.MkdirAll(req.OutputDir, 0755))
	must(os.WriteFile(filepath.Join(req.OutputDir, "transcript.json"), []byte(` + "`" + `{"schema_version":"1.0","segments":[{"start":0,"end":3,"text":"test"}]}` + "`" + `), 0644))
	must(os.WriteFile(filepath.Join(req.OutputDir, "captions.srt"), []byte("1\n00:00:00,000 --> 00:00:02,800\ntest\n"), 0644))
	must(os.WriteFile(filepath.Join(req.OutputDir, "captions.ass"), []byte("[Script Info]\nScriptType: v4.00+\n[V4+ Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\nStyle: Default,Arial,58,&H00FFFFFF,&H000000FF,&H64000000,&H96000000,1,0,0,0,100,100,0,0,1,4,1,2,80,80,180,1\n[Events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\nDialogue: 0,0:00:00.00,0:00:02.80,Default,,0,0,0,,test\n"), 0644))
	write(map[string]any{"schema_version": "1.0", "episode_id": req.EpisodeID, "provider": "fake-faster-whisper", "transcript_path": filepath.Join(req.OutputDir, "transcript.json"), "srt_path": filepath.Join(req.OutputDir, "captions.srt"), "ass_path": filepath.Join(req.OutputDir, "captions.ass"), "word_timestamps": true, "safe_zone": true, "sync": true})
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		fail(fmt.Sprintf("%s failed: %v: %s", name, err, string(data)))
	}
}

func fileHash(path string) string {
	data, err := os.ReadFile(path)
	must(err)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func write(v any) {
	data, err := json.Marshal(v)
	must(err)
	fmt.Println(string(data))
}

func must(err error) {
	if err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
`
