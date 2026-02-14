package classify

import (
	"testing"

	"github.com/jackbonatakis/differ/internal/config"
)

func newClassifier(categories map[string]config.CategoryConfig) *Classifier {
	return New(config.Config{Categories: categories})
}

func defaultClassifier() *Classifier {
	return newClassifier(nil)
}

func TestCategoryConstants(t *testing.T) {
	if Generated != "generated" {
		t.Errorf("Generated = %q, want %q", Generated, "generated")
	}
	if Docs != "docs" {
		t.Errorf("Docs = %q, want %q", Docs, "docs")
	}
	if Tests != "tests" {
		t.Errorf("Tests = %q, want %q", Tests, "tests")
	}
	if Source != "source" {
		t.Errorf("Source = %q, want %q", Source, "source")
	}
	if Other != "other" {
		t.Errorf("Other = %q, want %q", Other, "other")
	}
}

func TestGeneratedDirectories(t *testing.T) {
	c := defaultClassifier()
	paths := []string{
		"vendor/github.com/pkg/errors/errors.go",
		"node_modules/lodash/index.js",
		"dist/bundle.js",
		"build/output.js",
		"pkg/vendor/dep.go",
		"frontend/node_modules/react/index.js",
		"out/dist/app.js",
		"project/build/main.o",
	}
	for _, p := range paths {
		cat, _ := c.Classify(p)
		if cat != Generated {
			t.Errorf("Classify(%q) category = %q, want %q", p, cat, Generated)
		}
	}
}

func TestGeneratedLockfiles(t *testing.T) {
	c := defaultClassifier()
	lockfileNames := []string{
		"package-lock.json",
		"pnpm-lock.yaml",
		"yarn.lock",
		"go.sum",
		"Cargo.lock",
		"Gemfile.lock",
		"composer.lock",
		"poetry.lock",
		"Pipfile.lock",
		"bun.lockb",
		"flake.lock",
	}
	for _, lf := range lockfileNames {
		cat, _ := c.Classify(lf)
		if cat != Generated {
			t.Errorf("Classify(%q) category = %q, want %q", lf, cat, Generated)
		}
	}
}

func TestGeneratedLockfilesInSubdirs(t *testing.T) {
	c := defaultClassifier()
	cat, _ := c.Classify("frontend/package-lock.json")
	if cat != Generated {
		t.Errorf("Classify(\"frontend/package-lock.json\") = %q, want %q", cat, Generated)
	}
}

func TestDocsExtensions(t *testing.T) {
	c := defaultClassifier()
	paths := []string{
		"README.md",
		"CHANGELOG.markdown",
		"guide.mdx",
		"intro.rst",
		"manual.adoc",
		"notes.txt",
		"src/README.md",
	}
	for _, p := range paths {
		cat, _ := c.Classify(p)
		if cat != Docs {
			t.Errorf("Classify(%q) category = %q, want %q", p, cat, Docs)
		}
	}
}

func TestDocsDirectories(t *testing.T) {
	c := defaultClassifier()
	paths := []string{
		"docs/api.html",
		"docs/guide.go",
		"documentation/setup.py",
		"project/docs/readme.html",
		"project/documentation/index.html",
	}
	for _, p := range paths {
		cat, _ := c.Classify(p)
		if cat != Docs {
			t.Errorf("Classify(%q) category = %q, want %q", p, cat, Docs)
		}
	}
}

func TestTestDirectories(t *testing.T) {
	c := defaultClassifier()
	paths := []string{
		"test/unit/main.go",
		"tests/integration/api_test.go",
		"spec/models/user_spec.rb",
		"specs/features/login.feature",
		"__tests__/App.test.js",
		"src/__tests__/utils.test.ts",
		"pkg/test/helper.go",
	}
	for _, p := range paths {
		cat, _ := c.Classify(p)
		if cat != Tests {
			t.Errorf("Classify(%q) category = %q, want %q", p, cat, Tests)
		}
	}
}

