// Package nolint provides code analyzer wrapper that checks for existing
// "//lint:file-ignore ..." comment inside first comment group and skips entire file
// in case such comment exists.
//
// Approach is inspired by https://github.com/kyoh86/nolint but here it's much more simple.
package nolint

import (
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	noLintPrefix string = "//lint:file-ignore "
)

// WrapSlice wraps slice of analyzers.
func WrapSlice(as []*analysis.Analyzer) {
	for _, a := range as {
		Wrap(a)
	}
}

// Wrap overwrites analyzer's run function adding check for "//lint:file-ignore ".
// If such line exists in first comment group, it removes file from analysis.Pass.
// After comment check, usual run() logic is called if some (or all) files still remain.
func Wrap(a *analysis.Analyzer) {
	realRun := a.Run
	a.Run = func(pass *analysis.Pass) (any, error) {
		var newFiles []*ast.File
	Loop:
		for _, file := range pass.Files {
			if len(file.Comments) > 0 {
				for _, comment := range file.Comments[0].List {
					if strings.HasPrefix(comment.Text, noLintPrefix) && len(comment.Text) > len(noLintPrefix) {
						continue Loop
					}
				}
			}
			newFiles = append(newFiles, file)
		}
		if len(newFiles) == 0 {
			return reflect.TypeOf(a.ResultType), nil
		}
		pass.Files = newFiles
		return realRun(pass)
	}
}
