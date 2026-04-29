package artifacts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateEpisodeDirectoryPassesForCompleteBundle(t *testing.T) {
	dir := t.TempDir()
	for _, name := range RequiredEpisodeFiles {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("placeholder"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := ValidateEpisodeDirectory(dir); err != nil {
		t.Fatalf("expected complete bundle to validate: %v", err)
	}
}

func TestValidateEpisodeDirectoryFailsForMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	for _, name := range RequiredEpisodeFiles[1:] {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("placeholder"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	if err := ValidateEpisodeDirectory(dir); err == nil {
		t.Fatal("expected missing artifact to fail validation")
	}
}

func TestValidateEpisodeDirectoryFailsForFilePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(path, []byte("placeholder"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if err := ValidateEpisodeDirectory(path); err == nil {
		t.Fatal("expected file path to fail validation")
	}
}