func TestTestFilenamePatterns(t *testing.T) {
	c := defaultClassifier()
	tests := []struct {
		path string
		desc string
	}{
		// Generic *.test.* and *.spec.*
		{"src/utils.test.js", "JS test file"},
		{"src/utils.test.ts", "TS test file"},
		{"src/utils.test.jsx", "JSX test file"},
		{"src/utils.test.tsx", "TSX test file"},
		{"src/utils.spec.js", "JS spec file"},
		{"src/utils.spec.ts", "TS spec file"},
		{"src/utils.spec.jsx", "JSX spec file"},
		{"src/utils.spec.tsx", "TSX spec file"},
		// Go
		{"pkg/handler_test.go", "Go test file"},
		{"internal/classify/classify_test.go", "Go test file nested"},
		// Python
		{"test_main.py", "Python test_ prefix"},
		{"src/test_utils.py", "Python test_ prefix nested"},
		{"src/utils_test.py", "Python _test suffix"},
		// Java
		{"src/UserTest.java", "Java Test suffix"},
		{"src/UserTests.java", "Java Tests suffix"},
		// Kotlin
		{"src/UserTest.kt", "Kotlin Test suffix"},
		// Ruby
		{"spec/user_spec.rb", "Ruby _spec suffix"},
		{"test_helper.rb", "Ruby test_ prefix"},
	}
	for _, tt := range tests {
		cat, _ := c.Classify(tt.path)
		if cat != Tests {
			t.Errorf("Classify(%q) [%s] category = %q, want %q", tt.path, tt.desc, cat, Tests)
		}
	}
}

func TestSourceExtensions(t *testing.T) {
	c := defaultClassifier()
	tests := []struct {
		path string
		lang string
	}{
		{"main.go", "Go"},
		{"lib.rs", "Rust"},
		{"app.py", "Python"},
		{"index.js", "JavaScript"},
		{"index.mjs", "JavaScript"},
		{"app.ts", "TypeScript"},
		{"component.jsx", "JSX"},
		{"component.tsx", "TSX"},
		{"Main.java", "Java"},
		{"Main.kt", "Kotlin"},
		{"main.c", "C"},
		{"main.h", "C"},
		{"main.cpp", "C++"},
		{"main.hpp", "C++"},
		{"Program.cs", "C#"},
		{"index.php", "PHP"},
		{"app.rb", "Ruby"},
		{"app.swift", "Swift"},
		{"app.scala", "Scala"},
		{"script.sh", "Shell"},
		{"script.lua", "Lua"},
		{"script.pl", "Perl"},
		{"analysis.r", "R"},
		{"app.dart", "Dart"},
		{"app.ex", "Elixir"},
		{"app.erl", "Erlang"},
		{"app.hs", "Haskell"},
		{"app.ml", "OCaml"},
		{"app.clj", "Clojure"},
		{"app.groovy", "Groovy"},
		{"app.zig", "Zig"},
		{"app.nim", "Nim"},
		{"query.sql", "SQL"},
		{"index.html", "HTML"},
		{"styles.css", "CSS"},
		{"styles.scss", "CSS"},
		{"App.vue", "Vue"},
		{"App.svelte", "Svelte"},
		{"config.yaml", "YAML"},
		{"config.toml", "TOML"},
		{"data.json", "JSON"},
		{"layout.xml", "XML"},
		{"schema.proto", "Protobuf"},
		{"schema.graphql", "GraphQL"},
		{"main.tf", "Terraform"},
	}
	for _, tt := range tests {
		cat, lang := c.Classify(tt.path)
		if cat != Source {
			t.Errorf("Classify(%q) category = %q, want %q", tt.path, cat, Source)
		}
		if lang != tt.lang {
			t.Errorf("Classify(%q) language = %q, want %q", tt.path, lang, tt.lang)
		}
	}
}

func TestOtherCategory(t *testing.T) {
	c := defaultClassifier()
	paths := []string{
		"Makefile",
		".gitignore",
		"Dockerfile",
		"LICENSE",
		"file.bin",
		"image.png",
		"archive.tar.gz",
	}
	for _, p := range paths {
		cat, _ := c.Classify(p)
		if cat != Other {
			t.Errorf("Classify(%q) category = %q, want %q", p, cat, Other)
		}
	}
}

