package suppress

import (
	"go/ast"
	"go/token"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
	"golang.org/x/tools/go/analysis"
)

type lineSelectors map[int][]string

type fileIndex map[string]lineSelectors

type directiveCache struct {
	mu      sync.Mutex
	byFiles map[string]fileIndex
}

type runFunc func(*analysis.Pass) (any, error)

const (
	directivePrefix = "//nolint"
	selectAll       = "all"
	selectLegible   = "legibility"
)

func Wrap(list []*analysis.Analyzer) []*analysis.Analyzer {
	rules := rulesByAnalyzer()
	cache := &directiveCache{byFiles: map[string]fileIndex{}}
	wrapped := make([]*analysis.Analyzer, 0, len(list))
	for _, original := range list {
		clone := *original
		clone.Run = filterRun(original.Run, rules[original.Name], cache)
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

func filterRun(run runFunc, rule analyzers.Rule, cache *directiveCache) runFunc {
	return func(pass *analysis.Pass) (any, error) {
		directives := cache.forPass(pass)
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

func (cache *directiveCache) forPass(pass *analysis.Pass) fileIndex {
	key := packageKey(pass)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if directives, found := cache.byFiles[key]; found {
		return directives
	}

	directives := collect(pass)
	cache.byFiles[key] = directives
	return directives
}

func packageKey(pass *analysis.Pass) string {
	paths := []string{pass.Pkg.Path()}
	for _, file := range pass.Files {
		paths = append(paths, pass.Fset.Position(file.Pos()).Filename)
	}
	slices.Sort(paths[1:])

	return strings.Join(paths, "\x00")
}

func collect(pass *analysis.Pass) fileIndex {
	found := fileIndex{}
	for _, file := range pass.Files {
		addFile(found, pass, file)
	}

	return found
}

func addFile(found fileIndex, pass *analysis.Pass, file *ast.File) {
	var lines []string
	loaded := false
	for _, comment := range commentsIn(file) {
		selectors, isDirective := parse(comment.Text)
		if !isDirective {
			continue
		}
		position := pass.Fset.Position(comment.Slash)
		if !loaded {
			lines = sourceLines(pass, position.Filename)
			loaded = true
		}
		register(found, position, selectors, lines)
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

func register(found fileIndex, position token.Position, selectors []string, source []string) {
	if found[position.Filename] == nil {
		found[position.Filename] = lineSelectors{}
	}

	lines := found[position.Filename]
	lines[position.Line] = append(lines[position.Line], selectors...)
	if isOwnLine(source, position) {
		lines[position.Line+1] = append(lines[position.Line+1], selectors...)
	}
}

func sourceLines(pass *analysis.Pass, name string) []string {
	source, err := readSource(pass, name)
	if err != nil {
		return nil
	}

	return strings.Split(string(source), "\n")
}

func isOwnLine(lines []string, position token.Position) bool {
	if position.Line < 1 {
		return false
	}
	if position.Line > len(lines) {
		return false
	}
	if position.Column < 1 {
		return false
	}

	line := lines[position.Line-1]
	column := position.Column - 1
	if column > len(line) {
		return false
	}
	prefix := line[:column]

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
