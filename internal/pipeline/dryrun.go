package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/publishing"
	"github.com/AnimusHQ/news/internal/verification"
)

// DryRunReport is a local, safe summary of the current pipeline readiness.
type DryRunReport struct {
	EpisodeDir           string
	ArtifactsValid       bool
	ValidationIssues     []artifacts.ValidationIssue
	CouncilConsensus     council.Consensus
	CouncilSelected      []string
	CouncilDissent       int
	CouncilBlockers      int
	VerificationDecision string
	VerificationBlockers int
	PublishVisibility    artifacts.PublishVisibility
	PublishDraftID       string
	WorkflowReached      []string
	Warnings             []string
	Blockers             []string
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
	if r.VerificationDecision != "" {
		fmt.Fprintf(&b, "verification_decision: %s\n", r.VerificationDecision)
		fmt.Fprintf(&b, "verification_blocker_count: %d\n", r.VerificationBlockers)
	}
	if r.PublishDraftID != "" {
		fmt.Fprintf(&b, "publish_visibility: %s\n", r.PublishVisibility)
		fmt.Fprintf(&b, "publish_draft_id: %s\n", r.PublishDraftID)
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
			"load_claims",
			"local_multimodel_council",
			"deterministic_claim_verification",
			"generate_publish_pack",
			"dry_run_publish_private_draft",
			"human_qa_required",
			"public_publish_blocked_by_design",
		},
		Warnings: []string{
			"no external model providers are called; council uses deterministic local mock providers",
			"publishing uses local dry-run adapter only; no network call or upload is performed",
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

	claimsFile, err := artifacts.LoadClaimsFile(filepath.Join(episodeDir, "claims.json"))
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("claim loading failed: %v", err))
		return report, err
	}

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

	verificationReport, err := verification.VerifyClaims(claimsFile.Claims, councilResult.Report)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("claim verification failed: %v", err))
		return report, err
	}
	report.VerificationDecision = verificationReport.Decision
	report.VerificationBlockers = len(verificationReport.BlockingIssues)
	if verificationReport.Decision != "approved" {
		report.Warnings = append(report.Warnings, "claim verification requires revision before production publication")
	}

	pack, err := publishing.GeneratePack(publishing.PackInput{
		EpisodeID:     "episode-0001",
		Title:         "What Happens After git push?",
		Summary:       "A source-grounded dry-run publish pack for the pilot episode.",
		Visibility:    artifacts.PublishVisibilityPrivate,
		HumanApproved: false,
		CTA:           "Join the Animus open-source community and follow the source-backed production path.",
	})
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("publish pack generation failed: %v", err))
		return report, err
	}

	publishResult, err := publishing.DryRunAdapter{}.UploadPrivateDraft(context.Background(), pack)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("dry-run publishing failed: %v", err))
		return report, err
	}
	report.PublishVisibility = publishResult.Visibility
	report.PublishDraftID = publishResult.DraftID

	return report, nil
}
