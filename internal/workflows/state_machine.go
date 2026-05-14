package workflows

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// EpisodeState is the canonical lifecycle state for an episode.
type EpisodeState string

const (
	StateBacklog         EpisodeState = "backlog"
	StateCandidate       EpisodeState = "candidate"
	StateApprovedTopic   EpisodeState = "approved_topic"
	StateResearching     EpisodeState = "researching"
	StateResearchReady   EpisodeState = "research_ready"
	StateDrafting        EpisodeState = "drafting"
	StateVerifying       EpisodeState = "verifying"
	StateHumanQA         EpisodeState = "human_qa"
	StateStoryboarding   EpisodeState = "storyboarding"
	StateAssetProduction EpisodeState = "asset_production"
	StateRendering       EpisodeState = "rendering"
	StateProductionQA    EpisodeState = "production_qa"
	StateScheduled       EpisodeState = "scheduled"
	StatePublished       EpisodeState = "published"
	StateMonitored       EpisodeState = "monitored"
	StateArchived        EpisodeState = "archived"
	StateBlocked         EpisodeState = "blocked"
)

// ActorKind identifies who initiated a lifecycle transition.
type ActorKind string

const (
	ActorHuman    ActorKind = "human"
	ActorSystem   ActorKind = "system"
	ActorModel    ActorKind = "model"
	ActorWorkflow ActorKind = "workflow"
)

// TransitionMetadata captures the audit metadata required for every transition.
type TransitionMetadata struct {
	Reason    string
	Actor     string
	ActorKind ActorKind
	At        time.Time
}

// StateTransition is an immutable transition record for an episode lifecycle.
type StateTransition struct {
	From          EpisodeState `json:"from"`
	To            EpisodeState `json:"to"`
	Reason        string       `json:"reason"`
	Actor         string       `json:"actor"`
	ActorKind     ActorKind    `json:"actor_kind"`
	At            time.Time    `json:"at"`
	HumanRequired bool         `json:"human_required"`
}

// EpisodeStateMachine enforces the canonical episode lifecycle transition table.
type EpisodeStateMachine struct {
	Current     EpisodeState      `json:"current"`
	History     []StateTransition `json:"history"`
	blockedFrom EpisodeState
}

type transitionRule struct {
	humanRequired bool
}

var canonicalStates = map[EpisodeState]struct{}{
	StateBacklog:         {},
	StateCandidate:       {},
	StateApprovedTopic:   {},
	StateResearching:     {},
	StateResearchReady:   {},
	StateDrafting:        {},
	StateVerifying:       {},
	StateHumanQA:         {},
	StateStoryboarding:   {},
	StateAssetProduction: {},
	StateRendering:       {},
	StateProductionQA:    {},
	StateScheduled:       {},
	StatePublished:       {},
	StateMonitored:       {},
	StateArchived:        {},
	StateBlocked:         {},
}

var allowedTransitions = map[EpisodeState]map[EpisodeState]transitionRule{
	StateBacklog: {
		StateCandidate: {},
	},
	StateCandidate: {
		StateApprovedTopic: {humanRequired: true},
		StateArchived:      {},
	},
	StateApprovedTopic: {
		StateResearching: {},
	},
	StateResearching: {
		StateResearchReady: {},
	},
	StateResearchReady: {
		StateDrafting: {},
	},
	StateDrafting: {
		StateVerifying: {},
	},
	StateVerifying: {
		StateHumanQA: {},
	},
	StateHumanQA: {
		StateStoryboarding: {humanRequired: true},
	},
	StateStoryboarding: {
		StateAssetProduction: {},
	},
	StateAssetProduction: {
		StateRendering: {},
	},
	StateRendering: {
		StateProductionQA: {},
	},
	StateProductionQA: {
		StateScheduled: {humanRequired: true},
	},
	StateScheduled: {
		StatePublished: {},
	},
	StatePublished: {
		StateMonitored: {},
	},
	StateMonitored: {
		StateArchived: {},
	},
}

// NewEpisodeStateMachine creates a lifecycle state machine at a valid state.
func NewEpisodeStateMachine(initial EpisodeState) (EpisodeStateMachine, error) {
	if !initial.Valid() {
		return EpisodeStateMachine{}, fmt.Errorf("unknown episode state %q", initial)
	}
	return EpisodeStateMachine{Current: initial}, nil
}

// Valid reports whether the state is part of the canonical lifecycle.
func (s EpisodeState) Valid() bool {
	_, ok := canonicalStates[s]
	return ok
}

