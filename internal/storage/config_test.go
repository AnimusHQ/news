package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultLocalConfigValidates(t *testing.T) {
	config := DefaultLocalConfig(t.TempDir())
	if err := config.Validate(); err != nil {
		t.Fatalf("expected local config to validate: %v", err)
	}
}

func TestPostgresS3ConfigValidatesWithReferences(t *testing.T) {
	config := validBackendConfig()
	if err := config.Validate(); err != nil {
		t.Fatalf("expected postgres/s3 config to validate: %v", err)
	}
}

func TestPostgresS3ConfigRejectsRawSecretValues(t *testing.T) {
	config := validBackendConfig()
	config.ObjectStore.SecretKeyRef = "secret=" + strings.Repeat("a", 20)
	err := config.Validate()
	if err == nil {
		t.Fatal("expected raw secret-looking value to fail")
	}
	if !strings.Contains(err.Error(), "credential reference") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresS3ConfigRequiresReferencePrefixes(t *testing.T) {
	config := validBackendConfig()
	config.Postgres.DSNRef = "POSTGRES_DSN"
	err := config.Validate()
	if err == nil {
		t.Fatal("expected unprefixed reference to fail")
	}
	if !strings.Contains(err.Error(), "env:") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadBackendConfigFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "storage.yaml")
	if err := os.WriteFile(path, []byte(`mode: postgres_s3
postgres:
  dsn_ref: env:ANIMUS_POSTGRES_DSN
  require_tls: true
  max_open_conns: 8
  migration_table: animus_schema_migrations
object_store:
  endpoint_ref: env:ANIMUS_S3_ENDPOINT
  bucket: animus-artifacts
  region: local
  access_key_ref: env:ANIMUS_S3_ACCESS_KEY
  secret_key_ref: env:ANIMUS_S3_SECRET_KEY
  force_path_style: true
  server_side_encryption: true
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	config, err := LoadBackendConfig(path)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if config.Mode != BackendPostgresS3 {
		t.Fatalf("unexpected mode: %s", config.Mode)
	}
	if !config.Postgres.RequireTLS {
		t.Fatal("expected require_tls to decode")
	}
}

func TestMigrationPlanContainsExpectedTablesAndNoCredentialMaterial(t *testing.T) {
	plan, err := MigrationPlan(validBackendConfig())
	if err != nil {
		t.Fatalf("migration plan failed: %v", err)
	}
	if len(plan) != 3 {
		t.Fatalf("expected 3 migration steps, got %d", len(plan))
	}
	joined := ""
	for _, step := range plan {
		joined += step.SQL + "\n"
	}
	for _, text := range []string{"episode_state", "artifact_refs", "animus_schema_migrations"} {
		if !strings.Contains(joined, text) {
			t.Fatalf("expected migration plan to contain %s: %s", text, joined)
		}
	}
	if strings.Contains(joined, "ANIMUS_") || strings.Contains(joined, "env:") || strings.Contains(joined, "secretref:") {
		t.Fatalf("migration plan should not contain credential references: %s", joined)
	}
}

func TestMigrationPlanRequiresPostgresS3Mode(t *testing.T) {
	_, err := MigrationPlan(DefaultLocalConfig(t.TempDir()))
	if err == nil {
		t.Fatal("expected local config migration plan to fail")
	}
}

func validBackendConfig() BackendConfig {
	return BackendConfig{
		Mode: BackendPostgresS3,
		Postgres: PostgresConfig{
			DSNRef:         "env:ANIMUS_POSTGRES_DSN",
			RequireTLS:     true,
			MaxOpenConns:   8,
			MigrationTable: "animus_schema_migrations",
		},
		ObjectStore: ObjectStoreConfig{
			EndpointRef:          "env:ANIMUS_S3_ENDPOINT",
			Bucket:               "animus-artifacts",
			Region:               "local",
			AccessKeyRef:         "env:ANIMUS_S3_ACCESS_KEY",
			SecretKeyRef:         "env:ANIMUS_S3_SECRET_KEY",
			ForcePathStyle:       true,
			ServerSideEncryption: true,
		},
	}
}
