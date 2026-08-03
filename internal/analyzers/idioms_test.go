package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestNoGetterPrefixReportsGetAccessor(t *testing.T) {
	source := readTestSource(t, "no_getter_prefix.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-getter-prefix"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG042 no-getter-prefix")
}

func TestNoUnderscoreNamesReportsUnderscoreIdentifiers(t *testing.T) {
	source := readTestSource(t, "no_underscore_names.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-underscore-names"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG043 no-underscore-names")
}

func TestNoUnderscoreNamesSkipsTestFiles(t *testing.T) {
	source := readTestSource(t, "no_underscore_names.go")

	analyzer := analyzerByRule(t, "no-underscore-names")
	diagnostics := runAnalyzer(t, analyzer, "p_test.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoUnderscoreNamesReportsRangeDeclarations(t *testing.T) {
	source := `package p
func visit(values []int) {
	for value_index := range values {
		_ = value_index
	}
}
`
	analyzer := analyzerByRule(t, "no-underscore-names")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestPreferInitialismCasingReportsMixedCaseInitialisms(t *testing.T) {
	source := readTestSource(t, "prefer_initialism_casing.go")

	analyzer := analyzerByRule(t, "prefer-initialism-casing")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "Capitalize the initialism as URL.")
	requireDiagnostic(t, diagnostics, "Capitalize the initialism as ID.")
}

func TestPreferInitialismCasingReportsRangeDeclarations(t *testing.T) {
	source := `package p
func visit(values []int) {
	for userId := range values {
		_ = userId
	}
}
`
	analyzer := analyzerByRule(t, "prefer-initialism-casing")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoPackageNameStutterReportsRepeatedPackagePrefix(t *testing.T) {
	source := readTestSource(t, "no_package_name_stutter.go")

	analyzer := analyzerByRule(t, "no-package-name-stutter")
	diagnostics := runAnalyzer(t, analyzer, "orders/orders.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
	requireDiagnostic(t, diagnostics, "LEG045 no-package-name-stutter")
}

func TestNoPackageNameStutterReportsExportedValues(t *testing.T) {
	source := `package orders
const OrdersLimit = 1
var OrdersDefault = 2
`
	analyzer := analyzerByRule(t, "no-package-name-stutter")
	diagnostics := runAnalyzer(t, analyzer, "orders/orders.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
}

func TestNoGenericPackageNamesReportsUtilPackage(t *testing.T) {
	source := readTestSource(t, "no_generic_package_names.go")

	analyzer := analyzerByRule(t, "no-generic-package-names")
	diagnostics := runAnalyzer(t, analyzer, "internal/util/util.go", source)
	requireDiagnostic(t, diagnostics, "LEG046 no-generic-package-names")
}

func TestNoGenericPackageNamesReportsOncePerPackage(t *testing.T) {
	fileSet := token.NewFileSet()
	firstSource := "package util\nfunc one() {}"
	secondSource := "package util\nfunc two() {}"
	first := parseAnalyzerFile(t, fileSet, "one.go", firstSource)
	second := parseAnalyzerFile(t, fileSet, "two.go", secondSource)
	files := []*ast.File{first, second}
	analyzer := analyzerByRule(t, "no-generic-package-names")
	diagnostics := runAnalyzerWithFiles(t, analyzer, fileSet, files)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestPreferRangeLoopReportsIndexCounterLoop(t *testing.T) {
	source := readTestSource(t, "prefer_range_loop.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-range-loop"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG047 prefer-range-loop")
}

func TestPreferRangeLoopAllowsBytewiseStringLoop(t *testing.T) {
	source := `package p
func visit(text string) {
	for index := 0; index < len(text); index++ {
		_ = text[index]
	}
}
`
	analyzer := analyzerByRule(t, "prefer-range-loop")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestPreferRangeLoopAllowsGrowingSliceLoop(t *testing.T) {
	source := `package p
func grow(values []int) {
	for index := 0; index < len(values); index++ {
		values = append(values, index)
	}
}
`
	analyzer := analyzerByRule(t, "prefer-range-loop")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestPreferRangeLoopAllowsLoopThatCallsClosure(t *testing.T) {
	source := `package p
func grow(values []int) {
	growValues := func() { values = append(values, 1) }
	for index := 0; index < len(values); index++ {
		growValues()
	}
}
`
	analyzer := analyzerByRule(t, "prefer-range-loop")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoRedundantBreakReportsTrailingBreak(t *testing.T) {
	source := readTestSource(t, "no_redundant_break.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-redundant-break"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG048 no-redundant-break")
}

func TestNoNakedReturnsReportsNakedReturnInLongFunction(t *testing.T) {
	maxLines := 5
	settings := Settings{MaxNakedReturnLines: &maxLines}
	source := readTestSource(t, "no_naked_returns.go")

	analyzer := analyzerByRuleWithSettings(t, "no-naked-returns", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "LEG049 no-naked-returns")
}

func TestNoNakedReturnsAllowsShortFunctions(t *testing.T) {
	maxLines := 20
	settings := Settings{MaxNakedReturnLines: &maxLines}
	source := readTestSource(t, "no_naked_returns.go")

	analyzer := analyzerByRuleWithSettings(t, "no-naked-returns", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestPreferLowercaseErrorStringsReportsCapitalizedMessage(t *testing.T) {
	source := readTestSource(t, "prefer_lowercase_error_strings.go")

	analyzer := analyzerByRule(t, "prefer-lowercase-error-strings")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "Start error strings with a lowercase word.")
}

func TestPreferLowercaseErrorStringsReportsTrailingPunctuation(t *testing.T) {
	source := readTestSource(t, "prefer_lowercase_error_strings_punctuation.go")

	analyzer := analyzerByRule(t, "prefer-lowercase-error-strings")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "Drop the trailing punctuation from error strings.")
}

func TestPreferLowercaseErrorStringsIgnoresUnimportedSelector(t *testing.T) {
	source := `package p
type maker struct{}
func (maker) New(string) error { return nil }
var errors maker
func load() error { return errors.New("Bad request") }
`
	analyzer := analyzerByRule(t, "prefer-lowercase-error-strings")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestPreferLowercaseErrorStringsHandlesImportAlias(t *testing.T) {
	source := `package p
import stderrors "errors"
func load() error { return stderrors.New("Bad request") }
`
	analyzer := analyzerByRule(t, "prefer-lowercase-error-strings")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestPreferLowercaseErrorStringsIgnoresShadowedImportAlias(t *testing.T) {
	source := `package p
import stderrors "errors"
var _ = stderrors.New
type maker struct{}
func (maker) New(string) error { return nil }
func load(stderrors maker) error { return stderrors.New("Bad request") }
`
	fileSet := token.NewFileSet()
	mode := parser.ParseComments | parser.SkipObjectResolution
	file := parseAnalyzerFileWithMode(t, fileSet, "p.go", source, mode)
	files := []*ast.File{file}
	analyzer := analyzerByRule(t, "prefer-lowercase-error-strings")
	diagnostics := runAnalyzerWithFiles(t, analyzer, fileSet, files)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestMaxFunctionParamsReportsLongSignature(t *testing.T) {
	source := readTestSource(t, "max_function_params.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "max-function-params"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG051 max-function-params")
}

func TestMaxFunctionParamsReportsAPITypeSignatures(t *testing.T) {
	source := `package p
type Sender interface {
	Send(one int, two int, three int, four int, five int, six int)
}
type Callback func(one int, two int, three int, four int, five int, six int)
`
	analyzer := analyzerByRule(t, "max-function-params")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
}
