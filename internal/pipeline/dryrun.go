package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	claimextractor "github.com/AnimusHQ/news/internal/claims"
	"github.com/AnimusHQ/news/internal/council"
	"github.com/AnimusHQ/news/internal/productionqa"
	"github.com/AnimusHQ/news/internal/publishing"
	"github.com/AnimusHQ/news/internal/qa"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/research"
	"github.com/AnimusHQ/news/internal/storyboard"
	"github.com/AnimusHQ/news/internal/verification"
)

// DryRunReport is a local, safe summary of the current pipeline readiness.
type DryRunReport struct {
	EpisodeDir            string
	ArtifactsValid        bool
	ValidationIssues      []artifacts.ValidationIssue
	ResearchValid         bool
	ResearchBlockers      int
	GeneratedClaimCount   int
	CouncilConsensus      council.Consensus
	CouncilSelected       []string
	CouncilDissent        int
	CouncilBlockers       int
	VerificationDecision  string
	VerificationBlockers  int
	HumanQARecommendation artifacts.HumanDecision
	HumanQAUnresolved     int
	HumanQABlockers       int
	StoryboardStatus      string
	StoryboardSceneCount  int
	RenderStatus          string
	RenderOutputPath      string
	ProductionQAStatus    string
	ProductionQADecision  string
	PublishVisibility     artifacts.PublishVisibility
	PublishDraftID        string
	WorkflowReached       []string
	Warnings              []string
	Blockers              []string
}

