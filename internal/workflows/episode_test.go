package workflows

import (
	"testing"

	"github.com/AnimusHQ/news/internal/activities"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestEpisodeLifecycleWorkflowCompletesDryRun(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerEpisodeWorkflowTestHandlers(env)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(HumanQADecisionSignalName, "approve")
	}, 0)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ReleaseApprovalSignalName, "approve")
	}, 0)

	env.ExecuteWorkflow(EpisodeLifecycleWorkflow, EpisodeWorkflowInput{
		EpisodeID:  "episode-test",
		EpisodeDir: "test-fixture",
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("expected workflow to complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow failed: %v", err)
	}

	var result EpisodeWorkflowResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if result.State != "dry_run_complete" {
		t.Fatalf("expected dry_run_complete, got %s", result.State)
	}
}

func TestEpisodeLifecycleWorkflowBlocksOnHumanQARejection(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerEpisodeWorkflowTestHandlers(env)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(HumanQADecisionSignalName, "block")
	}, 0)

	env.ExecuteWorkflow(EpisodeLifecycleWorkflow, EpisodeWorkflowInput{
		EpisodeID:  "episode-test",
		EpisodeDir: "test-fixture",
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("expected workflow to complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow returned unexpected error: %v", err)
	}

	var result EpisodeWorkflowResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if result.State != "blocked" {
		t.Fatalf("expected blocked, got %s", result.State)
	}
}

func TestEpisodeLifecycleWorkflowBlocksOnReleaseRejection(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerEpisodeWorkflowTestHandlers(env)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(HumanQADecisionSignalName, "approve")
	}, 0)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(ReleaseApprovalSignalName, "block")
	}, 0)

	env.ExecuteWorkflow(EpisodeLifecycleWorkflow, EpisodeWorkflowInput{
		EpisodeID:  "episode-test",
		EpisodeDir: "test-fixture",
	})

	if !env.IsWorkflowCompleted() {
		t.Fatal("expected workflow to complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow returned unexpected error: %v", err)
	}

	var result EpisodeWorkflowResult
	if err := env.GetWorkflowResult(&result); err != nil {
		t.Fatalf("get workflow result: %v", err)
	}
	if result.State != "blocked" {
		t.Fatalf("expected blocked, got %s", result.State)
	}
}

type workflowTestEnvironment interface {
	RegisterWorkflow(workflowFunc any)
	RegisterActivityWithOptions(activityFunc any, options activity.RegisterOptions)
}

func registerEpisodeWorkflowTestHandlers(env workflowTestEnvironment) {
	env.RegisterWorkflow(EpisodeLifecycleWorkflow)
	env.RegisterActivityWithOptions(func(string) (string, error) {
		return "artifact validation passed", nil
	}, activity.RegisterOptions{Name: "ValidateEpisodeActivity"})
	env.RegisterActivityWithOptions(activities.MockCouncilActivity, activity.RegisterOptions{Name: "MockCouncilActivity"})
	env.RegisterActivityWithOptions(activities.ProductionQAActivity, activity.RegisterOptions{Name: "ProductionQAActivity"})
	env.RegisterActivityWithOptions(activities.DryRunPublishActivity, activity.RegisterOptions{Name: "DryRunPublishActivity"})
}
