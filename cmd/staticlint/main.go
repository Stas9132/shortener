// Package staticlint ...
package staticlint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// Analyzer ...
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "report calls to os.Exit in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			// find main function
			var fn *ast.FuncDecl
			var ok bool
			if fn, ok = n.(*ast.FuncDecl); !ok || fn.Name.Name != "main" {
				return true
			}
			// find calls to os.Exit
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				var call *ast.CallExpr
				if call, ok = n.(*ast.CallExpr); !ok {
					return true
				}
				var sel *ast.SelectorExpr
				if sel, ok = call.Fun.(*ast.SelectorExpr); !ok {
					return true
				}
				if id, ok := sel.X.(*ast.Ident); !ok || id.Name != "os" || sel.Sel.Name != "Exit" {
					return true
				}
				// report diagnostic
				pass.Reportf(call.Pos(), "do not use os.Exit in main function")
				return true
			})
			return true
		})
	}
	return nil, nil
}
