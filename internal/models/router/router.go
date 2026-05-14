package router

import (
	"fmt"
	"slices"
	"sort"

	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/providers"
)

// Options controls model selection policy.
type Options struct {
	// AllowDegraded permits degraded models to be selected when healthy active
	// models are unavailable or lower ranked. It should be false for strict
	// production gates unless an explicit fallback policy allows degradation.
	AllowDegraded  bool
	ProviderHealth map[string]providers.HealthState
	FallbackPolicy providers.FallbackPolicy
}

// Router selects models for a task without calling any provider.
type Router struct {
	registry []models.ModelRecord
	options  Options
}

func New(registry []models.ModelRecord) Router {
	return NewWithOptions(registry, Options{})
}

func NewWithOptions(registry []models.ModelRecord, options Options) Router {
	return Router{registry: append([]models.ModelRecord(nil), registry...), options: options}
}

func (r Router) Route(req models.TaskRequest) (models.RoutingDecision, error) {
	if req.Capability == "" {
		return models.RoutingDecision{}, fmt.Errorf("task capability is required")
	}
	if req.Modality == "" {
		return models.RoutingDecision{}, fmt.Errorf("task modality is required")
	}
	if req.PrivacyTier == "" {
		return models.RoutingDecision{}, fmt.Errorf("task privacy tier is required")
	}

	var candidates []models.ModelRecord
	var rejected []models.RejectedModel
	var fallbackReasons []string
	for _, model := range r.registry {
		health := providerHealth(model.Provider, r.options.ProviderHealth)
		healthDecision := providers.Evaluate(model.Provider, health, fallbackPolicy(r.options))
		if !healthDecision.Allowed {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: healthDecision.Reason})
			continue
		}
		if health != "" && health != providers.HealthHealthy {
			fallbackReasons = append(fallbackReasons, healthDecision.Reason)
		}
		if model.Status == models.ModelStatusDisabled {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: "model disabled"})
			continue
		}
		if model.Status == models.ModelStatusDegraded && !r.options.AllowDegraded {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: "model degraded and fallback policy disallows degraded selection"})
			continue
		}
		if !hasCapability(model.Capabilities, req.Capability) {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: "missing capability"})
			continue
		}
		if !hasModality(model.Modalities, req.Modality) {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: "missing modality"})
			continue
		}
		if !privacyAllowed(model.PrivacyTier, req.PrivacyTier) {
			rejected = append(rejected, models.RejectedModel{ModelID: model.ID, Reason: "privacy tier not allowed"})
			continue
		}
		candidates = append(candidates, model)
	}

	if len(candidates) == 0 {
		return models.RoutingDecision{Rejected: rejected, Policy: "no_candidate"}, fmt.Errorf("no model candidates available")
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].QualityScore > candidates[j].QualityScore
	})

	count := 1
	policy := "single_model"
	switch req.RiskLevel {
	case models.RiskMedium:
		count = min(2, len(candidates))
		policy = "primary_plus_reviewer"
	case models.RiskHigh, models.RiskCritical:
		count = min(3, len(candidates))
		policy = "multimodel_council"
	}

	return models.RoutingDecision{
		Selected:        append([]models.ModelRecord(nil), candidates[:count]...),
		Rejected:        rejected,
		Policy:          policy,
		FallbackReasons: fallbackReasons,
	}, nil
}

func providerHealth(provider string, health map[string]providers.HealthState) providers.HealthState {
	if len(health) == 0 {
		return providers.HealthHealthy
	}
	state, ok := health[provider]
	if !ok {
		return providers.HealthUnknown
	}
	return state
}

func fallbackPolicy(options Options) providers.FallbackPolicy {
	policy := options.FallbackPolicy
	if options.AllowDegraded {
		policy.AllowDegraded = true
	}
	return policy
}

func hasCapability(capabilities []models.Capability, target models.Capability) bool {
	return slices.Contains(capabilities, target)
}

func hasModality(modalities []models.Modality, target models.Modality) bool {
	return slices.Contains(modalities, target)
}

func privacyAllowed(modelTier models.PrivacyTier, taskTier models.PrivacyTier) bool {
	if modelTier == models.PrivacyTierLocalOnly {
		return true
	}
	order := map[models.PrivacyTier]int{
		models.PrivacyTierPublic:           1,
		models.PrivacyTierInternalApproved: 2,
		models.PrivacyTierRestricted:       3,
		models.PrivacyTierLocalOnly:        4,
	}
	return order[modelTier] >= order[taskTier]
}
