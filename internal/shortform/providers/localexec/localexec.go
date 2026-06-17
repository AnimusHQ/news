// Package localexec contains shared safety helpers for local execution adapters.
// It intentionally does not run commands itself; provider adapters own their
// exec.CommandContext calls behind activity boundaries.
package localexec

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const hashPrefix = "sha256:"

var safeSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// SafeSegment validates a single filesystem path segment controlled by repo
// code, such as an episode id or platform name.
func SafeSegment(value, field string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	if strings.ContainsAny(value, `/\`) || value == "." || value == ".." {
		return fmt.Errorf("%s must be a single path segment", field)
	}
	if !safeSegmentPattern.MatchString(value) {
		return fmt.Errorf("%s contains unsupported characters", field)
	}
	return nil
}

// ResolveUnder resolves candidate under root and rejects traversal or escaping.
// Absolute candidates are accepted only when they are still inside root.
func ResolveUnder(root, candidate, field string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("%s root must be configured", field)
	}
	if candidate == "" {
		return "", fmt.Errorf("%s path must not be empty", field)
	}
	if strings.ContainsRune(candidate, 0) {
		return "", fmt.Errorf("%s path contains NUL", field)
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("%s root: %w", field, err)
	}
	var target string
	if filepath.IsAbs(candidate) {
		target = filepath.Clean(candidate)
	} else {
		target = filepath.Join(rootAbs, candidate)
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("%s path: %w", field, err)
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return "", fmt.Errorf("%s path containment: %w", field, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%s path escapes configured root", field)
	}
	return targetAbs, nil
}

// ExistingFileUnder resolves and stat-checks a file under root.
func ExistingFileUnder(root, candidate, field string) (string, error) {
	path, err := ResolveUnder(root, candidate, field)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s path is not readable: %w", field, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s path must be a file", field)
	}
	return path, nil
}

// ExistingDirUnder resolves and stat-checks a directory under root.
func ExistingDirUnder(root, candidate, field string) (string, error) {
	path, err := ResolveUnder(root, candidate, field)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("%s path is not readable: %w", field, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s path must be a directory", field)
	}
	return path, nil
}

// EnsureOutputDir creates an output directory below root using safe path
// segments only.
func EnsureOutputDir(root string, segments ...string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("output root must be configured")
	}
	cleanSegments := make([]string, 0, len(segments))
	for i, segment := range segments {
		if err := SafeSegment(segment, fmt.Sprintf("output segment %d", i)); err != nil {
			return "", err
		}
		cleanSegments = append(cleanSegments, segment)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("output root: %w", err)
	}
	path := filepath.Join(append([]string{rootAbs}, cleanSegments...)...)
	rel, err := filepath.Rel(rootAbs, path)
	if err != nil {
		return "", fmt.Errorf("output path containment: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("output path escapes configured root")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}
	return path, nil
}

// FileSHA256 returns a sha256: content hash over raw file bytes.
func FileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hashPrefix + hex.EncodeToString(sum[:]), nil
}

// Redact replaces configured sensitive values and caps noisy command output.
func Redact(text string, secrets ...string) string {
	out := text
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[REDACTED]")
	}
	const max = 4096
	if len(out) > max {
		return out[:max] + "...[truncated]"
	}
	return out
}
