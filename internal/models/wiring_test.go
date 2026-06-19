package models_test

import (
	"testing"

	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/models/registry"
	"github.com/AnimusHQ/news/internal/models/router"
)

// sampleRegistry is a small, valid in-memory model registry that covers the
// task categories the router is expected to resolve. It is deterministic and
// requires no files, network, or secrets.
func sampleRegistry() []models.ModelRecord {
	return []models.ModelRecord{
		{
			ID: "m-research", Provider: "mock", Version: "1", Status: models.ModelStatusActive,
			PrivacyTier: models.PrivacyTierPublic, Modalities: []models.Modality{models.ModalityText},
			Capabilities: []models.Capability{models.CapabilityResearchSynthesis, models.CapabilityScriptWriting},
			QualityScore: 0.9,
		},
		{
			ID: "m-verify", Provider: "mock", Version: "1", Status: models.ModelStatusActive,
			PrivacyTier: models.PrivacyTierPublic, Modalities: []models.Modality{models.ModalityText, models.ModalityCode},
			Capabilities: []models.Capability{models.CapabilityTechnicalVerification, models.CapabilityEditorialReview},
			QualityScore: 0.8,
		},
		{
			ID: "m-vision", Provider: "mock", Version: "1", Status: models.ModelStatusActive,
			PrivacyTier: models.PrivacyTierPublic, Modalities: []models.Modality{models.ModalityVision},
			Capabilities: []models.Capability{models.CapabilityVisualReasoning, models.CapabilityStoryboardPlanning},
			QualityScore: 0.7,
		},
	}
}

// TestRegistryAndRouterResolveExpectedCategories validates the wiring between the
// registry record shape and the router: for each task category the registry
// declares, the router must resolve at least one model whose capability matches
// the request. This guards that the models subsystem stays internally consistent.
func TestRegistryAndRouterResolveExpectedCategories(t *testing.T) {
	reg := sampleRegistry()
	if err := registry.Validate(reg); err != nil {
		t.Fatalf("sample registry must be valid: %v", err)
	}
	r := router.New(reg)

	cases := []struct {
		capability models.Capability
		modality   models.Modality
		wantModel  string
	}{
		{models.CapabilityResearchSynthesis, models.ModalityText, "m-research"},
		{models.CapabilityTechnicalVerification, models.ModalityText, "m-verify"},
		{models.CapabilityVisualReasoning, models.ModalityVision, "m-vision"},
	}
	for _, tc := range cases {
		decision, err := r.Route(models.TaskRequest{
			TaskID:      "t-" + string(tc.capability),
			Capability:  tc.capability,
			RiskLevel:   models.RiskLow,
			Modality:    tc.modality,
			PrivacyTier: models.PrivacyTierPublic,
		})
		if err != nil {
			t.Fatalf("router failed to resolve %s: %v", tc.capability, err)
		}
		if len(decision.Selected) == 0 {
			t.Fatalf("router selected no model for %s", tc.capability)
		}
		if decision.Selected[0].ID != tc.wantModel {
			t.Fatalf("capability %s resolved to %s, want %s", tc.capability, decision.Selected[0].ID, tc.wantModel)
		}
	}
}

// TestRouterRejectsUnsatisfiableCategory confirms the router fails closed when no
// registered model can serve the requested capability.
func TestRouterRejectsUnsatisfiableCategory(t *testing.T) {
	r := router.New(sampleRegistry())
	_, err := r.Route(models.TaskRequest{
		TaskID:      "t-analytics",
		Capability:  models.CapabilityAnalytics, // not declared by any sample model
		RiskLevel:   models.RiskLow,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err == nil {
		t.Fatal("router must fail closed when no model has the requested capability")
	}
}
