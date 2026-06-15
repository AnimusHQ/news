package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCredentialResolverResolvesEnvironmentReference(t *testing.T) {
	t.Setenv("ANIMUS_TEST_DSN", "postgres://localhost/animus")

	value, err := (CredentialResolver{}).Resolve(context.Background(), "env:ANIMUS_TEST_DSN")
	if err != nil {
		t.Fatalf("resolve env credential: %v", err)
	}
	if value.Value() != "postgres://localhost/animus" {
		t.Fatalf("unexpected credential value")
	}
}

func TestCredentialResolverResolvesFileReference(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credential.txt")
	if err := os.WriteFile(path, []byte("s3.local.internal\n"), 0o600); err != nil {
		t.Fatalf("write credential file: %v", err)
	}

	value, err := (CredentialResolver{}).Resolve(context.Background(), "file:"+path)
	if err != nil {
		t.Fatalf("resolve file credential: %v", err)
	}
	if value.Value() != "s3.local.internal" {
		t.Fatalf("unexpected credential value")
	}
}

func TestCredentialResolverRequiresInjectedSecretResolver(t *testing.T) {
	_, err := (CredentialResolver{}).Resolve(context.Background(), "secretref:animus/storage")
	if err == nil {
		t.Fatal("expected missing secret resolver to fail")
	}
	if !strings.Contains(err.Error(), "injected resolver") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCredentialResolverUsesInjectedSecretResolver(t *testing.T) {
	value, err := (CredentialResolver{SecretResolver: staticSecretResolver{value: "runtime-only-value"}}).Resolve(context.Background(), "secretref:animus/storage")
	if err != nil {
		t.Fatalf("resolve secretref credential: %v", err)
	}
	if value.Value() != "runtime-only-value" {
		t.Fatalf("unexpected credential value")
	}
}

func TestCredentialResolverRejectsInvalidReferences(t *testing.T) {
	tests := []string{
		"ANIMUS_TEST_DSN",
		"env:",
		"env:ANIMUS TEST DSN",
		"file:",
	}
	for _, ref := range tests {
		t.Run(ref, func(t *testing.T) {
			_, err := (CredentialResolver{}).Resolve(context.Background(), ref)
			if err == nil {
				t.Fatal("expected invalid reference to fail")
			}
		})
	}
}

func TestCredentialValueRedactsStringGoStringAndJSON(t *testing.T) {
	value, err := NewCredentialValue("runtime-only-value")
	if err != nil {
		t.Fatalf("new credential value: %v", err)
	}
	if fmt.Sprint(value) != "[REDACTED]" {
		t.Fatalf("String leaked credential")
	}
	if fmt.Sprintf("%#v", value) != "[REDACTED]" {
		t.Fatalf("GoString leaked credential")
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal credential: %v", err)
	}
	if string(data) != `"[REDACTED]"` {
		t.Fatalf("JSON leaked credential: %s", data)
	}
}

type staticSecretResolver struct {
	value string
}

func (r staticSecretResolver) ResolveSecretRef(context.Context, string) (CredentialValue, error) {
	return NewCredentialValue(r.value)
}
