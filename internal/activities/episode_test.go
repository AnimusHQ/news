package activities

import (
	"context"
	"strings"
	"testing"
)

// TestPlaceholderActivitiesAreOfflineAndDeterministic exercises the long-form
// activities the worker registers (MockCouncilActivity, ProductionQAActivity,
// DryRunPublishActivity). They must return deterministic, non-empty results with
// no network, secrets, or live provider calls — they are safe placeholders, not
// real provider execution.
func TestPlaceholderActivitiesAreOfflineAndDeterministic(t *testing.T) {
	ctx := context.Background()

	council, err := MockCouncilActivity(ctx, "episode-0001")
	if err != nil {
		t.Fatalf("MockCouncilActivity error: %v", err)
	}
	if strings.TrimSpace(council) == "" {
		t.Fatal("MockCouncilActivity returned empty result")
	}

	qa, err := ProductionQAActivity(ctx, "episode-0001")
	if err != nil {
		t.Fatalf("ProductionQAActivity error: %v", err)
	}
	if strings.TrimSpace(qa) == "" {
		t.Fatal("ProductionQAActivity returned empty result")
	}

	publish, err := DryRunPublishActivity(ctx, "episode-0001")
	if err != nil {
		t.Fatalf("DryRunPublishActivity error: %v", err)
	}
	// The publish placeholder must never imply a real upload.
	if !strings.Contains(publish, "dry-run") {
		t.Fatalf("DryRunPublishActivity must report a dry-run, got %q", publish)
	}

	// Determinism: the same input yields the same output.
	again, err := MockCouncilActivity(ctx, "episode-0001")
	if err != nil || again != council {
		t.Fatalf("MockCouncilActivity not deterministic: %q vs %q (err=%v)", council, again, err)
	}
}
