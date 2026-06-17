package artifacts

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func validateArtifactSchema(report *ValidationReport, name string, path string) {
	switch name {
	case "topic.yaml":
		validateTopicArtifact(report, name, path)
	case "research_pack.json":
		validateResearchPackArtifact(report, name, path)
	case "claims.json":
		validateClaimsArtifact(report, name, path)
	case "verification_report.json":
		validateVerificationReportArtifact(report, name, path)
	case "multimodel_approval_report.json":
		validateMultimodelApprovalArtifact(report, name, path)
	case "human_qa_report.json":
		validateHumanQAArtifact(report, name, path)
	case "storyboard.yaml":
		validateStoryboardArtifact(report, name, path)
	case "asset_manifest.json":
		validateAssetManifestArtifact(report, name, path)
	case "render_manifest.json":
		validateRenderManifestArtifact(report, name, path)
	case "production_qa_report.json":
		validateProductionQAArtifact(report, name, path)
	case "publish_manifest.json":
		validatePublishManifestArtifact(report, name, path)
	case "analytics_report.json":
		validateAnalyticsReportArtifact(report, name, path)
	}
}

func validArtifactStatus(status string) bool {
	switch ArtifactStatus(strings.TrimSpace(status)) {
	case ArtifactStatusDraft, ArtifactStatusInReview, ArtifactStatusApproved,
		ArtifactStatusRejected, ArtifactStatusSuperseded, ArtifactStatusLocked:
		return true
	default:
		return false
	}
}

type topicForSchemaValidation struct {
	TitleWorking     string            `json:"title_working" yaml:"title_working"`
	Format           string            `json:"format" yaml:"format"`
	Audience         map[string]string `json:"audience" yaml:"audience"`
	Scores           map[string]int    `json:"scores" yaml:"scores"`
	OperatorDecision struct {
		Decision string `json:"decision" yaml:"decision"`
		Notes    string `json:"notes" yaml:"notes"`
	} `json:"operator_decision" yaml:"operator_decision"`
}

func validateTopicArtifact(report *ValidationReport, file string, path string) {
	var topic topicForSchemaValidation
	if !decodeForSchema(report, file, path, &topic) {
		return
	}
	requireString(report, file, "title_working", topic.TitleWorking)
	requireString(report, file, "format", topic.Format)
	if strings.TrimSpace(topic.Audience["primary"]) == "" {
		report.add(file, "audience.primary", "primary audience is required")
	}
	requiredScores := []string{
		"educational_value",
		"evergreen_value",
		"community_fit",
		"visual_potential",
		"production_cost",
		"factual_risk",
		"funnel_value",
	}
	for _, key := range requiredScores {
		value, ok := topic.Scores[key]
		if !ok {
			report.add(file, "scores."+key, "score is required")
			continue
		}
		if value < 1 || value > 10 {
			report.add(file, "scores."+key, "score must be between 1 and 10")
		}
	}
	requireString(report, file, "operator_decision.decision", topic.OperatorDecision.Decision)
}

func validateResearchPackArtifact(report *ValidationReport, file string, path string) {
	pack, err := LoadResearchPackFile(path)
	if err != nil {
		report.add(file, "", err.Error())
		return
	}
	requireString(report, file, "core_question", pack.CoreQuestion)
	requireNonEmpty(report, file, "learning_objectives", len(pack.LearningObjectives))
	validateSources(report, file, pack.Sources)
	for i, snippet := range pack.SourceSnippets {
		field := fmt.Sprintf("source_snippets[%d]", i)
		requireString(report, file, field+".source_id", snippet.SourceID)
		requireString(report, file, field+".text", snippet.Text)
		validateEvidenceLocator(report, file, field+".locator", snippet.Locator, []string{snippet.SourceID}, false)
	}
	for i, candidate := range pack.ClaimCandidates {
		field := fmt.Sprintf("claim_candidates[%d]", i)
		requireString(report, file, field+".text", candidate.Text)
		requireNonEmpty(report, file, field+".source_ids", len(candidate.SourceIDs))
	}
}

