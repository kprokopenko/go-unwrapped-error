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

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		funcDecl := node.(*ast.FuncDecl)

		res := funcDecl.Type.Results
		if res == nil {
			return
		}

		has := false
		for _, result := range res.List {
			if ident, ok := result.Type.(*ast.Ident); ok {
				has = ident.Name == "error"
				break
			}
		}

		if !has {
			return
		}

		positions := CountUnwrappedReturns(funcDecl.Body)
		if len(positions) < 2 {
			return
		}

		for _, pos := range positions {
			pass.Reportf(pos, "unwrapped error")
		}

		// params := funcDecl.Type.Params.List
		// if len(params) < 2 { // [0] must be format (string), [1] must be args (...interface{})
		// 	return
		// }

		// formatParamType, ok := params[len(params)-2].Type.(*ast.Ident)
		// if !ok { // first param type isn't identificator so it can't be of type "string"
		// 	return
		// }

		// if formatParamType.Name != "string" { // first param (format) type is not string
		// 	return
		// }

		// if formatParamNames := params[len(params)-2].Names; len(formatParamNames) == 0 || formatParamNames[len(formatParamNames)-1].Name != "format" {
		// 	return
		// }

		// argsParamType, ok := params[len(params)-1].Type.(*ast.Ellipsis)
		// if !ok { // args are not ellipsis (...args)
		// 	return
		// }

		// elementType, ok := argsParamType.Elt.(*ast.InterfaceType)
		// if !ok { // args are not of interface type, but we need interface{}
		// 	return
		// }

		// if elementType.Methods != nil && len(elementType.Methods.List) != 0 {
		// 	return // has >= 1 method in interface, but we need an empty interface "interface{}"
		// }

		// if strings.HasSuffix(funcDecl.Name.Name, "f") {
		// 	return
		// }

		// pass.Reportf(node.Pos(), "printf-like formatting function '%s' should be named '%sf'",
		// 	funcDecl.Name.Name, funcDecl.Name.Name)
	})

	return nil, nil
}
