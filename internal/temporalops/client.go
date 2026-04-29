package temporalops

import (
	"context"
	"fmt"

	"github.com/AnimusHQ/news/internal/worker"
	"github.com/AnimusHQ/news/internal/workflows"
	"go.temporal.io/sdk/client"
)

// Config configures Temporal client operations.
type Config struct {
	TemporalAddress string
	Namespace       string
	TaskQueue       string
}

func dial(cfg Config) (client.Client, error) {
	address := cfg.TemporalAddress
	if address == "" {
		address = client.DefaultHostPort
	}
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = client.DefaultNamespace
	}
	return client.Dial(client.Options{HostPort: address, Namespace: namespace})
}

// StartEpisode starts the durable episode workflow.
func StartEpisode(ctx context.Context, cfg Config, episodeID string, episodeDir string) (client.WorkflowRun, error) {
	if episodeID == "" {
		return nil, fmt.Errorf("episode id is required")
	}
	if episodeDir == "" {
		return nil, fmt.Errorf("episode directory is required")
	}
	c, err := dial(cfg)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	taskQueue := cfg.TaskQueue
	if taskQueue == "" {
		taskQueue = worker.DefaultTaskQueue
	}

	return c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "animus-news-" + episodeID,
		TaskQueue: taskQueue,
	}, workflows.EpisodeLifecycleWorkflow, workflows.EpisodeWorkflowInput{EpisodeID: episodeID, EpisodeDir: episodeDir})
}

// SignalHumanQA sends the human QA decision signal.
func SignalHumanQA(ctx context.Context, cfg Config, workflowID string, decision string) error {
	return signal(ctx, cfg, workflowID, workflows.HumanQADecisionSignalName, decision)
}

// SignalRelease sends the release approval decision signal.
func SignalRelease(ctx context.Context, cfg Config, workflowID string, decision string) error {
	return signal(ctx, cfg, workflowID, workflows.ReleaseApprovalSignalName, decision)
}

func signal(ctx context.Context, cfg Config, workflowID string, signalName string, value string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow id is required")
	}
	if value == "" {
		return fmt.Errorf("signal value is required")
	}
	c, err := dial(cfg)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.SignalWorkflow(ctx, workflowID, "", signalName, value)
}

// QueryEpisodeState queries the current workflow state.
func QueryEpisodeState(ctx context.Context, cfg Config, workflowID string) (workflows.EpisodeWorkflowState, error) {
	if workflowID == "" {
		return workflows.EpisodeWorkflowState{}, fmt.Errorf("workflow id is required")
	}
	c, err := dial(cfg)
	if err != nil {
		return workflows.EpisodeWorkflowState{}, err
	}
	defer c.Close()

	value, err := c.QueryWorkflow(ctx, workflowID, "", workflows.GetEpisodeStateQueryName)
	if err != nil {
		return workflows.EpisodeWorkflowState{}, err
	}
	var state workflows.EpisodeWorkflowState
	if err := value.Get(&state); err != nil {
		return workflows.EpisodeWorkflowState{}, err
	}
	return state, nil
}
