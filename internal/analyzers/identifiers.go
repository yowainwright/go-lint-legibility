package analyzers

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var goInitialisms = map[string]bool{
	"ACL":   true,
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"RPC":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UID":   true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"UUID":  true,
	"XML":   true,
	"XSS":   true,
}

var getterPrefixes = []string{"Get", "get"}

func newNoGetterPrefix() ruleSpec {
	return newAnalyzer(
		"LEG042",
		"no-getter-prefix",
		"Avoid the Get prefix on accessor names.",
		func(pass *analysis.Pass) (any, error) {
			checkGetterPrefixes(pass)
			return nil, nil
		},
	)
}

func newNoUnderscoreNames() ruleSpec {
	return newAnalyzer(
		"LEG043",
		"no-underscore-names",
		"Prefer MixedCaps over underscores in declared names.",
		func(pass *analysis.Pass) (any, error) {
			checkUnderscoreNames(pass)
			return nil, nil
		},
	)
}

func newPreferInitialismCasing() ruleSpec {
	return newAnalyzer(
		"LEG044",
		"prefer-initialism-casing",
		"Keep initialisms fully capitalized in declared names.",
		func(pass *analysis.Pass) (any, error) {
			checkInitialismCasing(pass)
			return nil, nil
		},
	)
}

func checkGetterPrefixes(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if ok {
				checkGetterName(pass, decl)
			}

			return true
		})
	}
}

func checkGetterName(pass *analysis.Pass, decl *ast.FuncDecl) {
	if !hasGetterPrefix(decl.Name.Name) {
		return
	}

	if !isGetterSignature(decl.Type) {
		return
	}

	report(
		pass,
		decl.Name,
		"LEG042",
		"no-getter-prefix",
		"Drop the Get prefix; name the accessor after the value it returns.",
	)
}

func hasGetterPrefix(name string) bool {
	for _, prefix := range getterPrefixes {
		remainder, found := strings.CutPrefix(name, prefix)
		if found {
			return startsUppercase(remainder)
		}
	}

	return false
}

func isGetterSignature(funcType *ast.FuncType) bool {
	if fieldCount(funcType.Params) != 0 {
		return false
	}

	return fieldCount(funcType.Results) > 0
}

func checkUnderscoreNames(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if skipsNameChecks(pass, file) {
			continue
		}

		inspectDeclaredNames(file, func(identifier *ast.Ident) {
			checkUnderscoreName(pass, identifier)
		})
	}
}

func checkUnderscoreName(pass *analysis.Pass, identifier *ast.Ident) {
	if !strings.Contains(identifier.Name, "_") {
		return
	}

	report(
		pass,
		identifier,
		"LEG043",
		"no-underscore-names",
		"Write multiword names as MixedCaps instead of using underscores.",
	)
}

func checkInitialismCasing(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}

		inspectDeclaredNames(file, func(identifier *ast.Ident) {
			checkInitialismName(pass, identifier)
		})
	}
}

func checkInitialismName(pass *analysis.Pass, identifier *ast.Ident) {
	initialism, found := misusedInitialism(identifier.Name)
	if !found {
		return
	}

	report(
		pass,
		identifier,
		"LEG044",
		"prefer-initialism-casing",
		"Capitalize the initialism as "+initialism+".",
	)
}

func misusedInitialism(name string) (string, bool) {
	for _, word := range identifierWords(name) {
		initialism, found := initialismCasingFix(word)
		if found {
			return initialism, true
		}
	}

	return "", false
}

func initialismCasingFix(word string) (string, bool) {
	if !startsUppercase(word) {
		return "", false
	}

	upper := strings.ToUpper(word)
	if !goInitialisms[upper] {
		return "", false
	}

	if word == upper {
		return "", false
	}

	return upper, true
}

func identifierWords(name string) []string {
	words := make([]string, 0, len(name))
	start := 0
	for index, char := range name {
		if !startsNewWord(index, char) {
			continue
		}

		words = append(words, name[start:index])
		start = index
	}

	return append(words, name[start:])
}

func startsNewWord(index int, char rune) bool {
	if index == 0 {
		return false
	}

	return unicode.IsUpper(char)
}

func startsUppercase(text string) bool {
	if text == "" {
		return false
	}

	return unicode.IsUpper([]rune(text)[0])
}

func inspectDeclaredNames(file *ast.File, visit func(*ast.Ident)) {
	ast.Inspect(file, func(node ast.Node) bool {
		visitDeclaredNames(node, visit)
		return true
	})
}

func visitDeclaredNames(node ast.Node, visit func(*ast.Ident)) {
	for _, identifier := range declaredIdentifiers(node) {
		if identifier.Name != "_" {
			visit(identifier)
		}
	}
}

func declaredIdentifiers(node ast.Node) []*ast.Ident {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		return []*ast.Ident{typed.Name}
	case *ast.TypeSpec:
		return []*ast.Ident{typed.Name}
	case *ast.ValueSpec:
		return typed.Names
	case *ast.Field:
		return typed.Names
	case *ast.AssignStmt:
		return definedIdentifiers(typed)
	default:
		return nil
	}
}

func definedIdentifiers(stmt *ast.AssignStmt) []*ast.Ident {
	if stmt.Tok != token.DEFINE {
		return nil
	}

	identifiers := make([]*ast.Ident, 0, len(stmt.Lhs))
	for _, expression := range stmt.Lhs {
		identifiers = appendIdentifier(identifiers, expression)
	}

	return identifiers
}

func appendIdentifier(identifiers []*ast.Ident, expression ast.Expr) []*ast.Ident {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return identifiers
	}

	return append(identifiers, identifier)
}

func fieldCount(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}

	count := 0
	for _, field := range fields.List {
		count += fieldNameCount(field)
	}

	return count
}

func fieldNameCount(field *ast.Field) int {
	if len(field.Names) == 0 {
		return 1
	}

	return len(field.Names)
}

func skipsNameChecks(pass *analysis.Pass, file *ast.File) bool {
	if isTestFile(pass, file) {
		return true
	}

	return isGeneratedFile(file)
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	filename := pass.Fset.File(file.Pos()).Name()
	return strings.HasSuffix(filename, "_test.go")
}

func isGeneratedFile(file *ast.File) bool {
	for _, group := range file.Comments {
		if hasGeneratedMarker(file, group) {
			return true
		}
	}

	return false
}

func hasGeneratedMarker(file *ast.File, group *ast.CommentGroup) bool {
	isBeforePackage := group.Pos() < file.Package
	if !isBeforePackage {
		return false
	}

	return groupHasGeneratedComment(group)
}

func groupHasGeneratedComment(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if generatedCommentPattern.MatchString(comment.Text) {
			return true
		}
	}

	return false
}
