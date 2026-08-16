package analyzers

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

const errorStringPunctuation = ".!:"

func newPreferLowercaseErrorStrings() ruleSpec {
	return newInspectingAnalyzer(
		"LEG050",
		"prefer-lowercase-error-strings",
		"Prefer lowercase error strings without trailing punctuation.",
		func(pass *analysis.Pass) (any, error) {
			checkErrorStrings(pass)
			return nil, nil
		},
	)
}

func checkErrorStrings(pass *analysis.Pass) {
	constructorsByFile := make(map[*ast.File]map[string]bool, len(pass.Files))
	for _, file := range pass.Files {
		constructorsByFile[file] = errorStringConstructorNames(file)
	}

	callTypes := []ast.Node{(*ast.CallExpr)(nil)}
	inspectCursors(pass, callTypes, func(cursor syntaxCursor) {
		call := cursor.Node().(*ast.CallExpr)
		constructors := constructorsByFile[enclosingFile(cursor)]
		checkErrorStringCall(pass, call, constructors, enclosingFunction(cursor))
	})
}

func checkErrorStringCall(
	pass *analysis.Pass,
	call *ast.CallExpr,
	constructors map[string]bool,
	function ast.Node,
) {
	literal, ok := errorStringLiteral(call, constructors, function)
	if !ok {
		return
	}

	text, err := strconv.Unquote(literal.Value)
	if err != nil {
		return
	}

	checkErrorStringText(pass, literal, text)
}

func checkErrorStringText(pass *analysis.Pass, literal *ast.BasicLit, text string) {
	if text == "" {
		return
	}

	if startsWithCapitalizedWord(text) {
		reportErrorString(pass, literal, "Start error strings with a lowercase word.")
		return
	}

	if endsWithPunctuation(text) {
		reportErrorString(pass, literal, "Drop the trailing punctuation from error strings.")
	}
}

func reportErrorString(pass *analysis.Pass, literal *ast.BasicLit, message string) {
	report(pass, literal, "LEG050", "prefer-lowercase-error-strings", message)
}

func errorStringLiteral(
	call *ast.CallExpr,
	constructors map[string]bool,
	function ast.Node,
) (*ast.BasicLit, bool) {
	if !isErrorStringConstructor(call, constructors, function) {
		return nil, false
	}

	if len(call.Args) == 0 {
		return nil, false
	}

	literal, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return nil, false
	}

	return literal, literal.Kind == token.STRING
}

func isErrorStringConstructor(
	call *ast.CallExpr,
	constructors map[string]bool,
	function ast.Node,
) bool {
	pkg, method, found := errorStringSelector(call)
	if !found {
		return false
	}
	if pkg.Obj != nil {
		return false
	}
	if localNameShadowsCall(call, pkg.Name, function) {
		return false
	}

	qualified := pkg.Name + "." + method
	return constructors[qualified]
}

func errorStringSelector(call *ast.CallExpr) (*ast.Ident, string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, "", false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return nil, "", false
	}

	return pkg, selector.Sel.Name, true
}

func localNameShadowsCall(call *ast.CallExpr, name string, function ast.Node) bool {
	if function == nil {
		return false
	}

	return declarationCountBefore(function, name, call.Pos()) > 0
}

func errorStringConstructorNames(file *ast.File) map[string]bool {
	constructors := make(map[string]bool)
	for _, spec := range file.Imports {
		name, found := errorStringConstructorName(spec)
		if found {
			constructors[name] = true
		}
	}

	return constructors
}

func errorStringConstructorName(spec *ast.ImportSpec) (string, bool) {
	importPath, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return "", false
	}

	method, found := errorStringConstructorMethod(importPath)
	if !found {
		return "", false
	}

	name, found := errorStringImportName(spec, importPath)
	if !found {
		return "", false
	}

	qualified := name + "." + method
	return qualified, true
}

func errorStringConstructorMethod(importPath string) (string, bool) {
	switch importPath {
	case "errors":
		return "New", true
	case "fmt":
		return "Errorf", true
	default:
		return "", false
	}
}

func errorStringImportName(spec *ast.ImportSpec, importPath string) (string, bool) {
	if spec.Name == nil {
		return importPath, true
	}

	name := spec.Name.Name
	if name == "_" {
		return "", false
	}
	if name == "." {
		return "", false
	}

	return name, true
}

func startsWithCapitalizedWord(text string) bool {
	word := firstErrorStringWord(text)
	if !startsUppercase(word) {
		return false
	}

	if word == strings.ToUpper(word) {
		return false
	}

	return !hasInnerUppercase(word)
}

func firstErrorStringWord(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return ""
	}

	return strings.Trim(fields[0], errorStringPunctuation+",;")
}

func hasInnerUppercase(word string) bool {
	for index, char := range word {
		if isInnerUppercase(index, char) {
			return true
		}
	}

	return false
}

func isInnerUppercase(index int, char rune) bool {
	if index == 0 {
		return false
	}

	return unicode.IsUpper(char)
}

func endsWithPunctuation(text string) bool {
	last := text[len(text)-1:]
	return strings.Contains(errorStringPunctuation, last)
}
