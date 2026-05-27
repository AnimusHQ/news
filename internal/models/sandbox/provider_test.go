package sandbox

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/models/adapters"
)

func TestProviderFailsClosedWhenDisabled(t *testing.T) {
	client := &recordingClient{}
	_, err := Provider{Config: validConfig(), Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err != nil {
		t.Fatalf("valid provider failed: %v", err)
	}

	cfg := validConfig()
	cfg.Enabled = false
	_, err = Provider{Config: cfg, Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err == nil {
		t.Fatal("expected disabled provider to fail")
	}
	var unavailable adapters.ErrProviderUnavailable
	if !errors.As(err, &unavailable) {
		t.Fatalf("expected provider unavailable, got %T: %v", err, err)
	}
}

func TestProviderRequiresCredentialReference(t *testing.T) {
	cfg := validConfig()
	cfg.CredentialRef = ""
	_, err := Provider{Config: cfg, Client: &recordingClient{}}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err == nil {
		t.Fatal("expected missing credential ref to fail")
	}
	if !strings.Contains(err.Error(), "credential reference") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderRequiresCredentialReferencePrefix(t *testing.T) {
	client := &recordingClient{}
	cfg := validConfig()
	cfg.CredentialRef = "ANIMUS_SANDBOX_PROVIDER_TOKEN"
	_, err := Provider{Config: cfg, Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err == nil {
		t.Fatal("expected unprefixed credential ref to fail")
	}
	if !strings.Contains(err.Error(), "env:, secretref:, or file:") {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.calls != 0 {
		t.Fatalf("client should not be called, got %d calls", client.calls)
	}
}

func TestProviderRejectsSecretLikeCredentialReference(t *testing.T) {
	client := &recordingClient{}
	cfg := validConfig()
	cfg.CredentialRef = "token=" + strings.Repeat("a", 16) + "1234567890"
	_, err := Provider{Config: cfg, Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err == nil {
		t.Fatal("expected secret-like credential ref to fail")
	}
	if !strings.Contains(err.Error(), "credential value") {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.calls != 0 {
		t.Fatalf("client should not be called, got %d calls", client.calls)
	}
}

func TestProviderBlocksRestrictedDataBeforeClientCall(t *testing.T) {
	client := &recordingClient{}
	cfg := validConfig()
	cfg.AllowedPrivacyTiers = []models.PrivacyTier{models.PrivacyTierPublic, models.PrivacyTierRestricted}
	_, err := Provider{Config: cfg, Client: client}.Run(context.Background(), request(models.PrivacyTierRestricted))
	if err == nil {
		t.Fatal("expected restricted data to fail")
	}
	if client.calls != 0 {
		t.Fatalf("client should not be called, got %d calls", client.calls)
	}
}

func TestProviderBlocksSecretLikePromptBeforeClientCall(t *testing.T) {
	client := &recordingClient{}
	req := request(models.PrivacyTierPublic)
	req.Prompt = "api_key=" + strings.Repeat("a", 16) + "1234567890"
	_, err := Provider{Config: validConfig(), Client: client}.Run(context.Background(), req)
	if err == nil {
		t.Fatal("expected secret-like prompt to fail")
	}
	if client.calls != 0 {
		t.Fatalf("client should not be called, got %d calls", client.calls)
	}
}

func TestProviderRejectsInvalidClientOutput(t *testing.T) {
	client := &recordingClient{response: ClientResponse{Verdict: "maybe", Confidence: 0.9, Notes: "bad"}}
	_, err := Provider{Config: validConfig(), Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err == nil {
		t.Fatal("expected invalid output to fail")
	}
	var invalid adapters.ErrInvalidModelOutput
	if !errors.As(err, &invalid) {
		t.Fatalf("expected invalid model output, got %T: %v", err, err)
	}
}

func TestProviderNormalizesValidClientOutput(t *testing.T) {
	client := &recordingClient{response: ClientResponse{
		Verdict:    models.VerdictApproveWithSuggestions,
		Confidence: 0.82,
		Notes:      "Needs a tighter hook.",
		RawSummary: "Provider raw summary.",
	}}
	response, err := Provider{Config: validConfig(), Client: client}.Run(context.Background(), request(models.PrivacyTierPublic))
	if err != nil {
		t.Fatalf("run provider failed: %v", err)
	}
	if response.ModelID != "sandbox-reviewer" || response.Provider != "sandbox-provider" {
		t.Fatalf("unexpected provider identity: %+v", response)
	}
	if response.Review.Verdict != models.VerdictApproveWithSuggestions {
		t.Fatalf("unexpected verdict: %s", response.Review.Verdict)
	}
	if response.RawSummary != "Provider raw summary." {
		t.Fatalf("unexpected raw summary: %s", response.RawSummary)
	}
	if client.calls != 1 {
		t.Fatalf("expected one client call, got %d", client.calls)
	}
}

type recordingClient struct {
	calls    int
	response ClientResponse
}

func (c *recordingClient) Review(context.Context, ClientRequest) (ClientResponse, error) {
	c.calls++
	if c.response.Verdict == "" {
		return ClientResponse{Verdict: models.VerdictApprove, Confidence: 0.8, Notes: "sandbox approved"}, nil
	}
	return c.response, nil
}

func validConfig() Config {
	return Config{
		ProviderID:          "sandbox-provider",
		ModelID:             "sandbox-reviewer",
		Endpoint:            "https://sandbox.example.test/review",
		CredentialRef:       "env:ANIMUS_SANDBOX_PROVIDER_TOKEN",
		Enabled:             true,
		AllowedPrivacyTiers: []models.PrivacyTier{models.PrivacyTierPublic, models.PrivacyTierInternalApproved},
	}
}

func request(tier models.PrivacyTier) adapters.Request {
	return adapters.Request{
		Task: models.TaskRequest{
			TaskID:      "review-test",
			Capability:  models.CapabilityTechnicalVerification,
			RiskLevel:   models.RiskMedium,
			Modality:    models.ModalityText,
			PrivacyTier: tier,
		},
		Prompt:     "Review this source-grounded draft.",
		EpisodeID:  "episode-test",
		ArtifactID: "artifact-test",
	}
}
