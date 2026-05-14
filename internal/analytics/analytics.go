package analytics

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const (
	SchemaVersion = "1.0"
	fileStatus    = "draft"

	Window24h = "24h"
	Window72h = "72h"
	Window7d  = "7d"
)

// ImportRequest identifies the analytics slice to import.
type ImportRequest struct {
	EpisodeID string
	Window    string
}

// Adapter is the provider-agnostic analytics import interface.
type Adapter interface {
	Import(ctx context.Context, request ImportRequest) (ProviderRecord, error)
}

// FixtureAdapter returns deterministic provider records for tests and dry-runs.
type FixtureAdapter struct {
	Records map[string]ProviderRecord
}

func (a FixtureAdapter) Import(ctx context.Context, request ImportRequest) (ProviderRecord, error) {
	if err := ctx.Err(); err != nil {
		return ProviderRecord{}, err
	}
	if strings.TrimSpace(request.EpisodeID) == "" {
		return ProviderRecord{}, fmt.Errorf("episode id is required")
	}
	if !validWindow(request.Window) {
		return ProviderRecord{}, fmt.Errorf("unsupported analytics window: %s", request.Window)
	}
	key := request.EpisodeID + "|" + request.Window
	record, ok := a.Records[key]
	if !ok {
		return ProviderRecord{}, fmt.Errorf("analytics fixture not found for %s", key)
	}
	return record, nil
}

// ProviderRecord is provider-specific data after it has crossed the adapter
// boundary, but before canonical normalization.
type ProviderRecord struct {
	Provider  string
	EpisodeID string
	Window    string
	Metrics   ProviderMetrics
	Feedback  []FeedbackSignal
}

// ProviderMetrics uses pointers so missing provider fields can be reported.
type ProviderMetrics struct {
	CTR                        *float64
	Impressions                *int
	Views                      *int
	AverageViewDurationSeconds *int
	First30Retention           *float64
	CompletionRate             *float64
	SubscribersGained          *int
	CommentsCount              *int
	Shares                     *int
	Saves                      *int
	CommunityClicks            *int
	CostPerEpisode             *float64
}

// FeedbackSignal captures normalized viewer or community feedback.
type FeedbackSignal struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Count int    `json:"count,omitempty"`
}

// Input is canonical analytics data used by insight generation.
type Input struct {
	Provider          string
	EpisodeID         string
	Window            string
	Metrics           Metrics
	Feedback          []FeedbackSignal
	MissingMetricKeys []string
	DataQualityNotes  []string
}

// Metrics is the canonical analytics metric set.
type Metrics struct {
	CTR                        float64  `json:"ctr"`
	Impressions                int      `json:"impressions"`
	Views                      int      `json:"views"`
	AverageViewDurationSeconds int      `json:"average_view_duration_seconds"`
	First30Retention           float64  `json:"first_30s_retention"`
	CompletionRate             float64  `json:"completion_rate"`
	SubscribersGained          int      `json:"subscribers_gained"`
	CommentsCount              int      `json:"comments_count"`
	Shares                     int      `json:"shares"`
	Saves                      int      `json:"saves"`
	CommunityClicks            int      `json:"community_clicks"`
	CostPerEpisode             float64  `json:"cost_per_episode"`
	MissingMetricKeys          []string `json:"missing_metric_keys,omitempty"`
}

// Report is the canonical analytics_report.json shape.
type Report struct {
	SchemaVersion      string   `json:"schema_version"`
	EpisodeID          string   `json:"episode_id"`
	ArtifactID         string   `json:"artifact_id"`
	Status             string   `json:"status"`
	Window             string   `json:"window"`
	Metrics            Metrics  `json:"metrics"`
	Insights           []string `json:"insights"`
	RecommendedActions []string `json:"recommended_actions"`
	DataQualityNotes   []string `json:"data_quality_notes,omitempty"`
	AdvisoryOnly       bool     `json:"advisory_only"`
}