func TestOtherLanguageIsEmpty(t *testing.T) {
	c := defaultClassifier()
	_, lang := c.Classify("Makefile")
	if lang != "" {
		t.Errorf("Classify(\"Makefile\") language = %q, want empty", lang)
	}
}

func TestPriorityGeneratedBeforeDocs(t *testing.T) {
	// A .md file in vendor/ should be generated, not docs.
	c := defaultClassifier()
	cat, _ := c.Classify("vendor/README.md")
	if cat != Generated {
		t.Errorf("Classify(\"vendor/README.md\") = %q, want %q", cat, Generated)
	}
}

func TestPriorityGeneratedBeforeTests(t *testing.T) {
	// A test file in node_modules/ should be generated, not tests.
	c := defaultClassifier()
	cat, _ := c.Classify("node_modules/lodash/lodash.test.js")
	if cat != Generated {
		t.Errorf("Classify(\"node_modules/lodash/lodash.test.js\") = %q, want %q", cat, Generated)
	}
}

func TestPriorityGeneratedBeforeSource(t *testing.T) {
	c := defaultClassifier()
	cat, _ := c.Classify("vendor/github.com/pkg/errors/errors.go")
	if cat != Generated {
		t.Errorf("Classify(\"vendor/github.com/pkg/errors/errors.go\") = %q, want %q", cat, Generated)
	}
}

func TestPriorityDocsBeforeTests(t *testing.T) {
	// A .md file in a test directory should be docs (docs extension wins).
	c := defaultClassifier()
	cat, _ := c.Classify("tests/README.md")
	if cat != Docs {
		t.Errorf("Classify(\"tests/README.md\") = %q, want %q", cat, Docs)
	}
}

func TestPriorityDocsBeforeSource(t *testing.T) {
	// A file in docs/ with a source extension should be docs.
	c := defaultClassifier()
	cat, _ := c.Classify("docs/example.go")
	if cat != Docs {
		t.Errorf("Classify(\"docs/example.go\") = %q, want %q", cat, Docs)
	}
}

func TestPriorityTestsBeforeSource(t *testing.T) {
	// A Go test file should be tests, not source.
	c := defaultClassifier()
	cat, _ := c.Classify("pkg/handler_test.go")
	if cat != Tests {
		t.Errorf("Classify(\"pkg/handler_test.go\") = %q, want %q", cat, Tests)
	}
}

func TestLanguageDetectionForTests(t *testing.T) {
	c := defaultClassifier()
	_, lang := c.Classify("pkg/handler_test.go")
	if lang != "Go" {
		t.Errorf("Classify(\"pkg/handler_test.go\") language = %q, want %q", lang, "Go")
	}
	_, lang = c.Classify("src/utils.test.ts")
	if lang != "TypeScript" {
		t.Errorf("Classify(\"src/utils.test.ts\") language = %q, want %q", lang, "TypeScript")
	}
}

func TestLanguageDetectionForDocs(t *testing.T) {
	c := defaultClassifier()
	_, lang := c.Classify("README.md")
	if lang != "" {
		t.Errorf("Classify(\"README.md\") language = %q, want empty", lang)
	}
}

func TestLanguageDetectionForGenerated(t *testing.T) {
	c := defaultClassifier()
	_, lang := c.Classify("vendor/pkg/main.go")
	if lang != "Go" {
		t.Errorf("Classify(\"vendor/pkg/main.go\") language = %q, want %q", lang, "Go")
	}
}

func TestCustomGeneratedPatterns(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Generated: {
			Patterns: []string{"generated/", "*.pb.go"},
		},
	})
	tests := []struct {
		path string
		want string
	}{
		{"generated/api.go", Generated},
		{"pkg/generated/model.go", Generated},
		{"api.pb.go", Generated},
		// Built-in should still work.
		{"vendor/dep.go", Generated},
		{"go.sum", Generated},
	}
	for _, tt := range tests {
		cat, _ := c.Classify(tt.path)
		if cat != tt.want {
			t.Errorf("Classify(%q) = %q, want %q", tt.path, cat, tt.want)
		}
	}
}

