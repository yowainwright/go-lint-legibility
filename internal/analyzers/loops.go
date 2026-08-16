package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type collectionKind uint8

const (
	unknownCollection collectionKind = iota
	arrayOrSliceCollection
	stringCollection
)

func newPreferRangeLoop() ruleSpec {
	return newInspectingAnalyzer(
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
	loopTypes := []ast.Node{(*ast.ForStmt)(nil)}
	inspectCursors(pass, loopTypes, func(cursor syntaxCursor) {
		statement := cursor.Node().(*ast.ForStmt)
		checkIndexLoop(pass, statement, enclosingFunction(cursor))
	})
}

func checkIndexLoop(pass *analysis.Pass, stmt *ast.ForStmt, function ast.Node) {
	if !shouldPreferRange(stmt, function) {
		return
	}

	report(
		pass,
		stmt,
		"LEG047",
		"prefer-range-loop",
		"Use an index-only range clause for this stable array or slice.",
	)
}

func shouldPreferRange(stmt *ast.ForStmt, function ast.Node) bool {
	collection, ok := indexLoopCollectionName(stmt)
	if !ok {
		return false
	}

	if function == nil {
		return false
	}

	return rangeCollectionIsStable(function, stmt, collection)
}

func rangeCollectionIsStable(function ast.Node, stmt *ast.ForStmt, collection string) bool {
	kind := functionParameterCollectionKind(function, collection)
	if kind != arrayOrSliceCollection {
		return false
	}

	collectionDeclarations := declarationCountBefore(function, collection, stmt.Pos())
	if collectionDeclarations != 1 {
		return false
	}

	lenDeclarations := declarationCountBefore(function, "len", stmt.Pos())
	if lenDeclarations != 0 {
		return false
	}

	return !loopChangesCollection(stmt.Body, collection)
}

func indexLoopCollectionName(stmt *ast.ForStmt) (string, bool) {
	counter, ok := loopCounterName(stmt.Init)
	if !ok {
		return "", false
	}

	collection, ok := loopLengthCollectionName(stmt.Cond, counter)
	if !ok {
		return "", false
	}

	return collection, loopIncrementsCounter(stmt.Post, counter)
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

func loopLengthCollectionName(cond ast.Expr, counter string) (string, bool) {
	binary, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return "", false
	}

	if binary.Op != token.LSS {
		return "", false
	}

	if !isIdentNamed(binary.X, counter) {
		return "", false
	}

	return lenArgumentName(binary.Y)
}

func lenArgumentName(expression ast.Expr) (string, bool) {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return "", false
	}

	if len(call.Args) != 1 {
		return "", false
	}

	function, ok := call.Fun.(*ast.Ident)
	if !ok {
		return "", false
	}
	if !isBuiltinLen(function) {
		return "", false
	}

	return singleIdentName(call.Args)
}

func isBuiltinLen(identifier *ast.Ident) bool {
	if identifier.Name != "len" {
		return false
	}

	return identifier.Obj == nil
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

func functionParameterCollectionKind(function ast.Node, name string) collectionKind {
	funcType := functionType(function)
	if funcType == nil {
		return unknownCollection
	}
	if funcType.Params == nil {
		return unknownCollection
	}

	for _, field := range funcType.Params.List {
		if identifiersContainName(field.Names, name) {
			return collectionTypeKind(field.Type)
		}
	}

	return unknownCollection
}

func collectionTypeKind(expression ast.Expr) collectionKind {
	switch typed := expression.(type) {
	case *ast.ArrayType:
		return arrayOrSliceCollection
	case *ast.Ident:
		if typed.Name == "string" {
			return stringCollection
		}
	case *ast.ParenExpr:
		return collectionTypeKind(typed.X)
	}

	return unknownCollection
}

func loopChangesCollection(body *ast.BlockStmt, name string) bool {
	changed := false
	ast.Inspect(body, func(node ast.Node) bool {
		if changed {
			return false
		}

		changed = nodeChangesCollection(node, name)
		return !changed
	})

	return changed
}

func nodeChangesCollection(node ast.Node, name string) bool {
	switch typed := node.(type) {
	case *ast.AssignStmt:
		return expressionsContainName(typed.Lhs, name)
	case *ast.ValueSpec:
		return identifiersContainName(typed.Names, name)
	case *ast.RangeStmt:
		return rangeChangesCollection(typed, name)
	case *ast.UnaryExpr:
		return takesCollectionAddress(typed, name)
	case *ast.CallExpr:
		return true
	default:
		return false
	}
}

func rangeChangesCollection(stmt *ast.RangeStmt, name string) bool {
	if stmt.Tok != token.DEFINE {
		return false
	}

	return rangeContainsName(stmt, name)
}

func takesCollectionAddress(expression *ast.UnaryExpr, name string) bool {
	if expression.Op != token.AND {
		return false
	}

	return isIdentNamed(expression.X, name)
}

func expressionsContainName(expressions []ast.Expr, name string) bool {
	for _, expression := range expressions {
		if isIdentNamed(expression, name) {
			return true
		}
	}

	return false
}

func identifiersContainName(identifiers []*ast.Ident, name string) bool {
	for _, identifier := range identifiers {
		if identifier.Name == name {
			return true
		}
	}

	return false
}

func rangeContainsName(stmt *ast.RangeStmt, name string) bool {
	return isIdentNamed(stmt.Key, name) || isIdentNamed(stmt.Value, name)
}
