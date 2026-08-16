package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newMaxArrayChainDepth(settings Settings) ruleSpec {
	max := settings.maxArrayChainDepth()
	return newInspectingAnalyzer(
		"LEG011",
		"max-array-chain-depth",
		"Limit consecutive collection-style method chains.",
		func(pass *analysis.Pass) (any, error) {
			checkCallChains(pass, max)
			return nil, nil
		},
	)
}

func checkCallChains(pass *analysis.Pass, max int) {
	callTypes := []ast.Node{(*ast.CallExpr)(nil)}
	inspectCursors(pass, callTypes, func(cursor syntaxCursor) {
		call := cursor.Node().(*ast.CallExpr)
		checkCallChain(pass, parentNode(cursor), call, max)
	})
}

func checkCallChain(
	pass *analysis.Pass,
	parent ast.Node,
	call *ast.CallExpr,
	max int,
) {
	if isNestedChainCall(parent) {
		return
	}

	depth := callChainDepth(call)
	if depth <= max {
		return
	}

	reportCallChain(pass, call)
}

func reportCallChain(pass *analysis.Pass, call *ast.CallExpr) {
	report(
		pass,
		call,
		"LEG011",
		"max-array-chain-depth",
		"Method chain is too deep. Split it into named intermediate values.",
	)
}

func isNestedChainCall(parent ast.Node) bool {
	_, ok := parent.(*ast.SelectorExpr)
	return ok
}

func callChainDepth(expression ast.Expr) int {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return 0
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return 0
	}

	return 1 + callChainDepth(selector.X)
}