func validateSources(report *ValidationReport, file string, sources []Source) {
	requireNonEmpty(report, file, "sources", len(sources))
	seen := map[string]bool{}
	for i, source := range sources {
		field := fmt.Sprintf("sources[%d]", i)
		requireString(report, file, field+".source_id", source.ID)
		requireString(report, file, field+".title", source.Title)
		requireString(report, file, field+".uri", source.URI)
		requireString(report, file, field+".type", source.Type)
		requireString(report, file, field+".trust_level", source.TrustLevel)
		if source.ID != "" {
			if seen[source.ID] {
				report.add(file, field+".source_id", "source_id must be unique")
			}
			seen[source.ID] = true
		}
		if source.URI != "" && !isHTTPURI(source.URI) {
			report.add(file, field+".uri", "source uri must be absolute http(s)")
		}
		if source.Type != "" && !validSourceType(source.Type) {
			report.add(file, field+".type", "unsupported source type: "+source.Type)
		}
		if source.TrustLevel != "" && !validSourceTrust(source.TrustLevel) {
			report.add(file, field+".trust_level", "unsupported source trust level: "+source.TrustLevel)
		}
	}
}

func validateClaimsArtifact(report *ValidationReport, file string, path string) {
	validateClaimsCoverage(report, path)
	var claimSet claimsFileForValidation
	if !decodeForSchema(report, file, path, &claimSet) {
		return
	}
	for i, claim := range claimSet.Claims {
		field := claimField(i, claim)
		if strings.TrimSpace(claim.Type) == "" {
			report.add(file, field+".type", "claim type is required")
		}
		if !validClaimRisk(claim.RiskLevel) {
			report.add(file, field+".risk_level", "unsupported claim risk level: "+string(claim.RiskLevel))
		}
		if !validClaimStatus(claim.Status) {
			report.add(file, field+".verification_status", "unsupported claim verification status: "+string(claim.Status))
		}
		for j, locator := range claim.EvidenceLocators {
			validateEvidenceLocator(report, file, fmt.Sprintf("%s.evidence_locators[%d]", field, j), locator, claim.SourceIDs, true)
		}
	}
}

type verificationReportForSchemaValidation struct {
	Summary      string `json:"summary" yaml:"summary"`
	ClaimResults []struct {
		ClaimID string      `json:"claim_id" yaml:"claim_id"`
		Status  ClaimStatus `json:"status" yaml:"status"`
		Notes   string      `json:"notes" yaml:"notes"`
	} `json:"claim_results" yaml:"claim_results"`
	BlockingIssues []string `json:"blocking_issues" yaml:"blocking_issues"`
	Decision       string   `json:"decision" yaml:"decision"`
}

func validateVerificationReportArtifact(report *ValidationReport, file string, path string) {
	var verification verificationReportForSchemaValidation
	if !decodeForSchema(report, file, path, &verification) {
		return
	}
	requireString(report, file, "summary", verification.Summary)
	requireNonEmpty(report, file, "claim_results", len(verification.ClaimResults))
	validateDecision(report, file, "decision", verification.Decision, "approved", "request_revision", "block")
	for i, result := range verification.ClaimResults {
		field := fmt.Sprintf("claim_results[%d]", i)
		requireString(report, file, field+".claim_id", result.ClaimID)
		if !validClaimStatus(result.Status) {
			report.add(file, field+".status", "unsupported claim result status: "+string(result.Status))
		}
	}
	if verification.Decision != "approved" && len(verification.BlockingIssues) == 0 {
		report.add(file, "blocking_issues", "non-approved verification requires blocking issue details")
	}
}

type multimodelApprovalForSchemaValidation struct {
	ModelPanel []struct {
		ModelID    string  `json:"model_id" yaml:"model_id"`
		Provider   string  `json:"provider" yaml:"provider"`
		Task       string  `json:"task" yaml:"task"`
		Verdict    string  `json:"verdict" yaml:"verdict"`
		Confidence float64 `json:"confidence" yaml:"confidence"`
		Notes      string  `json:"notes" yaml:"notes"`
	} `json:"model_panel" yaml:"model_panel"`
	Consensus       string           `json:"consensus" yaml:"consensus"`
	Dissent         []map[string]any `json:"dissent" yaml:"dissent"`
	OperatorSummary string           `json:"operator_summary" yaml:"operator_summary"`
}

func validateMultimodelApprovalArtifact(report *ValidationReport, file string, path string) {
	var approval multimodelApprovalForSchemaValidation
	if !decodeForSchema(report, file, path, &approval) {
		return
	}
	if len(approval.ModelPanel) < 2 {
		report.add(file, "model_panel", "at least two model reviews are required")
	}
	validateDecision(report, file, "consensus", approval.Consensus, "approved", "approved_with_suggestions", "revision_required", "blocked")
	requireString(report, file, "operator_summary", approval.OperatorSummary)
	for i, review := range approval.ModelPanel {
		field := fmt.Sprintf("model_panel[%d]", i)
		requireString(report, file, field+".model_id", review.ModelID)
		requireString(report, file, field+".provider", review.Provider)
		requireString(report, file, field+".task", review.Task)
		validateDecision(report, file, field+".verdict", review.Verdict, "approve", "approve_with_suggestions", "request_revision", "block")
		if review.Confidence < 0 || review.Confidence > 1 {
			report.add(file, field+".confidence", "confidence must be between 0 and 1")
		}
	}
}

