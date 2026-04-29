package temporalops

import (
	"context"
	"strings"
	"testing"
)

func TestSignalHumanQARejectsInvalidDecisionBeforeDial(t *testing.T) {
	err := SignalHumanQA(context.Background(), Config{}, "workflow-id", "maybe")
	if err == nil {
		t.Fatal("expected invalid human QA decision to fail")
	}
	if !strings.Contains(err.Error(), "invalid human QA decision") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignalReleaseRejectsInvalidDecisionBeforeDial(t *testing.T) {
	err := SignalRelease(context.Background(), Config{}, "workflow-id", "maybe")
	if err == nil {
		t.Fatal("expected invalid release decision to fail")
	}
	if !strings.Contains(err.Error(), "invalid release decision") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartEpisodeRequiresEpisodeID(t *testing.T) {
	_, err := StartEpisode(context.Background(), Config{}, "", "episodes/0001-after-git-push")
	if err == nil {
		t.Fatal("expected missing episode ID to fail")
	}
}

func TestStartEpisodeRequiresEpisodeDir(t *testing.T) {
	_, err := StartEpisode(context.Background(), Config{}, "episode-0001", "")
	if err == nil {
		t.Fatal("expected missing episode directory to fail")
	}
}

func TestQueryEpisodeStateRequiresWorkflowID(t *testing.T) {
	_, err := QueryEpisodeState(context.Background(), Config{}, "")
	if err == nil {
		t.Fatal("expected missing workflow ID to fail")
	}
}
