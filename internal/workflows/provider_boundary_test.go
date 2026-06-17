package workflows

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShortFormWorkflowDoesNotCallExternalProviderBoundaries(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test file")
	}
	sourcePath := filepath.Join(filepath.Dir(file), "shortform.go")
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, forbidden := range []string{"davinci", "omnivoice", "mcp.", "exec.", "CommandContext"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("workflow source must not directly reference external provider boundary %q", forbidden)
		}
	}
}
