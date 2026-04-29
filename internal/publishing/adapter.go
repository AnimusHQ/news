package publishing

import (
	"context"
	"fmt"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// PublishResult is returned by publishing adapters.
type PublishResult struct {
	EpisodeID string    `json:"episode_id"`
	Provider  string    `json:"provider"`
	DraftID   string    `json:"draft_id"`
	Visibility artifacts.PublishVisibility `json:"visibility"`
	CreatedAt time.Time `json:"created_at"`
	Notes     []string  `json:"notes,omitempty"`
}

// Adapter is the provider-agnostic publishing interface.
type Adapter interface {
	UploadPrivateDraft(ctx context.Context, pack Pack) (PublishResult, error)
	ScheduleDraft(ctx context.Context, pack Pack) (PublishResult, error)
}

// DryRunAdapter records publishing intent without network calls or uploads.
type DryRunAdapter struct {
	Provider string
	Now      func() time.Time
}

func (a DryRunAdapter) UploadPrivateDraft(ctx context.Context, pack Pack) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if pack.EpisodeID == "" {
		return PublishResult{}, fmt.Errorf("episode id is required")
	}
	if pack.Visibility == artifacts.PublishVisibilityPublic {
		return PublishResult{}, fmt.Errorf("dry-run adapter refuses public uploads")
	}
	if pack.Visibility == artifacts.PublishVisibilityScheduled && !pack.HumanApproved {
		return PublishResult{}, fmt.Errorf("scheduled visibility requires human release approval")
	}
	return a.result(pack, "private draft dry-run created"), nil
}

func (a DryRunAdapter) ScheduleDraft(ctx context.Context, pack Pack) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if pack.EpisodeID == "" {
		return PublishResult{}, fmt.Errorf("episode id is required")
	}
	if !pack.HumanApproved {
		return PublishResult{}, fmt.Errorf("scheduling requires human release approval")
	}
	if pack.Visibility != artifacts.PublishVisibilityScheduled {
		return PublishResult{}, fmt.Errorf("schedule requires scheduled visibility")
	}
	return a.result(pack, "scheduled draft dry-run created"), nil
}

func (a DryRunAdapter) result(pack Pack, note string) PublishResult {
	now := time.Now
	if a.Now != nil {
		now = a.Now
	}
	provider := a.Provider
	if provider == "" {
		provider = "local-dry-run"
	}
	return PublishResult{
		EpisodeID: pack.EpisodeID,
		Provider:  provider,
		DraftID:   "dry-run-" + pack.EpisodeID,
		Visibility: pack.Visibility,
		CreatedAt: now(),
		Notes:     []string{note, "no network call performed", "no public upload performed"},
	}
}
