package mcp

import (
	"context"
	"strings"
	"testing"
)

func TestValidateResolveToolAllowlist(t *testing.T) {
	if err := ValidateResolveTool(ToolResolveHealthcheck); err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"run_arbitrary_python", "execute_script", "eval", "read_secret", "publish_after_render"}
	for _, tool := range forbidden {
		t.Run(tool, func(t *testing.T) {
			if err := ValidateResolveTool(tool); err == nil {
				t.Fatalf("expected %s to be refused", tool)
			}
		})
	}
}

func TestValidateEndpoint(t *testing.T) {
	if err := ValidateEndpoint("http://127.0.0.1:8989/mcp"); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"", "file:///tmp/mcp.sock", "http://"} {
		t.Run(raw, func(t *testing.T) {
			if err := ValidateEndpoint(raw); err == nil {
				t.Fatal("expected endpoint validation to fail")
			}
		})
	}
}

func TestDryRunClientRefusesUnknownTool(t *testing.T) {
	client := &DryRunClient{}
	_, err := client.Call(context.Background(), CallRequest{Tool: "execute_script"})
	if err == nil || !strings.Contains(err.Error(), "not allowlisted") {
		t.Fatalf("expected allowlist refusal, got %v", err)
	}
	if len(client.Calls) != 0 {
		t.Fatalf("forbidden call must not be recorded: %v", client.Calls)
	}
}
