package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// EpisodeWorkflowInput starts an episode production lifecycle.
type EpisodeWorkflowInput struct {
	EpisodeID  string
	EpisodeDir string
}

// EpisodeWorkflowResult summarizes the workflow execution.
type EpisodeWorkflowResult struct {
	EpisodeID string
	State     string
	Notes     []string
}

// HumanQADecisionSignalName is the signal used to continue after human QA.
const HumanQADecisionSignalName = "HumanQADecisionSignal"

// ReleaseApprovalSignalName is the signal used to continue after release approval.
const ReleaseApprovalSignalName = "ReleaseApprovalSignal"

// EpisodeLifecycleWorkflow is the canonical durable workflow for an episode.
//
// Workflow code must remain deterministic. Provider calls, file I/O, rendering,
// publishing, and model execution must be implemented as activities.
func EpisodeLifecycleWorkflow(ctx workflow.Context, input EpisodeWorkflowInput) (EpisodeWorkflowResult, error) {
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var validation string
	if err := workflow.ExecuteActivity(ctx, "ValidateEpisodeActivity", input.EpisodeDir).Get(ctx, &validation); err != nil {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"artifact validation failed"}}, err
	}

	var council string
	if err := workflow.ExecuteActivity(ctx, "MockCouncilActivity", input.EpisodeID).Get(ctx, &council); err != nil {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"mock council failed"}}, err
	}

	var humanDecision string
	humanSignal := workflow.GetSignalChannel(ctx, HumanQADecisionSignalName)
	humanSignal.Receive(ctx, &humanDecision)
	if humanDecision != "approve" && humanDecision != "approve_with_minor_edits" {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"human QA did not approve"}}, nil
	}

	var productionQA string
	if err := workflow.ExecuteActivity(ctx, "ProductionQAActivity", input.EpisodeID).Get(ctx, &productionQA); err != nil {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"production QA failed"}}, err
	}

	var releaseDecision string
	releaseSignal := workflow.GetSignalChannel(ctx, ReleaseApprovalSignalName)
	releaseSignal.Receive(ctx, &releaseDecision)
	if releaseDecision != "approve" {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"release approval denied"}}, nil
	}

	var publish string
	if err := workflow.ExecuteActivity(ctx, "DryRunPublishActivity", input.EpisodeID).Get(ctx, &publish); err != nil {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"dry-run publish failed"}}, err
	}

	return EpisodeWorkflowResult{
		EpisodeID: input.EpisodeID,
		State:     "dry_run_complete",
		Notes: []string{
			validation,
			council,
			"human QA approved",
			productionQA,
			"release approved",
			publish,
		},
	}, nil
}
