package architecture

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestBusinessModulesDoNotImportEachOther protects the vertical module
// boundary. Cross-module orchestration belongs outside internal/app modules.
func TestBusinessModulesDoNotImportEachOther(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test path")
	}
	appRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "internal", "app"))

	err := filepath.WalkDir(appRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range parsed.Imports {
			importPath, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, "cash-core/internal/app/") {
				t.Errorf("business module imports another app module: %s imports %s", path, importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect business module imports: %v", err)
	}
}
