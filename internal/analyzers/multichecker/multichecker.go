package multichecker

import (
	"github.com/adwski/shorty/internal/analyzers/exitcheck"
	"github.com/adwski/shorty/internal/analyzers/nolint"
	interfacebloat "github.com/sashamelentyev/interfacebloat/pkg/analyzer"
	"github.com/timonwong/loggercheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Run composes all analyzers and starts multichecker.
func Run() {
	checks := append(getStaticCheckAnalyzers(),
		loggercheck.NewAnalyzer(),
		interfacebloat.New(),
		exitcheck.New())

	nolint.WrapSlice(checks)
	multichecker.Main(checks...)
}

func getStaticCheckAnalyzers() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range quickfix.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	return checks
}
