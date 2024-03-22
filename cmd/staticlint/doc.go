/*
Staticlint is a custom linter for shorty project. It's built using go/analysis/multichecker
and consists of various linters.

# Usage

In simplest case it can be called just with usual package path argument

	$ go run cmd/staticlint/main.go ./...

Specific checks can be selected by specifying their name as argument

	$ go run cmd/staticlint/main.go -exitcheck ./...

All existing arguments (provided by multichecker) can be shown with

	$ go run cmd/staticlint/main.go -h

# Analyzers

Majority of analyzers are from staticcheck project:
  - Standard library checks (SA1 group)
  - Concurrency checks (SA2 group)
  - Test checks (SA3 group)
  - Useless code (SA4 group)
  - Correctness checks (SA5 group)
  - Performance checks (SA6 groups)
  - Dubious code (SA9 group)

Also included additional staticchecks analyzers:
  - Code simplifications (S1 group)
  - Style checks (ST group)
  - Quickfixes (QF group)

You can see details in the official doc https://staticcheck.dev/docs/checks/

Additional analyzers (that not part of staticcheck) are:

  - loggercheck (https://github.com/timonwong/loggercheck)
    Prevents misuse of loggers, including zap.

  - interfacebloat (https://github.com/sashamelentyev/interfacebloat)
    Detects interfaces that are too big, which is considered as anti-pattern.

# Exitcheck

Staticlint also includes custom analyzer that forbids direct os.Exit() usage
inside main function of main package.

# Nolint

Staticlint has an ability to skip entire files with "//lint:file-ignore ..." comment.
To be detected, this comment should be present in first comment group.

Skiping specific lines and/or checks is not supported.
*/
package main
