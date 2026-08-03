package analyzers

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const minStutterPackageNameLength = 3

var genericPackageNames = map[string]bool{
	"base":    true,
	"common":  true,
	"helper":  true,
	"helpers": true,
	"lib":     true,
	"libs":    true,
	"misc":    true,
	"shared":  true,
	"util":    true,
	"utils":   true,
}

func newNoPackageNameStutter() ruleSpec {
	return newAnalyzer(
		"LEG045",
		"no-package-name-stutter",
		"Avoid repeating the package name in exported names.",
		func(pass *analysis.Pass) (any, error) {
			checkPackageNameStutter(pass)
			return nil, nil
		},
	)
}

func newNoGenericPackageNames() ruleSpec {
	return newAnalyzer(
		"LEG046",
		"no-generic-package-names",
		"Avoid catch-all package names such as util or common.",
		func(pass *analysis.Pass) (any, error) {
			checkGenericPackageNames(pass)
			return nil, nil
		},
	)
}

func checkPackageNameStutter(pass *analysis.Pass) {
	for _, file := range pass.Files {
		prefix := stutterPrefix(file)
		if prefix == "" {
			continue
		}

		checkDeclNamesStutter(pass, file.Decls, prefix)
	}
}

func stutterPrefix(file *ast.File) string {
	name := strings.TrimSuffix(file.Name.Name, "_test")
	if name == "main" {
		return ""
	}

	if len(name) < minStutterPackageNameLength {
		return ""
	}

	return strings.ToLower(name)
}

func checkDeclNamesStutter(pass *analysis.Pass, decls []ast.Decl, prefix string) {
	for _, decl := range decls {
		checkDeclStutter(pass, decl, prefix)
	}
}

func checkDeclStutter(pass *analysis.Pass, decl ast.Decl, prefix string) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		checkFuncDeclStutter(pass, typed, prefix)
	case *ast.GenDecl:
		checkSpecNamesStutter(pass, typed.Specs, prefix)
	}
}

func checkFuncDeclStutter(pass *analysis.Pass, decl *ast.FuncDecl, prefix string) {
	if decl.Recv != nil {
		return
	}

	checkStutterName(pass, decl.Name, prefix)
}

func checkSpecNamesStutter(pass *analysis.Pass, specs []ast.Spec, prefix string) {
	for _, spec := range specs {
		checkSpecNameStutter(pass, spec, prefix)
	}
}

func checkSpecNameStutter(pass *analysis.Pass, spec ast.Spec, prefix string) {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		checkStutterName(pass, typed.Name, prefix)
	case *ast.ValueSpec:
		checkValueNamesStutter(pass, typed.Names, prefix)
	}
}

func checkValueNamesStutter(pass *analysis.Pass, names []*ast.Ident, prefix string) {
	for _, name := range names {
		checkStutterName(pass, name, prefix)
	}
}

func checkStutterName(pass *analysis.Pass, identifier *ast.Ident, prefix string) {
	if !stuttersPackageName(identifier.Name, prefix) {
		return
	}

	report(
		pass,
		identifier,
		"LEG045",
		"no-package-name-stutter",
		"Drop the package name from this name; callers already read it as a qualifier.",
	)
}

func stuttersPackageName(name string, prefix string) bool {
	if !ast.IsExported(name) {
		return false
	}

	lowered := strings.ToLower(name)
	if len(lowered) != len(name) {
		return false
	}

	if !strings.HasPrefix(lowered, prefix) {
		return false
	}

	return startsUppercase(name[len(prefix):])
}

func checkGenericPackageNames(pass *analysis.Pass) {
	if len(pass.Files) == 0 {
		return
	}

	file := pass.Files[0]
	name := strings.TrimSuffix(file.Name.Name, "_test")
	if !genericPackageNames[name] {
		return
	}

	reportGenericPackageName(pass, file.Name)
}

func reportGenericPackageName(pass *analysis.Pass, identifier *ast.Ident) {
	report(
		pass,
		identifier,
		"LEG046",
		"no-generic-package-names",
		"Name the package for what it provides instead of a catch-all bucket.",
	)
}
