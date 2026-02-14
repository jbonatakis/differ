package classify

import (
	"path/filepath"
	"strings"

	"github.com/jbonatakis/differ/internal/config"
)

// Category constants.
const (
	Generated = "generated"
	Docs      = "docs"
	Tests     = "tests"
	Source    = "source"
	Other     = "other"
)

// Classifier assigns a category and language to file paths.
type Classifier struct {
	customCategories map[string]config.CategoryConfig
}

// New creates a Classifier with optional custom category overrides from config.
func New(cfg config.Config) *Classifier {
	return &Classifier{
		customCategories: cfg.Categories,
	}
}

// Classify returns the category and detected language for a file path.
// Categories are evaluated in first-match priority order:
// generated > docs > tests > source > other.
func (c *Classifier) Classify(path string) (category string, language string) {
	// Normalize path separators.
	normalized := filepath.ToSlash(path)
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(base))

	if c.isGenerated(normalized, base) {
		return Generated, detectLanguage(ext)
	}
	if c.isDocs(normalized, ext) {
		return Docs, detectLanguage(ext)
	}
	if c.isTests(normalized, base) {
		return Tests, detectLanguage(ext)
	}
	if c.isSource(ext) {
		return Source, detectLanguage(ext)
	}
	return Other, detectLanguage(ext)
}

// Generated directories that indicate generated/vendored content.
var generatedDirs = []string{
	"vendor/",
	"node_modules/",
	"dist/",
	"build/",
}

// Lockfiles considered generated.
var lockfiles = map[string]bool{
	"package-lock.json": true,
	"pnpm-lock.yaml":   true,
	"yarn.lock":        true,
	"go.sum":           true,
	"cargo.lock":       true,
	"gemfile.lock":     true,
	"composer.lock":    true,
	"poetry.lock":      true,
	"pipfile.lock":     true,
	"bun.lockb":        true,
	"flake.lock":       true,
}

func (c *Classifier) isGenerated(normalized, base string) bool {
	// Check custom generated patterns first.
	if cc, ok := c.customCategories[Generated]; ok {
		if matchesCustom(normalized, base, cc) {
			return true
		}
	}

	// Check generated directories.
	for _, dir := range generatedDirs {
		if strings.HasPrefix(normalized, dir) || strings.Contains(normalized, "/"+dir) {
			return true
		}
	}

	// Check lockfiles (case-insensitive).
	if lockfiles[strings.ToLower(base)] {
		return true
	}

	return false
}

// Doc extensions.
var docExtensions = map[string]bool{
	".md":       true,
	".markdown": true,
	".mdx":      true,
	".rst":      true,
	".adoc":     true,
	".txt":      true,
}

// Doc directories.
var docDirs = []string{
	"docs/",
	"documentation/",
}

func (c *Classifier) isDocs(normalized, ext string) bool {
	if cc, ok := c.customCategories[Docs]; ok {
		if matchesCustom(normalized, filepath.Base(normalized), cc) {
			return true
		}
	}

	if docExtensions[ext] {
		return true
	}

	for _, dir := range docDirs {
		if strings.HasPrefix(normalized, dir) || strings.Contains(normalized, "/"+dir) {
			return true
		}
	}

	return false
}

// Test directories.
var testDirs = []string{
	"test/",
	"tests/",
	"spec/",
	"specs/",
	"__tests__/",
}

func (c *Classifier) isTests(normalized, base string) bool {
	if cc, ok := c.customCategories[Tests]; ok {
		if matchesCustom(normalized, base, cc) {
			return true
		}
	}

	// Check test directories.
	for _, dir := range testDirs {
		if strings.HasPrefix(normalized, dir) || strings.Contains(normalized, "/"+dir) {
			return true
		}
	}

	// Check filename patterns.
	lower := strings.ToLower(base)

	// Generic patterns: *.test.*, *.spec.*
	if strings.Contains(lower, ".test.") || strings.Contains(lower, ".spec.") {
		return true
	}

	// Go: *_test.go
	if strings.HasSuffix(lower, "_test.go") {
		return true
	}

	// Python: test_*.py, *_test.py
	if strings.HasSuffix(lower, ".py") {
		name := strings.TrimSuffix(lower, ".py")
		if strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test") {
			return true
		}
	}

	// Java: *Test.java, *Tests.java
	if strings.HasSuffix(base, "Test.java") || strings.HasSuffix(base, "Tests.java") {
		return true
	}

	// Kotlin: *Test.kt
	if strings.HasSuffix(base, "Test.kt") {
		return true
	}

	// Ruby: *_spec.rb, test_*.rb
	if strings.HasSuffix(lower, ".rb") {
		name := strings.TrimSuffix(lower, ".rb")
		if strings.HasSuffix(name, "_spec") || strings.HasPrefix(name, "test_") {
			return true
		}
	}

	return false
}

