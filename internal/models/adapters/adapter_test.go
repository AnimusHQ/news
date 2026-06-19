package adapters_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/models"
	"github.com/AnimusHQ/news/internal/models/adapters"
	"github.com/AnimusHQ/news/internal/models/mock"
	"github.com/AnimusHQ/news/internal/models/sandbox"
)

// Compile-time guards: the concrete provider implementations must satisfy the
// adapters.Provider interface. If a method signature drifts, the build breaks
// here rather than at a distant call site.
var (
	_ adapters.Provider = mock.Provider{}
	_ adapters.Provider = sandbox.Provider{}
)

// TestMockProviderSatisfiesProviderInterface exercises a concrete provider
// strictly through the adapters.Provider interface, proving the wiring between
// the interface and an implementation is sound and offline.
func TestMockProviderSatisfiesProviderInterface(t *testing.T) {
	var p adapters.Provider = mock.Provider{ProviderID: "local-mock", ModelID: "mock-model"}

	if p.ID() != "local-mock:mock-model" {
		t.Fatalf("unexpected provider id: %q", p.ID())
	}

	resp, err := p.Run(context.Background(), adapters.Request{
		Task:      models.TaskRequest{TaskID: "t1", Capability: models.CapabilityEditorialReview},
		Prompt:    "review this",
		EpisodeID: "episode-0001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Review.Verdict != models.VerdictApprove {
		t.Fatalf("expected default approve verdict, got %q", resp.Review.Verdict)
	}
	if resp.Provider == "" || resp.ModelID == "" {
		t.Fatalf("normalized response missing provider/model id: %+v", resp)
	}
}

// TestErrorTypesAreDistinguishable confirms the adapter error types carry their
// provider id and are matchable with errors.As, which routing/fallback code
// relies on to classify failures.
func TestErrorTypesAreDistinguishable(t *testing.T) {
	var unavailable error = adapters.ErrProviderUnavailable{ProviderID: "p1", Reason: "disabled"}
	var invalid error = adapters.ErrInvalidModelOutput{ProviderID: "p2", Reason: "bad verdict"}

	var ua adapters.ErrProviderUnavailable
	if !errors.As(unavailable, &ua) || ua.ProviderID != "p1" {
		t.Fatalf("ErrProviderUnavailable not matchable: %v", unavailable)
	}
	var io adapters.ErrInvalidModelOutput
	if !errors.As(invalid, &io) || io.ProviderID != "p2" {
		t.Fatalf("ErrInvalidModelOutput not matchable: %v", invalid)
	}
	if !strings.Contains(unavailable.Error(), "p1") || !strings.Contains(invalid.Error(), "p2") {
		t.Fatalf("error messages must name the provider: %q / %q", unavailable, invalid)
	}
}
