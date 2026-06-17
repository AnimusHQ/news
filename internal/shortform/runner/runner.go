// Package runner drives the short-form pipeline end-to-end on mock activities in
// a single process (no Temporal server required). It shares the exact activity
// and gate functions with ShortFormWorkflow (see docs/adr/0004), and persists
// every artifact, gate decision, and an audit log to a run directory.
package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AnimusHQ/news/internal/audit"
	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/activities"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
)

// Injection selects a deterministic failure to exercise the demo's block path.
type Injection string

const (
	// InjectNone runs the healthy happy path.
	InjectNone Injection = ""
	// InjectUnapprovedStoryboard withholds storyboard approval so the storyboard
	// image gate halts the pipeline (the §Phase 5 example).
	InjectUnapprovedStoryboard Injection = "unapproved_storyboard"
	// InjectRenderNoAudio injects a no-audio render so the render gate halts.
	InjectRenderNoAudio Injection = "render_no_audio"
	// InjectReleaseDenied denies the release approval.
	InjectReleaseDenied Injection = "release_denied"
)

// Config configures a demo run.
type Config struct {
	EpisodeID          string
	OutputDir          string
	Now                time.Time
	Inject             Injection
	StoryboardApprover string
	ReleaseApprover    string
	Scenes             []providers.SceneSpec
}

// Result summarizes a demo run.
type Result struct {
	EpisodeID   string            `json:"episode_id"`
	RunDir      string            `json:"run_dir"`
	State       string            `json:"state"`
	Blocked     bool              `json:"blocked"`
	BlockReason string            `json:"block_reason,omitempty"`
	Notes       []string          `json:"notes"`
	GateResults []gates.Result    `json:"gate_results"`
	Artifacts   map[string]string `json:"artifacts"` // kind -> content_hash
}

// DefaultScenes returns the demo scene set.
func DefaultScenes() []providers.SceneSpec {
	return []providers.SceneSpec{
		{SceneID: "scene-001", StartSec: 0, EndSec: 5, Prompt: "hook: what happens after git push"},
		{SceneID: "scene-002", StartSec: 5, EndSec: 12, Prompt: "explainer: CI pipeline kicks off"},
		{SceneID: "scene-003", StartSec: 12, EndSec: 20, Prompt: "payoff: deploy and rollback safety"},
	}
}

type run struct {
	cfg   Config
	dir   string
	acts  *activities.Activities
	sink  *audit.MemorySink
	res   Result
	seq   int
	clock time.Time
}

// Run executes the demo pipeline and writes all outputs under
// <OutputDir>/<EpisodeID>. It returns a Result; a blocked run is a normal
// (non-error) return, mirroring the gate semantics.
func Run(ctx context.Context, cfg Config) (Result, error) {
	if cfg.EpisodeID == "" {
		cfg.EpisodeID = "episode-0001"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = filepath.Join("dist", "demo")
	}
	if cfg.Now.IsZero() {
		cfg.Now = time.Now().UTC()
	}
	if cfg.StoryboardApprover == "" {
		cfg.StoryboardApprover = "human:editor"
	}
	if cfg.ReleaseApprover == "" {
		cfg.ReleaseApprover = "human:reviewer"
	}
	if len(cfg.Scenes) == 0 {
		cfg.Scenes = DefaultScenes()
	}

	dir := filepath.Join(cfg.OutputDir, cfg.EpisodeID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Result{}, err
	}

	defects := activities.MockDefects{}
	if cfg.Inject == InjectRenderNoAudio {
		defects.Render = providers.DefectRenderNoAudio
	}

	r := &run{
		cfg:   cfg,
		dir:   dir,
		acts:  activities.NewMockActivitiesWithDefects(defects),
		sink:  audit.NewMemorySink(),
		clock: cfg.Now,
		res: Result{
			EpisodeID: cfg.EpisodeID,
			RunDir:    dir,
			State:     "started",
			Notes:     []string{},
			Artifacts: map[string]string{},
		},
	}
	if err := r.execute(ctx); err != nil {
		return r.res, err
	}
	if err := r.finalize(); err != nil {
		return r.res, err
	}
	return r.res, nil
}

