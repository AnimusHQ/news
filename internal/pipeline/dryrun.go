package pipeline

import (
	"fmt"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
)

// DryRunReport is a local, safe summary of the current pipeline readiness.
type DryRunReport struct {
	EpisodeDir      string
	ArtifactsValid  bool
	WorkflowReached []string
	Warnings        []string
	Blockers        []string
}

func (r DryRunReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Animus News dry run\n")
	fmt.Fprintf(&b, "episode: %s\n", r.EpisodeDir)
	fmt.Fprintf(&b, "artifacts_valid: %t\n", r.ArtifactsValid)
	fmt.Fprintf(&b, "workflow_reached: %s\n", strings.Join(r.WorkflowReached, " -> "))
	if len(r.Warnings) > 0 {
		fmt.Fprintf(&b, "warnings:\n")
		for _, warning := range r.Warnings {
			fmt.Fprintf(&b, "  - %s\n", warning)
		}
	}
	if len(r.Blockers) > 0 {
		fmt.Fprintf(&b, "blockers:\n")
		for _, blocker := range r.Blockers {
			fmt.Fprintf(&b, "  - %s\n", blocker)
		}
	}
	return b.String()
}

// DryRun executes the local no-network MVP pipeline skeleton.
func DryRun(episodeDir string) (DryRunReport, error) {
	report := DryRunReport{
		EpisodeDir: episodeDir,
		WorkflowReached: []string{
			"validate_artifacts",
			"mock_research_ready",
			"mock_claim_verification",
			"mock_multimodel_council",
			"human_qa_required",
			"dry_run_publish_blocked_by_design",
		},
		Warnings: []string{
			"dry run currently validates structural artifact presence only; deep schema validation is a follow-up task",
			"no model providers are called; multimodel council is represented as a safe placeholder",
			"no publishing adapter is executed; public publishing is unavailable by design",
		},
	}

	if err := artifacts.ValidateEpisodeDirectory(episodeDir); err != nil {
		report.Blockers = append(report.Blockers, err.Error())
		return report, err
	}

	report.ArtifactsValid = true
	return report, nil
}
