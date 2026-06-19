package worker

import (
	"reflect"
	"strings"
	"testing"
	"time"

	shortformactivities "github.com/AnimusHQ/news/internal/shortform/activities"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/workflows"
	"go.temporal.io/sdk/testsuite"
)

// commercialVendors are vendor names that must never appear in a workflow-visible
// (registered) short-form activity name. The provider layer is the sanctioned
// vendor boundary; the workflow layer must stay capability-named (WI-4 / ADR-0014).
var commercialVendors = []string{"elevenlabs", "uploadpost", "upload_post", "seedance", "omnivoice", "chatterbox"}

// TestRegisteredShortFormActivityNamesAreVendorNeutral reflects over the activity
// set the worker registers (NewMockActivities is registered by struct, so every
// exported method becomes a registered Temporal activity name) and asserts no
// registered name encodes a commercial vendor. This permanently guards the
// capability-naming invariant.
func TestRegisteredShortFormActivityNamesAreVendorNeutral(t *testing.T) {
	typ := reflect.TypeOf(shortformactivities.NewMockActivities()) // *Activities
	if typ.NumMethod() == 0 {
		t.Fatal("expected exported activity methods on *Activities")
	}
	for i := 0; i < typ.NumMethod(); i++ {
		name := strings.ToLower(typ.Method(i).Name)
		for _, vendor := range commercialVendors {
			if strings.Contains(name, vendor) {
				t.Fatalf("registered activity %q encodes commercial vendor %q", typ.Method(i).Name, vendor)
			}
		}
	}
}

// TestCapabilityNamedActivitiesExist locks the capability names the workflow
// invokes so a rename cannot silently drop or vendor-name them.
func TestCapabilityNamedActivitiesExist(t *testing.T) {
	typ := reflect.TypeOf(shortformactivities.NewMockActivities())
	want := []string{
		"GenerateVisualShotsMock", "GenerateVisualShotsReal", "GenerateVoiceover",
		"GeneratePublishManifest", "PublishDryRun", "PublishSchedule",
	}
	for _, name := range want {
		if _, ok := typ.MethodByName(name); !ok {
			t.Fatalf("expected registered activity %q to exist", name)
		}
	}
}

// TestWorkerRegistrationMatchesWorkflowUsage drives the short-form workflow in the
// Temporal test environment using exactly the registration the worker performs
// (the ShortFormWorkflow plus the NewMockActivities activity set). If a registered
// activity name ever diverges from a name the workflow invokes, the test
// environment cannot resolve the activity and the workflow fails — so this test
// is the registration/usage contract guard. It is fully offline (mock activities,
// no network, no secrets).
func TestWorkerRegistrationMatchesWorkflowUsage(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestWorkflowEnvironment()
	env.SetStartTime(time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC))

	// Mirror worker.Run: register the same activity set the worker registers.
	env.RegisterActivity(shortformactivities.NewMockActivities())

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(workflows.StoryboardImageApprovalSignal, workflows.ApprovalSignal{Decision: "approve", Approver: "human:editor"})
	}, time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(workflows.ReleaseApprovalSignal, workflows.ApprovalSignal{Decision: "approve", Approver: "human:reviewer"})
	}, 2*time.Second)

	env.ExecuteWorkflow(workflows.ShortFormWorkflow, workflows.ShortFormInput{
		EpisodeID:            "episode-0001",
		Scenes:               []providers.SceneSpec{{SceneID: "scene-001", StartSec: 0, EndSec: 5}},
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
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("registered activities did not satisfy workflow usage: %v", err)
	}
	var res workflows.ShortFormResult
	if err := env.GetWorkflowResult(&res); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if res.Blocked || res.State != "published_dry_run_complete" {
		t.Fatalf("expected published_dry_run_complete, got state=%s blocked=%v", res.State, res.Blocked)
	}
}
