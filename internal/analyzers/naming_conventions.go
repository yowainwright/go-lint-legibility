package analyzers

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var genericFunctionPrefixes = []string{
	"config", "data", "info", "item", "object", "result", "settings", "stuff", "thing", "value",
}

var booleanPrefixes = []string{"can", "did", "has", "is", "should", "was", "were", "will"}

var booleanExceptions = map[string]bool{
	"debug": true, "disabled": true, "done": true, "empty": true, "enabled": true,
	"exists": true, "found": true, "invalid": true, "ok": true, "ready": true, "valid": true,
}

func newPreferVerbFunctionNames() ruleSpec {
	return newOptionalAnalyzer(
		"LEG052",
		"prefer-verb-function-names",
		"Prefer action verbs at the start of function names.",
		func(pass *analysis.Pass) (any, error) {
			checkFunctionNames(pass)
			return nil, nil
		},
	)
}

func newPreferBooleanPrefixes() ruleSpec {
	return newOptionalAnalyzer(
		"LEG053",
		"prefer-boolean-prefixes",
		"Prefer predicate or state prefixes for boolean names.",
		func(pass *analysis.Pass) (any, error) {
			checkBooleanNames(pass)
			return nil, nil
		},
	)
}

func checkFunctionNames(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			declaration, ok := node.(*ast.FuncDecl)
			if ok {
				checkFunctionName(pass, declaration)
			}

			return true
		})
	}
}

func checkFunctionName(pass *analysis.Pass, declaration *ast.FuncDecl) {
	name := declaration.Name.Name
	isAllowed := allowedFunctionName(name)
	if isAllowed {
		return
	}

	startsWithNoun := startsWithGenericNoun(name)
	if !startsWithNoun {
		return
	}

	report(
		pass,
		declaration.Name,
		"LEG052",
		"prefer-verb-function-names",
		"Start the function name with an action verb, such as getData.",
	)
}

func allowedFunctionName(name string) bool {
	switch name {
	case "init", "main", "New", "Open", "String", "Error", "Len", "Close", "Read", "Write":
		return true
	default:
		return false
	}
}

func startsWithGenericNoun(name string) bool {
	for _, prefix := range genericFunctionPrefixes {
		if hasWordPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func hasWordPrefix(name string, prefix string) bool {
	if !strings.HasPrefix(strings.ToLower(name), prefix) {
		return false
	}

	if len(name) == len(prefix) {
		return true
	}

	return unicode.IsUpper(rune(name[len(prefix)]))
}

func checkBooleanNames(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			checkBooleanNode(pass, node)

			return true
		})
	}
}

func checkBooleanNode(pass *analysis.Pass, node ast.Node) {
	switch typed := node.(type) {
	case *ast.Field:
		checkBooleanField(pass, typed.Type, typed.Names)
	case *ast.ValueSpec:
		checkBooleanField(pass, typed.Type, typed.Names)
	}
}

func checkBooleanField(pass *analysis.Pass, expression ast.Expr, names []*ast.Ident) {
	if !isBoolType(expression) {
		return
	}

	for _, name := range names {
		if hasBooleanPrefix(name.Name) {
			continue
		}

		report(
			pass,
			name,
			"LEG053",
			"prefer-boolean-prefixes",
			"Use a boolean predicate or state prefix such as is, has, can, or should.",
		)
	}
}

func isBoolType(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return false
	}

	return identifier.Name == "bool"
}

func hasBooleanPrefix(name string) bool {
	lowerName := strings.ToLower(name)
	if booleanExceptions[lowerName] {
		return true
	}

	for _, prefix := range booleanPrefixes {
		if hasWordPrefix(name, prefix) {
			return true
		}
	}

	return false
}
