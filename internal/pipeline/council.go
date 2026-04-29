package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/models/adapters"
	"github.com/AnimusHQ/news/internal/models/mock"
	"github.com/AnimusHQ/news/internal/models/registry"
	"github.com/AnimusHQ/news/internal/models/router"
)

const DefaultModelRegistryPath = "config/model-registry.example.yaml"

// CouncilDryRunResult is the deterministic local model council result used by
// safe dry runs. It does not call any external model provider.
type CouncilDryRunResult struct {
	Report       council.Report
	Selected     []string
	RegistryPath string
}

// RunLocalMockCouncil loads the example model registry, routes review tasks,
// executes deterministic mock providers, and aggregates their outputs.
func RunLocalMockCouncil(ctx context.Context, registryPath string) (CouncilDryRunResult, error) {
	if registryPath == "" {
		registryPath = DefaultModelRegistryPath
	}
	registryPath = resolveRegistryPath(registryPath)

	records, err := registry.LoadFile(registryPath)
	if err != nil {
		return CouncilDryRunResult{}, err
	}

	r := router.New(records)
	tasks := []models.TaskRequest{
		{
			TaskID:      "dry-run-technical-review",
			Capability:  models.CapabilityTechnicalVerification,
			RiskLevel:   models.RiskHigh,
			Modality:    models.ModalityText,
			PrivacyTier: models.PrivacyTierPublic,
			Description: "Verify pilot episode technical claims.",
		},
		{
			TaskID:      "dry-run-editorial-review",
			Capability:  models.CapabilityEditorialReview,
			RiskLevel:   models.RiskMedium,
			Modality:    models.ModalityText,
			PrivacyTier: models.PrivacyTierPublic,
			Description: "Review pilot episode clarity and pedagogy.",
		},
		{
			TaskID:      "dry-run-safety-review",
			Capability:  models.CapabilitySafetyReview,
			RiskLevel:   models.RiskHigh,
			Modality:    models.ModalityText,
			PrivacyTier: models.PrivacyTierPublic,
			Description: "Review pilot episode safety and release posture.",
		},
	}

	providers := make([]adapters.Provider, 0, len(tasks))
	selected := make([]string, 0, len(tasks))
	for _, task := range tasks {
		decision, err := r.Route(task)
		if err != nil {
			return CouncilDryRunResult{}, fmt.Errorf("route %s: %w", task.TaskID, err)
		}
		if len(decision.Selected) == 0 {
			return CouncilDryRunResult{}, fmt.Errorf("route %s selected no models", task.TaskID)
		}
		model := decision.Selected[0]
		selected = append(selected, model.ID)
		providers = append(providers, mockProviderForTask(model, task))
	}

	runner := council.NewRunner(providers)
	report, err := runner.Run(ctx, adapters.Request{
		Task:       models.TaskRequest{TaskID: "dry-run-council", Capability: models.CapabilityTechnicalVerification, RiskLevel: models.RiskHigh, Modality: models.ModalityText, PrivacyTier: models.PrivacyTierPublic},
		EpisodeID:  "episode-0001",
		ArtifactID: "dry-run-council",
	})
	if err != nil {
		return CouncilDryRunResult{}, err
	}

	return CouncilDryRunResult{Report: report, Selected: selected, RegistryPath: registryPath}, nil
}

func resolveRegistryPath(registryPath string) string {
	if filepath.IsAbs(registryPath) {
		return registryPath
	}
	if _, err := os.Stat(registryPath); err == nil {
		return registryPath
	}

	current, err := os.Getwd()
	if err != nil {
		return registryPath
	}
	for {
		candidate := filepath.Join(current, registryPath)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(current)
		if parent == current {
			return registryPath
		}
		current = parent
	}
}

func mockProviderForTask(model models.ModelRecord, task models.TaskRequest) adapters.Provider {
	provider := mock.Provider{
		ModelID:    model.ID,
		ProviderID: model.Provider,
		Task:       task.TaskID,
		Confidence: max(0.5, model.QualityScore),
	}

	switch task.Capability {
	case models.CapabilityTechnicalVerification:
		provider.Verdict = council.VerdictRequestRevision
		provider.Notes = "Pilot claims still use placeholder evidence locators; real source verification is required before production release."
	case models.CapabilityEditorialReview:
		provider.Verdict = council.VerdictApproveWithSuggestions
		provider.Notes = "Pilot narrative is structurally clear; tighten hook after real verification."
	case models.CapabilitySafetyReview:
		provider.Verdict = council.VerdictApproveWithSuggestions
		provider.Notes = "Dry-run posture is safe because public publishing remains disabled."
	default:
		provider.Verdict = council.VerdictApprove
		provider.Notes = "Deterministic mock approval."
	}

	return provider
}
