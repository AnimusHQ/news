package sandbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/models"
)

func TestHTTPClientSendsNormalizedJSONWithoutAuthorization(t *testing.T) {
	var received httpReviewRequest
	var authorizationHeader string
	var credentialRefHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader = r.Header.Get("Authorization")
		credentialRefHeader = r.Header.Get("X-Animus-Credential-Ref")
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"verdict":"approve_with_suggestions","confidence":0.72,"notes":"Looks usable.","raw_summary":"provider summary"}`))
	}))
	defer server.Close()

	response, err := HTTPClient{Client: server.Client()}.Review(context.Background(), httpClientRequest(server.URL))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if authorizationHeader != "" {
		t.Fatalf("credential reference must not be used as authorization, got %q", authorizationHeader)
	}
	if credentialRefHeader != "env:ANIMUS_SANDBOX_PROVIDER_TOKEN" {
		t.Fatalf("unexpected credential ref header: %q", credentialRefHeader)
	}
	if received.ProviderID != "sandbox-provider" || received.ModelID != "sandbox-reviewer" {
		t.Fatalf("unexpected provider identity: %+v", received)
	}
	if received.Task.TaskID != "review-test" || received.Task.PrivacyTier != models.PrivacyTierPublic {
		t.Fatalf("unexpected task payload: %+v", received.Task)
	}
	if received.EpisodeID != "episode-test" || received.ArtifactID != "artifact-test" {
		t.Fatalf("unexpected artifact context: %+v", received)
	}
	if received.CredentialRef != "env:ANIMUS_SANDBOX_PROVIDER_TOKEN" {
		t.Fatalf("unexpected credential ref body value: %q", received.CredentialRef)
	}
	if response.Verdict != models.VerdictApproveWithSuggestions || response.Confidence != 0.72 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.RawSummary != "provider summary" {
		t.Fatalf("unexpected raw summary: %s", response.RawSummary)
	}
}

func TestHTTPClientRejectsUnsupportedEndpointScheme(t *testing.T) {
	_, err := HTTPClient{}.Review(context.Background(), httpClientRequest("file:///tmp/sandbox-model"))
	if err == nil {
		t.Fatal("expected unsupported endpoint scheme to fail")
	}
	if !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClientRejectsNon2xxResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "provider unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := HTTPClient{Client: server.Client()}.Review(context.Background(), httpClientRequest(server.URL))
	if err == nil {
		t.Fatal("expected non-2xx response to fail")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClientRejectsMalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	_, err := HTTPClient{Client: server.Client()}.Review(context.Background(), httpClientRequest(server.URL))
	if err == nil {
		t.Fatal("expected malformed JSON response to fail")
	}
	if !strings.Contains(err.Error(), "decode sandbox model response") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func httpClientRequest(endpoint string) ClientRequest {
	return ClientRequest{
		ProviderID:    "sandbox-provider",
		ModelID:       "sandbox-reviewer",
		Endpoint:      endpoint,
		CredentialRef: "env:ANIMUS_SANDBOX_PROVIDER_TOKEN",
		Task: models.TaskRequest{
			TaskID:      "review-test",
			Capability:  models.CapabilityTechnicalVerification,
			RiskLevel:   models.RiskMedium,
			Modality:    models.ModalityText,
			PrivacyTier: models.PrivacyTierPublic,
			Description: "Review an educational IT media script.",
		},
		Prompt:      "Review this source-grounded draft.",
		EpisodeID:   "episode-test",
		ArtifactID:  "artifact-test",
		InputDigest: "sha256:test",
	}
}
