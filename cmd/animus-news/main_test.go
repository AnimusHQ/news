package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseValidateArgsPlainPath(t *testing.T) {
	jsonOutput, path, err := parseValidateArgs([]string{"artifact.json"})
	if err != nil {
		t.Fatalf("parse args failed: %v", err)
	}
	if jsonOutput {
		t.Fatal("expected plain path not to request JSON output")
	}
	if path != "artifact.json" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestParseValidateArgsJSONPath(t *testing.T) {
	jsonOutput, path, err := parseValidateArgs([]string{"--json", "artifact.json"})
	if err != nil {
		t.Fatalf("parse args failed: %v", err)
	}
	if !jsonOutput {
		t.Fatal("expected JSON output flag")
	}
	if path != "artifact.json" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestParseValidateArgsRejectsInvalidShape(t *testing.T) {
	_, _, err := parseValidateArgs([]string{"--json"})
	if err == nil {
		t.Fatal("expected invalid validate args to fail")
	}
}

func TestRunValidateCommandPlainOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "research_pack.json")
	writeCLIArtifact(t, path, validResearchPack())

	stdout := captureStdout(t, func() {
		if err := run([]string{"animus-news", "validate", path}); err != nil {
			t.Fatalf("run validate failed: %v", err)
		}
	})
	if !strings.Contains(stdout, "valid: "+path) {
		t.Fatalf("expected valid output, got %q", stdout)
	}
}

func TestRunValidateCommandJSONOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "research_pack.json")
	writeCLIArtifact(t, path, validResearchPack())

	stdout := captureStdout(t, func() {
		if err := run([]string{"animus-news", "validate", "--json", path}); err != nil {
			t.Fatalf("run validate --json failed: %v", err)
		}
	})
	if !strings.Contains(stdout, `"valid": true`) {
		t.Fatalf("expected JSON validation output, got %q", stdout)
	}
}

func TestRunValidateCommandFailsInvalidArtifact(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "research_pack.json")
	writeCLIArtifact(t, path, `{"schema_version":"1.0"}`)

	err := run([]string{"animus-news", "validate", path})
	if err == nil {
		t.Fatal("expected invalid artifact to fail")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer
	defer func() { os.Stdout = old }()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return buf.String()
}

func writeCLIArtifact(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
}

func validResearchPack() string {
	return `{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "created_at": "2026-04-29T00:00:00Z",
  "created_by": "test",
  "status": "draft",
  "core_question": "How does validation work?",
  "learning_objectives": ["Explain validation."],
  "sources": [
    {
      "source_id": "source-test",
      "title": "Test source",
      "uri": "https://example.com/source",
      "type": "official_docs",
      "trust_level": "primary"
    }
  ]
}`
}
