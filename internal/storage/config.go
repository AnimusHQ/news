package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/AnimusHQ/news/internal/security"
	"gopkg.in/yaml.v3"
)

type BackendMode string

const (
	BackendLocal      BackendMode = "local"
	BackendPostgresS3 BackendMode = "postgres_s3"
)

// BackendConfig describes storage backend wiring. It contains only references
// to credentials, never credential values.
type BackendConfig struct {
	Mode        BackendMode        `json:"mode" yaml:"mode"`
	Local       LocalBackendConfig `json:"local,omitempty" yaml:"local,omitempty"`
	Postgres    PostgresConfig     `json:"postgres,omitempty" yaml:"postgres,omitempty"`
	ObjectStore ObjectStoreConfig  `json:"object_store,omitempty" yaml:"object_store,omitempty"`
}

type LocalBackendConfig struct {
	Root string `json:"root" yaml:"root"`
}

type PostgresConfig struct {
	DSNRef         string `json:"dsn_ref" yaml:"dsn_ref"`
	RequireTLS     bool   `json:"require_tls" yaml:"require_tls"`
	MaxOpenConns   int    `json:"max_open_conns,omitempty" yaml:"max_open_conns,omitempty"`
	MigrationTable string `json:"migration_table,omitempty" yaml:"migration_table,omitempty"`
}

type ObjectStoreConfig struct {
	EndpointRef          string `json:"endpoint_ref" yaml:"endpoint_ref"`
	Bucket               string `json:"bucket" yaml:"bucket"`
	Region               string `json:"region,omitempty" yaml:"region,omitempty"`
	AccessKeyRef         string `json:"access_key_ref" yaml:"access_key_ref"`
	SecretKeyRef         string `json:"secret_key_ref" yaml:"secret_key_ref"`
	ForcePathStyle       bool   `json:"force_path_style,omitempty" yaml:"force_path_style,omitempty"`
	ServerSideEncryption bool   `json:"server_side_encryption,omitempty" yaml:"server_side_encryption,omitempty"`
}

type MigrationStep struct {
	ID          string `json:"id" yaml:"id"`
	Description string `json:"description" yaml:"description"`
	SQL         string `json:"sql" yaml:"sql"`
}

func DefaultLocalConfig(root string) BackendConfig {
	return BackendConfig{
		Mode:  BackendLocal,
		Local: LocalBackendConfig{Root: strings.TrimSpace(root)},
	}
}

func LoadBackendConfig(path string) (BackendConfig, error) {
	if strings.TrimSpace(path) == "" {
		return BackendConfig{}, fmt.Errorf("config path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return BackendConfig{}, err
	}
	var config BackendConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return BackendConfig{}, err
	}
	if err := config.Validate(); err != nil {
		return BackendConfig{}, err
	}
	return config, nil
}

func (c BackendConfig) Validate() error {
	switch c.Mode {
	case BackendLocal:
		if strings.TrimSpace(c.Local.Root) == "" {
			return fmt.Errorf("local.root is required")
		}
		return nil
	case BackendPostgresS3:
		return c.validatePostgresS3()
	default:
		return fmt.Errorf("unsupported storage backend mode: %s", c.Mode)
	}
}

func (c BackendConfig) validatePostgresS3() error {
	if err := requireCredentialRef("postgres.dsn_ref", c.Postgres.DSNRef); err != nil {
		return err
	}
	if c.Postgres.MaxOpenConns < 0 {
		return fmt.Errorf("postgres.max_open_conns must not be negative")
	}
	if c.Postgres.MigrationTable != "" && !safeIdentifier(c.Postgres.MigrationTable) {
		return fmt.Errorf("postgres.migration_table must be a simple identifier")
	}
	if err := requireCredentialRef("object_store.endpoint_ref", c.ObjectStore.EndpointRef); err != nil {
		return err
	}
	if strings.TrimSpace(c.ObjectStore.Bucket) == "" {
		return fmt.Errorf("object_store.bucket is required")
	}
	if strings.Contains(c.ObjectStore.Bucket, "/") || strings.Contains(c.ObjectStore.Bucket, "\\") {
		return fmt.Errorf("object_store.bucket must not contain path separators")
	}
	if err := requireCredentialRef("object_store.access_key_ref", c.ObjectStore.AccessKeyRef); err != nil {
		return err
	}
	if err := requireCredentialRef("object_store.secret_key_ref", c.ObjectStore.SecretKeyRef); err != nil {
		return err
	}
	return nil
}

func MigrationPlan(config BackendConfig) ([]MigrationStep, error) {
	if config.Mode != BackendPostgresS3 {
		return nil, fmt.Errorf("migration plan is only defined for %s mode", BackendPostgresS3)
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	migrationTable := strings.TrimSpace(config.Postgres.MigrationTable)
	if migrationTable == "" {
		migrationTable = "animus_schema_migrations"
	}
	return []MigrationStep{
		{
			ID:          "001_schema_migrations",
			Description: "Track applied storage schema migrations.",
			SQL:         "CREATE TABLE IF NOT EXISTS " + migrationTable + " (id TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now());",
		},
		{
			ID:          "002_episode_state",
			Description: "Persist durable episode state records.",
			SQL:         "CREATE TABLE IF NOT EXISTS episode_state (episode_id TEXT PRIMARY KEY, state TEXT NOT NULL, updated_at TIMESTAMPTZ NOT NULL, metadata JSONB NOT NULL DEFAULT '{}'::jsonb);",
		},
		{
			ID:          "003_artifact_refs",
			Description: "Persist immutable content-addressed artifact references.",
			SQL:         "CREATE TABLE IF NOT EXISTS artifact_refs (episode_id TEXT NOT NULL, artifact_name TEXT NOT NULL, content_hash TEXT NOT NULL, size_bytes BIGINT NOT NULL, uri TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL, metadata JSONB NOT NULL DEFAULT '{}'::jsonb, PRIMARY KEY (episode_id, artifact_name), UNIQUE (content_hash, artifact_name));",
		},
	}, nil
}

func requireCredentialRef(field string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if security.Redact(value) != value {
		return fmt.Errorf("%s must be a credential reference, not a credential value", field)
	}
	if strings.HasPrefix(value, "env:") || strings.HasPrefix(value, "secretref:") || strings.HasPrefix(value, "file:") {
		return nil
	}
	return fmt.Errorf("%s must use env:, secretref:, or file: reference prefix", field)
}

func safeIdentifier(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}
