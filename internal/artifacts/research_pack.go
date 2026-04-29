package artifacts

import (
	"fmt"
)

// ResearchPackFile is the persisted research_pack.json shape needed by the pipeline.
type ResearchPackFile struct {
	SchemaVersion            string   `json:"schema_version" yaml:"schema_version"`
	EpisodeID                string   `json:"episode_id" yaml:"episode_id"`
	ArtifactID               string   `json:"artifact_id" yaml:"artifact_id"`
	Status                   string   `json:"status" yaml:"status"`
	CoreQuestion             string   `json:"core_question" yaml:"core_question"`
	LearningObjectives       []string `json:"learning_objectives" yaml:"learning_objectives"`
	Sources                  []Source `json:"sources" yaml:"sources"`
	ForbiddenSimplifications []string `json:"forbidden_simplifications" yaml:"forbidden_simplifications"`
	VisualOpportunities      []string `json:"visual_opportunities" yaml:"visual_opportunities"`
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
