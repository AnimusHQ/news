package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanPathDetectsFakeToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.env")
	if err := os.WriteFile(path, []byte("API_KEY=abcdefghijklmnop1234567890\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	summary, err := ScanPath(dir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(summary.Findings) == 0 {
		t.Fatal("expected fake token finding")
	}
	if !summary.HasHighRiskFindings() {
		t.Fatal("expected high-risk finding")
	}
}

func TestScanPathIgnoresCleanFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte("hello world\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	summary, err := ScanPath(dir)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if len(summary.Findings) != 0 {
		t.Fatalf("expected no findings, got %+v", summary.Findings)
	}
}

func TestRedactPreservesKeyName(t *testing.T) {
	redacted := Redact("password=abcdefghijklmnop1234567890")
	if !strings.Contains(redacted, "password=[REDACTED]") {
		t.Fatalf("unexpected redaction: %s", redacted)
	}
	if strings.Contains(redacted, "abcdefghijklmnop") {
		t.Fatalf("secret value was not redacted: %s", redacted)
	}
}

func TestScanPathRequiresRoot(t *testing.T) {
	_, err := ScanPath("")
	if err == nil {
		t.Fatal("expected empty root to fail")
	}
}
