package mock

import (
	"context"

	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/models/adapters"
)

// Provider is a deterministic local provider used in tests and dry runs.
type Provider struct {
	ModelID    string
	ProviderID string
	Task       string
	Verdict    council.Verdict
	Confidence float64
	Notes      string
	Err        error
}

func (p Provider) ID() string {
	return p.ProviderID + ":" + p.ModelID
}

func (p Provider) Run(ctx context.Context, req adapters.Request) (adapters.Response, error) {
	if p.Err != nil {
		return adapters.Response{}, p.Err
	}
	select {
	case <-ctx.Done():
		return adapters.Response{}, ctx.Err()
	default:
	}

	providerID := p.ProviderID
	if providerID == "" {
		providerID = "local-mock"
	}
	modelID := p.ModelID
	if modelID == "" {
		modelID = "mock-model"
	}
	verdict := p.Verdict
	if verdict == "" {
		verdict = council.VerdictApprove
	}
	confidence := p.Confidence
	if confidence == 0 {
		confidence = 0.75
	}
	notes := p.Notes
	if notes == "" {
		notes = "deterministic mock review"
	}

	review := council.ModelReview{
		ModelID:    modelID,
		Provider:   providerID,
		Task:       p.Task,
		Verdict:    verdict,
		Confidence: confidence,
		Notes:      notes,
	}

	return adapters.Response{
		ModelID:    modelID,
		Provider:   providerID,
		Review:     review,
		RawSummary: notes,
	}, nil
}
