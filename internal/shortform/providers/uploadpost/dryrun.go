// Package uploadpost contains the Upload-Post publishing adapter boundary. M2
// implements dry-run request construction only; live publishing is refused.
package uploadpost

import (
	"context"
	"fmt"
	"time"

	"github.com/AnimusHQ/news/internal/shortform"
	"github.com/AnimusHQ/news/internal/shortform/gates"
	"github.com/AnimusHQ/news/internal/shortform/providers"
	"github.com/AnimusHQ/news/internal/shortform/providers/localexec"
)

const defaultVersion = "m2-dry-run"

// DryRunConfig controls the Upload-Post adapter. Enabled must be true and Mode
// must be dry_run, otherwise the provider fails closed.
type DryRunConfig struct {
	Enabled bool
	Mode    string
	APIKey  string
	Version string
}

// Redacted returns a copy safe for logs or diagnostics.
func (c DryRunConfig) Redacted() DryRunConfig {
	if c.APIKey != "" {
		c.APIKey = "[REDACTED]"
	}
	return c
}

// DryRunProvider builds an Upload-Post-shaped manifest without network calls.
type DryRunProvider struct {
	cfg DryRunConfig
}

// NewDryRunProvider builds a disabled-by-default Upload-Post dry-run adapter.
func NewDryRunProvider(cfg DryRunConfig) *DryRunProvider {
	return &DryRunProvider{cfg: cfg}
}

// UploadPostDryRun implements providers.PublishingProvider. It never uploads or
// schedules anything publicly.
func (p *DryRunProvider) UploadPostDryRun(_ context.Context, req providers.PublishRequest) (*shortform.UploadPostPublishManifest, error) {
	if !p.cfg.Enabled {
		return nil, fmt.Errorf("upload-post dry-run adapter is disabled")
	}
	if p.cfg.Mode != "" && p.cfg.Mode != "dry_run" {
		return nil, fmt.Errorf("upload-post live mode %q is refused in M2", p.cfg.Mode)
	}
	if err := validateDryRunRequest(req); err != nil {
		return nil, err
	}
	qaRef := req.ProductionQARef
	if qaRef == "" {
		qaRef = req.Release.ProductionQARef
	}
	manifest := &shortform.UploadPostPublishManifest{
		Envelope: shortform.Envelope{
			SchemaVersion: shortform.SchemaVersion,
			EpisodeID:     req.EpisodeID,
			ArtifactID:    fmt.Sprintf("%s-%s-v1", shortform.KindUploadPostPublishManifest, req.EpisodeID),
			CreatedAt:     rfc3339(req.Now),
			CreatedBy:     "system:upload-post-dry-run",
			SourceArtifacts: []string{
				req.Render.ArtifactID,
				req.Release.ArtifactID,
			},
			Status: shortform.StatusDraft,
		},
		Provider:             "upload_post",
		Mode:                 "dry_run",
		DryRun:               true,
		Platforms:            append([]string(nil), req.Release.Platforms...),
		Visibility:           req.Release.Visibility,
		ScheduledAt:          req.Release.ScheduledAt,
		AIDisclosureRequired: req.Release.AIDisclosureRequired,
		AIDisclosure:         req.Release.AIDisclosure,
		HumanReleaseApproval: req.Release.HumanReleaseApproval,
		ProductionQARef:      qaRef,
		ReleaseApprovalRef:   req.Release.ArtifactID,
	}
	if err := shortform.Stamp(manifest); err != nil {
		return nil, err
	}
	if issues := shortform.Validate(manifest); len(issues) != 0 {
		return nil, fmt.Errorf("upload-post dry-run manifest failed validation: %v", issues)
	}
	return manifest, nil
}

func validateDryRunRequest(req providers.PublishRequest) error {
	if req.EpisodeID == "" {
		return fmt.Errorf("episode_id is required")
	}
	if err := localexec.SafeSegment(req.EpisodeID, "episode_id"); err != nil {
		return err
	}
	if req.Render == nil {
		return fmt.Errorf("upload-post dry-run requires a render artifact")
	}
	if issues := shortform.Validate(req.Render); len(issues) != 0 {
		return fmt.Errorf("render artifact failed validation: %v", issues)
	}
	if req.Render.Status != shortform.StatusApproved {
		return fmt.Errorf("render artifact must be approved before dry-run publish")
	}
	for _, out := range req.Render.Outputs {
		if out.Status != shortform.StatusApproved {
			return fmt.Errorf("render output %q must be approved before dry-run publish", out.Platform)
		}
	}
	if req.ProductionQADecision != gates.ProductionQAApproved {
		return fmt.Errorf("production QA must be approved before dry-run publish")
	}
	if req.Release == nil {
		return fmt.Errorf("upload-post dry-run requires a release approval")
	}
	if issues := shortform.Validate(req.Release); len(issues) != 0 {
		return fmt.Errorf("release approval failed validation: %v", issues)
	}
	if req.Release.Status != shortform.StatusApproved || !req.Release.HumanReleaseApproval {
		return fmt.Errorf("human release approval is required before dry-run publish")
	}
	if req.Release.ApprovedBy == "" {
		return fmt.Errorf("release approval approver must be recorded")
	}
	if len(req.Release.Platforms) == 0 {
		return fmt.Errorf("release platforms must be explicit")
	}
	for _, platform := range req.Release.Platforms {
		if platform != "youtube" && platform != "tiktok" && platform != "instagram" {
			return fmt.Errorf("unsupported upload-post platform %q", platform)
		}
	}
	if req.Release.Visibility == "" {
		return fmt.Errorf("release visibility must be set")
	}
	if req.Release.Visibility == "scheduled" && req.Release.ScheduledAt == "" {
		return fmt.Errorf("scheduled release requires scheduled_at")
	}
	if req.Release.AIDisclosureRequired && req.Release.AIDisclosure == "" {
		return fmt.Errorf("AI disclosure is required before dry-run publish")
	}
	if req.ProductionQARef == "" && req.Release.ProductionQARef == "" {
		return fmt.Errorf("production QA reference is required")
	}
	return nil
}

func rfc3339(now time.Time) string {
	if now.IsZero() {
		now = time.Unix(0, 0).UTC()
	}
	return now.UTC().Format(time.RFC3339)
}

func versionOrDefault(version string) string {
	if version == "" {
		return defaultVersion
	}
	return version
}

var _ providers.PublishingProvider = (*DryRunProvider)(nil)
