package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type operatorMode int

type operatorLimit struct {
	max     int
	mode    operatorMode
	code    string
	rule    string
	message string
}

const (
	readabilityOperators operatorMode = iota
	booleanOperators
	computedValueOperators
)

func newMaxExpressionOperators(settings Settings) ruleSpec {
	max := settings.maxExpressionOperators()
	return newAnalyzer(
		"LEG001",
		"max-expression-operators",
		"Limit operators inside a single expression.",
		func(pass *analysis.Pass) (any, error) {
			checkExpressionContexts(pass, max, readabilityOperators)
			return nil, nil
		},
	)
}

func newHoistIfOperators(settings Settings) ruleSpec {
	max := settings.maxIfOperators()
	return newAnalyzer(
		"LEG002",
		"hoist-if-operators",
		"Prefer named booleans before operator-heavy conditions.",
		func(pass *analysis.Pass) (any, error) {
			checkIfConditions(pass, max)
			return nil, nil
		},
	)
}

func newNoComputedValues(settings Settings) ruleSpec {
	max := settings.maxComputedValueOperators()
	return newAnalyzer(
		"LEG012",
		"no-computed-values",
		"Prefer named values before returning computed expressions.",
		func(pass *analysis.Pass) (any, error) {
			checkComputedValues(pass, max)
			return nil, nil
		},
	)
}

func checkExpressionContexts(pass *analysis.Pass, max int, mode operatorMode) {
	seen := make(map[ast.Expr]bool)
	counts := make(map[ast.Expr]int)
	limit := expressionOperatorLimit(max, mode)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			for _, expression := range expressionContexts(node) {
				checkCachedOperatorLimit(pass, expression, seen, counts, limit)
			}
			return true
		})
	}
}

func expressionOperatorLimit(max int, mode operatorMode) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    mode,
		code:    "LEG001",
		rule:    "max-expression-operators",
		message: "Expression has too many operators. Extract named values.",
	}
}

func expressionContexts(node ast.Node) []ast.Expr {
	switch typed := node.(type) {
	case *ast.ReturnStmt:
		return typed.Results
	case *ast.AssignStmt:
		return typed.Rhs
	case *ast.ValueSpec:
		return typed.Values
	case *ast.IfStmt:
		return []ast.Expr{typed.Cond}
	case *ast.ForStmt:
		return []ast.Expr{typed.Cond}
	case *ast.CallExpr:
		return typed.Args
	default:
		return nil
	}
}

func checkIfConditions(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkIfConditionNode(pass, node, max)
			return true
		})
	}
}

func checkIfConditionNode(pass *analysis.Pass, node ast.Node, max int) {
	stmt, ok := node.(*ast.IfStmt)
	if !ok {
		return
	}

	seen := make(map[ast.Expr]bool)
	limit := ifOperatorLimit(max)
	checkOperatorLimit(pass, stmt.Cond, seen, limit)
}

func ifOperatorLimit(max int) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    booleanOperators,
		code:    "LEG002",
		rule:    "hoist-if-operators",
		message: "If condition has too many boolean operators. Hoist it into a named boolean.",
	}
}

func checkComputedValues(pass *analysis.Pass, max int) {
	seen := make(map[ast.Expr]bool)
	counts := make(map[ast.Expr]int)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkComputedNode(pass, node, max, seen, counts)
			return true
		})
	}
}

func checkComputedNode(
	pass *analysis.Pass,
	node ast.Node,
	max int,
	seen map[ast.Expr]bool,
	counts map[ast.Expr]int,
) {
	switch typed := node.(type) {
	case *ast.ReturnStmt:
		checkComputedExpressions(pass, typed.Results, max, seen, counts)
	case *ast.CompositeLit:
		checkCompositeValues(pass, typed.Elts, max, seen, counts)
	}
}

func checkCompositeValues(
	pass *analysis.Pass,
	expressions []ast.Expr,
	max int,
	seen map[ast.Expr]bool,
	counts map[ast.Expr]int,
) {
	for _, expression := range expressions {
		value := compositeValue(expression)
		checkComputedExpression(pass, value, max, seen, counts)
	}
}

func checkComputedExpressions(
	pass *analysis.Pass,
	expressions []ast.Expr,
	max int,
	seen map[ast.Expr]bool,
	counts map[ast.Expr]int,
) {
	for _, expression := range expressions {
		checkComputedExpression(pass, expression, max, seen, counts)
	}
}

func checkComputedExpression(
	pass *analysis.Pass,
	expression ast.Expr,
	max int,
	seen map[ast.Expr]bool,
	counts map[ast.Expr]int,
) {
	limit := computedOperatorLimit(max)
	checkCachedOperatorLimit(pass, expression, seen, counts, limit)
}

func computedOperatorLimit(max int) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    computedValueOperators,
		code:    "LEG012",
		rule:    "no-computed-values",
		message: "Computed value has too many operators. Extract it into a named value.",
	}
}

func compositeValue(expression ast.Expr) ast.Expr {
	keyValue, ok := expression.(*ast.KeyValueExpr)
	if !ok {
		return expression
	}

	return keyValue.Value
}

func checkOperatorLimit(
	pass *analysis.Pass,
	expression ast.Expr,
	seen map[ast.Expr]bool,
	limit operatorLimit,
) {
	if expressionAlreadyChecked(expression, seen) {
		return
	}

	seen[expression] = true
	reportOperatorLimit(pass, expression, countOperators(expression, limit.mode), limit)
}

