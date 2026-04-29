package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
)

// DryRunReport is a local, safe summary of the current pipeline readiness.
type DryRunReport struct {
	EpisodeDir        string
	ArtifactsValid    bool
	ValidationIssues  []artifacts.ValidationIssue
	CouncilConsensus  council.Consensus
	CouncilSelected   []string
	CouncilDissent    int
	CouncilBlockers   int
	WorkflowReached   []string
	Warnings          []string
	Blockers          []string
}

func (r DryRunReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Animus News dry run\n")
	fmt.Fprintf(&b, "episode: %s\n", r.EpisodeDir)
	fmt.Fprintf(&b, "artifacts_valid: %t\n", r.ArtifactsValid)
	if r.CouncilConsensus != "" {
		fmt.Fprintf(&b, "council_consensus: %s\n", r.CouncilConsensus)
		fmt.Fprintf(&b, "council_selected: %s\n", strings.Join(r.CouncilSelected, ", "))
		fmt.Fprintf(&b, "council_dissent_count: %d\n", r.CouncilDissent)
		fmt.Fprintf(&b, "council_blocker_count: %d\n", r.CouncilBlockers)
	}
	fmt.Fprintf(&b, "workflow_reached: %s\n", strings.Join(r.WorkflowReached, " -> "))
	if len(r.ValidationIssues) > 0 {
		fmt.Fprintf(&b, "validation_issues:\n")
		for _, issue := range r.ValidationIssues {
			location := issue.File
			if issue.Field != "" {
				location = location + ":" + issue.Field
			}
			fmt.Fprintf(&b, "  - %s: %s\n", location, issue.Message)
		}
	}
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
			"local_multimodel_council",
			"human_qa_required",
			"dry_run_publish_blocked_by_design",
		},
		Warnings: []string{
			"no external model providers are called; council uses deterministic local mock providers",
			"no publishing adapter is executed; public publishing is unavailable by design",
			"pilot artifacts are draft/dry-run artifacts and must not be treated as public-release approval",
		},
	}

	validation := artifacts.ValidateEpisodeDirectoryStrict(episodeDir)
	report.ValidationIssues = validation.Issues
	if !validation.Valid {
		report.Blockers = append(report.Blockers, "artifact validation failed")
		return report, artifacts.ValidateEpisodeDirectory(episodeDir)
	}
	report.ArtifactsValid = true

	councilResult, err := RunLocalMockCouncil(context.Background(), DefaultModelRegistryPath)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("local multimodel council failed: %v", err))
		return report, err
	}
	report.CouncilConsensus = councilResult.Report.Consensus
	report.CouncilSelected = councilResult.Selected
	report.CouncilDissent = len(councilResult.Report.Dissent)
	report.CouncilBlockers = len(councilResult.Report.BlockingObjections)
	if councilResult.Report.Consensus == council.ConsensusRevisionRequired {
		report.Warnings = append(report.Warnings, "local council requires revision before production publication")
	}
	if councilResult.Report.Consensus == council.ConsensusBlocked {
		report.Blockers = append(report.Blockers, "local council blocked the artifact")
		return report, fmt.Errorf("local council blocked the artifact")
	}

	return report, nil
}
