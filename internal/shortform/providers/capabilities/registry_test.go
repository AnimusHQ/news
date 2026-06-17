package capabilities

import (
	"strings"
	"testing"
)

func TestDefaultRegistryValidates(t *testing.T) {
	registry := DefaultRegistry()
	if err := registry.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(registry.List()) < 9 {
		t.Fatalf("expected M3 provider records, got %d", len(registry.List()))
	}
}

func TestUnknownAndDisabledProvidersFailClosed(t *testing.T) {
	registry := DefaultRegistry()
	if _, err := registry.Select("does_not_exist", TypeRender); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("expected unknown provider failure, got %v", err)
	}
	if _, err := registry.Select("ffmpeg", TypeRender); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled provider failure, got %v", err)
	}
}

func TestProviderCannotClaimApprovalOrPublishAuthority(t *testing.T) {
	registry := DefaultRegistry()
	for _, record := range registry.List() {
		if record.CanProduceApprovedArtifacts {
			t.Fatalf("%s must not produce approved artifacts", record.Name)
		}
		if record.CanPublish {
			t.Fatalf("%s must not publish live in M3", record.Name)
		}
	}
}

func TestEnabledDryRunProviderCanBeSelected(t *testing.T) {
	registry := DefaultRegistry()
	record, err := registry.Select("upload_post_dry_run", TypePublishing)
	if err != nil {
		t.Fatal(err)
	}
	if !record.SupportsDryRun || record.CanPublish {
		t.Fatalf("unexpected upload-post dry-run posture: %+v", record)
	}
}

func TestDaVinciAndOmniVoiceCapabilityFlags(t *testing.T) {
	registry := DefaultRegistry()
	davinci, ok := registry.Get("davinci_resolve_mcp")
	if !ok {
		t.Fatal("davinci_resolve_mcp missing")
	}
	if !davinci.RequiresGUI || !davinci.RequiresMCP || !davinci.SupportsDryRun || davinci.Enabled {
		t.Fatalf("unexpected DaVinci posture: %+v", davinci)
	}
	omni, ok := registry.Get("omnivoice")
	if !ok {
		t.Fatal("omnivoice missing")
	}
	if !omni.RequiresLocalBinary || !omni.RequiresHumanConsent || !omni.RequiresGPU || omni.Enabled {
		t.Fatalf("unexpected OmniVoice posture: %+v", omni)
	}
}
