package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findModuleRoot walks up from current working directory to find go.mod.
func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find module root containing go.mod")
		}
		dir = parent
	}
}

// isStdLib checks whether an import path belongs to the Go standard library.
// Stdlib packages do not contain '.' in the first path element and do not start with the module name.
func isStdLib(importPath string) bool {
	if importPath == "" || importPath == "C" {
		return false
	}
	firstElem := strings.Split(importPath, "/")[0]
	return !strings.Contains(firstElem, ".") && !strings.HasPrefix(importPath, "labshop")
}

// inspectImports parses all Go files in a given directory tree and returns
// a map of file paths to their imported package paths.
func inspectImports(t *testing.T, dir string) map[string][]string {
	t.Helper()
	fileImports := make(map[string][]string)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fileImports
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return fmt.Errorf("failed to parse file %s: %w", path, parseErr)
		}

		var imports []string
		for _, imp := range node.Imports {
			cleanPath, unquoteErr := strconv.Unquote(imp.Path.Value)
			if unquoteErr != nil {
				cleanPath = strings.Trim(imp.Path.Value, `"`)
			}
			imports = append(imports, cleanPath)
		}
		fileImports[path] = imports
		return nil
	})

	require.NoError(t, err, "failed to walk directory: %s", dir)
	return fileImports
}

// TestDomainOnlyImportsStdlib enforces Clean Architecture rule:
// internal/domain packages ONLY import Go stdlib packages (and internal domain packages),
// and specifically NEVER import net/http, database/sql, or external libraries.
func TestDomainOnlyImportsStdlib(t *testing.T) {
	rootDir := findModuleRoot(t)
	domainDir := filepath.Join(rootDir, "internal", "domain")

	files := inspectImports(t, domainDir)
	assert.NotEmpty(t, files, "internal/domain should contain at least one Go file")

	var violations []string
	for filePath, imports := range files {
		for _, imp := range imports {
			// Rule: no net/http or database/sql in domain
			if imp == "net/http" || imp == "database/sql" {
				violations = append(violations, fmt.Sprintf("%s imports forbidden stdlib package: %s", filePath, imp))
				continue
			}

			// Must be standard library or another domain package
			isDomainPkg := strings.HasPrefix(imp, "labshop/internal/domain")
			if !isStdLib(imp) && !isDomainPkg {
				violations = append(violations, fmt.Sprintf("%s imports non-stdlib / non-domain package: %s", filePath, imp))
			}
		}
	}

	assert.Empty(t, violations, "Clean Architecture violations found in internal/domain:\n%s", strings.Join(violations, "\n"))
}

// TestUsecaseOnlyImportsStdlibAndDomain enforces Clean Architecture rule:
// internal/usecase packages ONLY import Go stdlib and internal/domain packages.
// They must NEVER import adapters, infrastructure, or external third-party libraries.
func TestUsecaseOnlyImportsStdlibAndDomain(t *testing.T) {
	rootDir := findModuleRoot(t)
	usecaseDir := filepath.Join(rootDir, "internal", "usecase")

	files := inspectImports(t, usecaseDir)
	assert.NotEmpty(t, files, "internal/usecase should contain at least one Go file")

	var violations []string
	for filePath, imports := range files {
		for _, imp := range imports {
			isDomainPkg := strings.HasPrefix(imp, "labshop/internal/domain")
			isUsecasePkg := strings.HasPrefix(imp, "labshop/internal/usecase")

			if !isStdLib(imp) && !isDomainPkg && !isUsecasePkg {
				violations = append(violations, fmt.Sprintf("%s imports forbidden package: %s", filePath, imp))
			}
		}
	}

	assert.Empty(t, violations, "Clean Architecture violations found in internal/usecase:\n%s", strings.Join(violations, "\n"))
}

// TestArchRuleEnforcement verifies that our architecture checker correctly flags forbidden imports.
func TestArchRuleEnforcement(t *testing.T) {
	srcDomainBad := `package testdomain
import (
	"database/sql"
	"github.com/go-chi/chi/v5"
)
`
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "bad_domain.go", srcDomainBad, parser.ImportsOnly)
	require.NoError(t, err)

	var badImports []string
	for _, imp := range node.Imports {
		cleanPath, _ := strconv.Unquote(imp.Path.Value)
		if imp.Path.Value != "" {
			if cleanPath == "database/sql" || !isStdLib(cleanPath) {
				badImports = append(badImports, cleanPath)
			}
		}
	}
	assert.Contains(t, badImports, "database/sql")
	assert.Contains(t, badImports, "github.com/go-chi/chi/v5")
}