func (r DryRunReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Animus News dry run\n")
	fmt.Fprintf(&b, "episode: %s\n", r.EpisodeDir)
	fmt.Fprintf(&b, "artifacts_valid: %t\n", r.ArtifactsValid)
	fmt.Fprintf(&b, "research_valid: %t\n", r.ResearchValid)
	fmt.Fprintf(&b, "research_blocker_count: %d\n", r.ResearchBlockers)
	fmt.Fprintf(&b, "generated_claim_count: %d\n", r.GeneratedClaimCount)
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
	if r.HumanQARecommendation != "" {
		fmt.Fprintf(&b, "human_qa_recommendation: %s\n", r.HumanQARecommendation)
		fmt.Fprintf(&b, "human_qa_unresolved_claim_count: %d\n", r.HumanQAUnresolved)
		fmt.Fprintf(&b, "human_qa_blocker_count: %d\n", r.HumanQABlockers)
	}
	if r.StoryboardStatus != "" {
		fmt.Fprintf(&b, "storyboard_status: %s\n", r.StoryboardStatus)
		fmt.Fprintf(&b, "storyboard_scene_count: %d\n", r.StoryboardSceneCount)
	}
	if r.RenderStatus != "" {
		fmt.Fprintf(&b, "render_status: %s\n", r.RenderStatus)
		if r.RenderOutputPath != "" {
			fmt.Fprintf(&b, "render_output_path: %s\n", r.RenderOutputPath)
		}
	}
	if r.ProductionQAStatus != "" {
		fmt.Fprintf(&b, "production_qa_status: %s\n", r.ProductionQAStatus)
		if r.ProductionQADecision != "" {
			fmt.Fprintf(&b, "production_qa_decision: %s\n", r.ProductionQADecision)
		}
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
			"load_research_pack",
			"extract_claims_from_script",
			"audit_research_authority",
			"local_multimodel_council",
			"deterministic_claim_verification",
			"generate_human_qa_packet",
			"storyboard_gate_checked",
			"render_preview_gate_checked",
			"production_qa_gate_checked",
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

	researchFile, err := artifacts.LoadResearchPackFile(filepath.Join(episodeDir, "research_pack.json"))
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("research pack loading failed: %v", err))
		return report, err
	}

	claimExtraction, err := claimextractor.ExtractEpisode(episodeDir)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("claim extraction failed: %v", err))
		return report, err
	}
	report.GeneratedClaimCount = len(claimExtraction.ClaimsFile.Claims)
	report.Warnings = append(report.Warnings, claimExtraction.Warnings...)

	researchAudit, err := research.AuditPack(research.Pack{
		CoreQuestion:             researchFile.CoreQuestion,
		Sources:                  researchFile.Sources,
		LearningObjectives:       researchFile.LearningObjectives,
		ForbiddenSimplifications: researchFile.ForbiddenSimplifications,
		VisualOpportunities:      researchFile.VisualOpportunities,
	}, claimExtraction.ClaimsFile.Claims)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("research audit failed: %v", err))
		return report, err
	}
	report.ResearchValid = researchAudit.Valid
	report.ResearchBlockers = len(researchAudit.Blockers)
	report.Warnings = append(report.Warnings, researchAudit.Warnings...)
	if !researchAudit.Valid {
		report.Blockers = append(report.Blockers, researchAudit.Blockers...)
		return report, fmt.Errorf("research audit failed")
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

	verificationReport, err := verification.VerifyClaims(claimExtraction.ClaimsFile.Claims, councilResult.Report)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("claim verification failed: %v", err))
		return report, err
	}
	report.VerificationDecision = verificationReport.Decision
	report.VerificationBlockers = len(verificationReport.BlockingIssues)
	if verificationReport.Decision != "approved" {
		report.Warnings = append(report.Warnings, "claim verification requires revision before production publication")
	}

	qaPacket, err := qa.Generate(qa.Input{
		EpisodeID:         researchFile.EpisodeID,
		EpisodePurpose:    researchFile.CoreQuestion,
		EpisodeFormat:     "short educational explainer",
		ScriptPath:        filepath.Join(episodeDir, "script.md"),
		ResearchSummary:   researchSummary(researchFile),
		Claims:            claimExtraction.ClaimsFile.Claims,
		Verification:      verificationReport,
		Council:           councilResult.Report,
		QualityGateStatus: "dry_run",
	})
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("human QA packet generation failed: %v", err))
		return report, err
	}
	report.HumanQARecommendation = qaPacket.RecommendedDecision
	report.HumanQAUnresolved = len(qaPacket.UnresolvedClaims)
	report.HumanQABlockers = len(qaPacket.BlockingIssues)
	if qaPacket.RecommendedDecision == artifacts.HumanDecisionBlock {
		report.Blockers = append(report.Blockers, "human QA packet recommends blocking the episode")
		return report, fmt.Errorf("human QA packet recommends blocking the episode")
	}
	if qaPacket.RecommendedDecision != artifacts.HumanDecisionApprove {
		report.Warnings = append(report.Warnings, "human QA packet requires operator review before release")
	}

	var storyboardFile storyboard.File
	if canGenerateStoryboard(qaPacket.RecommendedDecision) {
		script, err := os.ReadFile(filepath.Join(episodeDir, "script.md"))
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("script loading for storyboard failed: %v", err))
			return report, err
		}
		storyboardFile, err = storyboard.Generate(storyboard.Input{
			EpisodeID:             researchFile.EpisodeID,
			ScriptMarkdown:        string(script),
			HumanQARecommendation: qaPacket.RecommendedDecision,
			Claims:                claimExtraction.ClaimsFile.Claims,
		})
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("storyboard generation failed: %v", err))
			return report, err
		}
		report.StoryboardStatus = "generated"
		report.StoryboardSceneCount = len(storyboardFile.Scenes)
	} else {
		report.StoryboardStatus = "skipped_by_human_qa_gate"
		report.Warnings = append(report.Warnings, "storyboard generation skipped because human QA has not approved this episode")
	}

	var renderResult render.Result
	if report.StoryboardStatus == "generated" {
		renderResult, err := render.GeneratePreview(render.Input{
			EpisodeID:  researchFile.EpisodeID,
			Storyboard: storyboardFile,
			OutputDir:  "dist",
		})
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("render preview generation failed: %v", err))
			return report, err
		}
		report.RenderStatus = "preview_generated"
		report.RenderOutputPath = renderResult.Preview.Path
	} else {
		report.RenderStatus = "skipped_by_storyboard_gate"
		report.Warnings = append(report.Warnings, "render preview generation skipped because storyboard was not generated")
	}

	if report.RenderStatus == "preview_generated" {
		productionQAReport, err := productionqa.Run(productionqa.Input{
			EpisodeID:             researchFile.EpisodeID,
			Render:                renderResult,
			Claims:                claimExtraction.ClaimsFile.Claims,
			Verification:          verificationReport,
			HumanQARecommendation: qaPacket.RecommendedDecision,
			PublishVisibility:     artifacts.PublishVisibilityPrivate,
		})
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("production QA failed: %v", err))
			return report, err
		}
		report.ProductionQAStatus = "checked"
		report.ProductionQADecision = productionQAReport.Decision
		if productionQAReport.Decision != productionqa.DecisionApproved {
			report.Warnings = append(report.Warnings, "production QA did not approve release")
		}
	} else {
		report.ProductionQAStatus = "skipped_by_render_gate"
		report.Warnings = append(report.Warnings, "production QA skipped because render preview was not generated")
	}

	pack, err := publishing.GeneratePack(publishing.PackInput{
		EpisodeID:     "episode-0001",
		Title:         "What Happens After git push?",
		Summary:       "A source-grounded dry-run publish pack for the pilot episode.",
		Sources:       researchFile.Sources,
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

func researchSummary(file artifacts.ResearchPackFile) string {
	var parts []string
	if file.CoreQuestion != "" {
		parts = append(parts, "Core question: "+file.CoreQuestion)
	}
	if len(file.LearningObjectives) > 0 {
		parts = append(parts, "Learning objectives: "+strings.Join(file.LearningObjectives, "; "))
	}
	if len(file.Sources) > 0 {
		parts = append(parts, fmt.Sprintf("Source count: %d", len(file.Sources)))
	}
	return strings.Join(parts, " | ")
}

func canGenerateStoryboard(decision artifacts.HumanDecision) bool {
	return decision == artifacts.HumanDecisionApprove || decision == artifacts.HumanDecisionApproveWithMinorEdits
}
