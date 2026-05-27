package sandbox

import (
	"context"
	"fmt"
	"strings"

	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/models/adapters"
	"github.com/AnimusHQ/news/internal/security"
)

// Client executes a sandbox model request. Real provider HTTP/SDK code must
// live behind this interface, never in routing or workflow code.
type Client interface {
	Review(ctx context.Context, request ClientRequest) (ClientResponse, error)
}

// Config keeps provider-specific settings as metadata only. CredentialRef is a
// reference such as an environment variable name, not a credential value.
type Config struct {
	ProviderID          string
	ModelID             string
	Endpoint            string
	CredentialRef       string
	Enabled             bool
	AllowedPrivacyTiers []models.PrivacyTier
}

// Provider is a sandbox adapter for future real model providers.
type Provider struct {
	Config Config
	Client Client
}

// ClientRequest is the provider-facing request after policy checks.
type ClientRequest struct {
	ProviderID    string
	ModelID       string
	Endpoint      string
	CredentialRef string
	Task          models.TaskRequest
	Prompt        string
	EpisodeID     string
	ArtifactID    string
	InputDigest   string
}

// ClientResponse is the provider-facing normalized response before final
// adapter validation.
type ClientResponse struct {
	Verdict    models.Verdict
	Confidence float64
	Notes      string
	RawSummary string
}

func (p Provider) ID() string {
	return providerID(p.Config) + ":" + modelID(p.Config)
}

// Run enforces sandbox safety policy, delegates execution to the injected
// client, and returns a provider-normalized model review.
func (p Provider) Run(ctx context.Context, req adapters.Request) (adapters.Response, error) {
	if err := ctx.Err(); err != nil {
		return adapters.Response{}, err
	}
	if err := p.validateConfig(); err != nil {
		return adapters.Response{}, err
	}
	if err := p.validateRequest(req); err != nil {
		return adapters.Response{}, err
	}

	response, err := p.Client.Review(ctx, ClientRequest{
		ProviderID:    providerID(p.Config),
		ModelID:       modelID(p.Config),
		Endpoint:      strings.TrimSpace(p.Config.Endpoint),
		CredentialRef: strings.TrimSpace(p.Config.CredentialRef),
		Task:          req.Task,
		Prompt:        req.Prompt,
		EpisodeID:     req.EpisodeID,
		ArtifactID:    req.ArtifactID,
		InputDigest:   req.InputDigest,
	})
	if err != nil {
		return adapters.Response{}, err
	}
	if err := validateClientResponse(response); err != nil {
		return adapters.Response{}, adapters.ErrInvalidModelOutput{ProviderID: providerID(p.Config), Reason: err.Error()}
	}

	notes := strings.TrimSpace(response.Notes)
	rawSummary := strings.TrimSpace(response.RawSummary)
	if rawSummary == "" {
		rawSummary = notes
	}
	review := models.ModelReview{
		ModelID:    modelID(p.Config),
		Provider:   providerID(p.Config),
		Task:       req.Task.TaskID,
		Verdict:    response.Verdict,
		Confidence: response.Confidence,
		Notes:      notes,
	}
	return adapters.Response{
		ModelID:    review.ModelID,
		Provider:   review.Provider,
		Review:     review,
		RawSummary: rawSummary,
	}, nil
}

func (p Provider) validateConfig() error {
	if !p.Config.Enabled {
		return adapters.ErrProviderUnavailable{ProviderID: providerID(p.Config), Reason: "sandbox provider is disabled"}
	}
	if strings.TrimSpace(p.Config.CredentialRef) == "" {
		return adapters.ErrProviderUnavailable{ProviderID: providerID(p.Config), Reason: "credential reference is required"}
	}
	if strings.TrimSpace(p.Config.Endpoint) == "" {
		return adapters.ErrProviderUnavailable{ProviderID: providerID(p.Config), Reason: "sandbox endpoint is required"}
	}
	if p.Client == nil {
		return adapters.ErrProviderUnavailable{ProviderID: providerID(p.Config), Reason: "sandbox client is not configured"}
	}
	return nil
}

func (p Provider) validateRequest(req adapters.Request) error {
	if !privacyAllowed(req.Task.PrivacyTier, p.Config.AllowedPrivacyTiers) {
		return adapters.ErrProviderUnavailable{
			ProviderID: providerID(p.Config),
			Reason:     fmt.Sprintf("privacy tier %s is not allowed for sandbox provider", req.Task.PrivacyTier),
		}
	}
	if req.Task.PrivacyTier == models.PrivacyTierRestricted || req.Task.PrivacyTier == models.PrivacyTierLocalOnly {
		return adapters.ErrProviderUnavailable{
			ProviderID: providerID(p.Config),
			Reason:     fmt.Sprintf("privacy tier %s must not leave local execution", req.Task.PrivacyTier),
		}
	}
	if strings.TrimSpace(req.Prompt) != "" && security.Redact(req.Prompt) != req.Prompt {
		return adapters.ErrProviderUnavailable{ProviderID: providerID(p.Config), Reason: "prompt contains secret-like text"}
	}
	return nil
}

func validateClientResponse(response ClientResponse) error {
	switch response.Verdict {
	case models.VerdictApprove, models.VerdictApproveWithSuggestions, models.VerdictRequestRevision, models.VerdictBlock:
	default:
		return fmt.Errorf("unsupported verdict %q", response.Verdict)
	}
	if response.Confidence < 0 || response.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	if strings.TrimSpace(response.Notes) == "" {
		return fmt.Errorf("notes are required")
	}
	return nil
}

func privacyAllowed(tier models.PrivacyTier, allowed []models.PrivacyTier) bool {
	if tier == "" {
		tier = models.PrivacyTierPublic
	}
	if len(allowed) == 0 {
		allowed = []models.PrivacyTier{models.PrivacyTierPublic}
	}
	for _, item := range allowed {
		if item == tier {
			return true
		}
	}
	return false
}

func providerID(config Config) string {
	if strings.TrimSpace(config.ProviderID) == "" {
		return "sandbox-provider"
	}
	return strings.TrimSpace(config.ProviderID)
}

func modelID(config Config) string {
	if strings.TrimSpace(config.ModelID) == "" {
		return "sandbox-model"
	}
	return strings.TrimSpace(config.ModelID)
}
