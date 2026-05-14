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
	ByModel    map[string]float64 `json:"by_model"`
	ByDay      map[string]float64 `json:"by_day,omitempty"`
	EventCount int                `json:"event_count"`
}

// BudgetAction is the normalized budget policy outcome.
type BudgetAction string

const (
	BudgetActionAllow           BudgetAction = "allow"
	BudgetActionWarn            BudgetAction = "warn"
	BudgetActionRequireApproval BudgetAction = "require_approval"
	BudgetActionBlock           BudgetAction = "block"
)

// BudgetPolicy describes cost thresholds for automation.
type BudgetPolicy struct {
	WarnAt            float64
	RequireApprovalAt float64
	BlockAt           float64
	Currency          string
}

// BudgetDecision describes cost policy output.
type BudgetDecision struct {
	Allowed bool
	Action  BudgetAction
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
		return Summary{ByStage: map[string]float64{}, ByProvider: map[string]float64{}, ByModel: map[string]float64{}, ByDay: map[string]float64{}}, nil
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
		ByModel:    map[string]float64{},
		ByDay:      map[string]float64{},
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
		if event.ModelID != "" {
			summary.ByModel[event.ModelID] += event.EstimatedCost
		}
		if !event.CreatedAt.IsZero() {
			summary.ByDay[event.CreatedAt.UTC().Format("2006-01-02")] += event.EstimatedCost
		}
		summary.EventCount++
	}
	return summary, nil
}

// CheckBudget returns whether the summary stays within maxCost.
func CheckBudget(summary Summary, maxCost float64) BudgetDecision {
	if maxCost < 0 {
		return BudgetDecision{Allowed: false, Action: BudgetActionBlock, Reason: "budget cannot be negative"}
	}
	if summary.Total > maxCost {
		return BudgetDecision{Allowed: false, Action: BudgetActionBlock, Reason: fmt.Sprintf("cost %.4f exceeds budget %.4f %s", summary.Total, maxCost, summary.Currency)}
	}
	return BudgetDecision{Allowed: true, Action: BudgetActionAllow, Reason: "within budget"}
}

// EvaluateBudget applies warn, approval, and block thresholds. Critical work is
// never auto-approved; callers still need the normal quality gates.
func EvaluateBudget(summary Summary, policy BudgetPolicy, critical bool) BudgetDecision {
	if policy.Currency != "" && summary.Currency != "" && policy.Currency != summary.Currency {
		return BudgetDecision{Allowed: false, Action: BudgetActionBlock, Reason: fmt.Sprintf("budget currency %s does not match summary currency %s", policy.Currency, summary.Currency)}
	}
	if policy.BlockAt > 0 && summary.Total >= policy.BlockAt {
		if critical {
			return BudgetDecision{Allowed: false, Action: BudgetActionRequireApproval, Reason: fmt.Sprintf("critical work cost %.4f reaches block threshold %.4f %s and requires approval", summary.Total, policy.BlockAt, summary.Currency)}
		}
		return BudgetDecision{Allowed: false, Action: BudgetActionBlock, Reason: fmt.Sprintf("cost %.4f reaches block threshold %.4f %s", summary.Total, policy.BlockAt, summary.Currency)}
	}
	if policy.RequireApprovalAt > 0 && summary.Total >= policy.RequireApprovalAt {
		return BudgetDecision{Allowed: false, Action: BudgetActionRequireApproval, Reason: fmt.Sprintf("cost %.4f requires approval at %.4f %s", summary.Total, policy.RequireApprovalAt, summary.Currency)}
	}
	if policy.WarnAt > 0 && summary.Total >= policy.WarnAt {
		return BudgetDecision{Allowed: true, Action: BudgetActionWarn, Reason: fmt.Sprintf("cost %.4f reaches warning threshold %.4f %s", summary.Total, policy.WarnAt, summary.Currency)}
	}
	return BudgetDecision{Allowed: true, Action: BudgetActionAllow, Reason: "within budget policy"}
}
