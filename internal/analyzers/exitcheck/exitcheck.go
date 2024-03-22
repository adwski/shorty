// Package exitcheck provides code analyzer for direct os.Exit() calls.
package exitcheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// New creates exitcheck analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "exitcheck",
		Doc:  "checks for direct os.Exit calls inside main func",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if file.Name.String() != "main" {
			continue
		}
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, ".go") {
			continue
		}
		ast.Inspect(file, inspect(pass))
	}
	return nil, nil //nolint:nilnil // looks like it's a pretty common return behaviour in analyzers.
}

func inspect(pass *analysis.Pass) func(node ast.Node) bool {
	var (
		insideMainFunc bool
	)
	return func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.ExprStmt:
			inspectExpr(x, pass, &insideMainFunc)
		case *ast.FuncDecl:
			if x.Name.String() == "main" {
				insideMainFunc = true
			} else {
				insideMainFunc = false
			}
		}
		return true
	}
}

func inspectExpr(x *ast.ExprStmt, pass *analysis.Pass, insideMainFunc *bool) {
	if !*insideMainFunc {
		return
	}
	call, ok := x.X.(*ast.CallExpr)
	if !ok {
		return
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !selector.Sel.IsExported() {
		return
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return
	}
	if selector.Sel.Name == "Exit" && pkg.String() == "os" {
		pass.Reportf(call.Pos(), "direct os.Exit call in main func")
	}
}
