package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newMaxFunctionLines(settings Settings) ruleSpec {
	max := settings.maxFunctionLines()
	return newAnalyzer(
		"LEG038",
		"max-function-lines",
		"Limit functions to a focused line budget.",
		func(pass *analysis.Pass) (any, error) {
			checkFunctionLines(pass, max)
			return nil, nil
		},
	)
}

func newMaxFunctionParams(settings Settings) ruleSpec {
	max := settings.maxFunctionParams()
	return newInspectingAnalyzer(
		"LEG051",
		"max-function-params",
		"Limit the number of parameters in a function signature.",
		func(pass *analysis.Pass) (any, error) {
			checkFunctionParams(pass, max)
			return nil, nil
		},
	)
}

func checkFunctionLines(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkFunctionLineBudget(pass, node, max)
			return true
		})
	}
}

func checkFunctionLineBudget(pass *analysis.Pass, node ast.Node, max int) {
	body := functionBody(node)
	if body == nil {
		return
	}

	lines := functionLineCount(pass, node)
	if lines <= max {
		return
	}

	report(
		pass,
		functionReportNode(node),
		"LEG038",
		"max-function-lines",
		"Function is too long. Split it into focused helpers.",
	)
}

func checkFunctionParams(pass *analysis.Pass, max int) {
	functionTypes := []ast.Node{(*ast.FuncType)(nil)}
	inspectCursors(pass, functionTypes, func(cursor syntaxCursor) {
		funcType := cursor.Node().(*ast.FuncType)
		checkFunctionParamBudget(pass, funcType, max, parentNode(cursor))
	})
}

func checkFunctionParamBudget(
	pass *analysis.Pass,
	funcType *ast.FuncType,
	max int,
	parent ast.Node,
) {
	if fieldCount(funcType.Params) <= max {
		return
	}

	report(
		pass,
		functionTypeReportNode(funcType, parent),
		"LEG051",
		"max-function-params",
		"Function takes too many parameters. Group related values into a struct.",
	)
}

func functionTypeReportNode(funcType *ast.FuncType, parent ast.Node) ast.Node {
	switch typed := parent.(type) {
	case *ast.FuncDecl:
		return typed.Name
	case *ast.Field:
		if len(typed.Names) > 0 {
			return typed.Names[0]
		}
	case *ast.TypeSpec:
		return typed.Name
	}

	return funcType
}

func functionLineCount(pass *analysis.Pass, node ast.Node) int {
	lines := nodeLineSpan(pass, node) - nestedFunctionLineSpan(pass, node)
	if lines < 1 {
		return 1
	}

	return lines
}

func nestedFunctionLineSpan(pass *analysis.Pass, node ast.Node) int {
	lines := 0
	ast.Inspect(node, func(child ast.Node) bool {
		isRootOrNil := child == nil || child == node
		if isRootOrNil {
			return true
		}

		return addNestedFunctionSpan(pass, child, &lines)
	})

	return lines
}

func addNestedFunctionSpan(pass *analysis.Pass, node ast.Node, lines *int) bool {
	if functionBody(node) == nil {
		return true
	}

	*lines += nodeLineSpan(pass, node)
	return false
}

func nodeLineSpan(pass *analysis.Pass, node ast.Node) int {
	start := pass.Fset.Position(node.Pos()).Line
	end := pass.Fset.Position(node.End()).Line
	lineSpan := end - start
	return lineSpan + 1
}

func functionReportNode(node ast.Node) ast.Node {
	decl, ok := node.(*ast.FuncDecl)
	if ok {
		return decl.Name
	}

	return node
}
