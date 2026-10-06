package analyzers

import (
	"go/ast"
	"go/parser"
	"testing"
)

func TestCachedOperatorCountsMatchUncachedCounts(t *testing.T) {
	source := `
outer(
	inner(a == b && c != d),
	func() bool { return a < b && c < d },
)
`
	expression, err := parser.ParseExpr(source)
	if err != nil {
		t.Fatal(err)
	}

	expressions := expressionsIn(expression)
	modes := []operatorMode{readabilityOperators, booleanOperators, computedValueOperators}
	for _, mode := range modes {
		assertOperatorCountParity(t, expressions, mode)
	}
}

func expressionsIn(root ast.Expr) []ast.Expr {
	var expressions []ast.Expr
	ast.Inspect(root, func(node ast.Node) bool {
		if expression, ok := node.(ast.Expr); ok {
			expressions = append(expressions, expression)
		}
		return true
	})
	return expressions
}

func assertOperatorCountParity(t *testing.T, expressions []ast.Expr, mode operatorMode) {
	t.Helper()
	counts := make(map[ast.Expr]int)
	for _, expression := range expressions {
		want := countOperators(expression, mode)
		got := countOperatorsWithCache(expression, mode, counts)
		if got != want {
			t.Errorf("cached count for %T = %d, want %d", expression, got, want)
		}
	}
}
