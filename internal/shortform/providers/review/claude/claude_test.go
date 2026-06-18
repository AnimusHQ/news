package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	scriptJSON = `{"schema_version":"1.0","episode_id":"animus-oss-001","verdict":"pass","production_readiness":86,"blocking_issues":[],"suggested_revisions":[],"approved_script_hash":"sha256:abc","can_continue_to_visual_generation":true}`
	finalJSON  = `{"schema_version":"1.0","episode_id":"animus-oss-001","verdict":"pass","production_readiness":88,"blocking_issues":[],"suggested_revisions":[],"can_release_candidate":true}`
)

// messagesAPIResponse builds a minimal Messages API response whose single text
// block carries the supplied body.
func messagesAPIResponse(text string) string {
	resp := map[string]any{
		"id":          "msg_test",
		"type":        "message",
		"role":        "assistant",
		"model":       "claude-opus-4-8",
		"stop_reason": "end_turn",
		"content":     []map[string]any{{"type": "thinking", "thinking": ""}, {"type": "text", "text": text}},
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 20},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func testClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := New(Config{APIKey: "sk-ant-test-key", BaseURL: baseURL, RetryBackoff: time.Millisecond})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return c
}

func TestReviewScriptSuccessValidatesHeadersAndJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" || r.Header.Get("anthropic-version") != anthropicVersion {
			t.Errorf("missing auth/version headers: %v", r.Header)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		io.WriteString(w, messagesAPIResponse(scriptJSON))
	}))
	defer srv.Close()

	raw, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "review this script")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	if got["verdict"] != "pass" {
		t.Fatalf("unexpected verdict: %v", got["verdict"])
	}
}

func TestReviewFinalSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, messagesAPIResponse(finalJSON))
	}))
	defer srv.Close()

	raw, err := testClient(t, srv.URL).Review(context.Background(), "final", "animus-oss-001", "review this candidate")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	if !strings.Contains(string(raw), "can_release_candidate") {
		t.Fatalf("final review missing can_release_candidate: %s", raw)
	}
}

func TestReviewToleratesCodeFences(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, messagesAPIResponse("```json\n"+scriptJSON+"\n```"))
	}))
	defer srv.Close()

	if _, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x"); err != nil {
		t.Fatalf("fenced JSON should be accepted: %v", err)
	}
}

func TestReviewRejectsMarkdownOnlyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, messagesAPIResponse("The script looks great, no JSON here."))
	}))
	defer srv.Close()

	if _, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x"); err == nil {
		t.Fatal("expected rejection of markdown-only response")
	}
}

func TestReviewRejectsSchemaMismatch(t *testing.T) {
	// Missing "verdict" and other required keys.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, messagesAPIResponse(`{"schema_version":"1.0","episode_id":"animus-oss-001"}`))
	}))
	defer srv.Close()

	_, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x")
	if err == nil || !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("expected schema mismatch rejection, got %v", err)
	}
}

func TestReviewRejectsEpisodeMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, messagesAPIResponse(scriptJSON)) // episode_id animus-oss-001
	}))
	defer srv.Close()

	_, err := testClient(t, srv.URL).Review(context.Background(), "script", "different-episode", "x")
	if err == nil || !strings.Contains(err.Error(), "episode_id") {
		t.Fatalf("expected episode mismatch rejection, got %v", err)
	}
}

func TestReviewRejectsRefusal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"type": "message", "role": "assistant", "stop_reason": "refusal", "content": []map[string]any{}}
		data, _ := json.Marshal(resp)
		w.Write(data)
	}))
	defer srv.Close()

	_, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x")
	if err == nil || !strings.Contains(err.Error(), "refusal") {
		t.Fatalf("expected refusal handling, got %v", err)
	}
}

func TestFromEnvMissingKeyFailsClosed(t *testing.T) {
	t.Setenv("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS", "1")
	t.Setenv("ANTHROPIC_API_KEY", "")
	if _, err := FromEnv(); err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("expected fail-closed on missing key, got %v", err)
	}
}

func TestFromEnvRequiresLiveCallGuard(t *testing.T) {
	t.Setenv("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS", "")
	t.Setenv("ANTHROPIC_API_KEY", "animus-fake-pilot-credential-0001")
	if _, err := FromEnv(); err == nil || !strings.Contains(err.Error(), "ANIMUS_ALLOW_LIVE_PROVIDER_CALLS") {
		t.Fatalf("expected fail-closed on missing live-call guard, got %v", err)
	}
}

func TestNewMissingKeyFailsClosed(t *testing.T) {
	if _, err := New(Config{}); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("expected fail-closed on missing key, got %v", err)
	}
}

func TestReviewRetriesTransientThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			io.WriteString(w, `{"error":{"message":"overloaded"}}`)
			return
		}
		io.WriteString(w, messagesAPIResponse(scriptJSON))
	}))
	defer srv.Close()

	if _, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x"); err != nil {
		t.Fatalf("expected retry success: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected 2 calls (one retry), got %d", got)
	}
}

func TestReviewDoesNotRetryClientError(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":{"message":"bad request"}}`)
	}))
	defer srv.Close()

	if _, err := testClient(t, srv.URL).Review(context.Background(), "script", "animus-oss-001", "x"); err == nil {
		t.Fatal("expected client error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("client errors must not retry, got %d calls", got)
	}
}

func TestReviewRedactsAPIKeyFromErrors(t *testing.T) {
	// Neutral, non-secret-shaped value: proves redaction without tripping the
	// repo secret scanner.
	leaked := "animus-fake-pilot-credential-0001"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		// Defensive: a provider that echoes the key must never reach logs.
		fmt.Fprintf(w, `{"error":{"message":"invalid key %s"}}`, leaked)
	}))
	defer srv.Close()

	c, err := New(Config{APIKey: leaked, BaseURL: srv.URL, RetryBackoff: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Review(context.Background(), "script", "animus-oss-001", "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), leaked) {
		t.Fatalf("api key leaked in error: %v", err)
	}
	if !strings.Contains(err.Error(), "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %v", err)
	}
}

func TestUnknownKindRejected(t *testing.T) {
	c := testClient(t, "http://example.invalid")
	if _, err := c.Review(context.Background(), "bogus", "ep", "x"); err == nil {
		t.Fatal("expected unknown kind rejection")
	}
}
