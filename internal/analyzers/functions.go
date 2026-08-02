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
	return newAnalyzer(
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
	parents := buildParentMap(pass.Files)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			funcType, ok := node.(*ast.FuncType)
			if ok {
				checkFunctionParamBudget(pass, funcType, max, parents)
			}

			return true
		})
	}
}

func checkFunctionParamBudget(
	pass *analysis.Pass,
	funcType *ast.FuncType,
	max int,
	parents map[ast.Node]ast.Node,
) {
	if fieldCount(funcType.Params) <= max {
		return
	}

	report(
		pass,
		functionTypeReportNode(funcType, parents),
		"LEG051",
		"max-function-params",
		"Function takes too many parameters. Group related values into a struct.",
	)
}

func functionTypeReportNode(funcType *ast.FuncType, parents map[ast.Node]ast.Node) ast.Node {
	switch parent := parents[funcType].(type) {
	case *ast.FuncDecl:
		return parent.Name
	case *ast.Field:
		if len(parent.Names) > 0 {
			return parent.Names[0]
		}
	case *ast.TypeSpec:
		return parent.Name
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
