// Package contenthash computes a deterministic, order-independent content hash
// for short-form artifacts.
//
// The hash is sha256 over a canonical JSON encoding of the artifact (object keys
// sorted recursively) with the top-level "content_hash" field excluded, so that
// an artifact hashes the same whether or not it already carries its own hash and
// regardless of map/struct field ordering. See docs/adr/0002.
package contenthash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// HashField is the envelope field excluded from the hash computation.
const HashField = "content_hash"

// Prefix is the algorithm prefix used on every emitted hash.
const Prefix = "sha256:"

// Compute returns the deterministic content hash of v as "sha256:<hex>".
//
// v is marshaled to JSON, decoded into a generic value, the top-level
// content_hash field (if any) is removed, and the result is canonicalized with
// sorted object keys before hashing.
func Compute(v any) (string, error) {
	generic, err := toGeneric(v)
	if err != nil {
		return "", err
	}
	if obj, ok := generic.(map[string]any); ok {
		delete(obj, HashField)
	}
	var buf bytes.Buffer
	if err := canonical(generic, &buf); err != nil {
		return "", err
	}
	sum := sha256.Sum256(buf.Bytes())
	return Prefix + hex.EncodeToString(sum[:]), nil
}

// Canonicalize returns the canonical JSON encoding (sorted keys) of v with the
// content_hash field removed. Exposed for tests and debugging.
func Canonicalize(v any) ([]byte, error) {
	generic, err := toGeneric(v)
	if err != nil {
		return nil, err
	}
	if obj, ok := generic.(map[string]any); ok {
		delete(obj, HashField)
	}
	var buf bytes.Buffer
	if err := canonical(generic, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Verify recomputes the content hash of v and compares it to expected.
func Verify(v any, expected string) error {
	got, err := Compute(v)
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("content hash mismatch: expected %s got %s", expected, got)
	}
	return nil
}

func toGeneric(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var generic any
	if err := json.Unmarshal(data, &generic); err != nil {
		return nil, err
	}
	return generic, nil
}

func canonical(v any, buf *bytes.Buffer) error {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64, string:
		encoded, err := json.Marshal(t)
		if err != nil {
			return err
		}
		buf.Write(encoded)
	case []any:
		buf.WriteByte('[')
		for i, elem := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := canonical(elem, buf); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for key := range t {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			encodedKey, err := json.Marshal(key)
			if err != nil {
				return err
			}
			buf.Write(encodedKey)
			buf.WriteByte(':')
			if err := canonical(t[key], buf); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("unsupported canonical type %T", v)
	}
	return nil
}
