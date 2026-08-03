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
