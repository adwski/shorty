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
		ast.Inspect(file, newInspector().inspectFunc(pass))
	}
	return nil, nil //nolint:nilnil // looks like it's a pretty common return behaviour in analyzers.
}

type inspector struct {
	insideMainFunc bool
}

func newInspector() *inspector {
	return &inspector{}
}

func (i *inspector) inspectFunc(pass *analysis.Pass) func(node ast.Node) bool {
	exprStmt := func(x *ast.ExprStmt) {
		if call, ok := x.X.(*ast.CallExpr); ok {
			if s, ok := call.Fun.(*ast.SelectorExpr); ok {
				if s.Sel.Name == "Exit" && i.insideMainFunc {
					pass.Reportf(call.Pos(), "direct os.Exit call in main func")
				}
			}
		}
	}
	funcDecl := func(x *ast.FuncDecl) {
		if x.Name.String() == "main" {
			i.insideMainFunc = true
		} else {
			i.insideMainFunc = false
		}
	}
	return func(node ast.Node) bool {
		switch x := node.(type) {
		case *ast.ExprStmt:
			exprStmt(x)
		case *ast.FuncDecl:
			funcDecl(x)
		}
		return true
	}
}
