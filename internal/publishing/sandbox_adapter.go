package publishing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// SandboxAdapter simulates a private/scheduled platform integration behind the
// publishing Adapter interface. It performs no network calls and cannot create
// public uploads.
type SandboxAdapter struct {
	Provider string
	Now      func() time.Time

	mu     sync.Mutex
	drafts map[string]DraftStatus
}

func (a *SandboxAdapter) ValidateMetadata(ctx context.Context, pack Pack) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if pack.EpisodeID == "" {
		return adapterError(AdapterErrorMetadataInvalid, "episode id is required")
	}
	if len(pack.TitleCandidates) == 0 {
		return adapterError(AdapterErrorMetadataInvalid, "at least one title candidate is required")
	}
	if pack.Visibility == artifacts.PublishVisibilityPublic {
		return adapterError(AdapterErrorVisibilityNotAllowed, "sandbox adapter refuses public uploads")
	}
	if pack.Visibility != "" &&
		pack.Visibility != artifacts.PublishVisibilityPrivate &&
		pack.Visibility != artifacts.PublishVisibilityScheduled {
		return adapterError(AdapterErrorVisibilityNotAllowed, fmt.Sprintf("unsupported visibility: %s", pack.Visibility))
	}
	if pack.Visibility == artifacts.PublishVisibilityScheduled && !pack.HumanApproved {
		return adapterError(AdapterErrorVisibilityNotAllowed, "scheduled visibility requires human release approval")
	}
	return nil
}

func (a *SandboxAdapter) UploadPrivateDraft(ctx context.Context, pack Pack) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if pack.Visibility == "" {
		pack.Visibility = artifacts.PublishVisibilityPrivate
	}
	if pack.Visibility != artifacts.PublishVisibilityPrivate {
		return PublishResult{}, adapterError(AdapterErrorVisibilityNotAllowed, "private draft upload requires private visibility")
	}
	if err := a.ValidateMetadata(ctx, pack); err != nil {
		return PublishResult{}, err
	}
	result := a.result(pack, "private sandbox draft created")
	a.saveDraft(result, "sandbox_private_draft")
	return result, nil
}

func (a *SandboxAdapter) ScheduleDraft(ctx context.Context, pack Pack) (PublishResult, error) {
	if err := ctx.Err(); err != nil {
		return PublishResult{}, err
	}
	if pack.Visibility != artifacts.PublishVisibilityScheduled {
		return PublishResult{}, adapterError(AdapterErrorVisibilityNotAllowed, "schedule requires scheduled visibility")
	}
	if !pack.HumanApproved {
		return PublishResult{}, adapterError(AdapterErrorVisibilityNotAllowed, "scheduling requires human release approval")
	}
	if err := a.ValidateMetadata(ctx, pack); err != nil {
		return PublishResult{}, err
	}
	result := a.result(pack, "scheduled sandbox draft created")
	a.saveDraft(result, "sandbox_scheduled")
	return result, nil
}

func (a *SandboxAdapter) GetDraftStatus(ctx context.Context, draftID string) (DraftStatus, error) {
	if err := ctx.Err(); err != nil {
		return DraftStatus{}, err
	}
	if draftID == "" {
		return DraftStatus{}, adapterError(AdapterErrorMetadataInvalid, "draft id is required")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.drafts == nil {
		return DraftStatus{}, adapterError(AdapterErrorProcessingFailed, "draft not found")
	}
	status, ok := a.drafts[draftID]
	if !ok {
		return DraftStatus{}, adapterError(AdapterErrorProcessingFailed, "draft not found")
	}
	return status, nil
}

func (a *SandboxAdapter) result(pack Pack, note string) PublishResult {
	return PublishResult{
		EpisodeID:  pack.EpisodeID,
		Provider:   a.provider(),
		DraftID:    "sandbox-" + pack.EpisodeID,
		Visibility: pack.Visibility,
		CreatedAt:  a.now(),
		Notes:      []string{note, "no network call performed", "no public upload performed"},
	}
}

func (a *SandboxAdapter) saveDraft(result PublishResult, status string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.drafts == nil {
		a.drafts = map[string]DraftStatus{}
	}
	a.drafts[result.DraftID] = DraftStatus{
		DraftID:    result.DraftID,
		Provider:   result.Provider,
		Visibility: result.Visibility,
		Status:     status,
		Notes:      append([]string(nil), result.Notes...),
	}
}

func (a *SandboxAdapter) provider() string {
	if a == nil || a.Provider == "" {
		return "platform-sandbox"
	}
	return a.Provider
}

func (a *SandboxAdapter) now() time.Time {
	if a != nil && a.Now != nil {
		return a.Now().UTC()
	}
	return time.Now().UTC()
}
