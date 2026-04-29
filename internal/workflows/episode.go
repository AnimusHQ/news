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

// EpisodeWorkflowState is returned by workflow queries.
type EpisodeWorkflowState struct {
	EpisodeID string   `json:"episode_id"`
	State     string   `json:"state"`
	Notes     []string `json:"notes"`
}

// HumanQADecisionSignalName is the signal used to continue after human QA.
const HumanQADecisionSignalName = "HumanQADecisionSignal"

// ReleaseApprovalSignalName is the signal used to continue after release approval.
const ReleaseApprovalSignalName = "ReleaseApprovalSignal"

// GetEpisodeStateQueryName returns current workflow state.
const GetEpisodeStateQueryName = "GetEpisodeStateQuery"

// EpisodeLifecycleWorkflow is the canonical durable workflow for an episode.
//
// Workflow code must remain deterministic. Provider calls, file I/O, rendering,
// publishing, and model execution must be implemented as activities.
func EpisodeLifecycleWorkflow(ctx workflow.Context, input EpisodeWorkflowInput) (EpisodeWorkflowResult, error) {
	state := EpisodeWorkflowState{EpisodeID: input.EpisodeID, State: "started"}
	if err := workflow.SetQueryHandler(ctx, GetEpisodeStateQueryName, func() (EpisodeWorkflowState, error) {
		return state, nil
	}); err != nil {
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: "blocked", Notes: []string{"failed to install state query handler"}}, err
	}

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

	state.State = "validating_artifacts"
	var validation string
	if err := workflow.ExecuteActivity(ctx, "ValidateEpisodeActivity", input.EpisodeDir).Get(ctx, &validation); err != nil {
		state.State = "blocked"
		state.Notes = append(state.Notes, "artifact validation failed")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, err
	}
	state.Notes = append(state.Notes, validation)

	state.State = "running_model_council"
	var council string
	if err := workflow.ExecuteActivity(ctx, "MockCouncilActivity", input.EpisodeID).Get(ctx, &council); err != nil {
		state.State = "blocked"
		state.Notes = append(state.Notes, "mock council failed")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, err
	}
	state.Notes = append(state.Notes, council)

	state.State = "awaiting_human_qa"
	var humanDecision string
	humanSignal := workflow.GetSignalChannel(ctx, HumanQADecisionSignalName)
	humanSignal.Receive(ctx, &humanDecision)
	if humanDecision != "approve" && humanDecision != "approve_with_minor_edits" {
		state.State = "blocked"
		state.Notes = append(state.Notes, "human QA did not approve")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, nil
	}
	state.Notes = append(state.Notes, "human QA approved")

	state.State = "production_qa"
	var productionQA string
	if err := workflow.ExecuteActivity(ctx, "ProductionQAActivity", input.EpisodeID).Get(ctx, &productionQA); err != nil {
		state.State = "blocked"
		state.Notes = append(state.Notes, "production QA failed")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, err
	}
	state.Notes = append(state.Notes, productionQA)

	state.State = "awaiting_release_approval"
	var releaseDecision string
	releaseSignal := workflow.GetSignalChannel(ctx, ReleaseApprovalSignalName)
	releaseSignal.Receive(ctx, &releaseDecision)
	if releaseDecision != "approve" {
		state.State = "blocked"
		state.Notes = append(state.Notes, "release approval denied")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, nil
	}
	state.Notes = append(state.Notes, "release approved")

	state.State = "dry_run_publishing"
	var publish string
	if err := workflow.ExecuteActivity(ctx, "DryRunPublishActivity", input.EpisodeID).Get(ctx, &publish); err != nil {
		state.State = "blocked"
		state.Notes = append(state.Notes, "dry-run publish failed")
		return EpisodeWorkflowResult{EpisodeID: input.EpisodeID, State: state.State, Notes: state.Notes}, err
	}
	state.Notes = append(state.Notes, publish)
	state.State = "dry_run_complete"

	return EpisodeWorkflowResult{
		EpisodeID: input.EpisodeID,
		State:     state.State,
		Notes:     state.Notes,
	}, nil
}