// Normalize converts provider data into canonical analytics input.
func Normalize(record ProviderRecord) (Input, error) {
	if strings.TrimSpace(record.EpisodeID) == "" {
		return Input{}, fmt.Errorf("episode id is required")
	}
	if !validWindow(record.Window) {
		return Input{}, fmt.Errorf("unsupported analytics window: %s", record.Window)
	}
	metrics, missing := normalizeMetrics(record.Metrics)
	notes := dataQualityNotes(metrics, missing)
	return Input{
		Provider:          defaultText(record.Provider, "fixture"),
		EpisodeID:         record.EpisodeID,
		Window:            record.Window,
		Metrics:           metrics,
		Feedback:          append([]FeedbackSignal(nil), record.Feedback...),
		MissingMetricKeys: missing,
		DataQualityNotes:  notes,
	}, nil
}

// ImportReport imports fixture/provider data and returns a canonical report.
func ImportReport(ctx context.Context, adapter Adapter, request ImportRequest) (Report, error) {
	record, err := adapter.Import(ctx, request)
	if err != nil {
		return Report{}, err
	}
	input, err := Normalize(record)
	if err != nil {
		return Report{}, err
	}
	return ReportFromInput(input), nil
}

// ReportFromInput creates a basic imported analytics report without editorial
// recommendations. Use GenerateInsightReport for advisory insights.
func ReportFromInput(input Input) Report {
	return Report{
		SchemaVersion:      SchemaVersion,
		EpisodeID:          input.EpisodeID,
		ArtifactID:         "analytics-report-" + input.EpisodeID + "-" + input.Window + "-v1",
		Status:             fileStatus,
		Window:             input.Window,
		Metrics:            input.Metrics,
		Insights:           []string{"Analytics imported from " + defaultText(input.Provider, "fixture") + " adapter."},
		RecommendedActions: []string{"advisory: review analytics with a human editor before changing future plans."},
		DataQualityNotes:   input.DataQualityNotes,
		AdvisoryOnly:       true,
	}
}

func normalizeMetrics(metrics ProviderMetrics) (Metrics, []string) {
	var missing []string
	out := Metrics{}
	assignFloat(metrics.CTR, &out.CTR, "ctr", &missing)
	assignInt(metrics.Impressions, &out.Impressions, "impressions", &missing)
	assignInt(metrics.Views, &out.Views, "views", &missing)
	assignInt(metrics.AverageViewDurationSeconds, &out.AverageViewDurationSeconds, "average_view_duration_seconds", &missing)
	assignFloat(metrics.First30Retention, &out.First30Retention, "first_30s_retention", &missing)
	assignFloat(metrics.CompletionRate, &out.CompletionRate, "completion_rate", &missing)
	assignInt(metrics.SubscribersGained, &out.SubscribersGained, "subscribers_gained", &missing)
	assignInt(metrics.CommentsCount, &out.CommentsCount, "comments_count", &missing)
	assignInt(metrics.Shares, &out.Shares, "shares", &missing)
	assignInt(metrics.Saves, &out.Saves, "saves", &missing)
	assignInt(metrics.CommunityClicks, &out.CommunityClicks, "community_clicks", &missing)
	assignFloat(metrics.CostPerEpisode, &out.CostPerEpisode, "cost_per_episode", &missing)
	sort.Strings(missing)
	out.MissingMetricKeys = missing
	return out, missing
}

func assignFloat(value *float64, target *float64, name string, missing *[]string) {
	if value == nil {
		*missing = append(*missing, name)
		return
	}
	*target = *value
}

func assignInt(value *int, target *int, name string, missing *[]string) {
	if value == nil {
		*missing = append(*missing, name)
		return
	}
	*target = *value
}

func dataQualityNotes(metrics Metrics, missing []string) []string {
	var notes []string
	if len(missing) > 0 {
		notes = append(notes, "missing metrics: "+strings.Join(missing, ", "))
	}
	if metrics.Views < 100 {
		notes = append(notes, "low sample size; treat insights as directional only")
	}
	return notes
}

func validWindow(window string) bool {
	switch window {
	case Window24h, Window72h, Window7d:
		return true
	default:
		return false
	}
}

func defaultText(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
