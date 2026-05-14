package cost

import (
	"testing"
	"time"
)

func TestAggregateCombinesCostEvents(t *testing.T) {
	summary, err := Aggregate([]Event{
		{EpisodeID: "episode-1", Stage: "model_council", Provider: "mock", ModelID: "model-a", OperationType: "review", EstimatedCost: 1.25, Currency: "USD", CreatedAt: time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)},
		{EpisodeID: "episode-1", Stage: "verification", Provider: "mock", ModelID: "model-a", OperationType: "verify", EstimatedCost: 2.75, Currency: "USD", CreatedAt: time.Date(2026, 5, 7, 11, 0, 0, 0, time.UTC)},
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
	if summary.ByModel["model-a"] != 4.0 {
		t.Fatalf("unexpected model cost")
	}
	if summary.ByDay["2026-05-07"] != 4.0 {
		t.Fatalf("unexpected day cost")
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

func TestEvaluateBudgetWarnsAndRequiresApproval(t *testing.T) {
	warn := EvaluateBudget(Summary{Total: 7, Currency: "USD"}, BudgetPolicy{WarnAt: 5, RequireApprovalAt: 10, BlockAt: 20, Currency: "USD"}, false)
	if !warn.Allowed || warn.Action != BudgetActionWarn {
		t.Fatalf("expected warning decision, got %+v", warn)
	}
	approval := EvaluateBudget(Summary{Total: 12, Currency: "USD"}, BudgetPolicy{WarnAt: 5, RequireApprovalAt: 10, BlockAt: 20, Currency: "USD"}, false)
	if approval.Allowed || approval.Action != BudgetActionRequireApproval {
		t.Fatalf("expected approval-required decision, got %+v", approval)
	}
}

func TestEvaluateBudgetBlocksNonCriticalAutomation(t *testing.T) {
	decision := EvaluateBudget(Summary{Total: 25, Currency: "USD"}, BudgetPolicy{BlockAt: 20, Currency: "USD"}, false)
	if decision.Allowed || decision.Action != BudgetActionBlock {
		t.Fatalf("expected block decision, got %+v", decision)
	}
	critical := EvaluateBudget(Summary{Total: 25, Currency: "USD"}, BudgetPolicy{BlockAt: 20, Currency: "USD"}, true)
	if critical.Allowed || critical.Action != BudgetActionRequireApproval {
		t.Fatalf("critical over-budget work should require approval, got %+v", critical)
	}
}
