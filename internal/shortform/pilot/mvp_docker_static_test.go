package pilot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/security"
)

func TestMVPDockerStatic(t *testing.T) {
	root := filepath.Clean("../../..")

	env := readRepoFile(t, root, ".env.mvp.example")
	for _, key := range []string{
		"ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=",
		"ANTHROPIC_API_KEY=",
		"SEEDANCE_API_KEY=",
		"CHATTERBOX_API_KEY=",
		"CHATTERBOX_VOICE=",
		"CHATTERBOX_VOICE_CONSENT_REFERENCE=",
	} {
		if !strings.Contains(env, key) {
			t.Fatalf(".env.mvp.example missing %s", key)
		}
	}
	for _, forbidden := range []string{
		"PROMPT",
		"TOPIC",
		"EPISODE_ID",
		"DURATION",
		"PLATFORMS",
		"VIDEO_",
		"ANIMUS_MVP_PROMPT",
		"ANIMUS_VISUAL_COMMAND",
		"ANIMUS_VOICE_COMMAND",
		"ANIMUS_FFMPEG_BINARY",
		"/workspace/episodes",
		"/app/scripts/providers",
	} {
		if strings.Contains(env, forbidden) {
			t.Fatalf(".env.mvp.example must not contain runtime/local key or path %q", forbidden)
		}
	}

	compose := readRepoFile(t, root, "docker-compose.mvp.yml")
	for _, required := range []string{
		"animus-news:",
		"chatterbox:",
		"./episodes:/workspace/episodes",
		"env_file:",
		".env.mvp.local",
		"condition: service_healthy",
	} {
		if !strings.Contains(compose, required) {
			t.Fatalf("docker-compose.mvp.yml missing %q", required)
		}
	}
	if !strings.Contains(compose, "PROMPT: ${PROMPT:-}") {
		t.Fatal("docker-compose.mvp.yml must pass PROMPT from runtime shell variables without a default")
	}
	assertNoThemeDefaults(t, "docker-compose.mvp.yml", compose)
	for _, forbidden := range []string{"runtime prompt here", "topic:", "cta:", "style:"} {
		if strings.Contains(strings.ToLower(compose), forbidden) {
			t.Fatalf("docker-compose.mvp.yml must not hardcode content value %q", forbidden)
		}
	}

	entrypoint := readRepoFile(t, root, "docker/animus_news_entrypoint.sh")
	for _, required := range []string{
		"require_var PROMPT",
		"ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1 is required",
		"go run ./cmd/animus-news pilot generate-real",
		"--subtitle-provider script-timing",
		`--out "/workspace/episodes/$EPISODE_ID"`,
	} {
		if !strings.Contains(entrypoint, required) {
			t.Fatalf("docker entrypoint missing %q", required)
		}
	}

	gitignore := readRepoFile(t, root, ".gitignore")
	if !strings.Contains(gitignore, ".env.mvp.local") {
		t.Fatal(".env.mvp.local must be explicitly gitignored")
	}

	for _, rel := range []string{
		".env.mvp.example",
		"docker-compose.mvp.yml",
		"docker/animus-news.Dockerfile",
		"docker/chatterbox-server.Dockerfile",
		"docker/chatterbox_server.py",
		"docker/animus_news_entrypoint.sh",
	} {
		assertNoHighRiskSecrets(t, filepath.Join(root, filepath.FromSlash(rel)))
	}
}

func TestMVPDockerEntrypointRejectsEmptyPrompt(t *testing.T) {
	root := filepath.Clean("../../..")
	entrypoint := filepath.Join(root, "docker", "animus_news_entrypoint.sh")
	cmd := exec.Command("bash", entrypoint)
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"EPISODE_ID=mvp-static-test",
		"PROMPT=   ",
		"LANGUAGE=ru",
		"DURATION=10s",
		"PLATFORMS=tiktok",
		"ANTHROPIC_API_KEY=static-test-key",
		"SEEDANCE_API_KEY=static-test-key",
		"ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1",
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("entrypoint accepted an empty PROMPT")
	}
	if !strings.Contains(string(output), "PROMPT is required") {
		t.Fatalf("expected PROMPT failure, got: %s", output)
	}
}

func assertNoHighRiskSecrets(t *testing.T, path string) {
	t.Helper()
	summary, err := security.ScanPath(path)
	if err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	if summary.HasHighRiskFindings() {
		t.Fatalf("high-risk secret findings in %s: %+v", path, summary.Findings)
	}
}
