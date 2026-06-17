package workflows

import (
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/activities"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"go.temporal.io/sdk/testsuite"
)

var fixedStart = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func sampleInput() ShortFormInput {
	return ShortFormInput{
		EpisodeID:            "episode-0001",
		Scenes:               []providers.SceneSpec{{SceneID: "scene-001", StartSec: 0, EndSec: 5}, {SceneID: "scene-002", StartSec: 5, EndSec: 12}},
		ScriptRef:            "script.md",
		ResearchPackRef:      "research_pack.json",
		ScriptApproved:       true,
		Claims:               []gates.ClaimRef{{ID: "c1", SourceIDs: []string{"s1"}}},
		Platforms:            []string{"youtube"},
		Visibility:           "private",
		AIDisclosureRequired: true,
		AIDisclosure:         "AI-generated visuals and synthetic voice.",
		Language:             "en",
		Operator:             "operator:ci",
	}
}

func approveBoth(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(StoryboardImageApprovalSignal, ApprovalSignal{Decision: "approve", Approver: "human:editor"})
	}, time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ReleaseApprovalSignal, ApprovalSignal{Decision: "approve", Approver: "human:reviewer"})
	}, 2*time.Second)
}

func runWorkflow(t *testing.T, defects activities.MockDefects, schedule func(*testsuite.TestWorkflowEnvironment)) (ShortFormResult, error) {
	t.Helper()
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.SetStartTime(fixedStart)
	env.RegisterActivity(activities.NewMockActivitiesWithDefects(defects))
	schedule(env)
	env.ExecuteWorkflow(ShortFormWorkflow, sampleInput())
	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	err := env.GetWorkflowError()
	var res ShortFormResult
	if err == nil {
		if rerr := env.GetWorkflowResult(&res); rerr != nil {
			t.Fatalf("decode result: %v", rerr)
		}
	}
	return res, err
}

func TestShortFormWorkflowHappyPath(t *testing.T) {
	res, err := runWorkflow(t, activities.MockDefects{}, approveBoth)
	if err != nil {
		t.Fatalf("unexpected workflow error: %v", err)
	}
	if res.Blocked {
		t.Fatalf("expected success, blocked: %s", res.BlockReason)
	}
	if res.State != "published_dry_run_complete" {
		t.Fatalf("unexpected terminal state: %s", res.State)
	}
	wantKinds := []string{
		shortform.KindStoryboardImageManifest, shortform.KindVisualShotManifest,
		shortform.KindVoiceoverManifest, shortform.KindSubtitleManifest,
		shortform.KindShortRenderManifest, shortform.KindProductionCandidate,
		shortform.KindReleaseApproval, shortform.KindUploadPostPublishManifest,
	}
	for _, kind := range wantKinds {
		if res.Artifacts[kind] == "" {
			t.Fatalf("missing stamped artifact for %s", kind)
		}
	}
	for _, g := range res.GateResults {
		if g.Blocked() {
			t.Fatalf("gate %s unexpectedly blocked: %v", g.Gate, g.Reasons)
		}
	}
	if len(res.GateResults) < 6 {
		t.Fatalf("expected the full gate sequence, got %d gates", len(res.GateResults))
	}
}

func TestShortFormWorkflowStoryboardRejectedBlocks(t *testing.T) {
	res, err := runWorkflow(t, activities.MockDefects{}, func(env *testsuite.TestWorkflowEnvironment) {
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(StoryboardImageApprovalSignal, ApprovalSignal{Decision: "reject", Approver: "human:editor"})
		}, time.Second)
	})
	if err != nil {
		t.Fatalf("gate block must not be a workflow error: %v", err)
	}
	if !res.Blocked || res.State != "storyboard_rejected" {
		t.Fatalf("expected storyboard_rejected, got state=%s blocked=%v", res.State, res.Blocked)
	}
}

func TestShortFormWorkflowReleaseDeniedBlocks(t *testing.T) {
	res, err := runWorkflow(t, activities.MockDefects{}, func(env *testsuite.TestWorkflowEnvironment) {
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(StoryboardImageApprovalSignal, ApprovalSignal{Decision: "approve", Approver: "human:editor"})
		}, time.Second)
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(ReleaseApprovalSignal, ApprovalSignal{Decision: "deny", Approver: "human:reviewer"})
		}, 2*time.Second)
	})
	if err != nil {
		t.Fatalf("release denial must not be a workflow error: %v", err)
	}
	if !res.Blocked || res.State != "release_denied" {
		t.Fatalf("expected release_denied, got state=%s blocked=%v", res.State, res.Blocked)
	}
}

func TestShortFormWorkflowRenderDefectBlocksAtRenderGate(t *testing.T) {
	res, err := runWorkflow(t, activities.MockDefects{Render: providers.DefectRenderNoAudio}, func(env *testsuite.TestWorkflowEnvironment) {
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(StoryboardImageApprovalSignal, ApprovalSignal{Decision: "approve", Approver: "human:editor"})
		}, time.Second)
	})
	if err != nil {
		t.Fatalf("render gate block must not be a workflow error: %v", err)
	}
	if !res.Blocked {
		t.Fatal("expected workflow to block on render gate")
	}
	last := res.GateResults[len(res.GateResults)-1]
	if last.Gate != "render" {
		t.Fatalf("expected block at render gate, got %s", last.Gate)
	}
}

func TestShortFormWorkflowProviderErrorPropagates(t *testing.T) {
	_, err := runWorkflow(t, activities.MockDefects{Voice: providers.DefectError}, func(env *testsuite.TestWorkflowEnvironment) {
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(StoryboardImageApprovalSignal, ApprovalSignal{Decision: "approve", Approver: "human:editor"})
		}, time.Second)
	})
	if err == nil {
		t.Fatal("expected an injected provider error to surface as a workflow error")
	}
}

// TestShortFormWorkflowReplayIsDeterministic runs the signal-driven workflow
// twice with a fixed start time. The test environment replays accumulated
// history at each workflow-task boundary (signals + activity completions), so a
// non-deterministic workflow would error here. Byte-identical results across
// runs reinforce determinism.
func TestShortFormWorkflowReplayIsDeterministic(t *testing.T) {
	res1, err1 := runWorkflow(t, activities.MockDefects{}, approveBoth)
	res2, err2 := runWorkflow(t, activities.MockDefects{}, approveBoth)
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if res1.State != res2.State || res1.Blocked != res2.Blocked {
		t.Fatalf("non-deterministic terminal state: %+v vs %+v", res1, res2)
	}
	for kind, hash := range res1.Artifacts {
		if res2.Artifacts[kind] != hash {
			t.Fatalf("non-deterministic artifact hash for %s: %s vs %s", kind, hash, res2.Artifacts[kind])
		}
	}
	if len(res1.GateResults) != len(res2.GateResults) {
		t.Fatalf("non-deterministic gate sequence length: %d vs %d", len(res1.GateResults), len(res2.GateResults))
	}
}
