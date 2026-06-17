package runner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/gates"
)

var demoNow = time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

func runDemo(t *testing.T, inject Injection) Result {
	t.Helper()
	res, err := Run(context.Background(), Config{
		EpisodeID: "episode-0001",
		OutputDir: t.TempDir(),
		Now:       demoNow,
		Inject:    inject,
	})
	if err != nil {
		t.Fatalf("runner error: %v", err)
	}
	return res
}

func TestRunnerHappyPathReachesTerminalState(t *testing.T) {
	res := runDemo(t, InjectNone)
	if res.Blocked {
		t.Fatalf("happy path must not block: %s", res.BlockReason)
	}
	if res.State != "published_dry_run_complete" {
		t.Fatalf("unexpected terminal state: %s", res.State)
	}

	wantKinds := []string{
		shortform.KindStoryboardImageManifest, shortform.KindVisualShotManifest,
		shortform.KindVoiceoverManifest, shortform.KindSubtitleManifest,
		shortform.KindShortRenderManifest, shortform.KindProductionCandidate,
		shortform.KindReleaseApproval, shortform.KindUploadPostPublishManifest,
	}
	for _, kind := range wantKinds {
		if res.Artifacts[kind] == "" {
			t.Fatalf("missing artifact %s", kind)
		}
		path := filepath.Join(res.RunDir, kind+".json")
		if issues := shortform.ValidateFile(path); len(issues) != 0 {
			t.Fatalf("persisted %s invalid: %v", kind, issues)
		}
	}
	for _, g := range res.GateResults {
		if g.Blocked() {
			t.Fatalf("gate %s unexpectedly blocked", g.Gate)
		}
	}

	for _, name := range []string{"gate_decisions.json", "audit.jsonl", "run_summary.json"} {
		if _, err := os.Stat(filepath.Join(res.RunDir, name)); err != nil {
			t.Fatalf("missing run output %s: %v", name, err)
		}
	}
}

func TestRunnerUnapprovedStoryboardHaltsAtStoryboardGate(t *testing.T) {
	res := runDemo(t, InjectUnapprovedStoryboard)
	if !res.Blocked {
		t.Fatal("expected the run to halt")
	}
	last := res.GateResults[len(res.GateResults)-1]
	if last.Gate != "storyboard_image" {
		t.Fatalf("expected halt at storyboard_image gate, got %s", last.Gate)
	}
	if !hasReason(last, "image_not_approved") {
		t.Fatalf("expected image_not_approved reason, got %v", last.Reasons)
	}
	// Downstream artifacts must not have been produced.
	if _, ok := res.Artifacts[shortform.KindVisualShotManifest]; ok {
		t.Fatal("visual shots must not be produced after a storyboard block")
	}
	// The block is recorded in the run summary and audit log.
	assertSummaryBlocked(t, res.RunDir, "storyboard_image")
	assertAuditContains(t, res.RunDir, "halted at storyboard_image gate")
}

func TestRunnerRenderNoAudioHaltsAtRenderGate(t *testing.T) {
	res := runDemo(t, InjectRenderNoAudio)
	if !res.Blocked {
		t.Fatal("expected the run to halt")
	}
	last := res.GateResults[len(res.GateResults)-1]
	if last.Gate != "render" {
		t.Fatalf("expected halt at render gate, got %s", last.Gate)
	}
}

func TestRunnerReleaseDeniedBlocks(t *testing.T) {
	res := runDemo(t, InjectReleaseDenied)
	if !res.Blocked || res.State != "release_denied" {
		t.Fatalf("expected release_denied, got state=%s blocked=%v", res.State, res.Blocked)
	}
	// Publish manifest must not exist.
	if _, ok := res.Artifacts[shortform.KindUploadPostPublishManifest]; ok {
		t.Fatal("publish manifest must not be produced when release is denied")
	}
}

func TestRunnerIsDeterministic(t *testing.T) {
	a := runDemo(t, InjectNone)
	b := runDemo(t, InjectNone)
	for kind, hash := range a.Artifacts {
		if b.Artifacts[kind] != hash {
			t.Fatalf("non-deterministic hash for %s: %s vs %s", kind, hash, b.Artifacts[kind])
		}
	}
}

func hasReason(r gates.Result, code string) bool {
	for _, reason := range r.Reasons {
		if reason.Code == code {
			return true
		}
	}
	return false
}

func assertSummaryBlocked(t *testing.T, dir, gate string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "run_summary.json"))
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	var summary Result
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if !summary.Blocked || !strings.Contains(summary.BlockReason, gate) {
		t.Fatalf("summary did not record block at %s: %+v", gate, summary)
	}
}

func assertAuditContains(t *testing.T, dir, substr string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	if !strings.Contains(string(data), substr) {
		t.Fatalf("audit log missing %q", substr)
	}
}
