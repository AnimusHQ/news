package cost

import "testing"

func TestAggregateCombinesCostEvents(t *testing.T) {
	summary, err := Aggregate([]Event{
		{EpisodeID: "episode-1", Stage: "model_council", Provider: "mock", OperationType: "review", EstimatedCost: 1.25, Currency: "USD"},
		{EpisodeID: "episode-1", Stage: "verification", Provider: "mock", OperationType: "verify", EstimatedCost: 2.75, Currency: "USD"},
	})
	if err != nil {
		t.Fatalf("aggregate failed: %v", err)
	}
	if summary.Total != 4.0 {
		t.Fatalf("expected total 4.0, got %f", summary.Total)
	}
	if summary.ByStage["model_council"] != 1.25 {
		t.Fatalf("unexpected model council stage cost")
	}
	if summary.ByProvider["mock"] != 4.0 {
		t.Fatalf("unexpected provider cost")
	}
	if summary.EventCount != 2 {
		t.Fatalf("expected 2 events, got %d", summary.EventCount)
	}
}

func TestAggregateRejectsMixedEpisodes(t *testing.T) {
	_, err := Aggregate([]Event{
		{EpisodeID: "episode-1", Stage: "model_council", OperationType: "review", EstimatedCost: 1, Currency: "USD"},
		{EpisodeID: "episode-2", Stage: "verification", OperationType: "verify", EstimatedCost: 1, Currency: "USD"},
	})
	if err == nil {
		t.Fatal("expected mixed episodes to fail")
	}
}

func TestAggregateRejectsNegativeCost(t *testing.T) {
	_, err := Aggregate([]Event{{EpisodeID: "episode-1", Stage: "model_council", OperationType: "review", EstimatedCost: -1, Currency: "USD"}})
	if err == nil {
		t.Fatal("expected negative cost to fail")
	}
}

func TestCheckBudgetBlocksOverBudget(t *testing.T) {
	decision := CheckBudget(Summary{Total: 10, Currency: "USD"}, 5)
	if decision.Allowed {
		t.Fatal("expected over-budget summary to be blocked")
	}
}

func TestCheckBudgetAllowsWithinBudget(t *testing.T) {
	decision := CheckBudget(Summary{Total: 5, Currency: "USD"}, 5)
	if !decision.Allowed {
		t.Fatalf("expected within-budget summary to be allowed: %s", decision.Reason)
	}
}
