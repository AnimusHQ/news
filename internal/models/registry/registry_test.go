package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileValidRegistry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	write(t, path, `models:
  - id: model-a
    provider: local-mock
    version: v1
    status: active
    privacy_tier: local_only
    modalities: [text]
    capabilities: [technical_verification]
    quality_score: 0.9
`)

	records, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load registry failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestLoadFileRejectsDuplicateID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	write(t, path, `models:
  - id: model-a
    provider: local-mock
    version: v1
    status: active
    privacy_tier: local_only
    modalities: [text]
    capabilities: [technical_verification]
  - id: model-a
    provider: local-mock
    version: v2
    status: active
    privacy_tier: local_only
    modalities: [text]
    capabilities: [technical_verification]
`)

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected duplicate model ID to fail")
	}
}

func TestLoadFileRejectsInvalidPrivacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	write(t, path, `models:
  - id: model-a
    provider: local-mock
    version: v1
    status: active
    privacy_tier: unknown
    modalities: [text]
    capabilities: [technical_verification]
`)

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected invalid privacy tier to fail")
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