func (r *run) execute(ctx context.Context) error {
	// Provenance (§4.1/§4.2).
	if r.gate(gates.ProvenanceGate(gates.ProvenanceInput{
		ScriptApproved: true, ResearchPackRef: "research_pack.json",
		Claims: []gates.ClaimRef{{ID: "claim-001", SourceIDs: []string{"source-001"}}},
	}), "provenance") {
		return nil
	}

	// Import storyboard images (draft).
	storyboard, err := r.acts.ImportStoryboardImages(ctx, activities.ImportStoryboardInput{
		EpisodeID: r.cfg.EpisodeID, Now: r.clock, Operator: "operator:ci", Scenes: r.cfg.Scenes,
	})
	if err != nil {
		return r.fail("storyboard_import_failed", err)
	}
	r.persist(storyboard)
	r.audit(audit.CategoryStateTransition, "system", "imported", "storyboard images imported (draft)")

	// Human-in-the-loop storyboard approval — withheld under injection.
	if r.cfg.Inject != InjectUnapprovedStoryboard {
		if _, err := r.acts.ApproveStoryboardImages(ctx, activities.ApproveStoryboardInput{Manifest: storyboard, Approver: r.cfg.StoryboardApprover, Now: r.clock}); err != nil {
			return r.fail("storyboard_approval_failed", err)
		}
		r.persist(storyboard)
		r.audit(audit.CategoryHumanQA, r.cfg.StoryboardApprover, "approve", "storyboard images approved")
		if r.gate(gates.SelfApprovalGate(gates.SelfApprovalInput{CreatedBy: storyboard.CreatedBy, ApproverID: r.cfg.StoryboardApprover, TargetStatus: shortform.StatusApproved}), "self_approval") {
			return nil
		}
	} else {
		r.note("INJECT: storyboard approval withheld")
	}
	if r.gate(gates.StoryboardImageGate(gates.StoryboardImageInput{Manifest: storyboard, RequiredScenes: sceneIDs(r.cfg.Scenes)}), "storyboard_image") {
		return nil
	}

	// Visual shots.
	shots, err := r.acts.GenerateMockVisualShots(ctx, activities.VisualShotsInput{EpisodeID: r.cfg.EpisodeID, Now: r.clock, Storyboard: storyboard})
	if err != nil {
		return r.fail("visual_shots_failed", err)
	}
	if _, err := r.acts.ApproveVisualShots(ctx, shots, r.clock); err != nil {
		return r.fail("visual_shots_approval_failed", err)
	}
	r.persist(shots)
	if r.gate(gates.VisualShotGate(gates.VisualShotInput{Manifest: shots, ApprovedImageHashes: approvedHashes(storyboard), KnownScenes: sceneSet(r.cfg.Scenes)}), "visual_shot") {
		return nil
	}

	// Voiceover.
	voiceover, err := r.acts.GenerateElevenLabsVoiceover(ctx, activities.VoiceoverInput{EpisodeID: r.cfg.EpisodeID, Now: r.clock, ScriptRef: "script.md", Language: "en"})
	if err != nil {
		return r.fail("voiceover_failed", err)
	}
	if _, err := r.acts.ApproveVoiceover(ctx, voiceover, r.clock); err != nil {
		return r.fail("voiceover_approval_failed", err)
	}
	r.persist(voiceover)

	// Subtitles.
	subtitles, err := r.acts.GenerateSubtitles(ctx, activities.SubtitlesInput{EpisodeID: r.cfg.EpisodeID, Now: r.clock, Voiceover: voiceover, Language: "en", WordTimestampsRequired: true})
	if err != nil {
		return r.fail("subtitles_failed", err)
	}
	if _, err := r.acts.ApproveSubtitles(ctx, subtitles, r.clock); err != nil {
		return r.fail("subtitles_approval_failed", err)
	}
	r.persist(subtitles)
	if r.gate(gates.SubtitleGate(gates.SubtitleInput{Manifest: subtitles}), "subtitle") {
		return nil
	}

	// Render + production QA.
	render, err := r.acts.RenderShortFinal(ctx, activities.RenderInput{EpisodeID: r.cfg.EpisodeID, Now: r.clock, Shots: shots, Voiceover: voiceover, Subtitles: subtitles, Platforms: []string{"master", "youtube"}})
	if err != nil {
		return r.fail("render_failed", err)
	}
	qa, err := r.acts.RunProductionQA(ctx, render)
	if err != nil {
		return r.fail("production_qa_failed", err)
	}
	r.audit(audit.CategoryProductionQA, "system:production-qa", qa.Decision, "production QA evaluated render")
	if _, err := r.acts.ApproveRenderOutputs(ctx, render, r.clock); err != nil {
		return r.fail("render_approval_failed", err)
	}
	r.persist(render)
	if r.gate(gates.RenderGate(gates.RenderInput{Manifest: render, ProductionQADecision: qa.Decision}), "render") {
		return nil
	}

	// Immutable production candidate.
	candidate, err := r.acts.AssembleProductionCandidate(ctx, activities.AssembleCandidateInput{
		EpisodeID: r.cfg.EpisodeID, CandidateID: r.cfg.EpisodeID + "-cand-001", Now: r.clock,
		Components: []shortform.ComponentRef{
			shortform.ComponentRefOf(storyboard), shortform.ComponentRefOf(shots),
			shortform.ComponentRefOf(voiceover), shortform.ComponentRefOf(subtitles), shortform.ComponentRefOf(render),
		},
	})
	if err != nil {
		return r.fail("candidate_assembly_failed", err)
	}
	r.persist(candidate)
	r.note("assembled locked production candidate")

	// Human-in-the-loop release approval — denied under injection.
	if r.cfg.Inject == InjectReleaseDenied {
		r.note("INJECT: release approval denied")
		return r.blockState("release_denied", "release approval was not granted")
	}

	release, err := r.acts.BuildReleaseApproval(ctx, shortform.BuildReleaseApprovalInput{
		EpisodeID: r.cfg.EpisodeID, CandidateID: candidate.CandidateID, Approver: r.cfg.ReleaseApprover, Now: r.clock,
		Platforms: []string{"youtube"}, Visibility: "private",
		AIDisclosureRequired: true, AIDisclosure: "AI-generated visuals and synthetic voice.", ProductionQARef: "production_qa_report.json",
	})
	if err != nil {
		return r.fail("release_build_failed", err)
	}
	r.persist(release)
	r.audit(audit.CategoryReleaseApproval, r.cfg.ReleaseApprover, "approve", "release approved by human")
	if r.gate(gates.MultiVerifierGate(gates.MultiVerifierInput{Verifiers: []string{r.cfg.StoryboardApprover, r.cfg.ReleaseApprover, "system:production-qa"}}), "multi_verifier") {
		return nil
	}

	// Guarded publish manifest + dry-run + release gate.
	publish, err := r.acts.GenerateUploadPostPublishManifest(ctx, activities.PublishManifestInput{
		EpisodeID: r.cfg.EpisodeID, Now: r.clock, Release: release, Render: render,
		ProductionQADecision: qa.Decision, ProductionQARef: "production_qa_report.json",
	})
	if err != nil {
		return r.fail("publish_manifest_failed", err)
	}
	r.persist(publish)
	dryRun, err := r.acts.UploadPostDryRun(ctx, publish)
	if err != nil {
		return r.fail("dry_run_failed", err)
	}
	r.audit(audit.CategoryPublishing, "system", "dry_run", dryRun.Detail)
	if r.gate(gates.ReleaseGate(gates.ReleaseInput{PublishManifest: publish, ProductionQADecision: qa.Decision, DryRunSucceeded: dryRun.OK}), "release") {
		return nil
	}

	r.res.State = "published_dry_run_complete"
	r.note("guarded upload-post dry-run complete (no upload performed)")
	r.audit(audit.CategoryStateTransition, "system", "published_dry_run_complete", "terminal state reached")
	return nil
}

