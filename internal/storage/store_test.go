package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLocalStoreWritesAndReadsCanonicalArtifact(t *testing.T) {
	store := newTestStore(t)
	content := validResearchPack()

	ref, err := store.PutArtifact(context.Background(), PutArtifactInput{
		EpisodeID:         "episode-test",
		ArtifactName:      "research_pack.json",
		Content:           content,
		ValidateCanonical: true,
		CreatedAt:         fixedTime(),
		Metadata:          map[string]string{"stage": "research"},
	})
	if err != nil {
		t.Fatalf("put artifact failed: %v", err)
	}
	if ref.ContentHash == "" || !strings.HasPrefix(ref.ContentHash, "sha256:") {
		t.Fatalf("expected sha256 ref, got %+v", ref)
	}
	if ref.SizeBytes != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), ref.SizeBytes)
	}

	got, err := store.GetArtifact(context.Background(), ref)
	if err != nil {
		t.Fatalf("get artifact failed: %v", err)
	}
	if string(got) != string(content) {
		t.Fatal("stored content changed")
	}

	refs, err := store.ListArtifacts(context.Background(), "episode-test")
	if err != nil {
		t.Fatalf("list artifacts failed: %v", err)
	}
	if len(refs) != 1 || refs[0].ArtifactName != "research_pack.json" {
		t.Fatalf("unexpected refs: %+v", refs)
	}
}

func TestLocalStoreRejectsInvalidCanonicalArtifact(t *testing.T) {
	store := newTestStore(t)
	_, err := store.PutArtifact(context.Background(), PutArtifactInput{
		EpisodeID:         "episode-test",
		ArtifactName:      "research_pack.json",
		Content:           []byte(`{"schema_version":"1.0","episode_id":"episode-test","artifact_id":"bad","status":"draft"}`),
		ValidateCanonical: true,
	})
	if !errors.Is(err, ErrValidationFailed) {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestLocalStoreArtifactRefsAreImmutable(t *testing.T) {
	store := newTestStore(t)
	input := PutArtifactInput{
		EpisodeID:         "episode-test",
		ArtifactName:      "research_pack.json",
		Content:           validResearchPack(),
		ValidateCanonical: true,
	}
	first, err := store.PutArtifact(context.Background(), input)
	if err != nil {
		t.Fatalf("put first artifact failed: %v", err)
	}

	second, err := store.PutArtifact(context.Background(), input)
	if err != nil {
		t.Fatalf("idempotent put failed: %v", err)
	}
	if first.ContentHash != second.ContentHash || first.URI != second.URI {
		t.Fatalf("expected idempotent ref, got first=%+v second=%+v", first, second)
	}

	input.Content = validResearchPackWithQuestion("What changed?")
	_, err = store.PutArtifact(context.Background(), input)
	if !errors.Is(err, ErrImmutableConflict) {
		t.Fatalf("expected immutable conflict, got %v", err)
	}
}

func TestLocalStoreEpisodeStateAndArtifactRefs(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	ref, err := store.PutArtifact(ctx, PutArtifactInput{
		EpisodeID:         "episode-test",
		ArtifactName:      "research_pack.json",
		Content:           validResearchPack(),
		ValidateCanonical: true,
	})
	if err != nil {
		t.Fatalf("put artifact failed: %v", err)
	}

	if err := store.SaveEpisode(ctx, EpisodeRecord{
		EpisodeID: "episode-test",
		State:     "research_ready",
		UpdatedAt: fixedTime(),
		Metadata:  map[string]string{"owner": "qa"},
	}); err != nil {
		t.Fatalf("save episode failed: %v", err)
	}
	if err := store.AddArtifactRef(ctx, "episode-test", ref); err != nil {
		t.Fatalf("add artifact ref failed: %v", err)
	}
	if err := store.AddArtifactRef(ctx, "episode-test", ref); err != nil {
		t.Fatalf("duplicate artifact ref should be idempotent: %v", err)
	}

	record, err := store.GetEpisode(ctx, "episode-test")
	if err != nil {
		t.Fatalf("get episode failed: %v", err)
	}
	if record.State != "research_ready" {
		t.Fatalf("expected research_ready, got %s", record.State)
	}
	if len(record.Artifacts) != 1 || record.Artifacts[0].ContentHash != ref.ContentHash {
		t.Fatalf("expected linked artifact ref, got %+v", record.Artifacts)
	}
}

func TestLocalStoreMissingArtifactReturnsNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.GetArtifact(context.Background(), ArtifactRef{
		EpisodeID:    "episode-test",
		ArtifactName: "research_pack.json",
		ContentHash:  "sha256:missing",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestLocalStoreRejectsPathTraversal(t *testing.T) {
	store := newTestStore(t)
	_, err := store.PutArtifact(context.Background(), PutArtifactInput{
		EpisodeID:    "../episode-test",
		ArtifactName: "research_pack.json",
		Content:      validResearchPack(),
	})
	if err == nil {
		t.Fatal("expected unsafe episode id to fail")
	}
	_, err = store.PutArtifact(context.Background(), PutArtifactInput{
		EpisodeID:    "episode-test",
		ArtifactName: "../research_pack.json",
		Content:      validResearchPack(),
	})
	if err == nil {
		t.Fatal("expected unsafe artifact name to fail")
	}
}

func newTestStore(t *testing.T) *LocalStore {
	t.Helper()
	store, err := NewLocalStore(t.TempDir())
	if err != nil {
		t.Fatalf("new local store failed: %v", err)
	}
	store.now = fixedTime
	return store
}

func validResearchPack() []byte {
	return validResearchPackWithQuestion("How does storage preserve artifacts?")
}

func validResearchPackWithQuestion(question string) []byte {
	return []byte(`{
  "schema_version": "1.0",
  "episode_id": "episode-test",
  "artifact_id": "research-test",
  "created_at": "2026-05-27T00:00:00Z",
  "created_by": "system:test",
  "status": "draft",
  "core_question": "` + question + `",
  "learning_objectives": ["Explain content-addressed storage."],
  "sources": [
    {
      "source_id": "source-test",
      "title": "Test source",
      "uri": "https://example.com/source",
      "type": "official_docs",
      "trust_level": "primary"
    }
  ]
}`)
}

func fixedTime() time.Time {
	return time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
}
