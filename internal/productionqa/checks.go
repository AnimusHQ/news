package productionqa

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AnimusHQ/news/internal/artifacts"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/verification"
)

const (
	SchemaVersion = "1.0"
	fileStatus    = "draft"

	DecisionApproved        = "approved"
	DecisionRequestRevision = "request_revision"
	DecisionBlock           = "block"
)

// Input contains all deterministic data needed for pre-publication QA.
type Input struct {
	EpisodeID                   string
	Render                      render.Result
	Claims                      []artifacts.Claim
	Verification                verification.Report
	HumanQARecommendation       artifacts.HumanDecision
	PublishVisibility           artifacts.PublishVisibility
	SyntheticDisclosureRequired bool
	SyntheticDisclosureStatus   string
	AvailableOutputs            map[string]string
}

// Report is the canonical production_qa_report.json shape.
type Report struct {
	SchemaVersion  string            `json:"schema_version"`
	EpisodeID      string            `json:"episode_id"`
	ArtifactID     string            `json:"artifact_id"`
	Status         string            `json:"status"`
	Checks         map[string]string `json:"checks"`
	BlockingIssues []string          `json:"blocking_issues,omitempty"`
	Decision       string            `json:"decision"`
}

// Run executes deterministic production QA checks without provider calls or
// platform APIs.
func Run(input Input) (Report, error) {
	episodeID := strings.TrimSpace(input.EpisodeID)
	if episodeID == "" {
		episodeID = strings.TrimSpace(input.Render.RenderManifest.EpisodeID)
	}
	if episodeID == "" {
		return Report{}, fmt.Errorf("episode id is required")
	}

	report := Report{
		SchemaVersion: SchemaVersion,
		EpisodeID:     episodeID,
		ArtifactID:    "production-qa-" + episodeID + "-v1",
		Status:        fileStatus,
		Checks:        map[string]string{},
		Decision:      DecisionApproved,
	}

	var blockers []string
	blockers = append(blockers, checkRenderManifest(input.Render.RenderManifest, episodeID, report.Checks)...)
	blockers = append(blockers, checkRenderOutputs(input.Render, input.AvailableOutputs, report.Checks)...)
	blockers = append(blockers, checkAssetManifest(input.Render.AssetManifest, episodeID, report.Checks)...)
	blockers = append(blockers, checkPublishIntent(input.PublishVisibility, report.Checks)...)
	blockers = append(blockers, checkSyntheticDisclosure(input, report.Checks)...)
	blockers = append(blockers, checkVerification(input.Claims, input.Verification, report.Checks)...)
	blockers = append(blockers, checkHumanQA(input.HumanQARecommendation, report.Checks)...)

	report.BlockingIssues = sortedUnique(blockers)
	if len(report.BlockingIssues) > 0 {
		report.Decision = DecisionRequestRevision
		if hasHardBlocker(report.BlockingIssues) {
			report.Decision = DecisionBlock
		}
	}
	return report, nil
}

func checkRenderManifest(manifest render.RenderManifest, episodeID string, checks map[string]string) []string {
	var blockers []string
	if manifest.SchemaVersion == "" || manifest.EpisodeID == "" || manifest.ArtifactID == "" || manifest.Status == "" {
		blockers = append(blockers, "render manifest is missing required metadata")
	}
	if manifest.EpisodeID != "" && manifest.EpisodeID != episodeID {
		blockers = append(blockers, "render manifest episode id does not match QA episode id")
	}
	if manifest.Renderer == "" || manifest.RendererVersion == "" {
		blockers = append(blockers, "render manifest is missing renderer metadata")
	}
	if len(manifest.Inputs) == 0 {
		blockers = append(blockers, "render manifest must list input artifacts")
	}
	if len(manifest.Outputs) == 0 {
		blockers = append(blockers, "render manifest must list required outputs")
	}
	setCheck(checks, "render_manifest", blockers)
	return blockers
}

func checkRenderOutputs(result render.Result, available map[string]string, checks map[string]string) []string {
	var blockers []string
	captionRequired := false
	captionPresent := false
	for _, output := range result.RenderManifest.Outputs {
		if output.Path == "" {
			blockers = append(blockers, "render output path is required")
			continue
		}
		if output.Hash == "" {
			blockers = append(blockers, output.Path+": render output hash is required")
		}
		if output.Path == result.Preview.Path {
			if result.Preview.Content == "" {
				blockers = append(blockers, output.Path+": preview output content is missing")
			}
			if result.Preview.Hash != "" && output.Hash != "" && result.Preview.Hash != output.Hash {
				blockers = append(blockers, output.Path+": render manifest hash does not match preview hash")
			}
		} else if hash, ok := available[output.Path]; !ok {
			blockers = append(blockers, output.Path+": required render output is missing")
		} else if hash != "" && output.Hash != "" && hash != output.Hash {
			blockers = append(blockers, output.Path+": available output hash does not match render manifest")
		}
		if outputRequiresCaption(output) {
			captionRequired = true
			if output.Path == result.Preview.Path || available[output.Path] != "" {
				captionPresent = true
			}
		}
	}
	setCheck(checks, "render_outputs", blockers)
	if captionRequired && !captionPresent {
		blockers = append(blockers, "caption or subtitle output is required but missing")
		checks["captions"] = "fail"
	} else if captionRequired {
		checks["captions"] = "pass"
	} else {
		checks["captions"] = "not_required"
	}
	return blockers
}

