package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AnimusHQ/news/internal/artifacts"
)

func TestDryRunPassesForCompleteBundle(t *testing.T) {
	dir := t.TempDir()
	for _, name := range artifacts.RequiredEpisodeFiles {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("placeholder"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	report, err := DryRun(dir)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if !report.ArtifactsValid {
		t.Fatal("expected artifacts to be valid")
	}
	if len(report.Blockers) != 0 {
		t.Fatalf("expected no blockers, got %v", report.Blockers)
	}
}

func TestDryRunFailsForMissingBundle(t *testing.T) {
	report, err := DryRun(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected dry run to fail for missing directory")
	}
	if len(report.Blockers) == 0 {
		t.Fatal("expected blockers to be populated")
	}
}
