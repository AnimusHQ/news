package council

import (
	"context"
	"testing"

	"github.com/AnimusHQ/news/internal/models/adapters"
	"github.com/AnimusHQ/news/internal/models/mock"
)

func TestRunnerAggregatesProviderReviews(t *testing.T) {
	runner := NewRunner([]adapters.Provider{
		mock.Provider{ModelID: "tech", ProviderID: "local", Task: "technical", Verdict: VerdictApprove, Confidence: 0.9},
		mock.Provider{ModelID: "editorial", ProviderID: "local", Task: "editorial", Verdict: VerdictApproveWithSuggestions, Confidence: 0.8, Notes: "tighten hook"},
	})

	report, err := runner.Run(context.Background(), adapters.Request{})
	if err != nil {
		t.Fatalf("runner failed: %v", err)
	}
	if report.Consensus != ConsensusApprovedWithSuggestions {
		t.Fatalf("expected approved_with_suggestions, got %s", report.Consensus)
	}
	if len(report.Reviews) != 2 {
		t.Fatalf("expected 2 reviews, got %d", len(report.Reviews))
	}
}

func TestRunnerStopsOnProviderError(t *testing.T) {
	runner := NewRunner([]adapters.Provider{
		mock.Provider{ModelID: "tech", ProviderID: "local", Verdict: VerdictApprove},
		mock.Provider{ModelID: "broken", ProviderID: "local", Err: adapters.ErrProviderUnavailable{ProviderID: "local", Reason: "test"}},
	})

	_, err := runner.Run(context.Background(), adapters.Request{})
	if err == nil {
		t.Fatal("expected provider error")
	}
}

func TestRunnerRequiresProvider(t *testing.T) {
	_, err := NewRunner(nil).Run(context.Background(), adapters.Request{})
	if err == nil {
		t.Fatal("expected empty runner to fail")
	}
}
