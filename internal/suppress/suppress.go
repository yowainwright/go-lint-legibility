package suppress

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
	"golang.org/x/tools/go/analysis"
)

type lineSelectors map[int][]string

type fileIndex map[string]lineSelectors

type runFunc func(*analysis.Pass) (any, error)

const (
	directivePrefix = "//nolint"
	selectAll       = "all"
	selectLegible   = "legibility"
)

func Wrap(list []*analysis.Analyzer) []*analysis.Analyzer {
	rules := rulesByAnalyzer()
	wrapped := make([]*analysis.Analyzer, 0, len(list))
	for _, original := range list {
		clone := *original
		clone.Run = filterRun(original.Run, rules[original.Name])
		wrapped = append(wrapped, &clone)
	}

	return wrapped
}

func rulesByAnalyzer() map[string]analyzers.Rule {
	rules := map[string]analyzers.Rule{}
	for _, rule := range analyzers.Rules() {
		rules[rule.Analyzer] = rule
	}

	return rules
}

func filterRun(run runFunc, rule analyzers.Rule) runFunc {
	return func(pass *analysis.Pass) (any, error) {
		directives := collect(pass)
		filtered := *pass
		filtered.Report = func(diagnostic analysis.Diagnostic) {
			position := pass.Fset.Position(diagnostic.Pos)
			selectors := directives[position.Filename][position.Line]
			if !covers(selectors, rule) {
				pass.Report(diagnostic)
			}
		}

		return run(&filtered)
	}
}

func collect(pass *analysis.Pass) fileIndex {
	found := fileIndex{}
	for _, file := range pass.Files {
		addFile(found, pass, file)
	}

	return found
}

func addFile(found fileIndex, pass *analysis.Pass, file *ast.File) {
	for _, comment := range commentsIn(file) {
		selectors, isDirective := parse(comment.Text)
		if isDirective {
			register(found, pass, comment, selectors)
		}
	}
}

func commentsIn(file *ast.File) []*ast.Comment {
	var comments []*ast.Comment
	for _, group := range file.Comments {
		comments = append(comments, group.List...)
	}

	return comments
}

func parse(text string) ([]string, bool) {
	rest, found := strings.CutPrefix(text, directivePrefix)
	if !found {
		return nil, false
	}
	if rest == "" {
		return []string{selectAll}, true
	}
	if rest[0] == ':' {
		return splitSelectors(rest[1:]), true
	}
	if strings.ContainsRune(" \t", rune(rest[0])) {
		return []string{selectAll}, true
	}

	return nil, false
}

func splitSelectors(rest string) []string {
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return nil
	}

	return strings.Split(fields[0], ",")
}

func register(found fileIndex, pass *analysis.Pass, comment *ast.Comment, selectors []string) {
	position := pass.Fset.Position(comment.Slash)
	if found[position.Filename] == nil {
		found[position.Filename] = lineSelectors{}
	}

	lines := found[position.Filename]
	lines[position.Line] = append(lines[position.Line], selectors...)
	if isOwnLine(pass, position) {
		lines[position.Line+1] = append(lines[position.Line+1], selectors...)
	}
}

func isOwnLine(pass *analysis.Pass, position token.Position) bool {
	source, err := readSource(pass, position.Filename)
	if err != nil {
		return false
	}

	lines := strings.Split(string(source), "\n")
	if position.Line > len(lines) {
		return false
	}

	prefix := lines[position.Line-1][:position.Column-1]
	return strings.TrimSpace(prefix) == ""
}

func readSource(pass *analysis.Pass, name string) ([]byte, error) {
	if pass.ReadFile != nil {
		return pass.ReadFile(name)
	}

	return os.ReadFile(name)
}

func covers(selectors []string, rule analyzers.Rule) bool {
	for _, selector := range selectors {
		if matches(selector, rule) {
			return true
		}
	}

	return false
}

func matches(selector string, rule analyzers.Rule) bool {
	switch strings.ToLower(strings.TrimSpace(selector)) {
	case selectAll, selectLegible, strings.ToLower(rule.Code), strings.ToLower(rule.Name):
		return true
	}

	return false
}
