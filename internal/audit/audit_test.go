package audit

import (
	"strings"
	"testing"
	"time"
)

func TestMemorySinkStoresStructuredRedactedEvent(t *testing.T) {
	sink := NewMemorySink()
	secret := fakeAuditSecret()
	event := NewEventAt(
		"event-1",
		CategorySecurity,
		"system:test",
		"episode-1",
		"scan",
		"token="+secret,
		time.Unix(100, 0).UTC(),
	)
	event.Metadata = map[string]string{"api_key": "api_key=" + secret}

	if err := sink.Append(event); err != nil {
		t.Fatalf("append event failed: %v", err)
	}
	stored := sink.Events()[0]
	if strings.Contains(stored.Reason, secret) {
		t.Fatalf("secret-like value was not redacted from reason: %s", stored.Reason)
	}
	if strings.Contains(stored.Metadata["api_key"], secret) {
		t.Fatalf("secret-like value was not redacted from metadata: %s", stored.Metadata["api_key"])
	}
}

func TestReleaseApprovalRequiresHumanActor(t *testing.T) {
	event := NewEventAt("event-1", CategoryReleaseApproval, "system:workflow", "episode-1", "approve", "release approved", time.Unix(100, 0).UTC())
	if err := event.Validate(); err == nil {
		t.Fatal("expected non-human release approval event to fail")
	}
	event.Actor = "human:operator"
	if err := event.Validate(); err != nil {
		t.Fatalf("expected human actor to satisfy release approval event: %v", err)
	}
}

func TestStateTransitionEventIsStructured(t *testing.T) {
	event := NewStateTransition("transition-1", "workflow:test", "episode-1", "started", "validating", "corr-1", "begin validation", time.Unix(100, 0).UTC())
	if err := event.Validate(); err != nil {
		t.Fatalf("state transition event should validate: %v", err)
	}
	if event.Category != CategoryStateTransition {
		t.Fatalf("unexpected category: %s", event.Category)
	}
	if event.Metadata["from_state"] != "started" || event.Metadata["to_state"] != "validating" {
		t.Fatalf("missing transition metadata: %+v", event.Metadata)
	}
}

func TestJSONLinesRedactsEvents(t *testing.T) {
	secret := fakeAuditSecret()
	event := NewEventAt("event-1", CategorySecurity, "system:test", "episode-1", "scan", "password="+secret, time.Unix(100, 0).UTC())
	lines, err := JSONLines([]Event{event})
	if err != nil {
		t.Fatalf("json lines failed: %v", err)
	}
	if strings.Contains(lines, secret) {
		t.Fatalf("json lines leaked secret-like value: %s", lines)
	}
}

func fakeAuditSecret() string {
	return strings.Repeat("a", 16) + "1234567890"
}
