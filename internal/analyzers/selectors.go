package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoDeepSelectorChain(settings Settings) ruleSpec {
	max := settings.maxSelectorChainDepth()
	return newInspectingAnalyzer(
		"LEG031",
		"no-deep-selector-chain",
		"Avoid deep selector or index chains without named intermediate values.",
		func(pass *analysis.Pass) (any, error) {
			checkSelectorChains(pass, max)
			return nil, nil
		},
	)
}

func checkSelectorChains(pass *analysis.Pass, max int) {
	inspectCursors(pass, nil, func(cursor syntaxCursor) {
		expression, ok := cursor.Node().(ast.Expr)
		if ok {
			checkSelectorChain(pass, parentNode(cursor), expression, max)
		}
	})
}

func checkSelectorChain(
	pass *analysis.Pass,
	parent ast.Node,
	expression ast.Expr,
	max int,
) {
	if isNestedAccessChain(parent, expression) {
		return
	}

	depth := accessChainDepth(expression)
	if depth <= max {
		return
	}

	reportSelectorChain(pass, expression)
}

func reportSelectorChain(pass *analysis.Pass, expression ast.Expr) {
	report(
		pass,
		expression,
		"LEG031",
		"no-deep-selector-chain",
		"Selector or index chain is too deep. Extract a named intermediate value.",
	)
}

func isNestedAccessChain(parent ast.Node, expression ast.Expr) bool {
	switch typed := parent.(type) {
	case *ast.SelectorExpr:
		return typed.X == expression
	case *ast.IndexExpr:
		return typed.X == expression
	case *ast.IndexListExpr:
		return typed.X == expression
	default:
		return false
	}
}

func accessChainDepth(expression ast.Expr) int {
	switch typed := expression.(type) {
	case *ast.SelectorExpr:
		return 1 + accessChainDepth(typed.X)
	case *ast.IndexExpr:
		return 1 + accessChainDepth(typed.X)
	case *ast.IndexListExpr:
		return 1 + accessChainDepth(typed.X)
	default:
		return 0
	}
}
