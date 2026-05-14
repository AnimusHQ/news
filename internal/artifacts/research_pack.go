package artifacts

import (
	"fmt"
)

// ResearchPackFile is the persisted research_pack.json shape needed by the pipeline.
type ResearchPackFile struct {
	SchemaVersion            string            `json:"schema_version" yaml:"schema_version"`
	EpisodeID                string            `json:"episode_id" yaml:"episode_id"`
	ArtifactID               string            `json:"artifact_id" yaml:"artifact_id"`
	Status                   string            `json:"status" yaml:"status"`
	CoreQuestion             string            `json:"core_question" yaml:"core_question"`
	Audience                 map[string]string `json:"audience,omitempty" yaml:"audience,omitempty"`
	LearningObjectives       []string          `json:"learning_objectives" yaml:"learning_objectives"`
	Sources                  []Source          `json:"sources" yaml:"sources"`
	RequiredTerms            []string          `json:"required_terms,omitempty" yaml:"required_terms,omitempty"`
	SourceSnippets           []SourceSnippet   `json:"source_snippets,omitempty" yaml:"source_snippets,omitempty"`
	ClaimCandidates          []ClaimCandidate  `json:"claim_candidates,omitempty" yaml:"claim_candidates,omitempty"`
	UnresolvedQuestions      []string          `json:"unresolved_questions,omitempty" yaml:"unresolved_questions,omitempty"`
	KnownControversies       []string          `json:"known_controversies,omitempty" yaml:"known_controversies,omitempty"`
	ForbiddenSimplifications []string          `json:"forbidden_simplifications" yaml:"forbidden_simplifications"`
	VisualOpportunities      []string          `json:"visual_opportunities" yaml:"visual_opportunities"`
	RecommendedCTA           string            `json:"recommended_cta,omitempty" yaml:"recommended_cta,omitempty"`
}

// SourceSnippet preserves a supplied source excerpt and locator.
type SourceSnippet struct {
	SourceID string          `json:"source_id" yaml:"source_id"`
	Locator  EvidenceLocator `json:"locator" yaml:"locator"`
	Text     string          `json:"text" yaml:"text"`
}

// ClaimCandidate is an early research-stage claim candidate. It is not
// verified evidence and must still pass claim extraction/verification gates.
type ClaimCandidate struct {
	Text             string            `json:"text" yaml:"text"`
	SourceIDs        []string          `json:"source_ids" yaml:"source_ids"`
	EvidenceLocators []EvidenceLocator `json:"evidence_locators,omitempty" yaml:"evidence_locators,omitempty"`
}

// LoadResearchPackFile loads research pack data from a JSON/YAML artifact path.
func LoadResearchPackFile(path string) (ResearchPackFile, error) {
	var file ResearchPackFile
	if err := decodeArtifact(path, &file); err != nil {
		return ResearchPackFile{}, fmt.Errorf("load research pack artifact: %w", err)
	}
	if len(file.Sources) == 0 {
		return ResearchPackFile{}, fmt.Errorf("research pack contains no sources")
	}
	return file, nil
}
