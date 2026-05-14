package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnimusHQ/news/internal/analytics"
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
	GeneratedOutputPaths  []string
	ProductionQAStatus    string
	ProductionQADecision  string
	ProductionQABlockers  []string
	PublishVisibility     artifacts.PublishVisibility
	PublishDraftID        string
	AnalyticsWindow       string
	AnalyticsInsightCount int
	WorkflowReached       []string
	Warnings              []string
	Blockers              []string
}

// DryRunOptions controls local fixture-only dry-run behavior.
type DryRunOptions struct {
	// UseApprovedFixtures consumes the canonical claims artifact and a local
	// approving council fixture. It is intended for integration tests that need
	// to exercise downstream generation without changing the real pilot gates.
	UseApprovedFixtures bool
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
	if len(r.GeneratedOutputPaths) > 0 {
		fmt.Fprintf(&b, "generated_output_paths:\n")
		for _, output := range r.GeneratedOutputPaths {
			fmt.Fprintf(&b, "  - %s\n", output)
		}
	}
	if r.ProductionQAStatus != "" {
		fmt.Fprintf(&b, "production_qa_status: %s\n", r.ProductionQAStatus)
		if r.ProductionQADecision != "" {
			fmt.Fprintf(&b, "production_qa_decision: %s\n", r.ProductionQADecision)
		}
		if len(r.ProductionQABlockers) > 0 {
			fmt.Fprintf(&b, "production_qa_blockers:\n")
			for _, blocker := range r.ProductionQABlockers {
				fmt.Fprintf(&b, "  - %s\n", blocker)
			}
		}
	}
	if r.PublishDraftID != "" {
		fmt.Fprintf(&b, "publish_visibility: %s\n", r.PublishVisibility)
		fmt.Fprintf(&b, "publish_draft_id: %s\n", r.PublishDraftID)
	}
	if r.AnalyticsWindow != "" {
		fmt.Fprintf(&b, "analytics_window: %s\n", r.AnalyticsWindow)
		fmt.Fprintf(&b, "analytics_insight_count: %d\n", r.AnalyticsInsightCount)
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
	return DryRunWithOptions(episodeDir, DryRunOptions{})
}

// DryRunWithOptions executes the local no-network MVP pipeline skeleton.
func DryRunWithOptions(episodeDir string, options DryRunOptions) (DryRunReport, error) {
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
			"import_fixture_analytics",
			"generate_analytics_insights",
			"human_qa_required",
			"public_publish_blocked_by_design",
			"final_summary",
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

	claimsFile, err := claimsForDryRun(episodeDir, options)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("claim preparation failed: %v", err))
		return report, err
	}
	report.GeneratedClaimCount = len(claimsFile.Claims)
	if options.UseApprovedFixtures {
		report.Warnings = append(report.Warnings, "approved fixture mode consumes local claims and council fixtures for downstream dry-run coverage")
	} else {
		claimExtraction, err := claimextractor.ExtractEpisode(episodeDir)
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("claim extraction failed: %v", err))
			return report, err
		}
		report.Warnings = append(report.Warnings, claimExtraction.Warnings...)
	}

	researchAudit, err := research.AuditPack(research.Pack{
		CoreQuestion:             researchFile.CoreQuestion,
		Sources:                  researchFile.Sources,
		LearningObjectives:       researchFile.LearningObjectives,
		ForbiddenSimplifications: researchFile.ForbiddenSimplifications,
		VisualOpportunities:      researchFile.VisualOpportunities,
	}, claimsFile.Claims)
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

	councilResult, err := councilForDryRun(context.Background(), options)
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

	qaPacket, err := qa.Generate(qa.Input{
		EpisodeID:         researchFile.EpisodeID,
		EpisodePurpose:    researchFile.CoreQuestion,
		EpisodeFormat:     "short educational explainer",
		ScriptPath:        filepath.Join(episodeDir, "script.md"),
		ResearchSummary:   researchSummary(researchFile),
		Claims:            claimsFile.Claims,
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
		if options.UseApprovedFixtures {
			transition := artifacts.ValidateTransition(episodeDir, artifacts.StateStoryboarding)
			if !transition.Valid {
				for _, issue := range transition.Issues {
					report.Blockers = append(report.Blockers, issue.Artifact+": "+issue.Reason)
				}
				return report, fmt.Errorf("approved fixture is missing human QA approval for storyboarding")
			}
		}
		script, err := os.ReadFile(filepath.Join(episodeDir, "script.md"))
		if err != nil {
			report.Blockers = append(report.Blockers, fmt.Sprintf("script loading for storyboard failed: %v", err))
			return report, err
		}
		storyboardFile, err = storyboard.Generate(storyboard.Input{
			EpisodeID:             researchFile.EpisodeID,
			ScriptMarkdown:        string(script),
			HumanQARecommendation: qaPacket.RecommendedDecision,
			Claims:                claimsFile.Claims,
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
		renderResult, err = render.GeneratePreview(render.Input{
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
		report.GeneratedOutputPaths = append(report.GeneratedOutputPaths, renderResult.Preview.Path)
	} else {
		report.RenderStatus = "skipped_by_storyboard_gate"
		report.Warnings = append(report.Warnings, "render preview generation skipped because storyboard was not generated")
	}

	if report.RenderStatus == "preview_generated" {
		productionQAReport, err := productionqa.Run(productionqa.Input{
			EpisodeID:             researchFile.EpisodeID,
			Render:                renderResult,
			Claims:                claimsFile.Claims,
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
		report.ProductionQABlockers = append([]string{}, productionQAReport.BlockingIssues...)
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
	report.GeneratedOutputPaths = append(report.GeneratedOutputPaths, "dry-run-publish:"+publishResult.DraftID)

	analyticsReport, err := fixtureAnalyticsReport(context.Background(), researchFile.EpisodeID)
	if err != nil {
		report.Blockers = append(report.Blockers, fmt.Sprintf("analytics fixture import failed: %v", err))
		return report, err
	}
	report.AnalyticsWindow = analyticsReport.Window
	report.AnalyticsInsightCount = len(analyticsReport.Insights)

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

func claimsForDryRun(episodeDir string, options DryRunOptions) (artifacts.ClaimsFile, error) {
	if options.UseApprovedFixtures {
		return artifacts.LoadClaimsFile(filepath.Join(episodeDir, "claims.json"))
	}
	claimExtraction, err := claimextractor.ExtractEpisode(episodeDir)
	if err != nil {
		return artifacts.ClaimsFile{}, err
	}
	return claimExtraction.ClaimsFile, nil
}

func councilForDryRun(ctx context.Context, options DryRunOptions) (CouncilDryRunResult, error) {
	if options.UseApprovedFixtures {
		return approvedCouncilFixture(), nil
	}
	return RunLocalMockCouncil(ctx, DefaultModelRegistryPath)
}

func approvedCouncilFixture() CouncilDryRunResult {
	reviews := []council.ModelReview{
		{ModelID: "mock-technical-approval", Provider: "local-fixture", Task: "technical verification", Verdict: council.VerdictApprove, Confidence: 0.9, Notes: "Fixture claims are source-backed for downstream dry-run coverage."},
		{ModelID: "mock-editorial-approval", Provider: "local-fixture", Task: "clarity and pedagogy", Verdict: council.VerdictApprove, Confidence: 0.86, Notes: "Fixture script is suitable for storyboard generation."},
		{ModelID: "mock-safety-approval", Provider: "local-fixture", Task: "safety and policy", Verdict: council.VerdictApprove, Confidence: 0.88, Notes: "Private dry-run posture is safe."},
	}
	report, err := council.Aggregate(reviews)
	if err != nil {
		return CouncilDryRunResult{}
	}
	return CouncilDryRunResult{
		Report:   report,
		Selected: []string{reviews[0].ModelID, reviews[1].ModelID, reviews[2].ModelID},
	}
}

func fixtureAnalyticsReport(ctx context.Context, episodeID string) (analytics.Report, error) {
	ctr := 0.05
	impressions := 5000
	views := 1200
	averageViewDurationSeconds := 210
	first30Retention := 0.62
	completionRate := 0.44
	subscribersGained := 17
	commentsCount := 8
	shares := 12
	saves := 20
	communityClicks := 48
	costPerEpisode := 120.0
	record := analytics.ProviderRecord{
		Provider:  "fixture-provider",
		EpisodeID: episodeID,
		Window:    analytics.Window72h,
		Metrics: analytics.ProviderMetrics{
			CTR:                        &ctr,
			Impressions:                &impressions,
			Views:                      &views,
			AverageViewDurationSeconds: &averageViewDurationSeconds,
			First30Retention:           &first30Retention,
			CompletionRate:             &completionRate,
			SubscribersGained:          &subscribersGained,
			CommentsCount:              &commentsCount,
			Shares:                     &shares,
			Saves:                      &saves,
			CommunityClicks:            &communityClicks,
			CostPerEpisode:             &costPerEpisode,
		},
	}
	adapter := analytics.FixtureAdapter{Records: map[string]analytics.ProviderRecord{
		episodeID + "|" + analytics.Window72h: record,
	}}
	imported, err := adapter.Import(ctx, analytics.ImportRequest{EpisodeID: episodeID, Window: analytics.Window72h})
	if err != nil {
		return analytics.Report{}, err
	}
	input, err := analytics.Normalize(imported)
	if err != nil {
		return analytics.Report{}, err
	}
	return analytics.GenerateInsightReport(input)
}
