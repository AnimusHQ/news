package adapters

import (
	"context"
	"fmt"

	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/models"
)

// Request is a provider-agnostic model execution request.
type Request struct {
	Task        models.TaskRequest
	Prompt      string
	ArtifactID  string
	EpisodeID   string
	InputDigest string
}

// Response is a provider-normalized output.
type Response struct {
	ModelID    string
	Provider   string
	Review     council.ModelReview
	RawSummary string
}

// Provider is implemented by concrete model providers and mocks.
type Provider interface {
	ID() string
	Run(ctx context.Context, req Request) (Response, error)
}

// ErrProviderUnavailable is returned when a provider cannot serve a request.
type ErrProviderUnavailable struct {
	ProviderID string
	Reason     string
}

func (e ErrProviderUnavailable) Error() string {
	return fmt.Sprintf("provider %s unavailable: %s", e.ProviderID, e.Reason)
}

// ErrInvalidModelOutput is returned when provider output cannot be normalized.
type ErrInvalidModelOutput struct {
	ProviderID string
	Reason     string
}

func (e ErrInvalidModelOutput) Error() string {
	return fmt.Sprintf("provider %s returned invalid output: %s", e.ProviderID, e.Reason)
}
