package providers

import "testing"

func TestEvaluateRejectsDisabledProvider(t *testing.T) {
	decision := Evaluate("provider-a", HealthDisabled, FallbackPolicy{AllowDegraded: true, AllowUnknown: true})
	if decision.Allowed {
		t.Fatalf("expected disabled provider to be rejected: %+v", decision)
	}
}

func TestEvaluateDegradedFollowsPolicy(t *testing.T) {
	rejected := Evaluate("provider-a", HealthDegraded, FallbackPolicy{})
	if rejected.Allowed {
		t.Fatalf("expected degraded provider to be rejected by default: %+v", rejected)
	}
	allowed := Evaluate("provider-a", HealthDegraded, FallbackPolicy{AllowDegraded: true})
	if !allowed.Allowed {
		t.Fatalf("expected degraded provider to be allowed by policy: %+v", allowed)
	}
}

func TestEvaluateDegradedCanRequireApproval(t *testing.T) {
	decision := Evaluate("provider-a", HealthDegraded, FallbackPolicy{AllowDegraded: true, RequireApprovalForDegraded: true})
	if decision.Allowed || !decision.RequiresApproval {
		t.Fatalf("expected degraded provider to require approval: %+v", decision)
	}
}

func TestEvaluateUnknownFollowsPolicy(t *testing.T) {
	rejected := Evaluate("provider-a", HealthUnknown, FallbackPolicy{})
	if rejected.Allowed {
		t.Fatalf("expected unknown provider to be rejected by default: %+v", rejected)
	}
	allowed := Evaluate("provider-a", HealthUnknown, FallbackPolicy{AllowUnknown: true})
	if !allowed.Allowed {
		t.Fatalf("expected unknown provider to be allowed by policy: %+v", allowed)
	}
}
