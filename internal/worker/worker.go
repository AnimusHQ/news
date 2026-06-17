package worker

import (
	"context"
	"fmt"

	"github.com/AnimusHQ/news/internal/activities"
	shortformactivities "github.com/AnimusHQ/news/internal/shortform/activities"
	"github.com/AnimusHQ/news/internal/workflows"
	"go.temporal.io/sdk/activity"
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

// Run starts a Temporal worker and blocks until the process is interrupted or
// the worker returns an error. The context is reserved for future graceful
// shutdown integration.
func Run(ctx context.Context, cfg Config) error {
	_ = ctx

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
	w.RegisterActivityWithOptions(activities.ValidateEpisodeActivity, activity.RegisterOptions{Name: "ValidateEpisodeActivity"})
	w.RegisterActivityWithOptions(activities.MockCouncilActivity, activity.RegisterOptions{Name: "MockCouncilActivity"})
	w.RegisterActivityWithOptions(activities.ProductionQAActivity, activity.RegisterOptions{Name: "ProductionQAActivity"})
	w.RegisterActivityWithOptions(activities.DryRunPublishActivity, activity.RegisterOptions{Name: "DryRunPublishActivity"})

	// Short-form pipeline (M1: mock activities).
	w.RegisterWorkflow(workflows.ShortFormWorkflow)
	w.RegisterActivity(shortformactivities.NewMockActivities())

	return w.Run(worker.InterruptCh())
}
