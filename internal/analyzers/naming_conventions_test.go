package analyzers

import "testing"

func TestPreferVerbFunctionNamesReportsPayloadFirstNames(t *testing.T) {
	settings := Settings{EnabledRules: []string{"prefer-verb-function-names"}}
	source := readTestSource(t, "prefer_verb_function_names.go")

	analyzer := analyzerByRuleWithSettings(t, "prefer-verb-function-names", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "LEG052 prefer-verb-function-names")
}

func TestPreferBooleanPrefixesReportsUnprefixedBooleanNames(t *testing.T) {
	settings := Settings{EnabledRules: []string{"prefer-boolean-prefixes"}}
	source := readTestSource(t, "prefer_boolean_prefixes.go")

	analyzer := analyzerByRuleWithSettings(t, "prefer-boolean-prefixes", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "LEG053 prefer-boolean-prefixes")
}

func TestPreferBooleanPrefixesSkipsBlankNamesAndBarePrefixes(t *testing.T) {
	settings := Settings{EnabledRules: []string{"prefer-boolean-prefixes"}}
	source := readTestSource(t, "prefer_boolean_prefixes_edge_cases.go")

	analyzer := analyzerByRuleWithSettings(t, "prefer-boolean-prefixes", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
}

func TestPreferVerbFunctionNamesAllowsCanonicalValue(t *testing.T) {
	settings := Settings{EnabledRules: []string{"prefer-verb-function-names"}}
	source := "package p\n\nfunc Value() any { return nil }\n"

	analyzer := analyzerByRuleWithSettings(t, "prefer-verb-function-names", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}
