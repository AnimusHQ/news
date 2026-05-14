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
	assertAdapterCode(t, err, AdapterErrorVisibilityNotAllowed)
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
	assertAdapterCode(t, err, AdapterErrorVisibilityNotAllowed)
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

func TestDryRunAdapterValidatesMetadata(t *testing.T) {
	err := DryRunAdapter{}.ValidateMetadata(context.Background(), Pack{})
	if err == nil {
		t.Fatal("expected missing episode id to fail")
	}
	assertAdapterCode(t, err, AdapterErrorMetadataInvalid)
}

func TestDryRunAdapterGetsDraftStatusOffline(t *testing.T) {
	status, err := DryRunAdapter{Provider: "test-provider"}.GetDraftStatus(context.Background(), "draft-1")
	if err != nil {
		t.Fatalf("get draft status failed: %v", err)
	}
	if status.Provider != "test-provider" {
		t.Fatalf("expected provider to be preserved, got %s", status.Provider)
	}
	if status.Status != "dry_run_created" {
		t.Fatalf("unexpected status: %s", status.Status)
	}
}

func TestDryRunAdapterStatusRequiresDraftID(t *testing.T) {
	_, err := DryRunAdapter{}.GetDraftStatus(context.Background(), "")
	if err == nil {
		t.Fatal("expected missing draft id to fail")
	}
	assertAdapterCode(t, err, AdapterErrorMetadataInvalid)
}

func assertAdapterCode(t *testing.T, err error, code AdapterErrorCode) {
	t.Helper()
	adapterErr, ok := err.(AdapterError)
	if !ok {
		t.Fatalf("expected AdapterError, got %T: %v", err, err)
	}
	if adapterErr.Code != code {
		t.Fatalf("expected code %s, got %s", code, adapterErr.Code)
	}
}