func validateHumanQAArtifact(report *ValidationReport, file string, path string) {
	validateHumanQAExplicit(report, path)
	var qa struct {
		Reviewer        string   `json:"reviewer" yaml:"reviewer"`
		Decision        string   `json:"decision" yaml:"decision"`
		Notes           string   `json:"notes" yaml:"notes"`
		RequiredChanges []string `json:"required_changes" yaml:"required_changes"`
	}
	if !decodeForSchema(report, file, path, &qa) {
		return
	}
	requireString(report, file, "reviewer", qa.Reviewer)
	validateDecision(report, file, "decision", qa.Decision, string(HumanDecisionApprove), string(HumanDecisionApproveWithMinorEdits), string(HumanDecisionRequestRevision), string(HumanDecisionBlock))
	if (qa.Decision == string(HumanDecisionRequestRevision) || qa.Decision == string(HumanDecisionBlock)) && len(qa.RequiredChanges) == 0 {
		report.add(file, "required_changes", "revision/block decisions require required_changes")
	}
}

type storyboardForSchemaValidation struct {
	Scenes []struct {
		SceneID    string `json:"scene_id" yaml:"scene_id"`
		TimeTarget string `json:"time_target" yaml:"time_target"`
		Narration  string `json:"narration" yaml:"narration"`
		Mascot     struct {
			Mode    string `json:"mode" yaml:"mode"`
			Emotion string `json:"emotion" yaml:"emotion"`
			Action  string `json:"action" yaml:"action"`
		} `json:"mascot" yaml:"mascot"`
		Visual struct {
			Type    string `json:"type" yaml:"type"`
			Content string `json:"content" yaml:"content"`
		} `json:"visual" yaml:"visual"`
		OnScreenText string `json:"on_screen_text" yaml:"on_screen_text"`
	} `json:"scenes" yaml:"scenes"`
}

func validateStoryboardArtifact(report *ValidationReport, file string, path string) {
	var storyboard storyboardForSchemaValidation
	if !decodeForSchema(report, file, path, &storyboard) {
		return
	}
	requireNonEmpty(report, file, "scenes", len(storyboard.Scenes))
	seen := map[string]bool{}
	for i, scene := range storyboard.Scenes {
		field := fmt.Sprintf("scenes[%d]", i)
		requireString(report, file, field+".scene_id", scene.SceneID)
		requireString(report, file, field+".time_target", scene.TimeTarget)
		requireString(report, file, field+".narration", scene.Narration)
		requireString(report, file, field+".mascot.mode", scene.Mascot.Mode)
		requireString(report, file, field+".mascot.emotion", scene.Mascot.Emotion)
		requireString(report, file, field+".mascot.action", scene.Mascot.Action)
		requireString(report, file, field+".visual.type", scene.Visual.Type)
		requireString(report, file, field+".visual.content", scene.Visual.Content)
		requireString(report, file, field+".on_screen_text", scene.OnScreenText)
		if scene.SceneID != "" {
			if seen[scene.SceneID] {
				report.add(file, field+".scene_id", "scene_id must be unique")
			}
			seen[scene.SceneID] = true
		}
	}
}

func validateAssetManifestArtifact(report *ValidationReport, file string, path string) {
	var manifest struct {
		Assets []struct {
			AssetID     string   `json:"asset_id" yaml:"asset_id"`
			Type        string   `json:"type" yaml:"type"`
			Path        string   `json:"path" yaml:"path"`
			GeneratedBy string   `json:"generated_by" yaml:"generated_by"`
			License     string   `json:"license" yaml:"license"`
			Hash        string   `json:"hash" yaml:"hash"`
			Provenance  []string `json:"provenance" yaml:"provenance"`
		} `json:"assets" yaml:"assets"`
	}
	if !decodeForSchema(report, file, path, &manifest) {
		return
	}
	requireNonEmpty(report, file, "assets", len(manifest.Assets))
	seen := map[string]bool{}
	for i, asset := range manifest.Assets {
		field := fmt.Sprintf("assets[%d]", i)
		requireString(report, file, field+".asset_id", asset.AssetID)
		requireString(report, file, field+".type", asset.Type)
		requireString(report, file, field+".path", asset.Path)
		requireString(report, file, field+".generated_by", asset.GeneratedBy)
		requireString(report, file, field+".license", asset.License)
		requireString(report, file, field+".hash", asset.Hash)
		if asset.AssetID != "" {
			if seen[asset.AssetID] {
				report.add(file, field+".asset_id", "asset_id must be unique")
			}
			seen[asset.AssetID] = true
		}
	}
}

