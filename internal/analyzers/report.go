package analyzers

import (
	"bytes"
	"go/ast"
	"go/printer"

	"golang.org/x/tools/go/analysis"
)

func report(pass *analysis.Pass, node ast.Node, code string, rule string, message string) {
	diagnosticMessage := code + " " + rule + ": " + message
	pass.Report(analysis.Diagnostic{
		Pos:     node.Pos(),
		End:     node.End(),
		Message: diagnosticMessage,
	})
}

func nodeText(pass *analysis.Pass, node ast.Node) string {
	var buffer bytes.Buffer
	err := printer.Fprint(&buffer, pass.Fset, node)
	if err != nil {
		return ""
	}

	return buffer.String()
}
