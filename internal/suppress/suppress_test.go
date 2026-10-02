package suppress

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"testing"

	"github.com/yowainwright/go-lint-legibility/internal/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestSuppressionInSource(t *testing.T) {
	settings := analyzers.Settings{EnabledRules: []string{"LEG002"}}
	wrapped := Wrap(analyzers.New(settings))
	if len(wrapped) != 1 {
		t.Fatalf("got %d analyzers, want 1", len(wrapped))
	}

	analysistest.Run(t, analysistest.TestData(), wrapped[0], "a")
}

func TestDirectiveCacheReadsSourceOnce(t *testing.T) {
	source := "package p\n\n//nolint\nfunc first() {}\n//nolint\nfunc second() {}\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "source.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	reads := 0
	pass := &analysis.Pass{
		Pkg:   types.NewPackage("example.test/p", "p"),
		Fset:  fset,
		Files: []*ast.File{file},
		ReadFile: func(string) ([]byte, error) {
			reads++
			return []byte(source), nil
		},
	}
	cache := &directiveCache{byFiles: map[string]fileIndex{}}
	cache.forPass(pass)
	cache.forPass(pass)
	if reads != 1 {
		t.Fatalf("source reads = %d, want 1", reads)
	}
}

type parseCase struct {
	name       string
	text       string
	want       []string
	isDirected bool
}

var parseCases = []parseCase{
	{"bare", "//nolint", []string{"all"}, true},
	{"bare with reason", "//nolint // because", []string{"all"}, true},
	{"one selector", "//nolint:LEG002", []string{"LEG002"}, true},
	{
		"several",
		"//nolint:legibility,gocritic // why",
		[]string{"legibility", "gocritic"},
		true,
	},
	{"empty list", "//nolint:", nil, true},
	{"not a directive", "// nolint", nil, false},
	{"longer word", "//nolintfoo", nil, false},
	{"ordinary comment", "// hello", nil, false},
}

func TestParse(t *testing.T) {
	for _, c := range parseCases {
		t.Run(c.name, func(t *testing.T) {
			got, isDirective := parse(c.text)
			matches := isDirective == c.isDirected && reflect.DeepEqual(got, c.want)
			if !matches {
				t.Fatalf(
					"parse(%q) = (%v, %v), want (%v, %v)",
					c.text,
					got,
					isDirective,
					c.want,
					c.isDirected,
				)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	rule := analyzers.Rule{Code: "LEG002", Name: "hoist-if-operators"}
	cases := []struct {
		selector string
		want     bool
	}{
		{"all", true},
		{"legibility", true},
		{"LEG002", true},
		{"leg002", true},
		{"hoist-if-operators", true},
		{"LEG003", false},
		{"gocritic", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.selector, func(t *testing.T) {
			if got := matches(c.selector, rule); got != c.want {
				t.Fatalf("matches(%q) = %v, want %v", c.selector, got, c.want)
			}
		})
	}
}
