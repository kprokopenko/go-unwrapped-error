package analyzer

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"golang.org/x/tools/go/analysis"
)

// TODO: xxx
var Analyzer = &analysis.Analyzer{
	Name:     "goprintffuncname",
	Doc:      "Checks that printf-like functions are named with `f` at the end.",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		if !hasErrors(funcDecl.Type.Results) {
			return
		}

		unwrappedErrorReturns := CountUnwrappedReturns(funcDecl.Body)

		if len(unwrappedErrorReturns) == 1 { // one unwrapped error is not an issue
			return
		}

		for _, pos := range unwrappedErrorReturns {
			pass.Reportf(pos, "unwrapped error")
		}
	})

	return nil, nil
}

func hasErrors(results *ast.FieldList) bool {
	if results == nil {
		return false
	}

	for _, result := range results.List {
		if ident, ok := result.Type.(*ast.Ident); ok && ident.Name == "error" {
			return true
		}
	}

	return false
}

func CountUnwrappedReturns(body *ast.BlockStmt) []token.Pos {
	rets := []token.Pos{}
	for _, stmt := range body.List {
		switch expr := stmt.(type) {
		case *ast.ReturnStmt:
			if !exprHasFuncWithName(expr.Results, []string{"Errorf", "New"}) && !exprHasNil(expr.Results) {
				rets = append(rets, stmt.Pos())
			}
		case *ast.IfStmt:
			rets = append(rets, CountUnwrappedReturns(expr.Body)...)
		}
	}

	return rets
}

func exprHasFuncWithName(exprs []ast.Expr, fnnames []string) bool {
	for _, result := range exprs {
		if call, ok := result.(*ast.CallExpr); ok {
			if fnsel, ok := call.Fun.(*ast.SelectorExpr); ok {
				for _, fnname := range fnnames {
					if fnsel.Sel.Name == fnname {
						return true
					}
				}
			}
		}
	}

	return false
}

func exprHasNil(exprs []ast.Expr) bool {
	for _, result := range exprs {
		if ident, ok := result.(*ast.Ident); ok && ident.Name == "nil" {
			return true
		}
	}

	return false
}
