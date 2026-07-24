package analyzers

import "testing"

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

func TestPreferInitialismCasingReportsMixedCaseInitialisms(t *testing.T) {
	source := readTestSource(t, "prefer_initialism_casing.go")

	analyzer := analyzerByRule(t, "prefer-initialism-casing")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "Capitalize the initialism as URL.")
	requireDiagnostic(t, diagnostics, "Capitalize the initialism as ID.")
}

func TestNoPackageNameStutterReportsRepeatedPackagePrefix(t *testing.T) {
	source := readTestSource(t, "no_package_name_stutter.go")

	analyzer := analyzerByRule(t, "no-package-name-stutter")
	diagnostics := runAnalyzer(t, analyzer, "orders/orders.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
	requireDiagnostic(t, diagnostics, "LEG045 no-package-name-stutter")
}

func TestNoGenericPackageNamesReportsUtilPackage(t *testing.T) {
	source := readTestSource(t, "no_generic_package_names.go")

	analyzer := analyzerByRule(t, "no-generic-package-names")
	diagnostics := runAnalyzer(t, analyzer, "internal/util/util.go", source)
	requireDiagnostic(t, diagnostics, "LEG046 no-generic-package-names")
}

func TestPreferRangeLoopReportsIndexCounterLoop(t *testing.T) {
	source := readTestSource(t, "prefer_range_loop.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-range-loop"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG047 prefer-range-loop")
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

func TestMaxFunctionParamsReportsLongSignature(t *testing.T) {
	source := readTestSource(t, "max_function_params.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "max-function-params"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG051 max-function-params")
}
