package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	benchmarkLegacyASTWalks        = 35
	benchmarkLegacyParentMapBuilds = 7
)

func BenchmarkASTTraversal(b *testing.B) {
	files := parseBenchmarkFiles(b)

	b.Run("repeated_ast_inspect", func(b *testing.B) {
		for b.Loop() {
			visits := visitWithASTInspect(files, benchmarkLegacyASTWalks)
			assertBenchmarkWork(b, visits)
		}
	})

	b.Run("shared_inspector", func(b *testing.B) {
		for b.Loop() {
			visits := visitWithInspector(files)
			assertBenchmarkWork(b, visits)
		}
	})
}

func BenchmarkParentLookup(b *testing.B) {
	files := parseBenchmarkFiles(b)

	b.Run("repeated_parent_maps", func(b *testing.B) {
		for b.Loop() {
			parents := buildParentMaps(files, benchmarkLegacyParentMapBuilds)
			assertBenchmarkWork(b, parents)
		}
	})

	b.Run("shared_inspector", func(b *testing.B) {
		for b.Loop() {
			parents := visitWithInspector(files)
			assertBenchmarkWork(b, parents)
		}
	})
}

func assertBenchmarkWork(b *testing.B, count int) {
	b.Helper()
	if count == 0 {
		b.Fatal("benchmark performed no work")
	}
}

func parseBenchmarkFiles(b *testing.B) []*ast.File {
	b.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		b.Fatal(err)
	}

	fileSet := token.NewFileSet()
	return parseBenchmarkEntries(b, fileSet, entries)
}

func parseBenchmarkEntries(
	b *testing.B,
	fileSet *token.FileSet,
	entries []os.DirEntry,
) []*ast.File {
	b.Helper()

	files := make([]*ast.File, 0, len(entries))
	for _, entry := range entries {
		file := parseBenchmarkEntry(b, fileSet, entry)
		if file != nil {
			files = append(files, file)
		}
	}
	return files
}

func parseBenchmarkEntry(b *testing.B, fileSet *token.FileSet, entry os.DirEntry) *ast.File {
	b.Helper()
	if !isAnalyzerSource(entry) {
		return nil
	}

	file, err := parser.ParseFile(fileSet, entry.Name(), nil, parser.SkipObjectResolution)
	if err != nil {
		b.Fatal(err)
	}
	return file
}

func isAnalyzerSource(entry os.DirEntry) bool {
	name := entry.Name()
	isGoFile := strings.HasSuffix(name, ".go")
	isTestFile := strings.HasSuffix(name, "_test.go")
	isSourceFile := !entry.IsDir() && isGoFile && !isTestFile
	return isSourceFile
}

func visitWithASTInspect(files []*ast.File, walks int) int {
	visits := 0
	for range walks {
		visits += visitFilesWithASTInspect(files)
	}
	return visits
}

func visitFilesWithASTInspect(files []*ast.File) int {
	visits := 0
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			if node != nil {
				visits++
			}
			return true
		})
	}
	return visits
}

func visitWithInspector(files []*ast.File) int {
	index := inspector.New(files)
	return visitInspectorParents(index)
}

func buildParentMaps(files []*ast.File, builds int) int {
	parents := 0
	for range builds {
		parentMap := buildParentMapForBenchmark(files)
		parents += len(parentMap)
	}
	return parents
}

func buildParentMapForBenchmark(files []*ast.File) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	for _, file := range files {
		astutil.Apply(file, func(cursor *astutil.Cursor) bool {
			node := cursor.Node()
			if node != nil {
				parents[node] = cursor.Parent()
			}
			return true
		}, nil)
	}
	return parents
}

func visitInspectorParents(index *inspector.Inspector) int {
	parents := 0
	for cursor := range index.Root().Preorder() {
		if cursor.Parent().Node() != nil {
			parents++
		}
	}
	return parents
}