func validateRenderManifestArtifact(report *ValidationReport, file string, path string) {
	var manifest struct {
		Renderer        string   `json:"renderer" yaml:"renderer"`
		RendererVersion string   `json:"renderer_version" yaml:"renderer_version"`
		Inputs          []string `json:"inputs" yaml:"inputs"`
		Outputs         []struct {
			Type            string `json:"type" yaml:"type"`
			Path            string `json:"path" yaml:"path"`
			DurationSeconds int    `json:"duration_seconds" yaml:"duration_seconds"`
			Resolution      string `json:"resolution" yaml:"resolution"`
			Hash            string `json:"hash" yaml:"hash"`
		} `json:"outputs" yaml:"outputs"`
	}
	if !decodeForSchema(report, file, path, &manifest) {
		return
	}
	requireString(report, file, "renderer", manifest.Renderer)
	requireString(report, file, "renderer_version", manifest.RendererVersion)
	requireNonEmpty(report, file, "inputs", len(manifest.Inputs))
	requireNonEmpty(report, file, "outputs", len(manifest.Outputs))
	for i, output := range manifest.Outputs {
		field := fmt.Sprintf("outputs[%d]", i)
		requireString(report, file, field+".type", output.Type)
		requireString(report, file, field+".path", output.Path)
		requireString(report, file, field+".resolution", output.Resolution)
		requireString(report, file, field+".hash", output.Hash)
		if output.DurationSeconds <= 0 {
			report.add(file, field+".duration_seconds", "duration_seconds must be positive")
		}
	}
}

func validateProductionQAArtifact(report *ValidationReport, file string, path string) {
	var qa struct {
		Checks         map[string]string `json:"checks" yaml:"checks"`
		BlockingIssues []string          `json:"blocking_issues" yaml:"blocking_issues"`
		Decision       string            `json:"decision" yaml:"decision"`
	}
	if !decodeForSchema(report, file, path, &qa) {
		return
	}
	if len(qa.Checks) == 0 {
		report.add(file, "checks", "production QA checks must not be empty")
	}
	requiredChecks := []string{"claims", "asset_provenance", "policy"}
	for _, check := range requiredChecks {
		if strings.TrimSpace(qa.Checks[check]) == "" {
			report.add(file, "checks."+check, "production QA check is required")
		}
	}
	validateDecision(report, file, "decision", qa.Decision, "approved", "request_revision", "block")
	if qa.Decision == "approved" && len(qa.BlockingIssues) > 0 {
		report.add(file, "blocking_issues", "approved production QA must not include blocking issues")
	}
	if qa.Decision != "approved" && len(qa.BlockingIssues) == 0 {
		report.add(file, "blocking_issues", "non-approved production QA requires blocking issues")
	}
}

func validatePublishManifestArtifact(report *ValidationReport, file string, path string) {
	validatePublishManifestSafety(report, path)
	var manifest struct {
		Platform             string            `json:"platform" yaml:"platform"`
		Visibility           PublishVisibility `json:"visibility" yaml:"visibility"`
		Title                string            `json:"title" yaml:"title"`
		DescriptionPath      string            `json:"description_path" yaml:"description_path"`
		ThumbnailPath        string            `json:"thumbnail_path" yaml:"thumbnail_path"`
		ScheduledAt          *string           `json:"scheduled_at" yaml:"scheduled_at"`
		HumanReleaseApproval bool              `json:"human_release_approval" yaml:"human_release_approval"`
	}
	if !decodeForSchema(report, file, path, &manifest) {
		return
	}
	requireString(report, file, "platform", manifest.Platform)
	validateDecision(report, file, "visibility", string(manifest.Visibility), string(PublishVisibilityPrivate), string(PublishVisibilityScheduled), string(PublishVisibilityPublic))
	requireString(report, file, "title", manifest.Title)
	requireString(report, file, "description_path", manifest.DescriptionPath)
	requireString(report, file, "thumbnail_path", manifest.ThumbnailPath)
	if manifest.Visibility == PublishVisibilityScheduled && (manifest.ScheduledAt == nil || strings.TrimSpace(*manifest.ScheduledAt) == "") {
		report.add(file, "scheduled_at", "scheduled visibility requires scheduled_at")
	}
	if manifest.Visibility == PublishVisibilityScheduled && !manifest.HumanReleaseApproval {
		report.add(file, "human_release_approval", "scheduled visibility requires human release approval")
	}
}

