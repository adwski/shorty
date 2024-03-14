// Package multichecker provides static code analyzer for shorty project.
//
// Analyzer uses following checks:
// - almost all checks from staticchec.io package (including stylechecks and quickfixes)
// - loggercheck https://github.com/timonwong/loggercheck
// - interfacebloat https://github.com/sashamelentyev/interfacebloat
// - exitcheck (local package)
//
// All checks comply to go/analysis pattern (i.e. implemented as analysis.Analyzer)
// and bundled together with go/analysis/multichecker.
//
// Each check is also wrapped with nolint wrapper to support ignoring specific files with comments,
// since multichecker itself provides no customizations to configure such behaviour.
// This is useful for generated code.
package multichecker
