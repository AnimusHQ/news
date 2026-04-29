package worker

import (
	"context"
	"fmt"

	"github.com/AnimusHQ/news/internal/activities"
	"github.com/AnimusHQ/news/internal/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const DefaultTaskQueue = "animus-news-episode"

// Config describes a local Temporal worker process.
type Config struct {
	TemporalAddress string
	Namespace       string
	TaskQueue       string
}

// Run starts a Temporal worker and blocks until context cancellation or worker error.
func Run(ctx context.Context, cfg Config) error {
	address := cfg.TemporalAddress
	if address == "" {
		address = client.DefaultHostPort
	}
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = client.DefaultNamespace
	}
	taskQueue := cfg.TaskQueue
	if taskQueue == "" {
		taskQueue = DefaultTaskQueue
	}

	c, err := client.Dial(client.Options{
		HostPort:  address,
		Namespace: namespace,
	})
	if err != nil {
		return fmt.Errorf("connect temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.EpisodeLifecycleWorkflow)
	w.RegisterActivityWithOptions(activities.ValidateEpisodeActivity, worker.RegisterActivityOptions{Name: "ValidateEpisodeActivity"})
	w.RegisterActivityWithOptions(activities.MockCouncilActivity, worker.RegisterActivityOptions{Name: "MockCouncilActivity"})
	w.RegisterActivityWithOptions(activities.ProductionQAActivity, worker.RegisterActivityOptions{Name: "ProductionQAActivity"})
	w.RegisterActivityWithOptions(activities.DryRunPublishActivity, worker.RegisterActivityOptions{Name: "DryRunPublishActivity"})

	return w.Run(worker.InterruptCh())
}
