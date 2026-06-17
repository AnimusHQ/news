// Package claude implements the L2 Claude API review provider for the
// short-form pilot. It performs the two existing review tasks — script_review
// and final_qa — through the Anthropic Messages API and returns the model's
// structured JSON verdict.
//
// The provider is transport plus strict JSON validation only. It is never an
// approval authority: the pilot owns gate decisions, artifact validation, and
// the script-hash binding (see internal/shortform/pilot/review.go). It fails
// closed when ANTHROPIC_API_KEY is unset, redacts the key from every error, and
// performs no network I/O in tests when ANIMUS_CLAUDE_BASE_URL points at a fake
// server.
//
// The client uses the Go standard library rather than the official Anthropic
// SDK on purpose: the repository enforces a no-new-dependencies invariant
// (docs/adr/0003) and verification must run fully offline. The wire contract is
// the documented Messages API (POST /v1/messages, x-api-key +
// anthropic-version: 2023-06-01, adaptive thinking, model default
// claude-opus-4-8). See docs/adr/0012 and docs/providers/CLAUDE_API.md.
package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

const (
	defaultModel        = "claude-opus-4-8"
	defaultBaseURL      = "https://api.anthropic.com"
	anthropicVersion    = "2023-06-01"
	defaultMaxTokens    = 4096
	defaultTimeout      = 60 * time.Second
	defaultRetryBackoff = 500 * time.Millisecond
	maxRetries          = 2
)

// Config configures the Claude review client. Only APIKey is required; the rest
// fall back to safe defaults.
type Config struct {
	APIKey       string
	Model        string
	BaseURL      string
	Timeout      time.Duration
	MaxTokens    int
	RetryBackoff time.Duration
	HTTPClient   *http.Client
}

// Client performs script and final-QA reviews against the Messages API.
type Client struct {
	cfg Config
}

// New builds a client, failing closed when no API key is configured.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("claude api review requires an API key (set ANTHROPIC_API_KEY)")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = defaultModel
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = defaultRetryBackoff
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}
	return &Client{cfg: cfg}, nil
}

// FromEnv builds a client from environment configuration. It fails closed when
// ANTHROPIC_API_KEY is unset so generate-real never silently degrades.
func FromEnv() (*Client, error) {
	key := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY"))
	if key == "" {
		return nil, fmt.Errorf("claude api review requires ANTHROPIC_API_KEY; set it or use --claude-review manual")
	}
	return New(Config{
		APIKey:    key,
		Model:     envOr("ANIMUS_CLAUDE_MODEL", defaultModel),
		BaseURL:   envOr("ANIMUS_CLAUDE_BASE_URL", defaultBaseURL),
		Timeout:   envDuration("ANIMUS_CLAUDE_TIMEOUT", defaultTimeout),
		MaxTokens: envInt("ANIMUS_CLAUDE_MAX_TOKENS", defaultMaxTokens),
	})
}

// Review performs a script or final review and returns the validated JSON
// object the model produced. The kind is "script" or "final". episodeID is the
// expected episode id; a mismatching response is rejected. The prompt is the
// rendered review request (the pilot's *_request.md content).
func (c *Client) Review(ctx context.Context, kind, episodeID, prompt string) (json.RawMessage, error) {
	var system string
	var requiredKeys []string
	switch kind {
	case "script":
		system = scriptSystemPrompt
		requiredKeys = []string{"schema_version", "episode_id", "verdict", "production_readiness", "blocking_issues", "suggested_revisions", "can_continue_to_visual_generation"}
	case "final":
		system = finalSystemPrompt
		requiredKeys = []string{"schema_version", "episode_id", "verdict", "production_readiness", "blocking_issues", "suggested_revisions", "can_release_candidate"}
	default:
		return nil, fmt.Errorf("unknown review kind %q", kind)
	}

	payload, err := json.Marshal(messagesRequest{
		Model:     c.cfg.Model,
		MaxTokens: c.cfg.MaxTokens,
		System:    system,
		Thinking:  &thinkingConfig{Type: "adaptive"},
		Messages:  []wireMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return nil, err
	}

	text, err := c.send(ctx, payload)
	if err != nil {
		return nil, err
	}
	obj, err := extractJSONObject(text)
	if err != nil {
		return nil, fmt.Errorf("claude %s review did not return JSON: %w", kind, err)
	}
	if err := validateShape(obj, requiredKeys, episodeID); err != nil {
		return nil, fmt.Errorf("claude %s review JSON is invalid: %w", kind, err)
	}
	return obj, nil
}

// send posts the request, retrying only safe transient failures (429, 5xx,
// network errors), and returns the concatenated text content of the response.
func (c *Client) send(ctx context.Context, payload []byte) (string, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/v1/messages"
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(c.cfg.RetryBackoff * time.Duration(attempt)):
			}
		}

		reqCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return "", err
		}
		req.Header.Set("content-type", "application/json")
		req.Header.Set("x-api-key", c.cfg.APIKey)
		req.Header.Set("anthropic-version", anthropicVersion)

		resp, err := c.cfg.HTTPClient.Do(req)
		if err != nil {
			cancel()
			lastErr = c.redact(fmt.Errorf("claude api request failed: %v", err))
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		switch {
		case resp.StatusCode == http.StatusOK:
			var mr messagesResponse
			if err := json.Unmarshal(body, &mr); err != nil {
				return "", c.redact(fmt.Errorf("claude api returned undecodable response: %v", err))
			}
			if mr.StopReason == "refusal" {
				return "", fmt.Errorf("claude declined the review request (stop_reason=refusal)")
			}
			var sb strings.Builder
			for _, block := range mr.Content {
				if block.Type == "text" {
					sb.WriteString(block.Text)
				}
			}
			if strings.TrimSpace(sb.String()) == "" {
				return "", fmt.Errorf("claude api returned an empty text response")
			}
			return sb.String(), nil
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
			lastErr = c.redact(fmt.Errorf("claude api transient error: status %d: %s", resp.StatusCode, snippet(body)))
			continue
		default:
			return "", c.redact(fmt.Errorf("claude api error: status %d: %s", resp.StatusCode, snippet(body)))
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("claude api failed")
	}
	return "", fmt.Errorf("claude api failed after %d attempts: %w", maxRetries+1, lastErr)
}

