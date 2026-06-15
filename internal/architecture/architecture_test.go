package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/AnimusHQ/news/internal/analytics"
)

const modulePath = "github.com/AnimusHQ/news"

func TestWorkflowProductionCodeAvoidsForbiddenSideEffectImports(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "workflows"))
	for _, parsed := range files {
		for _, imported := range parsed.imports() {
			if reason, forbidden := forbiddenWorkflowImport(imported); forbidden {
				t.Fatalf("%s imports %q: %s", parsed.rel, imported, reason)
			}
		}
	}
}

func TestWorkflowProductionCodeAvoidsDirectNondeterministicTimeCalls(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "workflows"))
	for _, parsed := range files {
		ast.Inspect(parsed.file, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "time" {
				return true
			}
			if forbiddenDirectTimeCall(selector.Sel.Name) {
				t.Fatalf("%s calls time.%s directly; workflow code must use Temporal workflow time APIs", parsed.rel, selector.Sel.Name)
			}
			return true
		})
	}
}

func TestAdaptersDoNotImportWorkflowPackages(t *testing.T) {
	root := repoRoot(t)
	for _, dir := range []string{
		filepath.Join(root, "internal", "models"),
		filepath.Join(root, "internal", "providers"),
		filepath.Join(root, "internal", "publishing"),
	} {
		files := parseProductionGoFiles(t, dir)
		for _, parsed := range files {
			for _, imported := range parsed.imports() {
				if imported == modulePath+"/internal/workflows" || strings.HasPrefix(imported, modulePath+"/internal/workflows/") {
					t.Fatalf("%s imports %q; adapters must not depend on workflow packages", parsed.rel, imported)
				}
			}
		}
	}
}

func TestPublishingCodeDoesNotExposeDirectPublicUploadPath(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "publishing"))
	for _, parsed := range files {
		for _, decl := range parsed.file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := strings.ToLower(fn.Name.Name)
			if strings.Contains(name, "public") && containsAny(name, "upload", "publish", "schedule") {
				t.Fatalf("%s exposes direct public publishing entrypoint %s", parsed.rel, fn.Name.Name)
			}
		}
	}
}

func TestPublishingAdaptersDoNotCreatePublicResults(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "publishing"))
	for _, parsed := range files {
		if !strings.Contains(filepath.Base(parsed.rel), "adapter") {
			continue
		}
		ast.Inspect(parsed.file, func(node ast.Node) bool {
			kv, ok := node.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Visibility" {
				return true
			}
			if exprIsPublishVisibilityPublic(kv.Value) {
				t.Fatalf("%s creates adapter result/status with public visibility", parsed.rel)
			}
			return true
		})
	}
}

func TestAnalyticsCodeDoesNotImportMutationBoundaries(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "analytics"))
	for _, parsed := range files {
		for _, imported := range parsed.imports() {
			if imported == modulePath+"/internal/publishing" ||
				imported == modulePath+"/internal/workflows" ||
				imported == modulePath+"/internal/storage" {
				t.Fatalf("%s imports %q; analytics must remain advisory and not mutate release state", parsed.rel, imported)
			}
		}
	}
}

func TestAnalyticsCodeDoesNotDisableAdvisoryOnly(t *testing.T) {
	root := repoRoot(t)
	files := parseProductionGoFiles(t, filepath.Join(root, "internal", "analytics"))
	for _, parsed := range files {
		ast.Inspect(parsed.file, func(node ast.Node) bool {
			kv, ok := node.(*ast.KeyValueExpr)
			if !ok {
				return true
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "AdvisoryOnly" {
				return true
			}
			value, ok := kv.Value.(*ast.Ident)
			if ok && value.Name == "false" {
				t.Fatalf("%s explicitly disables analytics AdvisoryOnly", parsed.rel)
			}
			return true
		})
	}
}

func TestAnalyticsReportConstructorsRemainAdvisoryOnly(t *testing.T) {
	input := analytics.Input{
		Provider:  "architecture-test",
		EpisodeID: "episode-test",
		Window:    analytics.Window72h,
		Metrics: analytics.Metrics{
			CTR:              0.02,
			Impressions:      2000,
			Views:            400,
			First30Retention: 0.35,
			CompletionRate:   0.30,
		},
	}
	imported := analytics.ReportFromInput(input)
	if !imported.AdvisoryOnly {
		t.Fatal("ReportFromInput must produce advisory-only reports")
	}
	insight, err := analytics.GenerateInsightReport(input)
	if err != nil {
		t.Fatalf("GenerateInsightReport failed: %v", err)
	}
	if !insight.AdvisoryOnly {
		t.Fatal("GenerateInsightReport must produce advisory-only reports")
	}
}

type parsedGoFile struct {
	rel  string
	file *ast.File
}

func (p parsedGoFile) imports() []string {
	imports := make([]string, 0, len(p.file.Imports))
	for _, spec := range p.file.Imports {
		imports = append(imports, strings.Trim(spec.Path.Value, `"`))
	}
	return imports
}

func parseProductionGoFiles(t *testing.T, dir string) []parsedGoFile {
	t.Helper()
	var files []parsedGoFile
	fset := token.NewFileSet()
	root := repoRoot(t)
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, parsedGoFile{rel: filepath.ToSlash(rel), file: parsed})
		return nil
	})
	if err != nil {
		t.Fatalf("parse production Go files under %s: %v", dir, err)
	}
	return files
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func containsAny(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func exprIsPublishVisibilityPublic(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "PublishVisibilityPublic" {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "artifacts"
}

func forbiddenWorkflowImport(imported string) (string, bool) {
	exact := map[string]string{
		"os":           "filesystem access belongs in activities",
		"io/ioutil":    "filesystem access belongs in activities",
		"net":          "network access belongs in activities",
		"net/http":     "network access belongs in activities",
		"database/sql": "database access belongs in activities",
		"math/rand":    "randomness is nondeterministic in workflows",
		"crypto/rand":  "randomness is nondeterministic in workflows",
	}
	if reason, ok := exact[imported]; ok {
		return reason, true
	}
	prefixes := map[string]string{
		"github.com/aws/aws-sdk-go":             "provider/object-store SDK calls belong in activities",
		"github.com/aws/aws-sdk-go-v2":          "provider/object-store SDK calls belong in activities",
		"github.com/jackc/pgx":                  "database calls belong in activities",
		"cloud.google.com/go":                   "provider/cloud SDK calls belong in activities",
		"github.com/Azure/azure-sdk-for-go":     "provider/cloud SDK calls belong in activities",
		modulePath + "/internal/storage":        "storage access belongs in activities",
		modulePath + "/internal/publishing":     "publishing access belongs in activities",
		modulePath + "/internal/models/sandbox": "provider execution belongs in activities",
	}
	for prefix, reason := range prefixes {
		if imported == prefix || strings.HasPrefix(imported, prefix+"/") {
			return reason, true
		}
	}
	return "", false
}

func forbiddenDirectTimeCall(name string) bool {
	switch name {
	case "Now", "Since", "Until", "After", "AfterFunc", "NewTimer", "NewTicker", "Sleep", "Tick":
		return true
	default:
		return false
	}
}
