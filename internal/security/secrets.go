package security

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Finding is a deterministic local security scan finding.
type Finding struct {
	Path     string `json:"path"`
	Pattern  string `json:"pattern"`
	Line     int    `json:"line"`
	HighRisk bool   `json:"high_risk"`
}

// ScanSummary summarizes a file tree scan.
type ScanSummary struct {
	Root     string    `json:"root"`
	Findings []Finding `json:"findings"`
}

var secretPatterns = []struct {
	name     string
	pattern  *regexp.Regexp
	highRisk bool
}{
	{name: "generic_api_key_assignment", pattern: regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s*[:=]\s*['\"]?[A-Za-z0-9_\-]{16,}`), highRisk: true},
	{name: "github_token", pattern: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`), highRisk: true},
	{name: "openai_key", pattern: regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`), highRisk: true},
	{name: "private_key_header", pattern: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`), highRisk: true},
}

// ScanPath scans a file or directory for common secret patterns.
func ScanPath(root string) (ScanSummary, error) {
	if root == "" {
		return ScanSummary{}, fmt.Errorf("scan root is required")
	}
	info, err := os.Stat(root)
	if err != nil {
		return ScanSummary{}, err
	}
	summary := ScanSummary{Root: root}
	if !info.IsDir() {
		findings, err := scanFile(root)
		if err != nil {
			return ScanSummary{}, err
		}
		summary.Findings = append(summary.Findings, findings...)
		return summary, nil
	}

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if shouldSkipFile(path) {
			return nil
		}
		findings, err := scanFile(path)
		if err != nil {
			return err
		}
		summary.Findings = append(summary.Findings, findings...)
		return nil
	})
	if err != nil {
		return ScanSummary{}, err
	}
	return summary, nil
}

func scanFile(path string) ([]Finding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) > 2*1024*1024 {
		return nil, nil
	}
	lines := strings.Split(string(data), "\n")
	var findings []Finding
	for i, line := range lines {
		for _, pattern := range secretPatterns {
			if pattern.pattern.MatchString(line) {
				findings = append(findings, Finding{Path: path, Pattern: pattern.name, Line: i + 1, HighRisk: pattern.highRisk})
			}
		}
	}
	return findings, nil
}

func shouldSkipFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".mp4", ".mov", ".webp", ".pdf":
		return true
	default:
		return false
	}
}

// Redact replaces known secret-like values while preserving surrounding text.
func Redact(text string) string {
	redacted := text
	for _, pattern := range secretPatterns {
		redacted = pattern.pattern.ReplaceAllStringFunc(redacted, func(match string) string {
			if strings.Contains(match, "=") {
				parts := strings.SplitN(match, "=", 2)
				return parts[0] + "=[REDACTED]"
			}
			if strings.Contains(match, ":") {
				parts := strings.SplitN(match, ":", 2)
				return parts[0] + ":[REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return redacted
}

// HasHighRiskFindings returns true if any finding is high risk.
func (s ScanSummary) HasHighRiskFindings() bool {
	for _, finding := range s.Findings {
		if finding.HighRisk {
			return true
		}
	}
	return false
}
