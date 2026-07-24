package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func newPreferRangeLoop() ruleSpec {
	return newAnalyzer(
		"LEG047",
		"prefer-range-loop",
		"Prefer a range clause over an index counter loop.",
		func(pass *analysis.Pass) (any, error) {
			checkIndexLoops(pass)
			return nil, nil
		},
	)
}

func checkIndexLoops(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.ForStmt)
			if ok {
				checkIndexLoop(pass, stmt)
			}

			return true
		})
	}
}

func checkIndexLoop(pass *analysis.Pass, stmt *ast.ForStmt) {
	if !isIndexCounterLoop(stmt) {
		return
	}

	report(
		pass,
		stmt,
		"LEG047",
		"prefer-range-loop",
		"Use a range clause instead of an index counter.",
	)
}

func isIndexCounterLoop(stmt *ast.ForStmt) bool {
	counter, ok := loopCounterName(stmt.Init)
	if !ok {
		return false
	}

	if !loopComparesLength(stmt.Cond, counter) {
		return false
	}

	return loopIncrementsCounter(stmt.Post, counter)
}

func loopCounterName(init ast.Stmt) (string, bool) {
	assign, ok := init.(*ast.AssignStmt)
	if !ok {
		return "", false
	}

	if assign.Tok != token.DEFINE {
		return "", false
	}

	if !isZeroLiteralValue(assign.Rhs) {
		return "", false
	}

	return singleIdentName(assign.Lhs)
}

func singleIdentName(expressions []ast.Expr) (string, bool) {
	if len(expressions) != 1 {
		return "", false
	}

	identifier, ok := expressions[0].(*ast.Ident)
	if !ok {
		return "", false
	}

	return identifier.Name, true
}

func isZeroLiteralValue(expressions []ast.Expr) bool {
	if len(expressions) != 1 {
		return false
	}

	literal, ok := expressions[0].(*ast.BasicLit)
	if !ok {
		return false
	}

	return literal.Value == "0"
}

func loopComparesLength(cond ast.Expr, counter string) bool {
	binary, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}

	if binary.Op != token.LSS {
		return false
	}

	if !isIdentNamed(binary.X, counter) {
		return false
	}

	return isLenCall(binary.Y)
}

func isLenCall(expression ast.Expr) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return false
	}

	if len(call.Args) != 1 {
		return false
	}

	return isIdentNamed(call.Fun, "len")
}

func loopIncrementsCounter(post ast.Stmt, counter string) bool {
	stmt, ok := post.(*ast.IncDecStmt)
	if !ok {
		return false
	}

	if stmt.Tok != token.INC {
		return false
	}

	return isIdentNamed(stmt.X, counter)
}

func isIdentNamed(expression ast.Expr, name string) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return false
	}

	return identifier.Name == name
}
