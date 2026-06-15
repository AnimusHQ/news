package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SecretRefResolver resolves secret manager references at runtime. Production
// deployments can inject an implementation without making the storage package
// depend on a concrete secret-manager SDK.
type SecretRefResolver interface {
	ResolveSecretRef(ctx context.Context, name string) (CredentialValue, error)
}

// CredentialResolver resolves credential references into in-memory values.
type CredentialResolver struct {
	LookupEnv      func(string) (string, bool)
	ReadFile       func(string) ([]byte, error)
	SecretResolver SecretRefResolver
}

// CredentialValue stores sensitive material. It intentionally redacts common
// string and JSON representations to avoid accidental logs or artifacts.
type CredentialValue struct {
	value string
}

// NewCredentialValue creates a redacting credential wrapper.
func NewCredentialValue(value string) (CredentialValue, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return CredentialValue{}, fmt.Errorf("resolved credential is empty")
	}
	return CredentialValue{value: value}, nil
}

// Value returns the raw credential. Callers must not log or serialize it.
func (v CredentialValue) Value() string {
	return v.value
}

func (v CredentialValue) String() string {
	return "[REDACTED]"
}

func (v CredentialValue) GoString() string {
	return "[REDACTED]"
}

func (v CredentialValue) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

// Resolve resolves env:, file:, or secretref: references at runtime.
func (r CredentialResolver) Resolve(ctx context.Context, ref string) (CredentialValue, error) {
	if err := ctx.Err(); err != nil {
		return CredentialValue{}, err
	}
	if err := requireCredentialRef("credential_ref", ref); err != nil {
		return CredentialValue{}, err
	}
	ref = strings.TrimSpace(ref)
	switch {
	case strings.HasPrefix(ref, "env:"):
		return r.resolveEnv(strings.TrimSpace(strings.TrimPrefix(ref, "env:")))
	case strings.HasPrefix(ref, "file:"):
		return r.resolveFile(strings.TrimSpace(strings.TrimPrefix(ref, "file:")))
	case strings.HasPrefix(ref, "secretref:"):
		return r.resolveSecretRef(ctx, strings.TrimSpace(strings.TrimPrefix(ref, "secretref:")))
	default:
		return CredentialValue{}, fmt.Errorf("unsupported credential reference")
	}
}

func (r CredentialResolver) resolveEnv(name string) (CredentialValue, error) {
	if !safeEnvName(name) {
		return CredentialValue{}, fmt.Errorf("env credential reference name is invalid")
	}
	lookup := r.LookupEnv
	if lookup == nil {
		lookup = os.LookupEnv
	}
	value, ok := lookup(name)
	if !ok {
		return CredentialValue{}, fmt.Errorf("env credential reference is not set")
	}
	return NewCredentialValue(value)
}

func (r CredentialResolver) resolveFile(path string) (CredentialValue, error) {
	if strings.TrimSpace(path) == "" {
		return CredentialValue{}, fmt.Errorf("file credential reference path is required")
	}
	readFile := r.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}
	data, err := readFile(path)
	if err != nil {
		return CredentialValue{}, fmt.Errorf("read file credential reference: %w", err)
	}
	return NewCredentialValue(string(data))
}

func (r CredentialResolver) resolveSecretRef(ctx context.Context, name string) (CredentialValue, error) {
	if strings.TrimSpace(name) == "" {
		return CredentialValue{}, fmt.Errorf("secretref credential reference name is required")
	}
	if r.SecretResolver == nil {
		return CredentialValue{}, fmt.Errorf("secretref credential reference requires an injected resolver")
	}
	value, err := r.SecretResolver.ResolveSecretRef(ctx, name)
	if err != nil {
		return CredentialValue{}, fmt.Errorf("resolve secretref credential reference: %w", err)
	}
	if strings.TrimSpace(value.Value()) == "" {
		return CredentialValue{}, fmt.Errorf("resolved credential is empty")
	}
	return value, nil
}

func safeEnvName(value string) bool {
	if value == "" {
		return false
	}
	for i, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '_' {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
