package publishing

import (
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestGeneratePackDefaultsToPrivate(t *testing.T) {
	pack, err := GeneratePack(PackInput{
		EpisodeID: "episode-1",
		Title:     "What Happens After git push?",
		Summary:   "A source-grounded explanation.",
		Sources: []artifacts.Source{{
			ID:    "git-docs",
			Title: "Git documentation",
			URI:   "https://git-scm.com/doc",
		}},
	})
	if err != nil {
		t.Fatalf("generate pack failed: %v", err)
	}
	if pack.Visibility != artifacts.PublishVisibilityPrivate {
		t.Fatalf("expected private visibility, got %s", pack.Visibility)
	}
	if !strings.Contains(pack.Description, "Sources:") {
		t.Fatalf("expected sources in description: %s", pack.Description)
	}
}

func TestGeneratePackRejectsPublicWithoutApproval(t *testing.T) {
	_, err := GeneratePack(PackInput{
		EpisodeID:   "episode-1",
		Title:       "What Happens After git push?",
		Summary:     "A source-grounded explanation.",
		Visibility:  artifacts.PublishVisibilityPublic,
		HumanApproved: false,
	})
	if err == nil {
		t.Fatal("expected public visibility without approval to fail")
	}
}

func TestGeneratePackAllowsPublicWithApproval(t *testing.T) {
	pack, err := GeneratePack(PackInput{
		EpisodeID:     "episode-1",
		Title:         "What Happens After git push?",
		Summary:       "A source-grounded explanation.",
		Visibility:    artifacts.PublishVisibilityPublic,
		HumanApproved: true,
	})
	if err != nil {
		t.Fatalf("expected approved public pack to be generated: %v", err)
	}
	if pack.Visibility != artifacts.PublishVisibilityPublic {
		t.Fatalf("expected public visibility, got %s", pack.Visibility)
	}
}

func TestGeneratePackRequiresTitle(t *testing.T) {
	_, err := GeneratePack(PackInput{EpisodeID: "episode-1"})
	if err == nil {
		t.Fatal("expected missing title to fail")
	}
}
