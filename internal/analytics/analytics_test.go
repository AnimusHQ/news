package analytics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestFixtureAnalyticsImportWorksOffline(t *testing.T) {
	adapter := FixtureAdapter{Records: map[string]ProviderRecord{
		"episode-1|72h": fixtureRecord("episode-1", Window72h),
	}}
	report, err := ImportReport(context.Background(), adapter, ImportRequest{EpisodeID: "episode-1", Window: Window72h})
	if err != nil {
		t.Fatalf("import report failed: %v", err)
	}
	if report.Metrics.Views != 1200 {
		t.Fatalf("expected normalized views, got %d", report.Metrics.Views)
	}
	if !report.AdvisoryOnly {
		t.Fatal("analytics report must be advisory only")
	}
}

func TestAnalyticsReportValidates(t *testing.T) {
	input, err := Normalize(fixtureRecord("episode-1", Window72h))
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	report, err := GenerateInsightReport(input)
	if err != nil {
		t.Fatalf("generate insight report failed: %v", err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal analytics report: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "analytics_report.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write analytics report: %v", err)
	}
	validation := artifacts.ValidatePath(path)
	if !validation.Valid {
		t.Fatalf("expected analytics report to validate: %+v", validation.Issues)
	}
}

func TestNormalizeHandlesMissingMetricFieldsExplicitly(t *testing.T) {
	record := fixtureRecord("episode-1", Window24h)
	record.Metrics.CTR = nil
	record.Metrics.Views = nil
	input, err := Normalize(record)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if len(input.MissingMetricKeys) == 0 {
		t.Fatal("expected missing metric keys")
	}
	if !contains(input.MissingMetricKeys, "ctr") || !contains(input.MissingMetricKeys, "views") {
		t.Fatalf("expected missing ctr and views, got %+v", input.MissingMetricKeys)
	}
	if len(input.DataQualityNotes) == 0 {
		t.Fatal("expected data quality notes for missing metrics")
	}
}

func TestNormalizeRejectsMalformedProviderData(t *testing.T) {
	_, err := Normalize(ProviderRecord{Window: Window72h, Metrics: fixtureMetrics()})
	if err == nil {
		t.Fatal("expected missing episode id to fail")
	}
	_, err = Normalize(ProviderRecord{EpisodeID: "episode-1", Window: "13d", Metrics: fixtureMetrics()})
	if err == nil {
		t.Fatal("expected unsupported window to fail")
	}
}

func TestLowRetentionProducesPacingRecommendation(t *testing.T) {
	input := normalizedFixture(t)
	first30 := 0.31
	input.Metrics.First30Retention = first30
	report, err := GenerateInsightReport(input)
	if err != nil {
		t.Fatalf("generate insight report failed: %v", err)
	}
	if !containsText(report.RecommendedActions, "review hook clarity") {
		t.Fatalf("expected pacing recommendation, got %+v", report.RecommendedActions)
	}
}

func TestLowCTRRecommendationAvoidsClickbait(t *testing.T) {
	input := normalizedFixture(t)
	input.Metrics.Impressions = 2000
	input.Metrics.CTR = 0.02
	report, err := GenerateInsightReport(input)
	if err != nil {
		t.Fatalf("generate insight report failed: %v", err)
	}
	if !containsText(report.RecommendedActions, "without misleading clickbait") {
		t.Fatalf("expected non-clickbait CTR recommendation, got %+v", report.RecommendedActions)
	}
	if !containsText(report.RecommendedActions, "misleading clickbait is prohibited") {
		t.Fatalf("expected explicit clickbait prohibition, got %+v", report.RecommendedActions)
	}
}

func TestFactualCorrectionSignalProducesCorrectionRecommendation(t *testing.T) {
	input := normalizedFixture(t)
	input.Feedback = []FeedbackSignal{{Type: "factual_correction", Text: "Source range may be wrong.", Count: 2}}
	report, err := GenerateInsightReport(input)
	if err != nil {
		t.Fatalf("generate insight report failed: %v", err)
	}
	if !containsText(report.RecommendedActions, "open a correction review") {
		t.Fatalf("expected correction recommendation, got %+v", report.RecommendedActions)
	}
}

func normalizedFixture(t *testing.T) Input {
	t.Helper()
	input, err := Normalize(fixtureRecord("episode-1", Window72h))
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	return input
}

func fixtureRecord(episodeID string, window string) ProviderRecord {
	return ProviderRecord{
		Provider:  "fixture-provider",
		EpisodeID: episodeID,
		Window:    window,
		Metrics:   fixtureMetrics(),
	}
}

func fixtureMetrics() ProviderMetrics {
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
	return ProviderMetrics{
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
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsText(values []string, target string) bool {
	for _, value := range values {
		if strings.Contains(value, target) {
			return true
		}
	}
	return false
}
