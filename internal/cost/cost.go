package cost

import (
	"fmt"
	"time"
)

// Event records estimated cost for a unit of work.
type Event struct {
	EpisodeID     string    `json:"episode_id"`
	Stage         string    `json:"stage"`
	Provider      string    `json:"provider"`
	ModelID       string    `json:"model_id"`
	OperationType string    `json:"operation_type"`
	InputUnits    float64   `json:"input_units"`
	OutputUnits   float64   `json:"output_units"`
	EstimatedCost float64   `json:"estimated_cost"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

// Summary aggregates cost events.
type Summary struct {
	EpisodeID  string             `json:"episode_id"`
	Currency   string             `json:"currency"`
	Total      float64            `json:"total"`
	ByStage    map[string]float64 `json:"by_stage"`
	ByProvider map[string]float64 `json:"by_provider"`
	EventCount int                `json:"event_count"`
}

// BudgetDecision describes cost policy output.
type BudgetDecision struct {
	Allowed bool
	Reason  string
}

// Validate checks event safety and required fields.
func (e Event) Validate() error {
	if e.EpisodeID == "" {
		return fmt.Errorf("episode id is required")
	}
	if e.Stage == "" {
		return fmt.Errorf("stage is required")
	}
	if e.OperationType == "" {
		return fmt.Errorf("operation type is required")
	}
	if e.EstimatedCost < 0 {
		return fmt.Errorf("estimated cost cannot be negative")
	}
	if e.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	return nil
}

// Aggregate combines cost events into a summary.
func Aggregate(events []Event) (Summary, error) {
	if len(events) == 0 {
		return Summary{ByStage: map[string]float64{}, ByProvider: map[string]float64{}}, nil
	}
	first := events[0]
	if err := first.Validate(); err != nil {
		return Summary{}, err
	}
	summary := Summary{
		EpisodeID:  first.EpisodeID,
		Currency:   first.Currency,
		ByStage:    map[string]float64{},
		ByProvider: map[string]float64{},
	}
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return Summary{}, err
		}
		if event.EpisodeID != summary.EpisodeID {
			return Summary{}, fmt.Errorf("cannot aggregate multiple episode IDs: %s and %s", summary.EpisodeID, event.EpisodeID)
		}
		if event.Currency != summary.Currency {
			return Summary{}, fmt.Errorf("cannot aggregate multiple currencies: %s and %s", summary.Currency, event.Currency)
		}
		summary.Total += event.EstimatedCost
		summary.ByStage[event.Stage] += event.EstimatedCost
		if event.Provider != "" {
			summary.ByProvider[event.Provider] += event.EstimatedCost
		}
		summary.EventCount++
	}
	return summary, nil
}

// CheckBudget returns whether the summary stays within maxCost.
func CheckBudget(summary Summary, maxCost float64) BudgetDecision {
	if maxCost < 0 {
		return BudgetDecision{Allowed: false, Reason: "budget cannot be negative"}
	}
	if summary.Total > maxCost {
		return BudgetDecision{Allowed: false, Reason: fmt.Sprintf("cost %.4f exceeds budget %.4f %s", summary.Total, maxCost, summary.Currency)}
	}
	return BudgetDecision{Allowed: true, Reason: "within budget"}
}
