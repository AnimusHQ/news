package artifacts

import "fmt"

// ClaimsFile is the persisted claims artifact shape.
type ClaimsFile struct {
	SchemaVersion string  `json:"schema_version" yaml:"schema_version"`
	EpisodeID     string  `json:"episode_id" yaml:"episode_id"`
	ArtifactID    string  `json:"artifact_id" yaml:"artifact_id"`
	Status        string  `json:"status" yaml:"status"`
	Claims        []Claim `json:"claims" yaml:"claims"`
}

// LoadClaimsFile loads claims from a JSON/YAML artifact path.
func LoadClaimsFile(path string) (ClaimsFile, error) {
	var file ClaimsFile
	if err := decodeArtifact(path, &file); err != nil {
		return ClaimsFile{}, fmt.Errorf("load claims artifact: %w", err)
	}
	if len(file.Claims) == 0 {
		return ClaimsFile{}, fmt.Errorf("claims artifact contains no claims")
	}
	return file, nil
}