func checkCachedOperatorLimit(
	pass *analysis.Pass,
	expression ast.Expr,
	seen map[ast.Expr]bool,
	counts map[ast.Expr]int,
	limit operatorLimit,
) {
	if expressionAlreadyChecked(expression, seen) {
		return
	}

	seen[expression] = true
	count := countOperatorsWithCache(expression, limit.mode, counts)
	reportOperatorLimit(pass, expression, count, limit)
}

func reportOperatorLimit(pass *analysis.Pass, expression ast.Expr, count int, limit operatorLimit) {
	if count <= limit.max {
		return
	}

	report(pass, expression, limit.code, limit.rule, limit.message)
}

func expressionAlreadyChecked(expression ast.Expr, seen map[ast.Expr]bool) bool {
	if expression == nil {
		return true
	}

	return seen[expression]
}

func countOperators(expression ast.Expr, mode operatorMode) int {
	count := 0
	ast.Inspect(expression, func(node ast.Node) bool {
		if _, isFunctionLiteral := node.(*ast.FuncLit); isFunctionLiteral {
			return false
		}

		count += operatorWeight(node, mode)
		return true
	})
	return count
}

func countOperatorsWithCache(
	expression ast.Expr,
	mode operatorMode,
	counts map[ast.Expr]int,
) int {
	if count, found := counts[expression]; found {
		return count
	}
	if _, isFunctionLiteral := expression.(*ast.FuncLit); isFunctionLiteral {
		counts[expression] = 0
		return 0
	}

	var stack []operatorCountFrame
	ast.Inspect(expression, func(node ast.Node) bool {
		return accumulateOperatorCount(node, mode, &stack, counts)
	})

	return counts[expression]
}

type operatorCountFrame struct {
	node  ast.Node
	count int
}

func accumulateOperatorCount(
	node ast.Node,
	mode operatorMode,
	stack *[]operatorCountFrame,
	counts map[ast.Expr]int,
) bool {
	if node == nil {
		finishOperatorCount(stack, counts)
		return true
	}

	if shouldSkipOperatorCountNode(node, stack, counts) {
		return false
	}

	pushOperatorCountFrame(node, mode, stack)
	return true
}

func pushOperatorCountFrame(node ast.Node, mode operatorMode, stack *[]operatorCountFrame) {
	*stack = append(*stack, operatorCountFrame{
		node:  node,
		count: operatorWeight(node, mode),
	})
}

func shouldSkipOperatorCountNode(
	node ast.Node,
	stack *[]operatorCountFrame,
	counts map[ast.Expr]int,
) bool {
	if addCachedOperatorCount(node, stack, counts) {
		return true
	}

	return cacheFunctionLiteralCount(node, counts)
}

func addCachedOperatorCount(
	node ast.Node,
	stack *[]operatorCountFrame,
	counts map[ast.Expr]int,
) bool {
	expression, ok := node.(ast.Expr)
	if !ok {
		return false
	}

	count, found := counts[expression]
	if !found {
		return false
	}
	addOperatorCount(stack, count)
	return true
}

func cacheFunctionLiteralCount(node ast.Node, counts map[ast.Expr]int) bool {
	functionLiteral, ok := node.(*ast.FuncLit)
	if !ok {
		return false
	}

	counts[functionLiteral] = 0
	return true
}

func finishOperatorCount(stack *[]operatorCountFrame, counts map[ast.Expr]int) {
	index := len(*stack) - 1
	frame := (*stack)[index]
	*stack = (*stack)[:index]

	if expression, ok := frame.node.(ast.Expr); ok {
		counts[expression] = frame.count
	}
	addOperatorCount(stack, frame.count)
}

func addOperatorCount(stack *[]operatorCountFrame, count int) {
	if len(*stack) == 0 {
		return
	}

	index := len(*stack) - 1
	(*stack)[index].count += count
}

func operatorWeight(node ast.Node, mode operatorMode) int {
	switch typed := node.(type) {
	case *ast.BinaryExpr:
		return binaryOperatorWeight(typed.Op, mode)
	case *ast.UnaryExpr:
		return unaryOperatorWeight(typed.Op, mode)
	default:
		return 0
	}
}

func binaryOperatorWeight(operator token.Token, mode operatorMode) int {
	if mode == booleanOperators {
		return logicalOperatorWeight(operator)
	}

	if isReadabilityOperator(operator) {
		return 1
	}

	if mode != computedValueOperators {
		return 0
	}

	if isArithmeticOperator(operator) {
		return 1
	}

	return 0
}

func unaryOperatorWeight(operator token.Token, mode operatorMode) int {
	if mode == booleanOperators {
		return 0
	}

	if operator == token.NOT {
		return 1
	}

	return 0
}

func logicalOperatorWeight(operator token.Token) int {
	isLogicalOperator := operator == token.LAND || operator == token.LOR
	if isLogicalOperator {
		return 1
	}

	return 0
}

func isReadabilityOperator(operator token.Token) bool {
	isLogicalOperator := logicalOperatorWeight(operator) == 1
	if isLogicalOperator {
		return true
	}

	return isComparisonOperator(operator)
}

func isComparisonOperator(operator token.Token) bool {
	switch operator {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		return true
	default:
		return false
	}
}

func isArithmeticOperator(operator token.Token) bool {
	switch operator {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		return true
	default:
		return false
	}
}
