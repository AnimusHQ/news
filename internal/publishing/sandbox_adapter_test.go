package publishing

import (
	"context"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestSandboxAdapterCreatesPrivateDraftStatus(t *testing.T) {
	adapter := &SandboxAdapter{Provider: "sandbox-youtube", Now: fixedPublishTime}
	result, err := adapter.UploadPrivateDraft(context.Background(), validSandboxPack(artifacts.PublishVisibilityPrivate, false))
	if err != nil {
		t.Fatalf("upload private draft failed: %v", err)
	}
	if result.Provider != "sandbox-youtube" {
		t.Fatalf("unexpected provider: %s", result.Provider)
	}
	if result.Visibility != artifacts.PublishVisibilityPrivate {
		t.Fatalf("unexpected visibility: %s", result.Visibility)
	}
	status, err := adapter.GetDraftStatus(context.Background(), result.DraftID)
	if err != nil {
		t.Fatalf("get status failed: %v", err)
	}
	if status.Status != "sandbox_private_draft" {
		t.Fatalf("unexpected status: %s", status.Status)
	}
}

func TestSandboxAdapterSchedulesOnlyApprovedDrafts(t *testing.T) {
	adapter := &SandboxAdapter{Now: fixedPublishTime}
	_, err := adapter.ScheduleDraft(context.Background(), validSandboxPack(artifacts.PublishVisibilityScheduled, false))
	if err == nil {
		t.Fatal("expected scheduling without human approval to fail")
	}
	assertAdapterCode(t, err, AdapterErrorVisibilityNotAllowed)

	result, err := adapter.ScheduleDraft(context.Background(), validSandboxPack(artifacts.PublishVisibilityScheduled, true))
	if err != nil {
		t.Fatalf("schedule approved draft failed: %v", err)
	}
	status, err := adapter.GetDraftStatus(context.Background(), result.DraftID)
	if err != nil {
		t.Fatalf("get scheduled status failed: %v", err)
	}
	if status.Status != "sandbox_scheduled" {
		t.Fatalf("unexpected status: %s", status.Status)
	}
}

func TestSandboxAdapterRefusesPublicVisibility(t *testing.T) {
	adapter := &SandboxAdapter{}
	_, err := adapter.UploadPrivateDraft(context.Background(), validSandboxPack(artifacts.PublishVisibilityPublic, true))
	if err == nil {
		t.Fatal("expected public upload to fail")
	}
	assertAdapterCode(t, err, AdapterErrorVisibilityNotAllowed)
}

func TestSandboxAdapterValidatesMetadata(t *testing.T) {
	adapter := &SandboxAdapter{}
	err := adapter.ValidateMetadata(context.Background(), Pack{EpisodeID: "episode-1", Visibility: artifacts.PublishVisibilityPrivate})
	if err == nil {
		t.Fatal("expected missing title candidate to fail")
	}
	assertAdapterCode(t, err, AdapterErrorMetadataInvalid)
}

func TestSandboxAdapterStatusRequiresKnownDraft(t *testing.T) {
	_, err := (&SandboxAdapter{}).GetDraftStatus(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected missing draft to fail")
	}
	assertAdapterCode(t, err, AdapterErrorProcessingFailed)
}

func validSandboxPack(visibility artifacts.PublishVisibility, approved bool) Pack {
	return Pack{
		EpisodeID:       "episode-1",
		TitleCandidates: []string{"What Happens After git push?"},
		Description:     "A source-grounded sandbox draft.",
		Visibility:      visibility,
		HumanApproved:   approved,
	}
}

func fixedPublishTime() time.Time {
	return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
}
