package research

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/sources"
)

// BuilderInput contains only explicitly supplied research material.
type BuilderInput struct {
	EpisodeID                string
	ArtifactID               string
	CoreQuestion             string
	Audience                 map[string]string
	Sources                  []artifacts.Source
	Snippets                 []artifacts.SourceSnippet
	LearningObjectives       []string
	RequiredTerms            []string
	UnresolvedQuestions      []string
	KnownControversies       []string
	ForbiddenSimplifications []string
	VisualOpportunities      []string
	RecommendedCTA           string
	HighRiskTopic            bool
}

// BuildResult includes the generated pack and deterministic diagnostics.
type BuildResult struct {
	Pack     artifacts.ResearchPackFile
	Warnings []string
	Blockers []string
}

// BuildPack constructs a draft research pack without network access or source
// fabrication.
func BuildPack(input BuilderInput) (BuildResult, error) {
	if strings.TrimSpace(input.EpisodeID) == "" {
		return BuildResult{}, fmt.Errorf("episode id is required")
	}
	if strings.TrimSpace(input.CoreQuestion) == "" {
		return BuildResult{}, fmt.Errorf("core question is required")
	}
	registry, err := sources.NewRegistry(input.Sources)
	if err != nil {
		return BuildResult{}, err
	}

	rankedSources := registry.Rank()
	result := BuildResult{
		Pack: artifacts.ResearchPackFile{
			SchemaVersion:            "1.0",
			EpisodeID:                input.EpisodeID,
			ArtifactID:               artifactID(input),
			Status:                   string(artifacts.ArtifactStatusDraft),
			CoreQuestion:             strings.TrimSpace(input.CoreQuestion),
			Audience:                 copyStringMap(input.Audience),
			LearningObjectives:       sortedNonEmpty(input.LearningObjectives),
			Sources:                  rankedSources,
			RequiredTerms:            sortedNonEmpty(input.RequiredTerms),
			SourceSnippets:           normalizeSnippets(input.Snippets),
			ClaimCandidates:          claimCandidates(input.Snippets),
			UnresolvedQuestions:      sortedNonEmpty(input.UnresolvedQuestions),
			KnownControversies:       sortedNonEmpty(input.KnownControversies),
			ForbiddenSimplifications: sortedNonEmpty(input.ForbiddenSimplifications),
			VisualOpportunities:      sortedNonEmpty(input.VisualOpportunities),
			RecommendedCTA:           strings.TrimSpace(input.RecommendedCTA),
		},
	}

	if len(result.Pack.LearningObjectives) == 0 {
		result.Warnings = append(result.Warnings, "no learning objectives supplied")
	}
	if len(result.Pack.ForbiddenSimplifications) == 0 {
		result.Warnings = append(result.Warnings, "no forbidden simplifications supplied")
	}
	if len(result.Pack.SourceSnippets) == 0 {
		result.Blockers = append(result.Blockers, "at least one source snippet is required")
	}
	if input.HighRiskTopic && !hasPrimarySource(rankedSources) {
		result.Blockers = append(result.Blockers, "high-risk topic requires at least one primary source")
	}
	for _, snippet := range result.Pack.SourceSnippets {
		if _, ok := registry.Get(snippet.SourceID); !ok {
			result.Blockers = append(result.Blockers, fmt.Sprintf("snippet references unknown source id: %s", snippet.SourceID))
		}
	}
	sort.Strings(result.Warnings)
	sort.Strings(result.Blockers)
	return result, nil
}

func artifactID(input BuilderInput) string {
	if strings.TrimSpace(input.ArtifactID) != "" {
		return strings.TrimSpace(input.ArtifactID)
	}
	return "research-pack-" + input.EpisodeID + "-generated-v1"
}

func normalizeSnippets(snippets []artifacts.SourceSnippet) []artifacts.SourceSnippet {
	out := append([]artifacts.SourceSnippet(nil), snippets...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SourceID != out[j].SourceID {
			return out[i].SourceID < out[j].SourceID
		}
		if out[i].Locator.Section != out[j].Locator.Section {
			return out[i].Locator.Section < out[j].Locator.Section
		}
		return out[i].Locator.Range < out[j].Locator.Range
	})
	return out
}

func claimCandidates(snippets []artifacts.SourceSnippet) []artifacts.ClaimCandidate {
	var candidates []artifacts.ClaimCandidate
	for _, snippet := range normalizeSnippets(snippets) {
		text := strings.TrimSpace(snippet.Text)
		if text == "" {
			continue
		}
		candidates = append(candidates, artifacts.ClaimCandidate{
			Text:             text,
			SourceIDs:        []string{snippet.SourceID},
			EvidenceLocators: []artifacts.EvidenceLocator{snippet.Locator},
		})
	}
	return candidates
}

func hasPrimarySource(items []artifacts.Source) bool {
	for _, source := range items {
		if strings.EqualFold(strings.TrimSpace(source.TrustLevel), string(sources.TrustPrimary)) {
			return true
		}
	}
	return false
}

func sortedNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := map[string]string{}
	for key, value := range values {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			out[key] = value
		}
	}
	return out
}
