package localexec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveUnderAllowsContainedRelativePath(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveUnder(root, filepath.Join("assets", "input.mp4"), "input")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, root) {
		t.Fatalf("resolved path outside root: %s", got)
	}
}

func TestResolveUnderRejectsTraversalAndOutsideAbsolutePath(t *testing.T) {
	root := t.TempDir()
	if _, err := ResolveUnder(root, "../outside.mp4", "input"); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
	outside := filepath.Join(t.TempDir(), "outside.mp4")
	if _, err := ResolveUnder(root, outside, "input"); err == nil {
		t.Fatal("expected outside absolute path to be rejected")
	}
}

func TestSafeSegmentRejectsPathSeparators(t *testing.T) {
	if err := SafeSegment("episode-0001", "episode_id"); err != nil {
		t.Fatal(err)
	}
	if err := SafeSegment("../episode", "episode_id"); err == nil {
		t.Fatal("expected path separator segment to be rejected")
	}
}
