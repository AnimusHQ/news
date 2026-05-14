package publishing

import (
	"context"
	"fmt"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// PublishResult is returned by publishing adapters.
type PublishResult struct {
	EpisodeID  string                      `json:"episode_id"`
	Provider   string                      `json:"provider"`
	DraftID    string                      `json:"draft_id"`
	Visibility artifacts.PublishVisibility `json:"visibility"`
	CreatedAt  time.Time                   `json:"created_at"`
	Notes      []string                    `json:"notes,omitempty"`
}

// AdapterErrorCode normalizes provider-specific publishing failures.
type AdapterErrorCode string

const (
	AdapterErrorAuthMissing          AdapterErrorCode = "auth_missing"
	AdapterErrorUploadFailed         AdapterErrorCode = "upload_failed"
	AdapterErrorProcessingFailed     AdapterErrorCode = "processing_failed"
	AdapterErrorPolicyBlocked        AdapterErrorCode = "policy_blocked"
	AdapterErrorVisibilityNotAllowed AdapterErrorCode = "visibility_not_allowed"
	AdapterErrorMetadataInvalid      AdapterErrorCode = "metadata_invalid"
)

// AdapterError is a normalized publishing adapter error.
type AdapterError struct {
	Code    AdapterErrorCode `json:"code"`
	Message string           `json:"message"`
}

func (e AdapterError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// DraftStatus is a provider-agnostic view of a draft state.
type DraftStatus struct {
	DraftID    string                      `json:"draft_id"`
	Provider   string                      `json:"provider"`
	Visibility artifacts.PublishVisibility `json:"visibility"`
	Status     string                      `json:"status"`
	Notes      []string                    `json:"notes,omitempty"`
}

// Adapter is the provider-agnostic publishing interface.
type Adapter interface {
	ValidateMetadata(ctx context.Context, pack Pack) error
	UploadPrivateDraft(ctx context.Context, pack Pack) (PublishResult, error)
	ScheduleDraft(ctx context.Context, pack Pack) (PublishResult, error)
	GetDraftStatus(ctx context.Context, draftID string) (DraftStatus, error)
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
	if err := a.ValidateMetadata(ctx, pack); err != nil {
		return PublishResult{}, err
	}
	if pack.Visibility == "" {
		pack.Visibility = artifacts.PublishVisibilityPrivate
	}
	return a.result(pack, "private draft dry-run created"), nil
}

func (a DryRunAdapter) ScheduleDraft(ctx context.Context, pack Pack) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if !pack.HumanApproved {
		return PublishResult{}, adapterError(AdapterErrorVisibilityNotAllowed, "scheduling requires human release approval")
	}
	if pack.Visibility != artifacts.PublishVisibilityScheduled {
		return PublishResult{}, adapterError(AdapterErrorVisibilityNotAllowed, "schedule requires scheduled visibility")
	}
	if err := a.ValidateMetadata(ctx, pack); err != nil {
		return PublishResult{}, err
	}
	return a.result(pack, "scheduled draft dry-run created"), nil
}

func (a DryRunAdapter) ValidateMetadata(ctx context.Context, pack Pack) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if pack.EpisodeID == "" {
		return adapterError(AdapterErrorMetadataInvalid, "episode id is required")
	}
	if pack.Visibility == artifacts.PublishVisibilityPublic {
		return adapterError(AdapterErrorVisibilityNotAllowed, "dry-run adapter refuses public uploads")
	}
	if pack.Visibility == artifacts.PublishVisibilityScheduled && !pack.HumanApproved {
		return adapterError(AdapterErrorVisibilityNotAllowed, "scheduled visibility requires human release approval")
	}
	if pack.Visibility != "" && pack.Visibility != artifacts.PublishVisibilityPrivate && pack.Visibility != artifacts.PublishVisibilityScheduled {
		return adapterError(AdapterErrorVisibilityNotAllowed, fmt.Sprintf("unsupported visibility: %s", pack.Visibility))
	}
	return nil
}

func (a DryRunAdapter) GetDraftStatus(ctx context.Context, draftID string) (DraftStatus, error) {
	if err := ctx.Err(); err != nil {
		return DraftStatus{}, err
	}
	if draftID == "" {
		return DraftStatus{}, adapterError(AdapterErrorMetadataInvalid, "draft id is required")
	}
	return DraftStatus{
		DraftID:    draftID,
		Provider:   a.provider(),
		Visibility: artifacts.PublishVisibilityPrivate,
		Status:     "dry_run_created",
		Notes:      []string{"offline dry-run status", "no platform lookup performed"},
	}, nil
}

func (a DryRunAdapter) result(pack Pack, note string) PublishResult {
	now := time.Now
	if a.Now != nil {
		now = a.Now
	}
	return PublishResult{
		EpisodeID:  pack.EpisodeID,
		Provider:   a.provider(),
		DraftID:    "dry-run-" + pack.EpisodeID,
		Visibility: pack.Visibility,
		CreatedAt:  now(),
		Notes:      []string{note, "no network call performed", "no public upload performed"},
	}
}

func (a DryRunAdapter) provider() string {
	if a.Provider == "" {
		return "local-dry-run"
	}
	return a.Provider
}

func adapterError(code AdapterErrorCode, message string) AdapterError {
	return AdapterError{Code: code, Message: message}
}