func checkAssetManifest(manifest render.AssetManifest, episodeID string, checks map[string]string) []string {
	var blockers []string
	if manifest.SchemaVersion == "" || manifest.EpisodeID == "" || manifest.ArtifactID == "" || manifest.Status == "" {
		blockers = append(blockers, "asset manifest is missing required metadata")
	}
	if manifest.EpisodeID != "" && manifest.EpisodeID != episodeID {
		blockers = append(blockers, "asset manifest episode id does not match QA episode id")
	}
	if len(manifest.Assets) == 0 {
		blockers = append(blockers, "asset manifest must contain assets")
	}
	for _, asset := range manifest.Assets {
		label := asset.AssetID
		if label == "" {
			label = asset.Path
		}
		if asset.AssetID == "" || asset.Type == "" || asset.Path == "" || asset.GeneratedBy == "" || asset.Hash == "" {
			blockers = append(blockers, label+": asset metadata is incomplete")
		}
		if len(asset.Provenance) == 0 {
			blockers = append(blockers, label+": asset provenance is required")
		}
		if strings.TrimSpace(asset.License) == "" || strings.EqualFold(asset.License, "unknown") {
			blockers = append(blockers, label+": asset license status is required")
		}
	}
	setCheck(checks, "asset_manifest", blockers)
	if hasAssetProvenanceBlocker(blockers) {
		checks["asset_provenance"] = "fail"
	} else {
		checks["asset_provenance"] = "pass"
	}
	return blockers
}

func checkPublishIntent(visibility artifacts.PublishVisibility, checks map[string]string) []string {
	if visibility == "" {
		visibility = artifacts.PublishVisibilityPrivate
	}
	if visibility == artifacts.PublishVisibilityPublic {
		checks["policy"] = "fail"
		return []string{"direct public publish intent is forbidden before explicit release approval"}
	}
	checks["policy"] = "pass"
	return nil
}

func checkSyntheticDisclosure(input Input, checks map[string]string) []string {
	if !input.SyntheticDisclosureRequired {
		checks["synthetic_disclosure"] = "not_required"
		return nil
	}
	status := strings.ToLower(strings.TrimSpace(input.SyntheticDisclosureStatus))
	if status == "present" || status == "declared" {
		checks["synthetic_disclosure"] = "pass"
		return nil
	}
	checks["synthetic_disclosure"] = "fail"
	return []string{"synthetic disclosure status is required"}
}

func checkVerification(claims []artifacts.Claim, report verification.Report, checks map[string]string) []string {
	claimByID := map[string]artifacts.Claim{}
	for _, claim := range claims {
		claimByID[claim.ID] = claim
	}

	var blockers []string
	for _, result := range report.ClaimResults {
		claim := claimByID[result.ClaimID]
		if claim.RiskLevel != artifacts.ClaimRiskHigh && claim.RiskLevel != artifacts.ClaimRiskCritical {
			continue
		}
		if result.Status == artifacts.ClaimStatusUnsupported ||
			result.Status == artifacts.ClaimStatusContradicted ||
			result.Status == artifacts.ClaimStatusNeedsHumanReview ||
			result.Status == artifacts.ClaimStatusPartiallySupported {
			blockers = append(blockers, result.ClaimID+": unresolved high-risk claim blocks production QA")
		}
	}
	if report.Decision != "" && report.Decision != "approved" {
		blockers = append(blockers, "verification decision is "+report.Decision)
	}
	setCheck(checks, "claims", blockers)
	return blockers
}

func checkHumanQA(decision artifacts.HumanDecision, checks map[string]string) []string {
	if decision == artifacts.HumanDecisionApprove || decision == artifacts.HumanDecisionApproveWithMinorEdits {
		checks["human_qa"] = "pass"
		return nil
	}
	checks["human_qa"] = "fail"
	if decision == "" {
		return []string{"human QA approval is required before production QA approval"}
	}
	return []string{"human QA decision is " + string(decision)}
}

func setCheck(checks map[string]string, key string, blockers []string) {
	if len(blockers) > 0 {
		checks[key] = "fail"
		return
	}
	checks[key] = "pass"
}

func outputRequiresCaption(output render.RenderOutput) bool {
	text := strings.ToLower(output.Type + " " + output.Path)
	return strings.Contains(text, "caption") || strings.Contains(text, "subtitle") || strings.Contains(text, ".vtt") || strings.Contains(text, ".srt")
}

func hasAssetProvenanceBlocker(blockers []string) bool {
	for _, blocker := range blockers {
		if strings.Contains(blocker, "asset provenance") || strings.Contains(blocker, "asset license") {
			return true
		}
	}
	return false
}

func hasHardBlocker(blockers []string) bool {
	for _, blocker := range blockers {
		if strings.Contains(blocker, "direct public publish") || strings.Contains(blocker, "human QA") {
			return true
		}
	}
	return false
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	compact := out[:0]
	var previous string
	for _, value := range out {
		if value == previous {
			continue
		}
		compact = append(compact, value)
		previous = value
	}
	return compact
}
