package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/models"
)

const (
	defaultHTTPClientTimeout = 30 * time.Second
	maxResponseBytes         = 1 << 20
	maxErrorBodyBytes        = 4 << 10
)

// HTTPClient executes sandbox model requests through a provider-neutral HTTP
// boundary. It never resolves credential references into credential values.
type HTTPClient struct {
	Client *http.Client
}

type httpReviewRequest struct {
	ProviderID    string          `json:"provider_id"`
	ModelID       string          `json:"model_id"`
	Task          httpTaskRequest `json:"task"`
	Prompt        string          `json:"prompt"`
	EpisodeID     string          `json:"episode_id,omitempty"`
	ArtifactID    string          `json:"artifact_id,omitempty"`
	InputDigest   string          `json:"input_digest,omitempty"`
	CredentialRef string          `json:"credential_ref,omitempty"`
}

type httpTaskRequest struct {
	TaskID      string             `json:"task_id"`
	Capability  models.Capability  `json:"capability"`
	RiskLevel   models.RiskLevel   `json:"risk_level"`
	Modality    models.Modality    `json:"modality"`
	PrivacyTier models.PrivacyTier `json:"privacy_tier"`
	EpisodeID   string             `json:"episode_id,omitempty"`
	ArtifactID  string             `json:"artifact_id,omitempty"`
	Description string             `json:"description,omitempty"`
}

type httpReviewResponse struct {
	Verdict    models.Verdict `json:"verdict"`
	Confidence float64        `json:"confidence"`
	Notes      string         `json:"notes"`
	RawSummary string         `json:"raw_summary,omitempty"`
}

// Review sends a normalized JSON request and maps the JSON response into the
// sandbox provider's normalized client response.
func (c HTTPClient) Review(ctx context.Context, request ClientRequest) (ClientResponse, error) {
	endpoint := strings.TrimSpace(request.Endpoint)
	if err := validateHTTPEndpoint(endpoint); err != nil {
		return ClientResponse{}, err
	}
	credentialRef := strings.TrimSpace(request.CredentialRef)
	if err := validateCredentialRef(credentialRef); err != nil {
		return ClientResponse{}, err
	}

	body, err := json.Marshal(newHTTPReviewRequest(request, credentialRef))
	if err != nil {
		return ClientResponse{}, fmt.Errorf("marshal sandbox model request: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return ClientResponse{}, fmt.Errorf("build sandbox model request: %w", err)
	}
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Animus-Credential-Ref", credentialRef)

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPClientTimeout}
	}
	response, err := client.Do(httpRequest)
	if err != nil {
		return ClientResponse{}, fmt.Errorf("sandbox model request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
		message := strings.TrimSpace(string(data))
		if message == "" {
			message = response.Status
		}
		return ClientResponse{}, fmt.Errorf("sandbox model endpoint returned status %d: %s", response.StatusCode, message)
	}

	var parsed httpReviewResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(&parsed); err != nil {
		return ClientResponse{}, fmt.Errorf("decode sandbox model response: %w", err)
	}
	return ClientResponse{
		Verdict:    parsed.Verdict,
		Confidence: parsed.Confidence,
		Notes:      parsed.Notes,
		RawSummary: parsed.RawSummary,
	}, nil
}

func newHTTPReviewRequest(request ClientRequest, credentialRef string) httpReviewRequest {
	episodeID := firstNonEmpty(request.EpisodeID, request.Task.EpisodeID)
	artifactID := firstNonEmpty(request.ArtifactID, request.Task.ArtifactID)
	task := httpTaskRequest{
		TaskID:      request.Task.TaskID,
		Capability:  request.Task.Capability,
		RiskLevel:   request.Task.RiskLevel,
		Modality:    request.Task.Modality,
		PrivacyTier: request.Task.PrivacyTier,
		EpisodeID:   episodeID,
		ArtifactID:  artifactID,
		Description: request.Task.Description,
	}
	return httpReviewRequest{
		ProviderID:    strings.TrimSpace(request.ProviderID),
		ModelID:       strings.TrimSpace(request.ModelID),
		Task:          task,
		Prompt:        request.Prompt,
		EpisodeID:     episodeID,
		ArtifactID:    artifactID,
		InputDigest:   strings.TrimSpace(request.InputDigest),
		CredentialRef: credentialRef,
	}
}

func validateHTTPEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("sandbox endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parse sandbox endpoint: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("sandbox endpoint must use http or https scheme")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("sandbox endpoint host is required")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
