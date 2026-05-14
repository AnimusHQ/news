package providers

import "fmt"

// HealthState describes provider availability for deterministic routing policy.
type HealthState string

const (
	HealthHealthy  HealthState = "healthy"
	HealthDegraded HealthState = "degraded"
	HealthDisabled HealthState = "disabled"
	HealthUnknown  HealthState = "unknown"
)

// FallbackPolicy controls whether non-healthy providers can be used.
type FallbackPolicy struct {
	AllowDegraded              bool
	AllowUnknown               bool
	RequireApprovalForDegraded bool
}

// Decision explains provider health policy output.
type Decision struct {
	Allowed          bool
	RequiresApproval bool
	Reason           string
}

// Evaluate returns a deterministic health decision for one provider.
func Evaluate(provider string, state HealthState, policy FallbackPolicy) Decision {
	if provider == "" {
		provider = "unknown-provider"
	}
	switch state {
	case "", HealthHealthy:
		return Decision{Allowed: true, Reason: provider + " is healthy"}
	case HealthDisabled:
		return Decision{Allowed: false, Reason: provider + " is disabled"}
	case HealthDegraded:
		if policy.RequireApprovalForDegraded {
			return Decision{Allowed: false, RequiresApproval: true, Reason: provider + " is degraded and requires human approval"}
		}
		if !policy.AllowDegraded {
			return Decision{Allowed: false, Reason: provider + " is degraded and fallback policy disallows degraded selection"}
		}
		return Decision{Allowed: true, Reason: provider + " is degraded but fallback policy allows degraded selection"}
	case HealthUnknown:
		if !policy.AllowUnknown {
			return Decision{Allowed: false, Reason: provider + " health is unknown and fallback policy disallows unknown selection"}
		}
		return Decision{Allowed: true, Reason: provider + " health is unknown but fallback policy allows unknown selection"}
	default:
		return Decision{Allowed: false, Reason: fmt.Sprintf("%s has unsupported health state %q", provider, state)}
	}
}
