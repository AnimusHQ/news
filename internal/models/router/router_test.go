package router

import (
	"testing"

	"github.com/AnimusHQ/news/internal/models"
)

func TestRouteLowRiskSelectsSingleModel(t *testing.T) {
	r := New([]models.ModelRecord{
		model("model-a", models.PrivacyTierPublic, 0.9),
		model("model-b", models.PrivacyTierPublic, 0.8),
	})

	decision, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskLow,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	if len(decision.Selected) != 1 {
		t.Fatalf("expected one selected model, got %d", len(decision.Selected))
	}
	if decision.Selected[0].ID != "model-a" {
		t.Fatalf("expected highest quality model-a, got %s", decision.Selected[0].ID)
	}
}

func TestRouteHighRiskSelectsCouncil(t *testing.T) {
	r := New([]models.ModelRecord{
		model("model-a", models.PrivacyTierPublic, 0.9),
		model("model-b", models.PrivacyTierPublic, 0.8),
		model("model-c", models.PrivacyTierPublic, 0.7),
		model("model-d", models.PrivacyTierPublic, 0.6),
	})

	decision, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskHigh,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}
	if decision.Policy != "multimodel_council" {
		t.Fatalf("expected council policy, got %s", decision.Policy)
	}
	if len(decision.Selected) != 3 {
		t.Fatalf("expected three selected models, got %d", len(decision.Selected))
	}
}

func TestRouteRejectsDisabledModel(t *testing.T) {
	r := New([]models.ModelRecord{
		{
			ID:           "disabled",
			Provider:     "test",
			Version:      "v1",
			Status:       models.ModelStatusDisabled,
			PrivacyTier:  models.PrivacyTierPublic,
			Modalities:   []models.Modality{models.ModalityText},
			Capabilities: []models.Capability{models.CapabilityTechnicalVerification},
			QualityScore: 1.0,
		},
	})

	_, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskLow,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err == nil {
		t.Fatal("expected disabled-only registry to fail")
	}
}

func TestRouteRejectsDegradedModelByDefault(t *testing.T) {
	degraded := model("degraded", models.PrivacyTierPublic, 0.9)
	degraded.Status = models.ModelStatusDegraded
	r := New([]models.ModelRecord{degraded})

	_, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskLow,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err == nil {
		t.Fatal("expected degraded-only registry to fail by default")
	}
}

func TestRouteAllowsDegradedModelWhenFallbackPolicyAllows(t *testing.T) {
	degraded := model("degraded", models.PrivacyTierPublic, 0.9)
	degraded.Status = models.ModelStatusDegraded
	r := NewWithOptions([]models.ModelRecord{degraded}, Options{AllowDegraded: true})

	decision, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskLow,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierPublic,
	})
	if err != nil {
		t.Fatalf("expected degraded fallback to be selectable: %v", err)
	}
	if len(decision.Selected) != 1 || decision.Selected[0].ID != "degraded" {
		t.Fatalf("expected degraded model selected, got %+v", decision.Selected)
	}
}

func TestRouteBlocksPrivacyMismatch(t *testing.T) {
	r := New([]models.ModelRecord{
		model("public-model", models.PrivacyTierPublic, 0.9),
	})

	_, err := r.Route(models.TaskRequest{
		Capability:  models.CapabilityTechnicalVerification,
		RiskLevel:   models.RiskHigh,
		Modality:    models.ModalityText,
		PrivacyTier: models.PrivacyTierRestricted,
	})
	if err == nil {
		t.Fatal("expected privacy mismatch to fail")
	}
}

func model(id string, privacy models.PrivacyTier, quality float64) models.ModelRecord {
	return models.ModelRecord{
		ID:           id,
		Provider:     "test",
		Version:      "v1",
		Status:       models.ModelStatusActive,
		PrivacyTier:  privacy,
		Modalities:   []models.Modality{models.ModalityText},
		Capabilities: []models.Capability{models.CapabilityTechnicalVerification},
		QualityScore: quality,
	}
}
