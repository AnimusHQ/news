package storage

import (
	"context"
	"fmt"
)

// Backend groups the storage interfaces used by the application.
type Backend struct {
	Mode      BackendMode
	Artifacts ArtifactStore
	Episodes  EpisodeRepository
}

// BackendOptions contains externally constructed clients for non-local modes.
type BackendOptions struct {
	PostgresS3 *ExternalBackendClients
}

// ExternalBackendClients lets deployment code provide concrete Postgres and
// S3-compatible implementations without making local tests depend on them.
type ExternalBackendClients struct {
	Artifacts ArtifactStore
	Episodes  EpisodeRepository
}

// NewBackend creates a validated storage backend. Local mode is fully
// repository-local; postgres_s3 mode fails closed unless clients are injected.
func NewBackend(ctx context.Context, config BackendConfig, options BackendOptions) (Backend, error) {
	if err := ctx.Err(); err != nil {
		return Backend{}, err
	}
	if err := config.Validate(); err != nil {
		return Backend{}, err
	}
	switch config.Mode {
	case BackendLocal:
		store, err := NewLocalStore(config.Local.Root)
		if err != nil {
			return Backend{}, err
		}
		return Backend{Mode: BackendLocal, Artifacts: store, Episodes: store}, nil
	case BackendPostgresS3:
		if options.PostgresS3 == nil || options.PostgresS3.Artifacts == nil || options.PostgresS3.Episodes == nil {
			return Backend{}, fmt.Errorf("%s backend requires injected artifact and episode clients", BackendPostgresS3)
		}
		return Backend{Mode: BackendPostgresS3, Artifacts: options.PostgresS3.Artifacts, Episodes: options.PostgresS3.Episodes}, nil
	default:
		return Backend{}, fmt.Errorf("unsupported storage backend mode: %s", config.Mode)
	}
}