// Source code extensions mapped to language names.
var sourceExtensions = map[string]string{
	// Go
	".go": "Go",
	// Rust
	".rs": "Rust",
	// Python
	".py": "Python", ".pyi": "Python", ".pyw": "Python",
	// JavaScript
	".js": "JavaScript", ".mjs": "JavaScript", ".cjs": "JavaScript",
	// TypeScript
	".ts": "TypeScript", ".mts": "TypeScript", ".cts": "TypeScript",
	// JSX/TSX
	".jsx": "JSX", ".tsx": "TSX",
	// Java
	".java": "Java",
	// Kotlin
	".kt": "Kotlin", ".kts": "Kotlin",
	// C
	".c": "C", ".h": "C",
	// C++
	".cpp": "C++", ".cc": "C++", ".cxx": "C++", ".hpp": "C++", ".hxx": "C++", ".hh": "C++",
	// C#
	".cs": "C#",
	// PHP
	".php": "PHP",
	// Ruby
	".rb": "Ruby", ".rake": "Ruby",
	// Swift
	".swift": "Swift",
	// Scala
	".scala": "Scala", ".sc": "Scala",
	// Shell
	".sh": "Shell", ".bash": "Shell", ".zsh": "Shell",
	// Lua
	".lua": "Lua",
	// Perl
	".pl": "Perl", ".pm": "Perl",
	// R
	".r": "R",
	// Dart
	".dart": "Dart",
	// Elixir
	".ex": "Elixir", ".exs": "Elixir",
	// Erlang
	".erl": "Erlang", ".hrl": "Erlang",
	// Haskell
	".hs": "Haskell", ".lhs": "Haskell",
	// OCaml
	".ml": "OCaml", ".mli": "OCaml",
	// Clojure
	".clj": "Clojure", ".cljs": "Clojure", ".cljc": "Clojure",
	// Groovy
	".groovy": "Groovy",
	// Zig
	".zig": "Zig",
	// Nim
	".nim": "Nim",
	// V
	".v": "V",
	// SQL
	".sql": "SQL",
	// HTML
	".html": "HTML", ".htm": "HTML",
	// CSS
	".css": "CSS", ".scss": "CSS", ".sass": "CSS", ".less": "CSS",
	// Vue
	".vue": "Vue",
	// Svelte
	".svelte": "Svelte",
	// YAML
	".yaml": "YAML", ".yml": "YAML",
	// TOML
	".toml": "TOML",
	// JSON
	".json": "JSON",
	// XML
	".xml": "XML",
	// Protobuf
	".proto": "Protobuf",
	// GraphQL
	".graphql": "GraphQL", ".gql": "GraphQL",
	// Terraform
	".tf": "Terraform", ".tfvars": "Terraform",
}

func (c *Classifier) isSource(ext string) bool {
	if cc, ok := c.customCategories[Source]; ok {
		for _, e := range cc.Extensions {
			cmpExt := strings.ToLower(e)
			if !strings.HasPrefix(cmpExt, ".") {
				cmpExt = "." + cmpExt
			}
			if ext == cmpExt {
				return true
			}
		}
	}
	_, ok := sourceExtensions[ext]
	return ok
}

// detectLanguage returns the language name for a given extension.
func detectLanguage(ext string) string {
	if lang, ok := sourceExtensions[ext]; ok {
		return lang
	}
	return ""
}

// matchesCustom checks if a file matches custom category patterns or extensions.
func matchesCustom(normalized, base string, cc config.CategoryConfig) bool {
	for _, pattern := range cc.Patterns {
		p := filepath.ToSlash(pattern)
		// Support glob patterns.
		if matched, _ := filepath.Match(p, base); matched {
			return true
		}
		// Support directory prefix patterns.
		if strings.HasSuffix(p, "/") {
			if strings.HasPrefix(normalized, p) || strings.Contains(normalized, "/"+p) {
				return true
			}
		}
		// Support substring matching for non-glob, non-directory patterns.
		if !strings.ContainsAny(p, "*?[") && !strings.HasSuffix(p, "/") {
			if strings.Contains(normalized, p) {
				return true
			}
		}
	}
	ext := strings.ToLower(filepath.Ext(base))
	for _, e := range cc.Extensions {
		cmpExt := strings.ToLower(e)
		if !strings.HasPrefix(cmpExt, ".") {
			cmpExt = "." + cmpExt
		}
		if ext == cmpExt {
			return true
		}
	}
	return false
}
