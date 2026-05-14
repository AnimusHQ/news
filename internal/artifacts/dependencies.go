package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LifecycleState identifies transition targets with artifact requirements.
type LifecycleState string

const (
	StateResearchReady LifecycleState = "research_ready"
	StateVerifying     LifecycleState = "verifying"
	StateStoryboarding LifecycleState = "storyboarding"
	StateRendering     LifecycleState = "rendering"
	StateScheduled     LifecycleState = "scheduled"
	StatePublished     LifecycleState = "published"
)

// DependencyIssue is a machine-readable dependency gate failure.
type DependencyIssue struct {
	Artifact string `json:"artifact"`
	Reason   string `json:"reason"`
}

// DependencyReport summarizes transition readiness.
type DependencyReport struct {
	State  LifecycleState    `json:"state"`
	Valid  bool              `json:"valid"`
	Issues []DependencyIssue `json:"issues,omitempty"`
}

type dependencyEnvelope struct {
	ArtifactEnvelope `json:",inline" yaml:",inline"`
	SourceArtifacts  []string `json:"source_artifacts" yaml:"source_artifacts"`
	ContentHash      string   `json:"content_hash" yaml:"content_hash"`
}

var transitionRequirements = map[LifecycleState][]string{
	StateResearchReady: {"research_pack.json"},
	StateVerifying:     {"claims.json", "script.md"},
	StateStoryboarding: {"human_qa_report.json"},
	StateRendering:     {"storyboard.yaml", "asset_manifest.json"},
	StateScheduled:     {"production_qa_report.json", "publish_manifest.json"},
	StatePublished:     {"publish_manifest.json", "analytics_report.json"},
}

// ValidateTransition validates the artifact dependencies required to enter a
// lifecycle state. It is intentionally stricter than draft bundle validation.
func ValidateTransition(dir string, state LifecycleState) DependencyReport {
	report := DependencyReport{State: state, Valid: true}
	required, ok := transitionRequirements[state]
	if !ok {
		report.add("", fmt.Sprintf("unknown lifecycle state: %s", state))
		return report
	}
	for _, artifact := range required {
		path := filepath.Join(dir, artifact)
		if _, err := os.Stat(path); err != nil {
			report.add(artifact, "required artifact is missing")
			continue
		}
		validation := ValidatePath(path)
		for _, issue := range validation.Issues {
			report.add(artifact, issue.Message)
		}
		if isMachineReadableArtifact(artifact) {
			envelope := dependencyEnvelope{}
			if err := decodeArtifact(path, &envelope); err != nil {
				report.add(artifact, fmt.Sprintf("cannot decode artifact dependencies: %v", err))
				continue
			}
			if envelope.Status == string(ArtifactStatusRejected) || envelope.Status == string(ArtifactStatusSuperseded) {
				report.add(artifact, fmt.Sprintf("required artifact status is %s", envelope.Status))
			}
			for _, dependency := range envelope.SourceArtifacts {
				if err := validateSourceDependency(dir, dependency); err != nil {
					report.add(artifact, err.Error())
				}
			}
		}
	}
	validateStateSpecificGate(dir, state, &report)
	return report
}

func validateStateSpecificGate(dir string, state LifecycleState, report *DependencyReport) {
	switch state {
	case StateStoryboarding:
		decision := humanDecisionForTransition(filepath.Join(dir, "human_qa_report.json"))
		if decision != string(HumanDecisionApprove) && decision != string(HumanDecisionApproveWithMinorEdits) {
			report.add("human_qa_report.json", "human QA approval is required for storyboarding")
		}
	case StateScheduled:
		decision := productionQADecision(filepath.Join(dir, "production_qa_report.json"))
		if decision != "approved" {
			report.add("production_qa_report.json", "approved production QA is required for scheduling")
		}
	case StatePublished:
		approval := publishApproval(filepath.Join(dir, "publish_manifest.json"))
		if !approval {
			report.add("publish_manifest.json", "human release approval is required for publication")
		}
	}
}

func validateSourceDependency(dir string, dependency string) error {
	artifact, expectedHash, hasHash := strings.Cut(strings.TrimSpace(dependency), "@")
	if artifact == "" {
		return fmt.Errorf("source artifact reference is empty")
	}
	path := filepath.Join(dir, artifact)
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("source artifact %s is missing", artifact)
	}
	if !hasHash || expectedHash == "" {
		return nil
	}
	actual, err := FileContentHash(path)
	if err != nil {
		return err
	}
	if actual != expectedHash {
		return fmt.Errorf("source artifact %s hash mismatch: expected %s got %s", artifact, expectedHash, actual)
	}
	return nil
}

// FileContentHash returns a deterministic sha256 content hash for dependency checks.
func FileContentHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func humanDecisionForTransition(path string) string {
	var qa humanQAForValidation
	if err := decodeArtifact(path, &qa); err != nil {
		return ""
	}
	return qa.Decision
}

func productionQADecision(path string) string {
	var report struct {
		Decision string `json:"decision" yaml:"decision"`
	}
	if err := decodeArtifact(path, &report); err != nil {
		return ""
	}
	return report.Decision
}

func publishApproval(path string) bool {
	var manifest publishManifestForValidation
	if err := decodeArtifact(path, &manifest); err != nil {
		return false
	}
	return manifest.HumanReleaseApproval
}

func (r *DependencyReport) add(artifact string, reason string) {
	r.Valid = false
	r.Issues = append(r.Issues, DependencyIssue{Artifact: artifact, Reason: reason})
}
