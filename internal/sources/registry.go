package sources

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// TrustLevel is the normalized authority level for a source.
type TrustLevel string

const (
	TrustPrimary   TrustLevel = "primary"
	TrustSecondary TrustLevel = "secondary"
	TrustCommunity TrustLevel = "community"
	TrustUnknown   TrustLevel = "unknown"
)

// Type is the normalized source category.
type Type string

const (
	TypeOfficialDocs        Type = "official_docs"
	TypeSpecification       Type = "specification"
	TypeSourceCode          Type = "source_code"
	TypeReleaseNotes        Type = "release_notes"
	TypeMaintainerStatement Type = "maintainer_statement"
	TypeEngineeringBlog     Type = "engineering_blog"
	TypeCommunityDiscussion Type = "community_discussion"
	TypeUnknown             Type = "unknown"
)

// Registry stores normalized sources by ID.
type Registry struct {
	byID map[string]artifacts.Source
}

// NewRegistry validates and indexes sources.
func NewRegistry(items []artifacts.Source) (Registry, error) {
	if len(items) == 0 {
		return Registry{}, fmt.Errorf("source registry requires at least one source")
	}
	byID := make(map[string]artifacts.Source, len(items))
	for _, item := range items {
		if err := ValidateSource(item); err != nil {
			return Registry{}, err
		}
		if _, exists := byID[item.ID]; exists {
			return Registry{}, fmt.Errorf("duplicate source id: %s", item.ID)
		}
		byID[item.ID] = item
	}
	return Registry{byID: byID}, nil
}

// ValidateSource checks minimum source metadata needed for provenance.
func ValidateSource(source artifacts.Source) error {
	if source.ID == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(source.Title) == "" {
		return fmt.Errorf("source %s title is required", source.ID)
	}
	if strings.TrimSpace(source.URI) == "" {
		return fmt.Errorf("source %s uri is required", source.ID)
	}
	if normalizedType(source.Type) == TypeUnknown {
		return fmt.Errorf("source %s has unsupported type %q", source.ID, source.Type)
	}
	if normalizedTrust(source.TrustLevel) == TrustUnknown {
		return fmt.Errorf("source %s has unsupported trust level %q", source.ID, source.TrustLevel)
	}
	return nil
}

// Get returns one source by ID.
func (r Registry) Get(id string) (artifacts.Source, bool) {
	source, ok := r.byID[id]
	return source, ok
}

// Rank returns sources ordered from most authoritative to least authoritative.
func (r Registry) Rank() []artifacts.Source {
	items := make([]artifacts.Source, 0, len(r.byID))
	for _, source := range r.byID {
		items = append(items, source)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return score(items[i]) > score(items[j])
	})
	return items
}

// SatisfiesHighRiskAuthority returns true when at least one cited source is
// primary authority. Community-only evidence cannot satisfy high-risk claims.
func (r Registry) SatisfiesHighRiskAuthority(sourceIDs []string) bool {
	for _, id := range sourceIDs {
		source, ok := r.Get(id)
		if !ok {
			continue
		}
		if normalizedTrust(source.TrustLevel) == TrustPrimary {
			return true
		}
	}
	return false
}

func score(source artifacts.Source) int {
	trustScore := map[TrustLevel]int{
		TrustPrimary:   100,
		TrustSecondary: 60,
		TrustCommunity: 20,
		TrustUnknown:   0,
	}[normalizedTrust(source.TrustLevel)]
	typeScore := map[Type]int{
		TypeSpecification:       30,
		TypeOfficialDocs:        28,
		TypeSourceCode:          26,
		TypeReleaseNotes:        24,
		TypeMaintainerStatement: 22,
		TypeEngineeringBlog:     12,
		TypeCommunityDiscussion: 4,
		TypeUnknown:             0,
	}[normalizedType(source.Type)]
	return trustScore + typeScore
}

func normalizedTrust(value string) TrustLevel {
	switch TrustLevel(strings.ToLower(strings.TrimSpace(value))) {
	case TrustPrimary:
		return TrustPrimary
	case TrustSecondary:
		return TrustSecondary
	case TrustCommunity:
		return TrustCommunity
	default:
		return TrustUnknown
	}
}

func normalizedType(value string) Type {
	switch Type(strings.ToLower(strings.TrimSpace(value))) {
	case TypeOfficialDocs:
		return TypeOfficialDocs
	case TypeSpecification:
		return TypeSpecification
	case TypeSourceCode:
		return TypeSourceCode
	case TypeReleaseNotes:
		return TypeReleaseNotes
	case TypeMaintainerStatement:
		return TypeMaintainerStatement
	case TypeEngineeringBlog:
		return TypeEngineeringBlog
	case TypeCommunityDiscussion:
		return TypeCommunityDiscussion
	default:
		return TypeUnknown
	}
}
