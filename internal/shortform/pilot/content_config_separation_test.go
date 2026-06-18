package pilot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentConfigurationSeparation(t *testing.T) {
	root := filepath.Clean("../../..")

	env := readRepoFile(t, root, ".env.example")
	for _, forbidden := range []string{"PROMPT", "TOPIC", "EPISODE_ID", "VIDEO_", "ANIMUS_MVP_PROMPT"} {
		if strings.Contains(env, forbidden) {
			t.Fatalf(".env.example must not contain runtime content key %q", forbidden)
		}
	}
	assertNoThemeDefaults(t, ".env.example", env)

	for _, rel := range []string{
		"scripts/providers/seedance2-visual-wrapper.example.py",
		"scripts/providers/chatterbox-voice-wrapper.example.py",
	} {
		body := readRepoFile(t, root, rel)
		assertNoThemeDefaults(t, rel, body)
		if !strings.Contains(body, "ANIMUS_ALLOW_LIVE_PROVIDER_CALLS") {
			t.Fatalf("%s must require ANIMUS_ALLOW_LIVE_PROVIDER_CALLS before live calls", rel)
		}
	}

	for _, rel := range []string{
		"internal/shortform/pilot/helpers.go",
		"internal/shortform/pilot/pipeline.go",
		"internal/shortform/pilot/types.go",
		"internal/shortform/pilot/render.go",
		"internal/shortform/pilot/review.go",
	} {
		assertNoThemeDefaults(t, rel, readRepoFile(t, root, rel))
	}
}

func TestRuntimePromptDrivesScriptAndShotPrompts(t *testing.T) {
	manifest := EpisodeManifest{
		EpisodeID:      "episode-test-001",
		OriginalPrompt: "Explain the runtime-provided subject without adding a fixed theme",
		Language:       "en",
		Duration:       "9s",
		DurationSec:    9,
	}
	script := buildScript(manifest)
	if !strings.Contains(script, manifest.OriginalPrompt) {
		t.Fatal("script must include the runtime prompt")
	}
	assertNoThemeDefaults(t, "script", script)

	shots := buildShotRequests(manifest)
	if len(shots) != 3 {
		t.Fatalf("expected 3 generic shot requests, got %d", len(shots))
	}
	for _, shot := range shots {
		if !strings.Contains(shot.Prompt, manifest.OriginalPrompt) {
			t.Fatalf("%s prompt must include runtime prompt: %q", shot.ShotID, shot.Prompt)
		}
		assertNoThemeDefaults(t, shot.ShotID, shot.Prompt)
	}
}

func readRepoFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func assertNoThemeDefaults(t *testing.T, name, body string) {
	t.Helper()
	lower := strings.ToLower(body)
	for _, forbidden := range []string{
		"animus-oss",
		"open-source",
		"open source",
		"developer ecosystem",
		"sustainable ecosystem",
		"sustainable ecosystems",
		"maintainers",
		"contributors",
		"release candidate handoff",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("%s contains hardcoded content theme %q", name, forbidden)
		}
	}
}
