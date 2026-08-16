package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func newPreferSwitchOverLongIfChain(settings Settings) ruleSpec {
	min := settings.minSwitchChainLength()
	return newInspectingAnalyzer(
		"LEG034",
		"prefer-switch-over-long-if-chain",
		"Prefer switch over long if chains that compare the same value.",
		func(pass *analysis.Pass) (any, error) {
			checkSwitchableIfChains(pass, min)
			return nil, nil
		},
	)
}

func newNoComplexIfInit(settings Settings) ruleSpec {
	max := settings.maxIfInitOperators()
	return newAnalyzer(
		"LEG036",
		"no-complex-if-init",
		"Avoid combining an if initializer with an operator-heavy condition.",
		func(pass *analysis.Pass) (any, error) {
			checkComplexIfInit(pass, max)
			return nil, nil
		},
	)
}

func checkSwitchableIfChains(pass *analysis.Pass, min int) {
	ifTypes := []ast.Node{(*ast.IfStmt)(nil)}
	inspectCursors(pass, ifTypes, func(cursor syntaxCursor) {
		statement := cursor.Node().(*ast.IfStmt)
		checkSwitchableIfChain(pass, parentNode(cursor), statement, min)
	})
}

func checkSwitchableIfChain(
	pass *analysis.Pass,
	parent ast.Node,
	stmt *ast.IfStmt,
	min int,
) {
	if isElseIf(parent, stmt) {
		return
	}

	if !hasSwitchableIfChainSubjects(pass, stmt, min) {
		return
	}

	reportSwitchableIfChain(pass, stmt)
}

func hasSwitchableIfChainSubjects(pass *analysis.Pass, stmt *ast.IfStmt, min int) bool {
	subjects := collectIfChainSubjects(pass, stmt)
	if len(subjects) < min {
		return false
	}

	return sameSubject(subjects)
}

func reportSwitchableIfChain(pass *analysis.Pass, stmt *ast.IfStmt) {
	report(
		pass,
		stmt,
		"LEG034",
		"prefer-switch-over-long-if-chain",
		"Use a switch for this repeated comparison chain.",
	)
}

func isElseIf(parent ast.Node, stmt *ast.IfStmt) bool {
	parentIf, ok := parent.(*ast.IfStmt)
	if !ok {
		return false
	}

	return parentIf.Else == stmt
}

func collectIfChainSubjects(pass *analysis.Pass, stmt *ast.IfStmt) []string {
	var subjects []string
	current := stmt
	for current != nil {
		subject, ok := equalitySubject(pass, current.Cond)
		if !ok {
			return subjects
		}

		subjects = append(subjects, subject)
		current, _ = current.Else.(*ast.IfStmt)
	}

	return subjects
}

func equalitySubject(pass *analysis.Pass, expression ast.Expr) (string, bool) {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return "", false
	}

	if binary.Op != token.EQL {
		return "", false
	}

	if isLiteral(binary.Y) {
		return nodeText(pass, binary.X), true
	}

	if isLiteral(binary.X) {
		return nodeText(pass, binary.Y), true
	}

	return "", false
}

func sameSubject(subjects []string) bool {
	if len(subjects) == 0 {
		return false
	}

	first := subjects[0]
	for _, subject := range subjects {
		if subject != first {
			return false
		}
	}

	return true
}

func checkComplexIfInit(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if ok {
				checkIfInitCondition(pass, stmt, max)
			}

			return true
		})
	}
}

func checkIfInitCondition(pass *analysis.Pass, stmt *ast.IfStmt, max int) {
	if stmt.Init == nil {
		return
	}

	operatorCount := countOperators(stmt.Cond, booleanOperators)
	if operatorCount <= max {
		return
	}

	report(
		pass,
		stmt.Cond,
		"LEG036",
		"no-complex-if-init",
		"Move the condition into a named boolean before this if.",
	)
}
