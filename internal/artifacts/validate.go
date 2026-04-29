package artifacts

import (
	"errors"
	"fmt"
	"strings"
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

// ValidateEpisodeDirectory validates a full episode bundle using the strict
// local validator. It returns a compact error for CLI/workflow usage.
func ValidateEpisodeDirectory(dir string) error {
	if dir == "" {
		return errors.New("episode directory is required")
	}

	report := ValidateEpisodeDirectoryStrict(dir)
	if report.Valid {
		return nil
	}

	messages := make([]string, 0, len(report.Issues))
	for _, issue := range report.Issues {
		location := issue.File
		if issue.Field != "" {
			location = fmt.Sprintf("%s:%s", location, issue.Field)
		}
		if location == "" {
			location = "episode"
		}
		messages = append(messages, fmt.Sprintf("%s: %s", location, issue.Message))
	}
	return fmt.Errorf("episode validation failed: %s", strings.Join(messages, "; "))
}
