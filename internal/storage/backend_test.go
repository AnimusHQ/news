package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestNewBackendCreatesUsableLocalBackend(t *testing.T) {
	backend, err := NewBackend(context.Background(), DefaultLocalConfig(t.TempDir()), BackendOptions{})
	if err != nil {
		t.Fatalf("new local backend: %v", err)
	}
	if backend.Mode != BackendLocal || backend.Artifacts == nil || backend.Episodes == nil {
		t.Fatalf("unexpected backend: %+v", backend)
	}

	if err := backend.Episodes.SaveEpisode(context.Background(), EpisodeRecord{EpisodeID: "episode-test", State: "backlog"}); err != nil {
		t.Fatalf("save episode through backend: %v", err)
	}
	record, err := backend.Episodes.GetEpisode(context.Background(), "episode-test")
	if err != nil {
		t.Fatalf("get episode through backend: %v", err)
	}
	if record.State != "backlog" {
		t.Fatalf("expected backlog, got %s", record.State)
	}
}

func TestNewBackendFailsClosedForPostgresS3WithoutInjectedClients(t *testing.T) {
	_, err := NewBackend(context.Background(), validBackendConfig(), BackendOptions{})
	if err == nil {
		t.Fatal("expected postgres_s3 backend without clients to fail")
	}
	if !strings.Contains(err.Error(), "requires injected") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewBackendAcceptsInjectedPostgresS3Clients(t *testing.T) {
	fake := &fakeBackendClient{}
	backend, err := NewBackend(context.Background(), validBackendConfig(), BackendOptions{
		PostgresS3: &ExternalBackendClients{
			Artifacts: fake,
			Episodes:  fake,
		},
	})
	if err != nil {
		t.Fatalf("new postgres_s3 backend with injected clients: %v", err)
	}
	if backend.Mode != BackendPostgresS3 || backend.Artifacts != fake || backend.Episodes != fake {
		t.Fatalf("unexpected backend: %+v", backend)
	}
}

type fakeBackendClient struct{}

func (fakeBackendClient) PutArtifact(context.Context, PutArtifactInput) (ArtifactRef, error) {
	return ArtifactRef{}, nil
}

func (fakeBackendClient) GetArtifact(context.Context, ArtifactRef) ([]byte, error) {
	return nil, ErrNotFound
}

func (fakeBackendClient) ListArtifacts(context.Context, string) ([]ArtifactRef, error) {
	return nil, nil
}

func (fakeBackendClient) SaveEpisode(context.Context, EpisodeRecord) error {
	return nil
}

func (fakeBackendClient) GetEpisode(context.Context, string) (EpisodeRecord, error) {
	return EpisodeRecord{}, errors.New("not implemented")
}

func (fakeBackendClient) AddArtifactRef(context.Context, string, ArtifactRef) error {
	return nil
}
