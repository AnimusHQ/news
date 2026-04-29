package mock

import (
	"context"
	"testing"

	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/models/adapters"
)

func TestProviderReturnsDeterministicReview(t *testing.T) {
	provider := Provider{
		ModelID:    "mock-tech",
		ProviderID: "local-mock",
		Task:       "technical verification",
		Verdict:    council.VerdictRequestRevision,
		Confidence: 0.64,
		Notes:      "needs source locator",
	}

	response, err := provider.Run(context.Background(), adapters.Request{})
	if err != nil {
		t.Fatalf("run provider: %v", err)
	}
	if response.Review.ModelID != "mock-tech" {
		t.Fatalf("unexpected model id: %s", response.Review.ModelID)
	}
	if response.Review.Verdict != council.VerdictRequestRevision {
		t.Fatalf("unexpected verdict: %s", response.Review.Verdict)
	}
	if response.Review.Notes != "needs source locator" {
		t.Fatalf("unexpected notes: %s", response.Review.Notes)
	}
}

func TestProviderReturnsConfiguredError(t *testing.T) {
	provider := Provider{Err: adapters.ErrProviderUnavailable{ProviderID: "local", Reason: "test"}}
	_, err := provider.Run(context.Background(), adapters.Request{})
	if err == nil {
		t.Fatal("expected configured error")
	}
}