// gate records a gate result and returns true if it blocked the run.
func (r *run) gate(result gates.Result, _ string) bool {
	r.res.GateResults = append(r.res.GateResults, result)
	if result.Blocked() {
		r.res.Blocked = true
		r.res.State = "blocked"
		r.res.BlockReason = result.Gate + ": " + firstReason(result)
		r.note("blocked at gate " + result.Gate)
		r.audit(audit.CategoryStateTransition, "system", "blocked", "halted at "+result.Gate+" gate")
		return true
	}
	r.audit(audit.CategoryStateTransition, "system", "pass", result.Gate+" gate passed")
	return false
}

func (r *run) blockState(state, reason string) error {
	r.res.Blocked = true
	r.res.State = state
	r.res.BlockReason = reason
	r.note("blocked: " + reason)
	r.audit(audit.CategoryStateTransition, "system", "blocked", reason)
	return nil
}

func (r *run) fail(state string, err error) error {
	r.res.Blocked = true
	r.res.State = state
	r.res.BlockReason = err.Error()
	r.note(state + ": " + err.Error())
	return nil // activity failures are recorded as a blocked run, not a hard error
}

func (r *run) persist(a shortform.Artifact) {
	if issues := shortform.Validate(a); len(issues) > 0 {
		r.note(fmt.Sprintf("WARNING: %s failed validation: %v", a.Kind(), issues))
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		r.note("persist marshal error: " + err.Error())
		return
	}
	path := filepath.Join(r.dir, a.Kind()+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		r.note("persist write error: " + err.Error())
		return
	}
	r.res.Artifacts[a.Kind()] = a.EnvelopeRef().ContentHash
}

