package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoNakedReturns(settings Settings) ruleSpec {
	max := settings.maxNakedReturnLines()
	return newAnalyzer(
		"LEG049",
		"no-naked-returns",
		"Avoid naked returns outside short functions.",
		func(pass *analysis.Pass) (any, error) {
			checkNakedReturns(pass, max)
			return nil, nil
		},
	)
}

func checkNakedReturns(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkFunctionNakedReturns(pass, node, max)
			return true
		})
	}
}

func checkFunctionNakedReturns(pass *analysis.Pass, node ast.Node, max int) {
	body := functionBody(node)
	if body == nil {
		return
	}

	if !hasNamedResults(node) {
		return
	}

	if functionLineCount(pass, node) <= max {
		return
	}

	reportNakedReturns(pass, body)
}

func reportNakedReturns(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		return inspectNakedReturn(pass, node)
	})
}

func inspectNakedReturn(pass *analysis.Pass, node ast.Node) bool {
	if functionBody(node) != nil {
		return false
	}

	stmt, ok := node.(*ast.ReturnStmt)
	if ok {
		checkNakedReturn(pass, stmt)
	}

	return true
}

func checkNakedReturn(pass *analysis.Pass, stmt *ast.ReturnStmt) {
	if len(stmt.Results) != 0 {
		return
	}

	report(
		pass,
		stmt,
		"LEG049",
		"no-naked-returns",
		"Return the result values explicitly in a function this long.",
	)
}

func hasNamedResults(node ast.Node) bool {
	funcType := functionType(node)
	if funcType == nil {
		return false
	}

	return hasNamedFields(funcType.Results)
}

func hasNamedFields(fields *ast.FieldList) bool {
	if fields == nil {
		return false
	}

	for _, field := range fields.List {
		if len(field.Names) > 0 {
			return true
		}
	}

	return false
}

func functionType(node ast.Node) *ast.FuncType {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		return typed.Type
	case *ast.FuncLit:
		return typed.Type
	default:
		return nil
	}
}
