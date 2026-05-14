package artifacts

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ArtifactEnvelope captures fields common to all JSON/YAML artifacts.
type ArtifactEnvelope struct {
	SchemaVersion   string   `json:"schema_version" yaml:"schema_version"`
	EpisodeID       string   `json:"episode_id" yaml:"episode_id"`
	ArtifactID      string   `json:"artifact_id" yaml:"artifact_id"`
	CreatedAt       string   `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy       string   `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	SourceArtifacts []string `json:"source_artifacts,omitempty" yaml:"source_artifacts,omitempty"`
	ContentHash     string   `json:"content_hash,omitempty" yaml:"content_hash,omitempty"`
	Status          string   `json:"status" yaml:"status"`
}

// ValidationIssue is a machine-readable artifact validation finding.
type ValidationIssue struct {
	File    string `json:"file"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ValidationReport summarizes structural and semantic validation for an episode.
type ValidationReport struct {
	EpisodeDir string            `json:"episode_dir"`
	Valid      bool              `json:"valid"`
	Issues     []ValidationIssue `json:"issues,omitempty"`
}

func (r *ValidationReport) add(file, field, message string) {
	r.Valid = false
	r.Issues = append(r.Issues, ValidationIssue{File: file, Field: field, Message: message})
}

// ValidateEpisodeDirectoryStrict validates the required artifact bundle and a
// minimal set of release-safety invariants. It intentionally does not require
// production approval for draft/dry-run bundles, but it does block unsafe
// publication defaults.
func ValidateEpisodeDirectoryStrict(dir string) ValidationReport {
	report := ValidationReport{EpisodeDir: dir, Valid: true}

	info, err := os.Stat(dir)
	if err != nil {
		report.add("", "episode_dir", fmt.Sprintf("episode directory not accessible: %v", err))
		return report
	}
	if !info.IsDir() {
		report.add("", "episode_dir", "episode path is not a directory")
		return report
	}

	for _, name := range RequiredEpisodeFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				report.add(name, "", "required artifact is missing")
				continue
			}
			report.add(name, "", fmt.Sprintf("cannot inspect artifact: %v", err))
			continue
		}

		if isMachineReadableArtifact(name) {
			var envelope ArtifactEnvelope
			if err := decodeArtifact(path, &envelope); err != nil {
				report.add(name, "", fmt.Sprintf("cannot decode artifact: %v", err))
				continue
			}
			validateEnvelope(&report, name, envelope)
			validateArtifactSchema(&report, name, path)
		}
	}

	return report
}

func isMachineReadableArtifact(name string) bool {
	return strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}

func decodeArtifact(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return json.Unmarshal(data, out)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, out)
	default:
		return fmt.Errorf("unsupported artifact extension: %s", filepath.Ext(path))
	}
}

func validateEnvelope(report *ValidationReport, file string, envelope ArtifactEnvelope) {
	if envelope.SchemaVersion == "" {
		report.add(file, "schema_version", "schema_version is required")
	} else if envelope.SchemaVersion != "1.0" {
		report.add(file, "schema_version", "unsupported schema_version: "+envelope.SchemaVersion)
	}
	if envelope.EpisodeID == "" {
		report.add(file, "episode_id", "episode_id is required")
	}
	if envelope.ArtifactID == "" {
		report.add(file, "artifact_id", "artifact_id is required")
	}
	if envelope.Status == "" {
		report.add(file, "status", "status is required")
	} else if !validArtifactStatus(envelope.Status) {
		report.add(file, "status", "unsupported artifact status: "+envelope.Status)
	}
	if envelope.CreatedAt != "" {
		if _, err := time.Parse(time.RFC3339, envelope.CreatedAt); err != nil {
			report.add(file, "created_at", "created_at must be RFC3339 when present")
		}
	}
	for i, source := range envelope.SourceArtifacts {
		if strings.TrimSpace(source) == "" {
			report.add(file, fmt.Sprintf("source_artifacts[%d]", i), "source artifact reference must not be empty")
		}
	}
	if envelope.ContentHash != "" && !strings.HasPrefix(envelope.ContentHash, "sha256:") {
		report.add(file, "content_hash", "content_hash must use sha256: prefix when present")
	}
}

type publishManifestForValidation struct {
	Visibility           PublishVisibility `json:"visibility" yaml:"visibility"`
	HumanReleaseApproval bool              `json:"human_release_approval" yaml:"human_release_approval"`
}

func validatePublishManifestSafety(report *ValidationReport, path string) {
	var manifest publishManifestForValidation
	if err := decodeArtifact(path, &manifest); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			report.add(filepath.Base(path), "", fmt.Sprintf("cannot decode publish manifest: %v", err))
		}
		return
	}

	if manifest.Visibility == "" {
		report.add(filepath.Base(path), "visibility", "visibility is required")
		return
	}
	if manifest.Visibility == PublishVisibilityPublic && !manifest.HumanReleaseApproval {
		report.add(filepath.Base(path), "visibility", "public visibility requires explicit human release approval")
	}
}

type humanQAForValidation struct {
	Decision string `json:"decision" yaml:"decision"`
}

func validateHumanQAExplicit(report *ValidationReport, path string) {
	var qa humanQAForValidation
	if err := decodeArtifact(path, &qa); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			report.add(filepath.Base(path), "", fmt.Sprintf("cannot decode human QA report: %v", err))
		}
		return
	}
	if qa.Decision == "" {
		report.add(filepath.Base(path), "decision", "human QA decision is required")
	}
}

type claimsFileForValidation struct {
	Claims []Claim `json:"claims" yaml:"claims"`
}

func validateClaimsCoverage(report *ValidationReport, path string) {
	var claimSet claimsFileForValidation
	if err := decodeArtifact(path, &claimSet); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			report.add(filepath.Base(path), "", fmt.Sprintf("cannot decode claims: %v", err))
		}
		return
	}
	if len(claimSet.Claims) == 0 {
		report.add(filepath.Base(path), "claims", "at least one claim is expected for pilot episode")
		return
	}
	for _, claim := range claimSet.Claims {
		if claim.ID == "" {
			report.add(filepath.Base(path), "claim_id", "claim_id is required")
		}
		if claim.Text == "" {
			report.add(filepath.Base(path), "text", "claim text is required")
		}
		if len(claim.SourceIDs) == 0 {
			report.add(filepath.Base(path), claim.ID, "claim must reference at least one source")
		}
		if (claim.RiskLevel == ClaimRiskHigh || claim.RiskLevel == ClaimRiskCritical) && len(claim.EvidenceLocators) == 0 {
			report.add(filepath.Base(path), claim.ID, "high/critical risk claim requires evidence locator")
		}
	}
}
