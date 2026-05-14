package analytics

import (
	"fmt"
	"strings"
)

const (
	lowCTRThreshold          = 0.035
	lowFirst30Threshold      = 0.45
	lowCompletionThreshold   = 0.35
	lowConversionThreshold   = 0.01
	highCostPerViewThreshold = 1.50
)

// GenerateInsightReport creates advisory recommendations from normalized
// analytics. It never mutates editorial metadata or approval gates.
func GenerateInsightReport(input Input) (Report, error) {
	if strings.TrimSpace(input.EpisodeID) == "" {
		return Report{}, fmt.Errorf("episode id is required")
	}
	if !validWindow(input.Window) {
		return Report{}, fmt.Errorf("unsupported analytics window: %s", input.Window)
	}

	report := ReportFromInput(input)
	report.Insights = nil
	report.RecommendedActions = nil

	addRetentionInsights(&report, input.Metrics)
	addCTRInsights(&report, input.Metrics)
	addCommunityInsights(&report, input.Metrics)
	addCostInsights(&report, input.Metrics)
	addCorrectionInsights(&report, input.Feedback)

	if len(report.Insights) == 0 {
		report.Insights = append(report.Insights, "No strong analytics signal yet; keep monitoring before changing future episodes.")
	}
	if len(report.RecommendedActions) == 0 {
		report.RecommendedActions = append(report.RecommendedActions, "advisory: keep current editorial plan; do not auto-change topics or metadata.")
	}
	report.RecommendedActions = append(report.RecommendedActions, "advisory: analytics cannot override source, QA, release, or correction gates.")
	return report, nil
}

func addRetentionInsights(report *Report, metrics Metrics) {
	if metrics.First30Retention > 0 && metrics.First30Retention < lowFirst30Threshold {
		report.Insights = append(report.Insights, "First 30 seconds retention is below target.")
		report.RecommendedActions = append(report.RecommendedActions, "advisory: review hook clarity, pacing, and first-scene visual payoff.")
	}
	if metrics.CompletionRate > 0 && metrics.CompletionRate < lowCompletionThreshold {
		report.Insights = append(report.Insights, "Completion rate is below target.")
		report.RecommendedActions = append(report.RecommendedActions, "advisory: review episode structure and trim slow transitions in future scripts.")
	}
}

func addCTRInsights(report *Report, metrics Metrics) {
	if metrics.Impressions < 100 || metrics.CTR == 0 || metrics.CTR >= lowCTRThreshold {
		return
	}
	report.Insights = append(report.Insights, "CTR is below target for the available impression sample.")
	report.RecommendedActions = append(report.RecommendedActions, "advisory: review title and thumbnail specificity without misleading clickbait.")
	report.RecommendedActions = append(report.RecommendedActions, "policy: misleading clickbait is prohibited even when CTR is low.")
}

func addCommunityInsights(report *Report, metrics Metrics) {
	if metrics.Views < 100 {
		return
	}
	conversion := float64(metrics.CommunityClicks) / float64(metrics.Views)
	if conversion < lowConversionThreshold {
		report.Insights = append(report.Insights, "Community conversion is below target.")
		report.RecommendedActions = append(report.RecommendedActions, "advisory: review CTA placement and make community value clearer.")
	}
}

func addCostInsights(report *Report, metrics Metrics) {
	if metrics.CostPerEpisode <= 0 || metrics.Views <= 0 {
		return
	}
	costPerView := metrics.CostPerEpisode / float64(metrics.Views)
	if costPerView > highCostPerViewThreshold {
		report.Insights = append(report.Insights, fmt.Sprintf("Cost per view is high for the %s window.", report.Window))
		report.RecommendedActions = append(report.RecommendedActions, "advisory: review production scope and reuse deterministic assets where appropriate.")
	}
}

func addCorrectionInsights(report *Report, feedback []FeedbackSignal) {
	for _, signal := range feedback {
		kind := strings.ToLower(strings.TrimSpace(signal.Type))
		if kind != "factual_correction" && kind != "source_dispute" {
			continue
		}
		report.Insights = append(report.Insights, "Viewer feedback includes a factual correction or source dispute signal.")
		report.RecommendedActions = append(report.RecommendedActions, "advisory: open a correction review before reusing this claim or metadata.")
		return
	}
}
