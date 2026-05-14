package activities

import (
	"context"

	"github.com/AnimusHQ/news/internal/artifacts"
	claimextractor "github.com/AnimusHQ/news/internal/claims"
	"github.com/AnimusHQ/news/internal/productionqa"
	"github.com/AnimusHQ/news/internal/qa"
	"github.com/AnimusHQ/news/internal/render"
	"github.com/AnimusHQ/news/internal/research"
	"github.com/AnimusHQ/news/internal/storyboard"
)

// ValidateEpisodeActivity validates the episode artifact bundle.
func ValidateEpisodeActivity(ctx context.Context, episodeDir string) (string, error) {
	if err := artifacts.ValidateEpisodeDirectory(episodeDir); err != nil {
		return "", err
	}
	return "artifact validation passed", nil
}

// ValidateTransitionActivity checks lifecycle-specific artifact dependencies.
func ValidateTransitionActivity(ctx context.Context, episodeDir string, state artifacts.LifecycleState) (artifacts.DependencyReport, error) {
	report := artifacts.ValidateTransition(episodeDir, state)
	if !report.Valid {
		return report, artifacts.ValidateReport(artifacts.ValidationReport{
			EpisodeDir: episodeDir,
			Valid:      false,
			Issues:     dependencyIssuesToValidation(report.Issues),
		})
	}
	return report, nil
}

// BuildResearchPackActivity builds a deterministic draft research pack from
// explicitly supplied source records and snippets.
func BuildResearchPackActivity(ctx context.Context, input research.BuilderInput) (research.BuildResult, error) {
	return research.BuildPack(input)
}

func dependencyIssuesToValidation(issues []artifacts.DependencyIssue) []artifacts.ValidationIssue {
	out := make([]artifacts.ValidationIssue, 0, len(issues))
	for _, issue := range issues {
		out = append(out, artifacts.ValidationIssue{File: issue.Artifact, Message: issue.Reason})
	}
	return out
}

// MockCouncilActivity is a safe placeholder for the future multimodel council.
func MockCouncilActivity(ctx context.Context, episodeID string) (string, error) {
	return "mock multimodel council approved with no external model calls", nil
}

// ExtractClaimsActivity extracts canonical claim candidates from script.md and
// research_pack.json. File I/O is intentionally activity-side, not workflow-side.
func ExtractClaimsActivity(ctx context.Context, episodeDir string) (artifacts.ClaimsFile, error) {
	result, err := claimextractor.ExtractEpisode(episodeDir)
	if err != nil {
		return artifacts.ClaimsFile{}, err
	}
	return result.ClaimsFile, nil
}

// GenerateHumanQAPacketActivity compiles a deterministic operator packet from
// upstream artifacts. It returns a recommendation, not human approval.
func GenerateHumanQAPacketActivity(ctx context.Context, input qa.Input) (qa.Packet, error) {
	return qa.Generate(input)
}

// GenerateStoryboardActivity produces a deterministic storyboard artifact from
// an approved script. It performs no rendering or provider calls.
func GenerateStoryboardActivity(ctx context.Context, input storyboard.Input) (storyboard.File, error) {
	return storyboard.Generate(input)
}

// GenerateRenderPreviewActivity turns a storyboard into a deterministic local
// preview and render manifest. It does not produce final video binaries.
func GenerateRenderPreviewActivity(ctx context.Context, input render.Input) (render.Result, error) {
	return render.GeneratePreview(input)
}

// RunProductionQAActivity evaluates generated render artifacts before any
// publishable release path. It is deterministic and offline.
func RunProductionQAActivity(ctx context.Context, input productionqa.Input) (productionqa.Report, error) {
	return productionqa.Run(input)
}

// ProductionQAActivity is a safe placeholder for future production QA checks.
func ProductionQAActivity(ctx context.Context, episodeID string) (string, error) {
	return "production QA placeholder passed", nil
}

// DryRunPublishActivity is intentionally non-public and no-network.
func DryRunPublishActivity(ctx context.Context, episodeID string) (string, error) {
	return "dry-run publish completed without upload", nil
}