func (r *run) audit(category audit.Category, actor, decision, reason string) {
	r.seq++
	event := audit.NewEventAt(fmt.Sprintf("demo-%s-%03d", r.cfg.EpisodeID, r.seq), category, actor, r.cfg.EpisodeID, decision, reason, r.clock)
	_ = r.sink.Append(event)
}

func (r *run) note(note string) { r.res.Notes = append(r.res.Notes, note) }

func (r *run) finalize() error {
	if r.res.State == "started" {
		r.res.State = "completed"
	}
	// gate_decisions.json
	if err := writeJSON(filepath.Join(r.dir, "gate_decisions.json"), r.res.GateResults); err != nil {
		return err
	}
	// audit.jsonl
	lines, err := audit.JSONLines(r.sink.Events())
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(r.dir, "audit.jsonl"), []byte(lines), 0o644); err != nil {
		return err
	}
	// run_summary.json
	return writeJSON(filepath.Join(r.dir, "run_summary.json"), r.res)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func firstReason(r gates.Result) string {
	if len(r.Reasons) == 0 {
		return string(r.Decision)
	}
	return r.Reasons[0].Code
}

func sceneIDs(scenes []providers.SceneSpec) []string {
	out := make([]string, 0, len(scenes))
	for _, s := range scenes {
		out = append(out, s.SceneID)
	}
	return out
}

func sceneSet(scenes []providers.SceneSpec) map[string]bool {
	out := make(map[string]bool, len(scenes))
	for _, s := range scenes {
		out[s.SceneID] = true
	}
	return out
}

func approvedHashes(m *shortform.StoryboardImageManifest) map[string]bool {
	out := map[string]bool{}
	for _, img := range m.Images {
		if img.Status == shortform.StatusApproved {
			out[img.ImageHash] = true
		}
	}
	return out
}
