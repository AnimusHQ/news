package sources

import (
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestRegistryRanksPrimarySourcesFirst(t *testing.T) {
	registry, err := NewRegistry([]artifacts.Source{
		{ID: "community", Title: "Forum", URI: "https://example.com/forum", Type: "community_discussion", TrustLevel: "community"},
		{ID: "official", Title: "Official Docs", URI: "https://example.com/docs", Type: "official_docs", TrustLevel: "primary"},
	})
	if err != nil {
		t.Fatalf("registry failed: %v", err)
	}
	ranked := registry.Rank()
	if ranked[0].ID != "official" {
		t.Fatalf("expected official source first, got %s", ranked[0].ID)
	}
}

func TestRegistryRejectsDuplicateSourceID(t *testing.T) {
	_, err := NewRegistry([]artifacts.Source{
		{ID: "same", Title: "A", URI: "https://example.com/a", Type: "official_docs", TrustLevel: "primary"},
		{ID: "same", Title: "B", URI: "https://example.com/b", Type: "official_docs", TrustLevel: "primary"},
	})
	if err == nil {
		t.Fatal("expected duplicate source id to fail")
	}
}

func TestRegistryRejectsUnsupportedSourceType(t *testing.T) {
	_, err := NewRegistry([]artifacts.Source{{ID: "x", Title: "X", URI: "https://example.com", Type: "random", TrustLevel: "primary"}})
	if err == nil {
		t.Fatal("expected unsupported type to fail")
	}
}

func TestHighRiskAuthorityRequiresPrimarySource(t *testing.T) {
	registry, err := NewRegistry([]artifacts.Source{
		{ID: "community", Title: "Forum", URI: "https://example.com/forum", Type: "community_discussion", TrustLevel: "community"},
		{ID: "blog", Title: "Blog", URI: "https://example.com/blog", Type: "engineering_blog", TrustLevel: "secondary"},
	})
	if err != nil {
		t.Fatalf("registry failed: %v", err)
	}
	if registry.SatisfiesHighRiskAuthority([]string{"community", "blog"}) {
		t.Fatal("expected community/secondary-only evidence not to satisfy high-risk authority")
	}
}

func TestHighRiskAuthorityAcceptsPrimarySource(t *testing.T) {
	registry, err := NewRegistry([]artifacts.Source{
		{ID: "official", Title: "Official Docs", URI: "https://example.com/docs", Type: "official_docs", TrustLevel: "primary"},
	})
	if err != nil {
		t.Fatalf("registry failed: %v", err)
	}
	if !registry.SatisfiesHighRiskAuthority([]string{"official"}) {
		t.Fatal("expected primary source to satisfy high-risk authority")
	}
}
