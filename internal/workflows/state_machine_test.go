package workflows

import (
	"strings"
	"testing"
	"time"
)

func TestEpisodeStateMachineHappyPathBacklogToArchived(t *testing.T) {
	m := mustStateMachine(t, StateBacklog)
	steps := []struct {
		to        EpisodeState
		actorKind ActorKind
	}{
		{StateCandidate, ActorWorkflow},
		{StateApprovedTopic, ActorHuman},
		{StateResearching, ActorWorkflow},
		{StateResearchReady, ActorWorkflow},
		{StateDrafting, ActorWorkflow},
		{StateVerifying, ActorWorkflow},
		{StateHumanQA, ActorWorkflow},
		{StateStoryboarding, ActorHuman},
		{StateAssetProduction, ActorWorkflow},
		{StateRendering, ActorWorkflow},
		{StateProductionQA, ActorWorkflow},
		{StateScheduled, ActorHuman},
		{StatePublished, ActorWorkflow},
		{StateMonitored, ActorSystem},
		{StateArchived, ActorSystem},
	}

	for _, step := range steps {
		if err := m.Transition(step.to, transitionMeta(step.actorKind)); err != nil {
			t.Fatalf("transition to %s failed: %v", step.to, err)
		}
	}

	if m.Current != StateArchived {
		t.Fatalf("expected archived, got %s", m.Current)
	}
	if len(m.History) != len(steps) {
		t.Fatalf("expected %d transitions, got %d", len(steps), len(m.History))
	}

	humanRequired := 0
	for _, transition := range m.History {
		if transition.HumanRequired {
			humanRequired++
			if transition.ActorKind != ActorHuman {
				t.Fatalf("human-required transition used %s actor: %+v", transition.ActorKind, transition)
			}
		}
	}
	if humanRequired != 3 {
		t.Fatalf("expected 3 human-required transitions, got %d", humanRequired)
	}
}

func TestEpisodeStateMachineRejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from EpisodeState
		to   EpisodeState
	}{
		{name: "backlog to published", from: StateBacklog, to: StatePublished},
		{name: "drafting to published", from: StateDrafting, to: StatePublished},
		{name: "human qa to rendering", from: StateHumanQA, to: StateRendering},
		{name: "production qa to published", from: StateProductionQA, to: StatePublished},
		{name: "scheduled to archived", from: StateScheduled, to: StateArchived},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mustStateMachine(t, tt.from)
			err := m.Transition(tt.to, transitionMeta(ActorHuman))
			if err == nil {
				t.Fatalf("expected %s -> %s to fail", tt.from, tt.to)
			}
		})
	}
}

func TestEpisodeStateMachineRequiresTransitionMetadata(t *testing.T) {
	tests := []struct {
		name string
		meta TransitionMetadata
		want string
	}{
		{name: "missing reason", meta: TransitionMetadata{Actor: "workflow:test", ActorKind: ActorWorkflow, At: fixedTransitionTime()}, want: "reason"},
		{name: "missing actor", meta: TransitionMetadata{Reason: "advance", ActorKind: ActorWorkflow, At: fixedTransitionTime()}, want: "actor"},
		{name: "invalid actor kind", meta: TransitionMetadata{Reason: "advance", Actor: "workflow:test", ActorKind: "provider", At: fixedTransitionTime()}, want: "actor kind"},
		{name: "missing timestamp", meta: TransitionMetadata{Reason: "advance", Actor: "workflow:test", ActorKind: ActorWorkflow}, want: "timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mustStateMachine(t, StateBacklog)
			err := m.Transition(StateCandidate, tt.meta)
			if err == nil {
				t.Fatal("expected metadata validation failure")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestEpisodeStateMachineLabelsAndEnforcesHumanRequiredTransitions(t *testing.T) {
	if !HumanRequiredTransition(StateHumanQA, StateStoryboarding) {
		t.Fatal("expected human QA to storyboarding to be human-required")
	}
	if HumanRequiredTransition(StateRendering, StateProductionQA) {
		t.Fatal("rendering to production QA should not be human-required")
	}

	m := mustStateMachine(t, StateHumanQA)
	if err := m.Transition(StateStoryboarding, transitionMeta(ActorWorkflow)); err == nil {
		t.Fatal("expected workflow actor to be rejected for human-required transition")
	}
	if err := m.Transition(StateStoryboarding, transitionMeta(ActorHuman)); err != nil {
		t.Fatalf("expected human actor to pass: %v", err)
	}
	if !m.History[0].HumanRequired {
		t.Fatal("expected transition record to label human-required gate")
	}
}

func TestEpisodeStateMachineRequiresExplicitBlockedAndUnblockedTransitions(t *testing.T) {
	m := mustStateMachine(t, StateDrafting)

	if err := m.Transition(StateBlocked, transitionMeta(ActorWorkflow)); err == nil {
		t.Fatal("expected direct transition to blocked to be rejected")
	}
	if err := m.Block(transitionMeta(ActorSystem)); err != nil {
		t.Fatalf("block failed: %v", err)
	}
	if m.Current != StateBlocked {
		t.Fatalf("expected blocked, got %s", m.Current)
	}
	if blockedFrom, ok := m.BlockedFrom(); !ok || blockedFrom != StateDrafting {
		t.Fatalf("expected blocked from drafting, got %s ok=%v", blockedFrom, ok)
	}

	if err := m.Transition(StateVerifying, transitionMeta(ActorWorkflow)); err == nil {
		t.Fatal("expected blocked episode to reject silent continuation")
	}
	if err := m.Unblock(transitionMeta(ActorModel)); err == nil {
		t.Fatal("expected model unblock to be rejected")
	}
	if err := m.Unblock(transitionMeta(ActorHuman)); err != nil {
		t.Fatalf("human unblock failed: %v", err)
	}
	if m.Current != StateDrafting {
		t.Fatalf("expected unblock to return to drafting, got %s", m.Current)
	}
	if err := m.Transition(StateVerifying, transitionMeta(ActorWorkflow)); err != nil {
		t.Fatalf("expected continuation after explicit unblock: %v", err)
	}
}

func TestAllowedNextStatesAreStable(t *testing.T) {
	got := AllowedNextStates(StateCandidate)
	want := []EpisodeState{StateApprovedTopic, StateArchived}
	if len(got) != len(want) {
		t.Fatalf("expected %d next states, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("next state %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func mustStateMachine(t *testing.T, initial EpisodeState) EpisodeStateMachine {
	t.Helper()
	m, err := NewEpisodeStateMachine(initial)
	if err != nil {
		t.Fatalf("new state machine: %v", err)
	}
	return m
}

func transitionMeta(actorKind ActorKind) TransitionMetadata {
	return TransitionMetadata{
		Reason:    "test transition",
		Actor:     string(actorKind) + ":test",
		ActorKind: actorKind,
		At:        fixedTransitionTime(),
	}
}

func fixedTransitionTime() time.Time {
	return time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
}
