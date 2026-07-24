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

var errorStringConstructors = map[string]bool{
	"errors.New": true,
	"fmt.Errorf": true,
}

func newPreferLowercaseErrorStrings() ruleSpec {
	return newAnalyzer(
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
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				checkErrorStringCall(pass, call)
			}

			return true
		})
	}
}

func checkErrorStringCall(pass *analysis.Pass, call *ast.CallExpr) {
	literal, ok := errorStringLiteral(call)
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

func errorStringLiteral(call *ast.CallExpr) (*ast.BasicLit, bool) {
	if !isErrorStringConstructor(call.Fun) {
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

func isErrorStringConstructor(fun ast.Expr) bool {
	selector, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	qualified := pkg.Name + "." + selector.Sel.Name
	return errorStringConstructors[qualified]
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
