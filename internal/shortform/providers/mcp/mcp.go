// Package mcp defines a constrained MCP boundary for provider adapters. It does
// not expose arbitrary MCP tool execution.
package mcp

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// DaVinci Resolve MCP tool allowlist.
const (
	ToolResolveHealthcheck               = "resolve.healthcheck"
	ToolResolveCreateProjectFromManifest = "resolve.create_project_from_manifest"
	ToolResolveImportAssetsFromManifest  = "resolve.import_assets_from_manifest"
	ToolResolveCreateVerticalTimeline    = "resolve.create_vertical_timeline"
	ToolResolveQueueRenderFromManifest   = "resolve.queue_render_from_manifest"
	ToolResolveStartRenderIfAllowed      = "resolve.start_render_if_allowed"
	ToolResolveGetRenderStatus           = "resolve.get_render_status"
	ToolResolveCollectRenderResult       = "resolve.collect_render_result"
	ToolResolveExportProjectArchive      = "resolve.export_project_archive"
)

var allowedResolveTools = map[string]bool{
	ToolResolveHealthcheck:               true,
	ToolResolveCreateProjectFromManifest: true,
	ToolResolveImportAssetsFromManifest:  true,
	ToolResolveCreateVerticalTimeline:    true,
	ToolResolveQueueRenderFromManifest:   true,
	ToolResolveStartRenderIfAllowed:      true,
	ToolResolveGetRenderStatus:           true,
	ToolResolveCollectRenderResult:       true,
	ToolResolveExportProjectArchive:      true,
}

// ValidateResolveTool fails closed for any tool outside the Resolve allowlist.
func ValidateResolveTool(tool string) error {
	if !allowedResolveTools[tool] {
		return fmt.Errorf("MCP tool %q is not allowlisted for DaVinci Resolve", tool)
	}
	if strings.Contains(tool, "script") || strings.Contains(tool, "eval") || strings.Contains(tool, "secret") {
		return fmt.Errorf("MCP tool %q is forbidden", tool)
	}
	return nil
}

// ValidateEndpoint checks the configured MCP URL shape without opening a
// network connection.
func ValidateEndpoint(raw string) error {
	if raw == "" {
		return fmt.Errorf("DaVinci Resolve MCP URL must be configured")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("DaVinci Resolve MCP URL is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("DaVinci Resolve MCP URL must use http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("DaVinci Resolve MCP URL must include a host")
	}
	return nil
}

// CallRequest is a constrained MCP tool call request.
type CallRequest struct {
	Tool    string         `json:"tool"`
	Timeout time.Duration  `json:"timeout,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// CallResponse is a constrained MCP tool call response.
type CallResponse struct {
	Tool   string         `json:"tool"`
	OK     bool           `json:"ok"`
	Status string         `json:"status,omitempty"`
	Values map[string]any `json:"values,omitempty"`
}

// Client is implemented by concrete or dry-run MCP clients.
type Client interface {
	Call(ctx context.Context, req CallRequest) (CallResponse, error)
}

// DryRunClient validates allowlisted calls and records them without network.
type DryRunClient struct {
	Calls []CallRequest
}

// Call implements Client.
func (c *DryRunClient) Call(_ context.Context, req CallRequest) (CallResponse, error) {
	if err := ValidateResolveTool(req.Tool); err != nil {
		return CallResponse{}, err
	}
	c.Calls = append(c.Calls, req)
	return CallResponse{Tool: req.Tool, OK: true, Status: "dry_run"}, nil
}
