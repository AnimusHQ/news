// Package capabilities records provider safety/capability metadata. It is
// descriptive and fail-closed; gates and artifact validation remain authoritative.
package capabilities

import (
	"fmt"
	"sort"
)

// ProviderType classifies a provider lane.
type ProviderType string

const (
	TypeStoryboardImage ProviderType = "storyboard_image"
	TypeVisualVideo     ProviderType = "visual_video"
	TypeVoice           ProviderType = "voice"
	TypeSubtitles       ProviderType = "subtitles"
	TypeRender          ProviderType = "render"
	TypePublishing      ProviderType = "publishing"
	TypeQA              ProviderType = "qa"
)

// Record describes one provider's safety posture.
type Record struct {
	Name                        string       `json:"name"`
	Type                        ProviderType `json:"type"`
	ModesSupported              []string     `json:"modes_supported"`
	Enabled                     bool         `json:"enabled"`
	RequiresNetwork             bool         `json:"requires_network"`
	RequiresLocalBinary         bool         `json:"requires_local_binary"`
	RequiresGPU                 bool         `json:"requires_gpu"`
	RequiresGUI                 bool         `json:"requires_gui"`
	RequiresMCP                 bool         `json:"requires_mcp"`
	RequiresPaidAPI             bool         `json:"requires_paid_api"`
	RequiresHumanConsent        bool         `json:"requires_human_consent"`
	CanProduceDraftArtifacts    bool         `json:"can_produce_draft_artifacts"`
	CanProduceApprovedArtifacts bool         `json:"can_produce_approved_artifacts"`
	CanPublish                  bool         `json:"can_publish"`
	SupportsDryRun              bool         `json:"supports_dry_run"`
	SupportedArtifactTypes      []string     `json:"supported_artifact_types"`
	KnownLimitations            []string     `json:"known_limitations"`
}

// Registry is an in-memory provider capability registry.
type Registry struct {
	records map[string]Record
}

// DefaultRegistry returns M3 provider capability metadata.
func DefaultRegistry() Registry {
	records := []Record{
		{
			Name: "mock", Type: TypeRender, ModesSupported: []string{"mock"}, Enabled: true,
			CanProduceDraftArtifacts: true, SupportsDryRun: true,
			SupportedArtifactTypes: []string{"storyboard_image_manifest", "visual_shot_manifest", "voiceover_manifest", "subtitle_manifest", "short_render_manifest", "uploadpost_publish_manifest"},
			KnownLimitations:       []string{"deterministic fake outputs only"},
		},
		{
			Name: "ffmpeg", Type: TypeRender, ModesSupported: []string{"local"}, Enabled: false,
			RequiresLocalBinary: true, CanProduceDraftArtifacts: true,
			SupportedArtifactTypes: []string{"short_render_manifest"},
			KnownLimitations:       []string{"local binary required", "output bytes may vary by FFmpeg build"},
		},
		{
			Name: "faster_whisper", Type: TypeSubtitles, ModesSupported: []string{"local_sidecar"}, Enabled: false,
			RequiresLocalBinary: true, RequiresGPU: false, CanProduceDraftArtifacts: true,
			SupportedArtifactTypes: []string{"subtitle_manifest"},
			KnownLimitations:       []string{"local model must be preinstalled", "no model downloads in verification"},
		},
		{
			Name: "upload_post_dry_run", Type: TypePublishing, ModesSupported: []string{"dry_run"}, Enabled: true,
			CanProduceDraftArtifacts: true, SupportsDryRun: true, CanPublish: false,
			SupportedArtifactTypes: []string{"uploadpost_publish_manifest"},
			KnownLimitations:       []string{"no live upload or scheduling in M3"},
		},
		{
			Name: "davinci_resolve_mcp", Type: TypeRender, ModesSupported: []string{"disabled", "dry_run", "local_mcp"}, Enabled: false,
			RequiresLocalBinary: false, RequiresGUI: true, RequiresMCP: true, CanProduceDraftArtifacts: true, SupportsDryRun: true,
			SupportedArtifactTypes: []string{"short_render_manifest"},
			KnownLimitations:       []string{"optional professional finishing lane", "no default GUI dependency", "allowlisted MCP tools only"},
		},
		{
			Name: "omnivoice", Type: TypeVoice, ModesSupported: []string{"disabled", "dry_run", "local_sidecar"}, Enabled: false,
			RequiresLocalBinary: true, RequiresGPU: true, RequiresHumanConsent: true, CanProduceDraftArtifacts: true, SupportsDryRun: true,
			SupportedArtifactTypes: []string{"voiceover_manifest"},
			KnownLimitations:       []string{"local model must be preinstalled", "voice reference use requires consent metadata"},
		},
		{
			Name: "planned_seedance", Type: TypeVisualVideo, ModesSupported: []string{"planned"}, Enabled: false,
			RequiresNetwork: true, RequiresPaidAPI: true, CanProduceDraftArtifacts: true,
			SupportedArtifactTypes: []string{"visual_shot_manifest"},
			KnownLimitations:       []string{"planned only; no M3 live calls"},
		},
		{
			Name: "planned_elevenlabs", Type: TypeVoice, ModesSupported: []string{"planned"}, Enabled: false,
			RequiresNetwork: true, RequiresPaidAPI: true, CanProduceDraftArtifacts: true,
			SupportedArtifactTypes: []string{"voiceover_manifest"},
			KnownLimitations:       []string{"planned only; no M3 live calls"},
		},
		{
			Name: "planned_uploadpost_live", Type: TypePublishing, ModesSupported: []string{"planned_live"}, Enabled: false,
			RequiresNetwork: true, RequiresPaidAPI: true, CanPublish: false,
			SupportedArtifactTypes: []string{"uploadpost_publish_manifest"},
			KnownLimitations:       []string{"planned only; live publish impossible in M3"},
		},
	}
	registry := Registry{records: map[string]Record{}}
	for _, record := range records {
		registry.records[record.Name] = record
	}
	return registry
}

// List returns sorted records.
func (r Registry) List() []Record {
	out := make([]Record, 0, len(r.records))
	for _, record := range r.records {
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns one provider record.
func (r Registry) Get(name string) (Record, bool) {
	record, ok := r.records[name]
	return record, ok
}

// Select returns an enabled provider of the requested type or fails closed.
func (r Registry) Select(name string, typ ProviderType) (Record, error) {
	record, ok := r.Get(name)
	if !ok {
		return Record{}, fmt.Errorf("provider %q is not registered", name)
	}
	if record.Type != typ {
		return Record{}, fmt.Errorf("provider %q has type %q, not %q", name, record.Type, typ)
	}
	if !record.Enabled {
		return Record{}, fmt.Errorf("provider %q is disabled", name)
	}
	if record.CanProduceApprovedArtifacts {
		return Record{}, fmt.Errorf("provider %q cannot claim approval authority", name)
	}
	if record.CanPublish {
		return Record{}, fmt.Errorf("provider %q cannot publish live in M3", name)
	}
	return record, nil
}

// Validate enforces M3 provider safety invariants.
func (r Registry) Validate() error {
	for _, record := range r.records {
		if record.Name == "" {
			return fmt.Errorf("provider record missing name")
		}
		if record.Type == "" {
			return fmt.Errorf("provider %q missing type", record.Name)
		}
		if record.CanProduceApprovedArtifacts {
			return fmt.Errorf("provider %q claims approval authority", record.Name)
		}
		if record.CanPublish {
			return fmt.Errorf("provider %q claims live publish authority", record.Name)
		}
	}
	return nil
}
