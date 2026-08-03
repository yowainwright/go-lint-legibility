package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func newNoRedundantBreak() ruleSpec {
	return newAnalyzer(
		"LEG048",
		"no-redundant-break",
		"Avoid a trailing break in a case clause.",
		func(pass *analysis.Pass) (any, error) {
			checkRedundantBreaks(pass)
			return nil, nil
		},
	)
}

func checkRedundantBreaks(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkRedundantBreakNode(pass, node)
			return true
		})
	}
}

func checkRedundantBreakNode(pass *analysis.Pass, node ast.Node) {
	statement, ok := node.(ast.Stmt)
	if !ok {
		return
	}

	body, ok := caseClauseBody(statement)
	if !ok {
		return
	}

	checkCaseClauseBreak(pass, body)
}

func checkCaseClauseBreak(pass *analysis.Pass, body []ast.Stmt) {
	if len(body) == 0 {
		return
	}

	last := body[len(body)-1]
	if !isPlainBreak(last) {
		return
	}

	report(
		pass,
		last,
		"LEG048",
		"no-redundant-break",
		"Case clauses already break. Remove this statement.",
	)
}

func isPlainBreak(statement ast.Stmt) bool {
	branch, ok := statement.(*ast.BranchStmt)
	if !ok {
		return false
	}

	if branch.Tok != token.BREAK {
		return false
	}

	return branch.Label == nil
}
