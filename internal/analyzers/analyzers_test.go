package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

func TestMaxExpressionOperatorsReportsComplexExpression(t *testing.T) {
	source := readTestSource(t, "max_expression_operators.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "max-expression-operators"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG001 max-expression-operators")
}

func TestPreferEarlyReturnReportsElseAfterReturn(t *testing.T) {
	source := readTestSource(t, "prefer_early_return.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-early-return"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG009 prefer-early-return")
}

func TestPreferObjectLookupReportsLongEqualityChain(t *testing.T) {
	source := readTestSource(t, "prefer_object_lookup.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-object-lookup"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG024 prefer-object-lookup")
}

func TestNoDeepSelectorChainReportsDeepChain(t *testing.T) {
	source := readTestSource(t, "no_deep_selector_chain.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-deep-selector-chain"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG031 no-deep-selector-chain")
}

func TestPreferSwitchOverLongIfChainReportsRepeatedComparison(t *testing.T) {
	source := readTestSource(t, "prefer_switch_over_long_if_chain.go")

	analyzer := analyzerByRule(t, "prefer-switch-over-long-if-chain")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG034 prefer-switch-over-long-if-chain")
}

func TestNoBoolLiteralArgsReportsBooleanArgument(t *testing.T) {
	source := readTestSource(t, "no_bool_literal_args.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-bool-literal-args"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG035 no-bool-literal-args")
}

func TestNoComplexIfInitReportsOperatorHeavyCondition(t *testing.T) {
	source := readTestSource(t, "no_complex_if_init.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-complex-if-init"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG036 no-complex-if-init")
}

func TestNoDeepCompositeLiteralArgReportsNestedLiteral(t *testing.T) {
	source := readTestSource(t, "no_deep_composite_literal_arg.go")

	analyzer := analyzerByRule(t, "no-deep-composite-literal-arg")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG037 no-deep-composite-literal-arg")
}

func TestMaxFunctionLinesReportsLongFunction(t *testing.T) {
	maxLines := 5
	settings := Settings{MaxFunctionLines: &maxLines}
	source := readTestSource(t, "max_function_lines.go")

	analyzer := analyzerByRuleWithSettings(t, "max-function-lines", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG038 max-function-lines")
}

func TestMaxFunctionLinesMeasuresNestedFunctionsSeparately(t *testing.T) {
	maxLines := 3
	settings := Settings{MaxFunctionLines: &maxLines}
	source := readTestSource(t, "max_function_lines_nested.go")

	analyzer := analyzerByRuleWithSettings(t, "max-function-lines", settings)
	diagnostics, fileSet := runAnalyzerWithFileSet(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
	requireDiagnostic(t, diagnostics, "LEG038 max-function-lines")
	requireDiagnosticLine(t, fileSet, diagnostics[0], 4)
}

func TestRequireFilenameMatchesDirnameReportsWhenEnabled(t *testing.T) {
	minDepth := 2
	settings := Settings{
		EnabledRules:         []string{"require-filename-matches-dirname"},
		MinDirnameMatchDepth: &minDepth,
	}
	source := readTestSource(t, "empty_function.go")

	analyzer := analyzerByRuleWithSettings(t, "require-filename-matches-dirname", settings)
	diagnostics := runAnalyzer(t, analyzer, "internal/orders/service.go", source)
	requireDiagnostic(t, diagnostics, "LEG025 require-filename-matches-dirname")
}

func TestSettingsDisableRule(t *testing.T) {
	settings := Settings{DisabledRules: []string{"prefer-early-return"}}
	analyzers := New(settings)

	for _, analyzer := range analyzers {
		if analyzer.Name == analysisName("prefer-early-return") {
			t.Fatal("prefer-early-return should be disabled")
		}
	}
}

func TestMaxFunctionLinesIsDefaultEnabled(t *testing.T) {
	analyzerByRule(t, "max-function-lines")
}

func TestRequireFilenameMatchesDirnameIsOptIn(t *testing.T) {
	for _, analyzer := range New(Settings{}) {
		if analyzer.Name == analysisName("require-filename-matches-dirname") {
			t.Fatal("require-filename-matches-dirname should be opt-in")
		}
	}
}

func TestNoUnmatchedCommentsIsOptInWithoutPolicy(t *testing.T) {
	for _, analyzer := range New(Settings{}) {
		if analyzer.Name == analysisName("no-unmatched-comments") {
			t.Fatal("no-unmatched-comments should be opt-in without a configured policy")
		}
	}
}

func TestSpecialRuleSelectorsIgnoreCase(t *testing.T) {
	selectors := []string{"All", "ALL", "leg"}
	for _, selector := range selectors {
		settings := Settings{EnabledRules: []string{selector}}
		analyzers := New(settings)
		if len(analyzers) == 0 {
			t.Fatalf("%q should enable analyzers", selector)
		}
	}
}

func TestAllAnalyzersUseSyntaxOnlyInputs(t *testing.T) {
	settings := Settings{EnabledRules: []string{"all"}}
	for _, analyzer := range New(settings) {
		//nolint:legibility // Requirements are a fixed, single-entry syntax dependency.
		for _, requirement := range analyzer.Requires {
			if requirement != inspect.Analyzer {
				t.Fatalf("%s requires non-syntax analyzer %s", analyzer.Name, requirement.Name)
			}
		}
	}
}

func TestAnalyzersAreValid(t *testing.T) {
	settings := Settings{EnabledRules: []string{"all"}}
	if err := analysis.Validate(New(settings)); err != nil {
		t.Fatal(err)
	}
}

func analyzerByRule(t *testing.T, name string) *analysis.Analyzer {
	t.Helper()

	for _, analyzer := range New(Settings{}) {
		if analyzer.Name == analysisName(name) {
			return analyzer
		}
	}

	t.Fatalf("missing analyzer for %s", name)
	return nil
}

func analyzerByRuleWithSettings(t *testing.T, name string, settings Settings) *analysis.Analyzer {
	t.Helper()

	for _, analyzer := range New(settings) {
		if analyzer.Name == analysisName(name) {
			return analyzer
		}
	}

	t.Fatalf("missing analyzer for %s", name)
	return nil
}

func runAnalyzer(
	t *testing.T,
	analyzer *analysis.Analyzer,
	filename string,
	source string,
) []analysis.Diagnostic {
	t.Helper()

	diagnostics, _ := runAnalyzerWithFileSet(t, analyzer, filename, source)
	return diagnostics
}

func runAnalyzerWithFileSet(
	t *testing.T,
	analyzer *analysis.Analyzer,
	filename string,
	source string,
) ([]analysis.Diagnostic, *token.FileSet) {
	t.Helper()

	var diagnostics []analysis.Diagnostic
	fileSet := token.NewFileSet()
	file := parseAnalyzerFile(t, fileSet, filename, source)
	pass := analyzerPass(fileSet, file, &diagnostics)
	runAnalysis(t, analyzer, pass)

	return diagnostics, fileSet
}

func runAnalyzerWithFiles(
	t *testing.T,
	analyzer *analysis.Analyzer,
	fileSet *token.FileSet,
	files []*ast.File,
) []analysis.Diagnostic {
	t.Helper()

	var diagnostics []analysis.Diagnostic
	pass := analyzerPassWithFiles(fileSet, files, &diagnostics)
	runAnalysis(t, analyzer, pass)

	return diagnostics
}

func parseAnalyzerFile(
	t *testing.T,
	fileSet *token.FileSet,
	filename string,
	source string,
) *ast.File {
	return parseAnalyzerFileWithMode(t, fileSet, filename, source, parser.ParseComments)
}

func parseAnalyzerFileWithMode(
	t *testing.T,
	fileSet *token.FileSet,
	filename string,
	source string,
	mode parser.Mode,
) *ast.File {
	t.Helper()

	file, err := parser.ParseFile(fileSet, filename, source, mode)
	if err != nil {
		t.Fatal(err)
	}

	return file
}

func analyzerPass(
	fileSet *token.FileSet,
	file *ast.File,
	diagnostics *[]analysis.Diagnostic,
) *analysis.Pass {
	files := []*ast.File{file}
	return analyzerPassWithFiles(fileSet, files, diagnostics)
}

func analyzerPassWithFiles(
	fileSet *token.FileSet,
	files []*ast.File,
	diagnostics *[]analysis.Diagnostic,
) *analysis.Pass {
	return &analysis.Pass{
		Fset:  fileSet,
		Files: files,
		Report: func(diagnostic analysis.Diagnostic) {
			*diagnostics = append(*diagnostics, diagnostic)
		},
	}
}

func runAnalysis(t *testing.T, analyzer *analysis.Analyzer, pass *analysis.Pass) {
	t.Helper()

	pass.ResultOf = runRequiredAnalyses(t, analyzer, pass)
	_, err := analyzer.Run(pass)
	if err != nil {
		t.Fatal(err)
	}
}

func runRequiredAnalyses(
	t *testing.T,
	analyzer *analysis.Analyzer,
	pass *analysis.Pass,
) map[*analysis.Analyzer]any {
	t.Helper()

	results := make(map[*analysis.Analyzer]any, len(analyzer.Requires))
	for _, requirement := range analyzer.Requires {
		result, err := requirement.Run(pass)
		if err != nil {
			t.Fatal(err)
		}
		results[requirement] = result
	}

	return results
}

func readTestSource(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("testdata", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return materializeTestSource(string(source))
}

func materializeTestSource(source string) string {
	firstSource := "chat" + "gpt"
	secondSource := "clau" + "de"
	shortSource := "a" + "i"
	replacer := strings.NewReplacer(
		"AUTOMATED_SOURCE_ONE", firstSource,
		"AUTOMATED_SOURCE_TWO", secondSource,
		"AUTOMATED_SOURCE_SHORT", shortSource,
	)
	return replacer.Replace(source)
}

func requireDiagnostic(t *testing.T, diagnostics []analysis.Diagnostic, text string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, text) {
			return
		}
	}

	t.Fatalf("missing diagnostic %q in %#v", text, diagnostics)
}

func requireDiagnosticsCount(t *testing.T, diagnostics []analysis.Diagnostic, want int) {
	t.Helper()

	if len(diagnostics) != want {
		t.Fatalf("diagnostic count = %d, want %d: %#v", len(diagnostics), want, diagnostics)
	}
}

func requireDiagnosticLine(
	t *testing.T,
	fileSet *token.FileSet,
	diagnostic analysis.Diagnostic,
	want int,
) {
	t.Helper()

	got := fileSet.Position(diagnostic.Pos).Line
	if got != want {
		t.Fatalf("diagnostic line = %d, want %d", got, want)
	}
}