func validateAnalyticsReportArtifact(report *ValidationReport, file string, path string) {
	var analytics struct {
		Window  string `json:"window" yaml:"window"`
		Metrics struct {
			CTR                        float64 `json:"ctr" yaml:"ctr"`
			AverageViewDurationSeconds float64 `json:"average_view_duration_seconds" yaml:"average_view_duration_seconds"`
			First30sRetention          float64 `json:"first_30s_retention" yaml:"first_30s_retention"`
			SubscriberDelta            int     `json:"subscriber_delta" yaml:"subscriber_delta"`
			CommunityClicks            int     `json:"community_clicks" yaml:"community_clicks"`
		} `json:"metrics" yaml:"metrics"`
		Insights           []string `json:"insights" yaml:"insights"`
		RecommendedActions []string `json:"recommended_actions" yaml:"recommended_actions"`
	}
	if !decodeForSchema(report, file, path, &analytics) {
		return
	}
	validateDecision(report, file, "window", analytics.Window, "dry_run", "24h", "72h", "7d")
	if analytics.Metrics.CTR < 0 || analytics.Metrics.CTR > 1 {
		report.add(file, "metrics.ctr", "ctr must be between 0 and 1")
	}
	if analytics.Metrics.First30sRetention < 0 || analytics.Metrics.First30sRetention > 1 {
		report.add(file, "metrics.first_30s_retention", "first_30s_retention must be between 0 and 1")
	}
	if analytics.Metrics.AverageViewDurationSeconds < 0 {
		report.add(file, "metrics.average_view_duration_seconds", "average_view_duration_seconds must not be negative")
	}
	if analytics.Metrics.CommunityClicks < 0 {
		report.add(file, "metrics.community_clicks", "community_clicks must not be negative")
	}
}

func decodeForSchema(report *ValidationReport, file string, path string, out any) bool {
	if err := decodeArtifact(path, out); err != nil {
		report.add(file, "", fmt.Sprintf("cannot decode %s: %v", filepath.Base(path), err))
		return false
	}
	return true
}

func requireString(report *ValidationReport, file string, field string, value string) {
	if strings.TrimSpace(value) == "" {
		report.add(file, field, field+" is required")
	}
}

func requireNonEmpty(report *ValidationReport, file string, field string, count int) {
	if count == 0 {
		report.add(file, field, field+" must not be empty")
	}
}

func validateDecision(report *ValidationReport, file string, field string, value string, allowed ...string) {
	value = strings.TrimSpace(value)
	if value == "" {
		report.add(file, field, field+" is required")
		return
	}
	for _, item := range allowed {
		if value == item {
			return
		}
	}
	report.add(file, field, "unsupported value: "+value)
}

func validateEvidenceLocator(report *ValidationReport, file string, field string, locator EvidenceLocator, sourceIDs []string, requireKnownSource bool) {
	requireString(report, file, field+".source_id", locator.SourceID)
	if strings.TrimSpace(locator.Section) == "" && strings.TrimSpace(locator.Range) == "" && strings.TrimSpace(locator.QuoteHash) == "" {
		report.add(file, field, "evidence locator requires section, range, or quote_hash")
	}
	if requireKnownSource && locator.SourceID != "" && !containsValue(sourceIDs, locator.SourceID) {
		report.add(file, field+".source_id", "locator source_id must match claim source_ids")
	}
}

func validClaimRisk(risk ClaimRisk) bool {
	switch risk {
	case ClaimRiskLow, ClaimRiskMedium, ClaimRiskHigh, ClaimRiskCritical:
		return true
	default:
		return false
	}
}

func validClaimStatus(status ClaimStatus) bool {
	switch status {
	case ClaimStatusSupported, ClaimStatusPartiallySupported, ClaimStatusUnsupported, ClaimStatusContradicted, ClaimStatusNeedsHumanReview, ClaimStatusRemoved:
		return true
	default:
		return false
	}
}

func validSourceType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "official_docs", "specification", "source_code", "release_notes", "maintainer_statement", "engineering_blog", "community_discussion":
		return true
	default:
		return false
	}
}

func validSourceTrust(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "primary", "secondary", "community":
		return true
	default:
		return false
	}
}

func isHTTPURI(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func claimField(index int, claim Claim) string {
	if strings.TrimSpace(claim.ID) != "" {
		return claim.ID
	}
	return fmt.Sprintf("claims[%d]", index)
}

func containsValue(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
