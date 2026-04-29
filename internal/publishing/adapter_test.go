package publishing

import (
	"context"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestDryRunAdapterUploadsPrivateDraft(t *testing.T) {
	adapter := DryRunAdapter{Now: func() time.Time { return time.Unix(100, 0).UTC() }}
	result, err := adapter.UploadPrivateDraft(context.Background(), Pack{
		EpisodeID:     "episode-1",
		Visibility:    artifacts.PublishVisibilityPrivate,
		HumanApproved: false,
	})
	if err != nil {
		t.Fatalf("upload private draft failed: %v", err)
	}
	if result.DraftID != "dry-run-episode-1" {
		t.Fatalf("unexpected draft id: %s", result.DraftID)
	}
	if result.Visibility != artifacts.PublishVisibilityPrivate {
		t.Fatalf("unexpected visibility: %s", result.Visibility)
	}
}

func TestDryRunAdapterRejectsPublicUpload(t *testing.T) {
	_, err := DryRunAdapter{}.UploadPrivateDraft(context.Background(), Pack{
		EpisodeID:  "episode-1",
		Visibility: artifacts.PublishVisibilityPublic,
	})
	if err == nil {
		t.Fatal("expected public upload to fail")
	}
}

func TestDryRunAdapterScheduleRequiresApproval(t *testing.T) {
	_, err := DryRunAdapter{}.ScheduleDraft(context.Background(), Pack{
		EpisodeID:     "episode-1",
		Visibility:    artifacts.PublishVisibilityScheduled,
		HumanApproved: false,
	})
	if err == nil {
		t.Fatal("expected schedule without approval to fail")
	}
}

func TestDryRunAdapterSchedulesApprovedDraft(t *testing.T) {
	_, err := DryRunAdapter{}.ScheduleDraft(context.Background(), Pack{
		EpisodeID:     "episode-1",
		Visibility:    artifacts.PublishVisibilityScheduled,
		HumanApproved: true,
	})
	if err != nil {
		t.Fatalf("expected approved scheduled draft: %v", err)
	}
}