// Valid reports whether the actor kind is allowed in transition metadata.
func (a ActorKind) Valid() bool {
	switch a {
	case ActorHuman, ActorSystem, ActorModel, ActorWorkflow:
		return true
	default:
		return false
	}
}

// Transition advances to the next canonical state if the transition is allowed.
func (m *EpisodeStateMachine) Transition(to EpisodeState, meta TransitionMetadata) error {
	if to == StateBlocked {
		return fmt.Errorf("use Block for explicit blocked transitions")
	}
	if m.Current == StateBlocked {
		return fmt.Errorf("episode is blocked; use Unblock before continuing")
	}
	if err := validateTransitionMetadata(meta); err != nil {
		return err
	}
	if !to.Valid() {
		return fmt.Errorf("unknown episode state %q", to)
	}

	rule, ok := allowedTransitions[m.Current][to]
	if !ok {
		return fmt.Errorf("invalid episode transition %s -> %s", m.Current, to)
	}
	if rule.humanRequired && meta.ActorKind != ActorHuman {
		return fmt.Errorf("transition %s -> %s requires a human actor", m.Current, to)
	}

	m.appendTransition(to, meta, rule.humanRequired)
	return nil
}

// Block moves any active non-terminal state into blocked with explicit metadata.
func (m *EpisodeStateMachine) Block(meta TransitionMetadata) error {
	if m.Current == StateBlocked {
		return fmt.Errorf("episode is already blocked")
	}
	if m.Current == StateArchived {
		return fmt.Errorf("archived episodes cannot be blocked")
	}
	if err := validateTransitionMetadata(meta); err != nil {
		return err
	}

	m.blockedFrom = m.Current
	m.appendTransition(StateBlocked, meta, false)
	return nil
}

// Unblock returns a blocked episode to the state where it was blocked.
func (m *EpisodeStateMachine) Unblock(meta TransitionMetadata) error {
	if m.Current != StateBlocked {
		return fmt.Errorf("episode is not blocked")
	}
	if m.blockedFrom == "" {
		return fmt.Errorf("blocked episode has no recorded source state")
	}
	if err := validateTransitionMetadata(meta); err != nil {
		return err
	}
	if meta.ActorKind != ActorHuman && meta.ActorKind != ActorSystem {
		return fmt.Errorf("unblock requires human or system actor")
	}

	to := m.blockedFrom
	m.blockedFrom = ""
	m.appendTransition(to, meta, false)
	return nil
}

// BlockedFrom returns the state where the episode entered blocked.
func (m EpisodeStateMachine) BlockedFrom() (EpisodeState, bool) {
	if m.blockedFrom == "" {
		return "", false
	}
	return m.blockedFrom, true
}

// HumanRequiredTransition reports whether a transition requires a human actor.
func HumanRequiredTransition(from EpisodeState, to EpisodeState) bool {
	rule, ok := allowedTransitions[from][to]
	return ok && rule.humanRequired
}

// AllowedNextStates returns the allowed non-blocking transitions from a state.
func AllowedNextStates(from EpisodeState) []EpisodeState {
	next := make([]EpisodeState, 0, len(allowedTransitions[from]))
	for state := range allowedTransitions[from] {
		next = append(next, state)
	}
	sort.Slice(next, func(i int, j int) bool {
		return next[i] < next[j]
	})
	return next
}

func validateTransitionMetadata(meta TransitionMetadata) error {
	if strings.TrimSpace(meta.Reason) == "" {
		return fmt.Errorf("transition reason is required")
	}
	if strings.TrimSpace(meta.Actor) == "" {
		return fmt.Errorf("transition actor is required")
	}
	if !meta.ActorKind.Valid() {
		return fmt.Errorf("transition actor kind %q is invalid", meta.ActorKind)
	}
	if meta.At.IsZero() {
		return fmt.Errorf("transition timestamp is required")
	}
	return nil
}

func (m *EpisodeStateMachine) appendTransition(to EpisodeState, meta TransitionMetadata, humanRequired bool) {
	from := m.Current
	m.Current = to
	m.History = append(m.History, StateTransition{
		From:          from,
		To:            to,
		Reason:        strings.TrimSpace(meta.Reason),
		Actor:         strings.TrimSpace(meta.Actor),
		ActorKind:     meta.ActorKind,
		At:            meta.At,
		HumanRequired: humanRequired,
	})
}
