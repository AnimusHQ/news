package activities

import (
	"context"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// ValidateEpisodeActivity validates the episode artifact bundle.
func ValidateEpisodeActivity(ctx context.Context, episodeDir string) (string, error) {
	if err := artifacts.ValidateEpisodeDirectory(episodeDir); err != nil {
		return "", err
	}
	return "artifact validation passed", nil
}

// MockCouncilActivity is a safe placeholder for the future multimodel council.
func MockCouncilActivity(ctx context.Context, episodeID string) (string, error) {
	return "mock multimodel council approved with no external model calls", nil
}

// ProductionQAActivity is a safe placeholder for future production QA checks.
func ProductionQAActivity(ctx context.Context, episodeID string) (string, error) {
	return "production QA placeholder passed", nil
}

// DryRunPublishActivity is intentionally non-public and no-network.
func DryRunPublishActivity(ctx context.Context, episodeID string) (string, error) {
	return "dry-run publish completed without upload", nil
}
