package filter

import (
	"testing"

	"github.com/jackbonatakis/differ/internal/parser"
)

func fs(path string) parser.FileStat {
	return parser.FileStat{Path: path, Added: 1, Deleted: 0, Churn: 1}
}

func paths(stats []parser.FileStat) []string {
	out := make([]string, len(stats))
	for i, s := range stats {
		out[i] = s.Path
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIncludeOnly(t *testing.T) {
	input := []parser.FileStat{
		fs("src/main.go"),
		fs("docs/readme.md"),
		fs("test/foo_test.go"),
	}
	cfg := FilterConfig{Include: []string{"src/**"}}
	got := paths(Filter(input, cfg, nil))
	want := []string{"src/main.go"}
	if !eq(got, want) {
		t.Errorf("include-only: got %v, want %v", got, want)
	}
}

func TestExcludeOnly(t *testing.T) {
	input := []parser.FileStat{
		fs("src/main.go"),
		fs("vendor/lib.go"),
	}
	cfg := FilterConfig{Exclude: []string{"vendor/**"}}
	got := paths(Filter(input, cfg, nil))
	want := []string{"src/main.go"}
	if !eq(got, want) {
		t.Errorf("exclude-only: got %v, want %v", got, want)
	}
}

func TestIncludeAndExclude(t *testing.T) {
	input := []parser.FileStat{
		fs("src/main.go"),
		fs("vendor/lib.go"),
		fs("README.md"),
	}
	cfg := FilterConfig{
		Include: []string{"**/*.go"},
		Exclude: []string{"vendor/**"},
	}
	got := paths(Filter(input, cfg, nil))
	want := []string{"src/main.go"}
	if !eq(got, want) {
		t.Errorf("include+exclude: got %v, want %v", got, want)
	}
}

func TestCategoryFilter(t *testing.T) {
	input := []parser.FileStat{
		fs("src/main.go"),
		fs("docs/readme.md"),
		fs("test/foo_test.go"),
	}
	categories := map[string]string{
		"src/main.go":      "source",
		"docs/readme.md":   "docs",
		"test/foo_test.go": "tests",
	}
	catFn := func(path string) string { return categories[path] }
	cfg := FilterConfig{Categories: []string{"source"}}
	got := paths(Filter(input, cfg, catFn))
	want := []string{"src/main.go"}
	if !eq(got, want) {
		t.Errorf("category filter: got %v, want %v", got, want)
	}
}

func TestNoFilters(t *testing.T) {
	input := []parser.FileStat{
		fs("a.go"),
		fs("b.md"),
		fs("c.js"),
	}
	cfg := FilterConfig{}
	got := paths(Filter(input, cfg, nil))
	want := []string{"a.go", "b.md", "c.js"}
	if !eq(got, want) {
		t.Errorf("no filters: got %v, want %v", got, want)
	}
}

func TestDoublestarGlob(t *testing.T) {
	input := []parser.FileStat{
		fs("internal/pkg/test_util.go"),
		fs("internal/other/main.go"),
		fs("cmd/main.go"),
	}
	cfg := FilterConfig{Include: []string{"internal/**/test_*.go"}}
	got := paths(Filter(input, cfg, nil))
	want := []string{"internal/pkg/test_util.go"}
	if !eq(got, want) {
		t.Errorf("doublestar glob: got %v, want %v", got, want)
	}
}

func TestCategoryWithIncludeExclude(t *testing.T) {
	input := []parser.FileStat{
		fs("src/app.go"),
		fs("src/app_test.go"),
		fs("vendor/dep.go"),
		fs("docs/guide.md"),
	}
	categories := map[string]string{
		"src/app.go":      "source",
		"src/app_test.go": "tests",
		"vendor/dep.go":   "generated",
		"docs/guide.md":   "docs",
	}
	catFn := func(path string) string { return categories[path] }
	cfg := FilterConfig{
		Include:    []string{"src/**"},
		Exclude:    []string{"*_test.go"},
		Categories: []string{"source"},
	}
	got := paths(Filter(input, cfg, catFn))
	want := []string{"src/app.go"}
	if !eq(got, want) {
		t.Errorf("combined all filters: got %v, want %v", got, want)
	}
}

func TestEmptyInput(t *testing.T) {
	cfg := FilterConfig{Include: []string{"*.go"}}
	got := Filter(nil, cfg, nil)
	if got != nil {
		t.Errorf("empty input: expected nil, got %v", got)
	}
}

func TestMultipleIncludePatterns(t *testing.T) {
	input := []parser.FileStat{
		fs("main.go"),
		fs("style.css"),
		fs("readme.md"),
		fs("app.js"),
	}
	cfg := FilterConfig{Include: []string{"*.go", "*.js"}}
	got := paths(Filter(input, cfg, nil))
	want := []string{"main.go", "app.js"}
	if !eq(got, want) {
		t.Errorf("multiple include patterns: got %v, want %v", got, want)
	}
}

func TestMultipleCategories(t *testing.T) {
	input := []parser.FileStat{
		fs("src/main.go"),
		fs("docs/readme.md"),
		fs("test/foo_test.go"),
	}
	categories := map[string]string{
		"src/main.go":      "source",
		"docs/readme.md":   "docs",
		"test/foo_test.go": "tests",
	}
	catFn := func(path string) string { return categories[path] }
	cfg := FilterConfig{Categories: []string{"source", "docs"}}
	got := paths(Filter(input, cfg, catFn))
	want := []string{"src/main.go", "docs/readme.md"}
	if !eq(got, want) {
		t.Errorf("multiple categories: got %v, want %v", got, want)
	}
}