// redact removes the API key from any error text before it can be logged.
func (c *Client) redact(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s", localexec.Redact(err.Error(), c.cfg.APIKey))
}

func snippet(body []byte) string {
	const max = 512
	s := strings.TrimSpace(string(body))
	if len(s) > max {
		return s[:max] + "...[truncated]"
	}
	return s
}

// extractJSONObject pulls the first balanced JSON object out of the model text,
// tolerating surrounding prose or code fences but rejecting non-JSON responses.
func extractJSONObject(text string) (json.RawMessage, error) {
	s := strings.TrimSpace(text)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < start {
		return nil, fmt.Errorf("no JSON object found in model response")
	}
	candidate := s[start : end+1]
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(candidate), &probe); err != nil {
		return nil, fmt.Errorf("model response is not valid JSON: %w", err)
	}
	return json.RawMessage(candidate), nil
}

// validateShape enforces the required keys and the schema_version/episode_id
// binding so a malformed or mismatched response fails closed rather than
// reaching the gate.
func validateShape(obj json.RawMessage, requiredKeys []string, episodeID string) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(obj, &m); err != nil {
		return err
	}
	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			return fmt.Errorf("missing required field %q", key)
		}
	}
	var schemaVersion string
	_ = json.Unmarshal(m["schema_version"], &schemaVersion)
	if schemaVersion != "1.0" {
		return fmt.Errorf("schema_version must be \"1.0\", got %q", schemaVersion)
	}
	var responseEpisode string
	_ = json.Unmarshal(m["episode_id"], &responseEpisode)
	if responseEpisode != episodeID {
		return fmt.Errorf("episode_id %q does not match expected %q", responseEpisode, episodeID)
	}
	var verdict string
	_ = json.Unmarshal(m["verdict"], &verdict)
	if strings.TrimSpace(verdict) == "" {
		return fmt.Errorf("verdict must be a non-empty string")
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if n, err := strconv.Atoi(value); err == nil && n > 0 {
		return n
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	if d, err := time.ParseDuration(value); err == nil && d > 0 {
		return d
	}
	return fallback
}

// ----- Messages API wire types (documented contract subset) -----

type messagesRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Thinking  *thinkingConfig `json:"thinking,omitempty"`
	Messages  []wireMessage   `json:"messages"`
}

type thinkingConfig struct {
	Type string `json:"type"`
}

type wireMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesResponse struct {
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Model      string         `json:"model"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

const scriptSystemPrompt = `You are a production QA reviewer for an educational IT short-form video pipeline.
Review the supplied script for factual accuracy, suitability for visual generation, and editorial/safety policy.
Respond with a single JSON object only. Do not include prose, explanations, or markdown code fences.
Use exactly these keys:
- "schema_version": the string "1.0"
- "episode_id": the episode id from the request
- "verdict": "pass" or "fail"
- "production_readiness": integer 0-100
- "blocking_issues": array of strings (empty if none)
- "suggested_revisions": array of strings (empty if none)
- "approved_script_hash": echo the script hash provided in the request
- "can_continue_to_visual_generation": boolean
If there are any blocking issues, set "verdict" to "fail" and "can_continue_to_visual_generation" to false.`

const finalSystemPrompt = `You are a production QA reviewer for an educational IT short-form video pipeline.
Review the release candidate described in the request for coherence, factual accuracy, and editorial/safety policy.
Respond with a single JSON object only. Do not include prose, explanations, or markdown code fences.
Use exactly these keys:
- "schema_version": the string "1.0"
- "episode_id": the episode id from the request
- "verdict": "pass" or "fail"
- "production_readiness": integer 0-100
- "blocking_issues": array of strings (empty if none)
- "suggested_revisions": array of strings (empty if none)
- "can_release_candidate": boolean
This review does not authorize any public publishing. If there are any blocking issues, set "verdict" to "fail" and "can_release_candidate" to false.`
