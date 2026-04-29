package council

import "testing"

func TestAggregateApprovesUnanimousReviews(t *testing.T) {
	report, err := Aggregate([]ModelReview{
		{ModelID: "a", Provider: "test", Verdict: VerdictApprove},
		{ModelID: "b", Provider: "test", Verdict: VerdictApprove},
	})
	if err != nil {
		t.Fatalf("aggregate failed: %v", err)
	}
	if report.Consensus != ConsensusApproved {
		t.Fatalf("expected approved consensus, got %s", report.Consensus)
	}
	if len(report.Dissent) != 0 {
		t.Fatalf("expected no dissent, got %d", len(report.Dissent))
	}
}

func TestAggregatePreservesRevisionDissent(t *testing.T) {
	report, err := Aggregate([]ModelReview{
		{ModelID: "a", Provider: "test", Verdict: VerdictApprove},
		{ModelID: "b", Provider: "test", Verdict: VerdictRequestRevision, Notes: "unsupported claim"},
	})
	if err != nil {
		t.Fatalf("aggregate failed: %v", err)
	}
	if report.Consensus != ConsensusRevisionRequired {
		t.Fatalf("expected revision required, got %s", report.Consensus)
	}
	if len(report.Dissent) != 1 {
		t.Fatalf("expected dissent to be preserved")
	}
}

func TestAggregateBlocksOnAnyBlocker(t *testing.T) {
	report, err := Aggregate([]ModelReview{
		{ModelID: "a", Provider: "test", Verdict: VerdictApprove},
		{ModelID: "safety", Provider: "test", Verdict: VerdictBlock, Notes: "unsafe"},
	})
	if err != nil {
		t.Fatalf("aggregate failed: %v", err)
	}
	if report.Consensus != ConsensusBlocked {
		t.Fatalf("expected blocked consensus, got %s", report.Consensus)
	}
	if len(report.BlockingObjections) != 1 {
		t.Fatalf("expected blocking objection")
	}
}

func TestAggregateRejectsUnknownVerdict(t *testing.T) {
	_, err := Aggregate([]ModelReview{{ModelID: "a", Provider: "test", Verdict: Verdict("strange")}})
	if err == nil {
		t.Fatal("expected unknown verdict to fail")
	}
}
