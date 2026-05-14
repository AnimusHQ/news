package artifacts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

var knownArtifactFiles = map[string]bool{
	"topic.yaml":                      true,
	"research_pack.json":              true,
	"claims.json":                     true,
	"editorial_brief.md":              true,
	"script.md":                       true,
	"verification_report.json":        true,
	"multimodel_approval_report.json": true,
	"human_qa_report.json":            true,
	"storyboard.yaml":                 true,
	"asset_manifest.json":             true,
	"render_manifest.json":            true,
	"production_qa_report.json":       true,
	"publish_manifest.json":           true,
	"analytics_report.json":           true,
}

// ValidatePath validates either one canonical artifact or a complete episode
// directory. It is used by the CLI and keeps all checks local/deterministic.
func ValidatePath(path string) ValidationReport {
	report := ValidationReport{EpisodeDir: path, Valid: true}
	if path == "" {
		report.add("", "path", "path is required")
		return report
	}
	info, err := os.Stat(path)
	if err != nil {
		report.add(filepath.Base(path), "", fmt.Sprintf("path not accessible: %v", err))
		return report
	}
	if info.IsDir() {
		return ValidateEpisodeDirectoryStrict(path)
	}
	validateArtifactFile(&report, path)
	return report
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
	return validationError(report)
}

func validationError(report ValidationReport) error {
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
	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

func ValidateReport(report ValidationReport) error {
	return validationError(report)
}

func validateArtifactFile(report *ValidationReport, path string) {
	name := filepath.Base(path)
	if !knownArtifactFiles[name] {
		report.add(name, "", "unknown canonical artifact filename")
		return
	}
	if name == "script.md" || name == "editorial_brief.md" {
		data, err := os.ReadFile(path)
		if err != nil {
			report.add(name, "", fmt.Sprintf("cannot read markdown artifact: %v", err))
			return
		}
		if strings.TrimSpace(string(data)) == "" {
			report.add(name, "", "markdown artifact must not be empty")
		}
		return
	}
	if !isMachineReadableArtifact(name) {
		report.add(name, "", "unsupported artifact type")
		return
	}

	var envelope ArtifactEnvelope
	if err := decodeArtifact(path, &envelope); err != nil {
		report.add(name, "", fmt.Sprintf("cannot decode artifact: %v", err))
		return
	}
	validateEnvelope(report, name, envelope)
	validateArtifactSchema(report, name, path)
}
