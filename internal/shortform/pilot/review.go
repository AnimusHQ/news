package pilot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
	"github.com/AnimusHQ/news/internal/shortform/providers/review/claude"
)

// ReviewClient produces a structured Claude review (script or final QA) as raw
// JSON. It is transport-only: the pilot validates the JSON, owns the gate
// decision (scriptReviewPassed / finalReviewPassed), and binds the script hash.
// A review provider can never approve an artifact or publish.
type ReviewClient interface {
	Review(ctx context.Context, kind, episodeID, prompt string) (json.RawMessage, error)
}

// reviewClient returns the injected client (tests) or builds the real Claude
// client from the environment. Building fails closed when ANTHROPIC_API_KEY is
// unset, so generate-real never silently falls back to a mock.
func (r Runner) reviewClient() (ReviewClient, error) {
	if r.ReviewClient != nil {
		return r.ReviewClient, nil
	}
	return claude.FromEnv()
}

// ensureAPIScriptReview runs the Claude API script review when --claude-review
// api is selected and no response exists yet. The model owns the editorial
// verdict; the pilot owns (and rewrites) approved_script_hash so the review is
// bound to the exact script.md on disk. A non-pass verdict is written through
// and blocks at the existing gate; a transport/JSON failure fails closed.
func (r Runner) ensureAPIScriptReview(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	if manifest.Providers.ClaudeReview != "api" {
		return nil
	}
	respPath := filepath.Join(episodeDir, "claude_script_review_response.json")
	if fileExists(respPath) {
		return nil
	}
	client, err := r.reviewClient()
	if err != nil {
		return err
	}
	prompt, err := os.ReadFile(filepath.Join(episodeDir, "claude_script_review_request.md"))
	if err != nil {
		return err
	}
	raw, err := client.Review(ctx, "script", manifest.EpisodeID, string(prompt))
	if err != nil {
		return fmt.Errorf("claude api script review failed: %w", err)
	}
	review, err := decodeReviewResponse(raw, manifest.EpisodeID)
	if err != nil {
		return fmt.Errorf("claude api script review: %w", err)
	}
	scriptHash, err := localexec.FileSHA256(filepath.Join(episodeDir, "script.md"))
	if err != nil {
		return err
	}
	review.ApprovedScriptHash = scriptHash
	if err := writeJSON(respPath, review); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageClaudeScriptReview, "claude api script review generated, validated, and bound to script.md")
}

// ensureAPIFinalReview runs the Claude API final QA review when --claude-review
// api is selected and no response exists yet.
func (r Runner) ensureAPIFinalReview(ctx context.Context, episodeDir string, manifest EpisodeManifest) error {
	if manifest.Providers.ClaudeReview != "api" {
		return nil
	}
	respPath := filepath.Join(episodeDir, "final_review_response.json")
	if fileExists(respPath) {
		return nil
	}
	client, err := r.reviewClient()
	if err != nil {
		return err
	}
	prompt, err := os.ReadFile(filepath.Join(episodeDir, "final_review_request.md"))
	if err != nil {
		return err
	}
	raw, err := client.Review(ctx, "final", manifest.EpisodeID, string(prompt))
	if err != nil {
		return fmt.Errorf("claude api final review failed: %w", err)
	}
	review, err := decodeReviewResponse(raw, manifest.EpisodeID)
	if err != nil {
		return fmt.Errorf("claude api final review: %w", err)
	}
	if err := writeJSON(respPath, review); err != nil {
		return err
	}
	return r.appendAudit(episodeDir, StageClaudeFinalReview, "claude api final review generated and validated")
}

// decodeReviewResponse parses a provider response and rejects mismatched
// envelopes before they reach the gate.
func decodeReviewResponse(raw json.RawMessage, episodeID string) (ClaudeReviewResponse, error) {
	var review ClaudeReviewResponse
	if err := json.Unmarshal(raw, &review); err != nil {
		return ClaudeReviewResponse{}, fmt.Errorf("response is not valid review JSON: %w", err)
	}
	if review.SchemaVersion != SchemaVersion || review.EpisodeID != episodeID {
		return ClaudeReviewResponse{}, fmt.Errorf("response has invalid schema_version or episode_id")
	}
	return review, nil
}
