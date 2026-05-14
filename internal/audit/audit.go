package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/security"
)

// Category classifies an auditable production event.
type Category string

const (
	CategoryArtifactValidation Category = "artifact_validation"
	CategoryResearchAudit      Category = "research_audit"
	CategoryStateTransition    Category = "state_transition"
	CategoryModelRouting       Category = "model_routing"
	CategoryCouncilDecision    Category = "council_decision"
	CategoryClaimVerification  Category = "claim_verification"
	CategoryHumanQA            Category = "human_qa"
	CategoryProductionQA       Category = "production_qa"
	CategoryReleaseApproval    Category = "release_approval"
	CategoryPublishing         Category = "publishing"
	CategorySecurity           Category = "security"
	CategoryCost               Category = "cost"
	CategoryIncidentCorrection Category = "incident_correction"
)

// Event is the canonical structured audit record.
type Event struct {
	ID            string            `json:"id"`
	EpisodeID     string            `json:"episode_id,omitempty"`
	ArtifactID    string            `json:"artifact_id,omitempty"`
	Category      Category          `json:"category"`
	Actor         string            `json:"actor"`
	Decision      string            `json:"decision,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// Validate checks that an audit event is useful and attributable.
func (e Event) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("audit event id is required")
	}
	if e.Category == "" {
		return fmt.Errorf("audit event category is required")
	}
	if e.Actor == "" {
		return fmt.Errorf("audit event actor is required")
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("audit event created_at is required")
	}
	if e.Category == CategoryReleaseApproval && e.Decision == "approve" && !strings.HasPrefix(e.Actor, "human:") {
		return fmt.Errorf("release approval audit event requires human actor metadata")
	}
	return nil
}

// Redacted returns a copy with secret-like values removed from free-text fields.
func (e Event) Redacted() Event {
	out := e
	out.Decision = security.Redact(out.Decision)
	out.Reason = security.Redact(out.Reason)
	if len(e.Metadata) > 0 {
		out.Metadata = map[string]string{}
		for key, value := range e.Metadata {
			out.Metadata[key] = security.Redact(value)
		}
	}
	return out
}

// Sink stores audit events.
type Sink interface {
	Append(event Event) error
	Events() []Event
}

// MemorySink is deterministic and useful for local dry-runs/tests.
type MemorySink struct {
	events []Event
}

func NewMemorySink() *MemorySink {
	return &MemorySink{events: []Event{}}
}

func (s *MemorySink) Append(event Event) error {
	event = event.Redacted()
	if err := event.Validate(); err != nil {
		return err
	}
	s.events = append(s.events, event)
	return nil
}

func (s *MemorySink) Events() []Event {
	return append([]Event(nil), s.events...)
}

// JSONLines renders events as JSON Lines for logs or local artifacts.
func JSONLines(events []Event) (string, error) {
	out := ""
	for _, event := range events {
		event = event.Redacted()
		if err := event.Validate(); err != nil {
			return "", err
		}
		encoded, err := json.Marshal(event)
		if err != nil {
			return "", err
		}
		out += string(encoded) + "\n"
	}
	return out, nil
}

// NewEvent creates a validated event with stable caller-provided id.
func NewEvent(id string, category Category, actor string, episodeID string, decision string, reason string) Event {
	return NewEventAt(id, category, actor, episodeID, decision, reason, time.Now().UTC())
}

// NewEventAt creates an event with caller-provided time for deterministic tests
// and Temporal workflow replay.
func NewEventAt(id string, category Category, actor string, episodeID string, decision string, reason string, createdAt time.Time) Event {
	return Event{
		ID:        id,
		Category:  category,
		Actor:     actor,
		EpisodeID: episodeID,
		Decision:  decision,
		Reason:    reason,
		CreatedAt: createdAt.UTC(),
	}
}

// NewStateTransition records a workflow or orchestration state change.
func NewStateTransition(id string, actor string, episodeID string, from string, to string, correlationID string, reason string, createdAt time.Time) Event {
	event := NewEventAt(id, CategoryStateTransition, actor, episodeID, to, reason, createdAt)
	event.CorrelationID = correlationID
	event.Metadata = map[string]string{
		"from_state": from,
		"to_state":   to,
	}
	return event
}
