package artifacts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var RequiredEpisodeFiles = []string{
	"topic.yaml",
	"research_pack.json",
	"claims.json",
	"editorial_brief.md",
	"script.md",
	"verification_report.json",
	"multimodel_approval_report.json",
	"human_qa_report.json",
	"storyboard.yaml",
	"asset_manifest.json",
	"render_manifest.json",
	"production_qa_report.json",
	"publish_manifest.json",
	"analytics_report.json",
}

// ValidateEpisodeDirectory performs structural validation for an episode bundle.
// Deeper schema validation is intentionally left to follow-up tasks once all
// canonical Go validators are implemented.
func ValidateEpisodeDirectory(dir string) error {
	if dir == "" {
		return errors.New("episode directory is required")
	}

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("episode directory not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("episode path is not a directory: %s", dir)
	}

	var missing []string
	for _, name := range RequiredEpisodeFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				missing = append(missing, name)
				continue
			}
			return fmt.Errorf("cannot inspect %s: %w", name, err)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("episode is missing required artifacts: %v", missing)
	}
	return nil
}
