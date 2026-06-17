package contenthash

import (
	"strings"
	"testing"
)

type sample struct {
	SchemaVersion string   `json:"schema_version"`
	ArtifactID    string   `json:"artifact_id"`
	ContentHash   string   `json:"content_hash,omitempty"`
	Tags          []string `json:"tags"`
	Count         int      `json:"count"`
}

func TestComputeIsStableAcrossRuns(t *testing.T) {
	s := sample{SchemaVersion: "1.0", ArtifactID: "a-1", Tags: []string{"x", "y"}, Count: 3}
	first, err := Compute(s)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	second, err := Compute(s)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if first != second {
		t.Fatalf("hash not stable: %s != %s", first, second)
	}
	if !strings.HasPrefix(first, Prefix) || len(first) != len(Prefix)+64 {
		t.Fatalf("unexpected hash shape: %s", first)
	}
}

func TestComputeExcludesContentHashField(t *testing.T) {
	without := sample{SchemaVersion: "1.0", ArtifactID: "a-1", Tags: []string{"x"}, Count: 1}
	withHash := without
	withHash.ContentHash = "sha256:" + strings.Repeat("a", 64)

	a, err := Compute(without)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	b, err := Compute(withHash)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if a != b {
		t.Fatalf("content_hash field must be excluded from hash: %s != %s", a, b)
	}
}

func TestComputeIsKeyOrderIndependent(t *testing.T) {
	ordered := map[string]any{"a": 1.0, "b": 2.0, "c": []any{"p", "q"}}
	reordered := map[string]any{"c": []any{"p", "q"}, "b": 2.0, "a": 1.0}
	a, err := Compute(ordered)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	b, err := Compute(reordered)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if a != b {
		t.Fatalf("hash must be key-order independent: %s != %s", a, b)
	}
}

func TestComputeRoundTripThroughHashField(t *testing.T) {
	s := sample{SchemaVersion: "1.0", ArtifactID: "a-1", Tags: []string{"x", "y"}, Count: 7}
	hash, err := Compute(s)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	// Stamp the hash onto the artifact, then verify it round-trips.
	s.ContentHash = hash
	if err := Verify(s, hash); err != nil {
		t.Fatalf("verify after stamping hash: %v", err)
	}
	// Recomputing on the stamped artifact yields the same hash.
	again, err := Compute(s)
	if err != nil {
		t.Fatalf("recompute: %v", err)
	}
	if again != hash {
		t.Fatalf("hash changed after stamping: %s != %s", again, hash)
	}
}

func TestVerifyDetectsTampering(t *testing.T) {
	s := sample{SchemaVersion: "1.0", ArtifactID: "a-1", Tags: []string{"x"}, Count: 1}
	hash, err := Compute(s)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	s.Count = 2
	if err := Verify(s, hash); err == nil {
		t.Fatal("expected verify to fail after mutation")
	}
}

func TestCanonicalizeRemovesHashAndSortsKeys(t *testing.T) {
	in := map[string]any{"content_hash": "sha256:deadbeef", "b": 1.0, "a": 2.0}
	out, err := Canonicalize(in)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}
	got := string(out)
	if strings.Contains(got, "content_hash") {
		t.Fatalf("canonical form must exclude content_hash: %s", got)
	}
	if got != `{"a":2,"b":1}` {
		t.Fatalf("unexpected canonical form: %s", got)
	}
}
