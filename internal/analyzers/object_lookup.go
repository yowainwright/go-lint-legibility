package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type lookupPart struct {
	key  string
	node ast.Node
}

func newPreferObjectLookup(settings Settings) ruleSpec {
	min := settings.minObjectLookupChainLength()
	return newInspectingAnalyzer(
		"LEG024",
		"prefer-object-lookup",
		"Prefer set or map lookups over long equality-or chains.",
		func(pass *analysis.Pass) (any, error) {
			checkObjectLookup(pass, min)
			return nil, nil
		},
	)
}

func checkObjectLookup(pass *analysis.Pass, min int) {
	binaryTypes := []ast.Node{(*ast.BinaryExpr)(nil)}
	inspectCursors(pass, binaryTypes, func(cursor syntaxCursor) {
		expression := cursor.Node().(*ast.BinaryExpr)
		checkObjectLookupExpression(pass, parentNode(cursor), expression, min)
	})
}

func checkObjectLookupExpression(
	pass *analysis.Pass,
	parent ast.Node,
	expression *ast.BinaryExpr,
	min int,
) {
	if !isTopLevelOrExpression(parent, expression) {
		return
	}

	parts := collectLookupParts(pass, expression)
	if !hasObjectLookupParts(parts, min) {
		return
	}

	reportObjectLookup(pass, expression)
}

func isTopLevelOrExpression(parent ast.Node, expression *ast.BinaryExpr) bool {
	if expression.Op != token.LOR {
		return false
	}

	return !isNestedOr(parent)
}

func hasObjectLookupParts(parts []lookupPart, min int) bool {
	hasEnoughParts := len(parts) >= min
	if !hasEnoughParts {
		return false
	}

	return sameLookupKey(parts)
}

func reportObjectLookup(pass *analysis.Pass, expression ast.Expr) {
	report(
		pass,
		expression,
		"LEG024",
		"prefer-object-lookup",
		"Use a set, map, or switch instead of a long equality-or chain.",
	)
}

func collectLookupParts(pass *analysis.Pass, expression ast.Expr) []lookupPart {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	if binary.Op == token.LOR {
		left := collectLookupParts(pass, binary.X)
		return append(left, collectLookupParts(pass, binary.Y)...)
	}

	return equalityLookupPart(pass, binary)
}

func equalityLookupPart(pass *analysis.Pass, expression *ast.BinaryExpr) []lookupPart {
	isEquality := expression.Op == token.EQL
	isInequality := expression.Op == token.NEQ
	isComparison := isEquality || isInequality
	if !isComparison {
		return nil
	}

	if isLiteral(expression.Y) {
		return []lookupPart{{key: nodeText(pass, expression.X), node: expression}}
	}

	if isLiteral(expression.X) {
		return []lookupPart{{key: nodeText(pass, expression.Y), node: expression}}
	}

	return nil
}

func isNestedOr(parent ast.Node) bool {
	parentExpression, ok := parent.(*ast.BinaryExpr)
	if !ok {
		return false
	}

	return parentExpression.Op == token.LOR
}

func sameLookupKey(parts []lookupPart) bool {
	if len(parts) == 0 {
		return false
	}

	first := parts[0].key
	for _, part := range parts {
		if part.key != first {
			return false
		}
	}

	return true
}

func isLiteral(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		return isLiteralIdentifier(typed.Name)
	default:
		return false
	}
}

func isLiteralIdentifier(name string) bool {
	switch name {
	case "true", "false", "nil":
		return true
	default:
		return false
	}
}
