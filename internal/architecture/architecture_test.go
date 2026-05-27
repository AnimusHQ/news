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
		parsed, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
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