func TestCustomDocsExtensions(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Docs: {
			Extensions: []string{".pdf", ".docx"},
		},
	})
	tests := []struct {
		path string
		want string
	}{
		{"manual.pdf", Docs},
		{"report.docx", Docs},
		// Built-in should still work.
		{"README.md", Docs},
	}
	for _, tt := range tests {
		cat, _ := c.Classify(tt.path)
		if cat != tt.want {
			t.Errorf("Classify(%q) = %q, want %q", tt.path, cat, tt.want)
		}
	}
}

func TestCustomTestsPatterns(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Tests: {
			Patterns: []string{"testing/"},
		},
	})
	cat, _ := c.Classify("testing/integration/api.go")
	if cat != Tests {
		t.Errorf("Classify(\"testing/integration/api.go\") = %q, want %q", cat, Tests)
	}
}

func TestCustomSourceExtensions(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Source: {
			Extensions: []string{"wasm"},
		},
	})
	cat, _ := c.Classify("app.wasm")
	if cat != Source {
		t.Errorf("Classify(\"app.wasm\") = %q, want %q", cat, Source)
	}
}

func TestCustomExtensionsWithDot(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Docs: {
			Extensions: []string{".org"},
		},
	})
	cat, _ := c.Classify("notes.org")
	if cat != Docs {
		t.Errorf("Classify(\"notes.org\") = %q, want %q", cat, Docs)
	}
}

func TestNewWithEmptyConfig(t *testing.T) {
	c := New(config.Config{})
	cat, _ := c.Classify("main.go")
	if cat != Source {
		t.Errorf("Classify(\"main.go\") with empty config = %q, want %q", cat, Source)
	}
}

func TestEdgeCaseEmptyPath(t *testing.T) {
	c := defaultClassifier()
	cat, lang := c.Classify("")
	if cat != Other {
		t.Errorf("Classify(\"\") category = %q, want %q", cat, Other)
	}
	if lang != "" {
		t.Errorf("Classify(\"\") language = %q, want empty", lang)
	}
}

func TestEdgeCaseNoExtension(t *testing.T) {
	c := defaultClassifier()
	cat, _ := c.Classify("Makefile")
	if cat != Other {
		t.Errorf("Classify(\"Makefile\") = %q, want %q", cat, Other)
	}
}

func TestEdgeCaseJsonIsSource(t *testing.T) {
	// package.json (not a lockfile) should be source.
	c := defaultClassifier()
	cat, _ := c.Classify("package.json")
	if cat != Source {
		t.Errorf("Classify(\"package.json\") = %q, want %q", cat, Source)
	}
}

func TestEdgeCaseGoSumIsGenerated(t *testing.T) {
	c := defaultClassifier()
	cat, _ := c.Classify("go.sum")
	if cat != Generated {
		t.Errorf("Classify(\"go.sum\") = %q, want %q", cat, Generated)
	}
}

func TestEdgeCaseBuildDirNotBuildFile(t *testing.T) {
	// "build.go" should not be classified as generated.
	c := defaultClassifier()
	cat, _ := c.Classify("build.go")
	if cat != Source {
		t.Errorf("Classify(\"build.go\") = %q, want %q", cat, Source)
	}
}

func TestEdgeCaseDistDirNotDistFile(t *testing.T) {
	c := defaultClassifier()
	cat, _ := c.Classify("distribute.py")
	if cat != Source {
		t.Errorf("Classify(\"distribute.py\") = %q, want %q", cat, Source)
	}
}

func TestCustomCategorySubstringPattern(t *testing.T) {
	c := newClassifier(map[string]config.CategoryConfig{
		Generated: {
			Patterns: []string{".generated."},
		},
	})
	cat, _ := c.Classify("api.generated.go")
	if cat != Generated {
		t.Errorf("Classify(\"api.generated.go\") = %q, want %q", cat, Generated)
	}
}
