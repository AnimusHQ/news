package workflows

import (
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/activities"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Signal and query names for the short-form workflow.
const (
	StoryboardImageApprovalSignal = "StoryboardImageApproval"
	ReleaseApprovalSignal         = "ReleaseApproval"
	GetShortFormStateQuery        = "GetShortFormState"
)

// ApprovalSignal is the payload for human-in-the-loop approval signals.
type ApprovalSignal struct {
	Decision string `json:"decision"` // "approve" or anything else to block
	Approver string `json:"approver"`
}

// ShortFormInput starts the short-form pipeline for an episode.
type ShortFormInput struct {
	EpisodeID            string                `json:"episode_id"`
	Scenes               []providers.SceneSpec `json:"scenes"`
	ScriptRef            string                `json:"script_ref"`
	ResearchPackRef      string                `json:"research_pack_ref"`
	ScriptApproved       bool                  `json:"script_approved"`
	Claims               []gates.ClaimRef      `json:"claims"`
	Platforms            []string              `json:"platforms"`
	Visibility           string                `json:"visibility"`
	ScheduledAt          string                `json:"scheduled_at"`
	AIDisclosureRequired bool                  `json:"ai_disclosure_required"`
	AIDisclosure         string                `json:"ai_disclosure"`
	Language             string                `json:"language"`
	Operator             string                `json:"operator"`
}

// ShortFormResult is the workflow outcome.
type ShortFormResult struct {
	EpisodeID   string            `json:"episode_id"`
	State       string            `json:"state"`
	Blocked     bool              `json:"blocked"`
	BlockReason string            `json:"block_reason,omitempty"`
	Notes       []string          `json:"notes"`
	GateResults []gates.Result    `json:"gate_results"`
	Artifacts   map[string]string `json:"artifacts"` // kind -> content_hash
}

// ShortFormState is returned by the workflow state query.
type ShortFormState struct {
	EpisodeID string   `json:"episode_id"`
	State     string   `json:"state"`
	Notes     []string `json:"notes"`
}

// ShortFormWorkflow drives the short-form video pipeline end-to-end on mock
// activities, enforcing every gate and waiting on human approval signals. The
// workflow is deterministic and replayable; all side effects live in activities.
func ShortFormWorkflow(ctx workflow.Context, in ShortFormInput) (ShortFormResult, error) {
	res := ShortFormResult{
		EpisodeID: in.EpisodeID,
		State:     "started",
		Notes:     []string{},
		Artifacts: map[string]string{},
	}
	if err := workflow.SetQueryHandler(ctx, GetShortFormStateQuery, func() (ShortFormState, error) {
		return ShortFormState{EpisodeID: res.EpisodeID, State: res.State, Notes: res.Notes}, nil
	}); err != nil {
		return res, err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
		},
	})

	var a *activities.Activities // nil receiver: used only to name registered activities
	now := workflow.Now(ctx)

	record := func(state, note string) {
		res.State = state
		if note != "" {
			res.Notes = append(res.Notes, note)
		}
	}
	gate := func(r gates.Result) bool {
		res.GateResults = append(res.GateResults, r)
		if r.Blocked() {
			res.Blocked = true
			res.BlockReason = r.Gate + ": " + firstReason(r)
			res.State = "blocked"
			res.Notes = append(res.Notes, "blocked at "+r.Gate)
		}
		return r.Blocked()
	}
	stamp := func(art shortform.Artifact) {
		res.Artifacts[art.Kind()] = art.EnvelopeRef().ContentHash
	}

	// §4.1/§4.2 provenance: no script downstream without a research pack; claims need sources.
	if gate(gates.ProvenanceGate(gates.ProvenanceInput{ScriptApproved: in.ScriptApproved, ResearchPackRef: in.ResearchPackRef, Claims: in.Claims})) {
		return res, nil
	}
	record("provenance_ok", "provenance gate passed")

	// Build prompt pack (deterministic) and import storyboard images (mock).
	var promptPack activities.StoryboardPromptPack
	if err := workflow.ExecuteActivity(ctx, a.BuildStoryboardPromptPack, activities.PromptPackInput{EpisodeID: in.EpisodeID, Scenes: in.Scenes}).Get(ctx, &promptPack); err != nil {
		return blockErr(res, "prompt_pack_failed", err)
	}

	var storyboard shortform.StoryboardImageManifest
	if err := workflow.ExecuteActivity(ctx, a.ImportStoryboardImages, activities.ImportStoryboardInput{EpisodeID: in.EpisodeID, Now: now, Operator: in.Operator, Scenes: in.Scenes}).Get(ctx, &storyboard); err != nil {
		return blockErr(res, "storyboard_import_failed", err)
	}
	stamp(&storyboard)
	record("storyboard_imported", "imported storyboard images (draft)")

	// Human-in-the-loop: storyboard image approval.
	record("awaiting_storyboard_approval", "waiting for storyboard image approval")
	sbApproval := receiveApproval(ctx, StoryboardImageApprovalSignal)
	if sbApproval.Decision != "approve" {
		return block(res, "storyboard_rejected", "storyboard image approval was not granted")
	}
	if err := workflow.ExecuteActivity(ctx, a.ApproveStoryboardImages, activities.ApproveStoryboardInput{Manifest: &storyboard, Approver: sbApproval.Approver, Now: now}).Get(ctx, &storyboard); err != nil {
		return blockErr(res, "storyboard_approval_failed", err)
	}
	stamp(&storyboard)

	// No self-approval (§4.6) + storyboard image gate (§8).
	if gate(gates.SelfApprovalGate(gates.SelfApprovalInput{CreatedBy: storyboard.CreatedBy, ApproverID: sbApproval.Approver, TargetStatus: shortform.StatusApproved})) {
		return res, nil
	}
	if gate(gates.StoryboardImageGate(gates.StoryboardImageInput{Manifest: &storyboard, RequiredScenes: sceneIDs(in.Scenes)})) {
		return res, nil
	}
	record("storyboard_approved", "storyboard images approved and gated")

	// Visual shots.
	var shots shortform.VisualShotManifest
	if err := workflow.ExecuteActivity(ctx, a.GenerateMockVisualShots, activities.VisualShotsInput{EpisodeID: in.EpisodeID, Now: now, Storyboard: &storyboard}).Get(ctx, &shots); err != nil {
		return blockErr(res, "visual_shots_failed", err)
	}
	if err := workflow.ExecuteActivity(ctx, a.ApproveVisualShots, &shots, now).Get(ctx, &shots); err != nil {
		return blockErr(res, "visual_shots_approval_failed", err)
	}
	stamp(&shots)
	if gate(gates.VisualShotGate(gates.VisualShotInput{Manifest: &shots, ApprovedImageHashes: approvedImageHashes(&storyboard), KnownScenes: sceneSet(in.Scenes)})) {
		return res, nil
	}
	record("visual_shots_approved", "visual shots generated, approved, and gated")

	// Voiceover.
	var voiceover shortform.VoiceoverManifest
	if err := workflow.ExecuteActivity(ctx, a.GenerateElevenLabsVoiceover, activities.VoiceoverInput{EpisodeID: in.EpisodeID, Now: now, ScriptRef: in.ScriptRef, Language: in.Language}).Get(ctx, &voiceover); err != nil {
		return blockErr(res, "voiceover_failed", err)
	}
	if err := workflow.ExecuteActivity(ctx, a.ApproveVoiceover, &voiceover, now).Get(ctx, &voiceover); err != nil {
		return blockErr(res, "voiceover_approval_failed", err)
	}
	stamp(&voiceover)
	record("voiceover_approved", "voiceover synthesized and approved")

	// Subtitles.
	var subtitles shortform.SubtitleManifest
	if err := workflow.ExecuteActivity(ctx, a.GenerateSubtitles, activities.SubtitlesInput{EpisodeID: in.EpisodeID, Now: now, Voiceover: &voiceover, Language: in.Language, WordTimestampsRequired: true}).Get(ctx, &subtitles); err != nil {
		return blockErr(res, "subtitles_failed", err)
	}
	if err := workflow.ExecuteActivity(ctx, a.ApproveSubtitles, &subtitles, now).Get(ctx, &subtitles); err != nil {
		return blockErr(res, "subtitles_approval_failed", err)
	}
	stamp(&subtitles)
	if gate(gates.SubtitleGate(gates.SubtitleInput{Manifest: &subtitles})) {
		return res, nil
	}
	record("subtitles_approved", "subtitles generated, approved, and gated")

	// Render + production QA.
	var render shortform.ShortRenderManifest
	if err := workflow.ExecuteActivity(ctx, a.RenderShortFinal, activities.RenderInput{EpisodeID: in.EpisodeID, Now: now, Shots: &shots, Voiceover: &voiceover, Subtitles: &subtitles, Platforms: in.Platforms}).Get(ctx, &render); err != nil {
		return blockErr(res, "render_failed", err)
	}
	var qa activities.ProductionQAResult
	if err := workflow.ExecuteActivity(ctx, a.RunProductionQA, &render).Get(ctx, &qa); err != nil {
		return blockErr(res, "production_qa_failed", err)
	}
	if err := workflow.ExecuteActivity(ctx, a.ApproveRenderOutputs, &render, now).Get(ctx, &render); err != nil {
		return blockErr(res, "render_approval_failed", err)
	}
	stamp(&render)
	if gate(gates.RenderGate(gates.RenderInput{Manifest: &render, ProductionQADecision: qa.Decision})) {
		return res, nil
	}
	record("render_approved", "render produced, QA "+qa.Decision+", and gated")

	// Assemble immutable production candidate from concrete component descriptors.
	components := []shortform.ComponentRef{
		shortform.ComponentRefOf(&storyboard), shortform.ComponentRefOf(&shots),
		shortform.ComponentRefOf(&voiceover), shortform.ComponentRefOf(&subtitles),
		shortform.ComponentRefOf(&render),
	}
	var candidate shortform.ProductionCandidate
	if err := workflow.ExecuteActivity(ctx, a.AssembleProductionCandidate, activities.AssembleCandidateInput{
		EpisodeID: in.EpisodeID, CandidateID: in.EpisodeID + "-cand-001", Now: now, Components: components,
	}).Get(ctx, &candidate); err != nil {
		return blockErr(res, "candidate_assembly_failed", err)
	}
	stamp(&candidate)
	record("production_candidate_locked", "assembled locked production candidate")

	// Human-in-the-loop: release approval.
	record("awaiting_release_approval", "waiting for release approval")
	relApproval := receiveApproval(ctx, ReleaseApprovalSignal)
	if relApproval.Decision != "approve" {
		return block(res, "release_denied", "release approval was not granted")
	}

	var release shortform.ReleaseApproval
	if err := workflow.ExecuteActivity(ctx, a.BuildReleaseApproval, shortform.BuildReleaseApprovalInput{
		EpisodeID: in.EpisodeID, CandidateID: candidate.CandidateID, Approver: relApproval.Approver, Now: now,
		Platforms: in.Platforms, Visibility: in.Visibility, ScheduledAt: in.ScheduledAt,
		AIDisclosureRequired: in.AIDisclosureRequired, AIDisclosure: in.AIDisclosure, ProductionQARef: "production_qa_report.json",
	}).Get(ctx, &release); err != nil {
		return blockErr(res, "release_build_failed", err)
	}
	stamp(&release)

	// No single-model final authority (§4.7): require >=2 distinct verifiers.
	if gate(gates.MultiVerifierGate(gates.MultiVerifierInput{Verifiers: []string{sbApproval.Approver, relApproval.Approver, "system:production-qa"}})) {
		return res, nil
	}

	// Guarded publish manifest + dry-run + release gate.
	var publish shortform.UploadPostPublishManifest
	if err := workflow.ExecuteActivity(ctx, a.GenerateUploadPostPublishManifest, activities.PublishManifestInput{
		EpisodeID: in.EpisodeID, Now: now, Release: &release, Render: &render,
		ProductionQADecision: qa.Decision, ProductionQARef: "production_qa_report.json",
	}).Get(ctx, &publish); err != nil {
		return blockErr(res, "publish_manifest_failed", err)
	}
	stamp(&publish)

	var dryRun activities.DryRunResult
	if err := workflow.ExecuteActivity(ctx, a.UploadPostDryRun, &publish).Get(ctx, &dryRun); err != nil {
		return blockErr(res, "dry_run_failed", err)
	}
	if gate(gates.ReleaseGate(gates.ReleaseInput{PublishManifest: &publish, ProductionQADecision: qa.Decision, DryRunSucceeded: dryRun.OK})) {
		return res, nil
	}

	record("published_dry_run_complete", "guarded upload-post dry-run complete (no upload performed)")
	return res, nil
}

func receiveApproval(ctx workflow.Context, signalName string) ApprovalSignal {
	var sig ApprovalSignal
	workflow.GetSignalChannel(ctx, signalName).Receive(ctx, &sig)
	return sig
}

func block(res ShortFormResult, state, reason string) (ShortFormResult, error) {
	res.State = state
	res.Blocked = true
	res.BlockReason = reason
	res.Notes = append(res.Notes, "blocked: "+reason)
	return res, nil
}

func blockErr(res ShortFormResult, state string, err error) (ShortFormResult, error) {
	res.State = state
	res.Blocked = true
	res.BlockReason = err.Error()
	res.Notes = append(res.Notes, state+": "+err.Error())
	return res, err
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

func approvedImageHashes(m *shortform.StoryboardImageManifest) map[string]bool {
	out := map[string]bool{}
	for _, img := range m.Images {
		if img.Status == shortform.StatusApproved {
			out[img.ImageHash] = true
		}
	}
	return out
}
