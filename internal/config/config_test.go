package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNoConfigFiles(t *testing.T) {
	// Use a temp dir with no config files at all.
	tmp := t.TempDir()
	cfg, err := load("", tmp, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Empty != "exclude" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "exclude")
	}
	if cfg.Sort != "churn" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "churn")
	}
	if cfg.Include != nil {
		t.Errorf("Include = %v, want nil", cfg.Include)
	}
	if cfg.Exclude != nil {
		t.Errorf("Exclude = %v, want nil", cfg.Exclude)
	}
	if cfg.Categories != nil {
		t.Errorf("Categories = %v, want nil", cfg.Categories)
	}
}

func TestLoadRepoOnly(t *testing.T) {
	tmp := t.TempDir()
	writeYAML(t, filepath.Join(tmp, ".differ.yml"), `
include:
  - "src/**"
exclude:
  - "vendor/**"
empty: include
sort: path
categories:
  docs:
    extensions: [".md", ".txt"]
    patterns: ["docs/**"]
`)

	cfg, err := load("", tmp, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertSlice(t, "Include", cfg.Include, []string{"src/**"})
	assertSlice(t, "Exclude", cfg.Exclude, []string{"vendor/**"})
	if cfg.Empty != "include" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "include")
	}
	if cfg.Sort != "path" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "path")
	}
	cat, ok := cfg.Categories["docs"]
	if !ok {
		t.Fatal("expected docs category")
	}
	assertSlice(t, "docs.Extensions", cat.Extensions, []string{".md", ".txt"})
	assertSlice(t, "docs.Patterns", cat.Patterns, []string{"docs/**"})
}

func TestLoadGlobalOnly(t *testing.T) {
	tmp := t.TempDir()
	globalFile := filepath.Join(tmp, "global.yml")
	writeYAML(t, globalFile, `
exclude:
  - "node_modules/**"
sort: path
`)

	cfg, err := load(globalFile, "", Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertSlice(t, "Exclude", cfg.Exclude, []string{"node_modules/**"})
	if cfg.Sort != "path" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "path")
	}
	// Defaults should still apply for unset fields.
	if cfg.Empty != "exclude" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "exclude")
	}
}

func TestLoadBothWithMerge(t *testing.T) {
	tmp := t.TempDir()

	// Global config.
	globalFile := filepath.Join(tmp, "global.yml")
	writeYAML(t, globalFile, `
exclude:
  - "node_modules/**"
sort: path
empty: include
categories:
  docs:
    extensions: [".md"]
`)

	// Repo config — should override global for overlapping fields.
	repoDir := filepath.Join(tmp, "repo")
	os.MkdirAll(repoDir, 0o755)
	writeYAML(t, filepath.Join(repoDir, ".differ.yml"), `
exclude:
  - "vendor/**"
sort: churn
categories:
  docs:
    extensions: [".rst"]
  tests:
    patterns: ["test/**"]
`)

	cfg, err := load(globalFile, repoDir, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Repo exclude overrides global exclude.
	assertSlice(t, "Exclude", cfg.Exclude, []string{"vendor/**"})
	// Repo sort overrides global sort.
	if cfg.Sort != "churn" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "churn")
	}
	// Empty comes from global (repo didn't set it).
	if cfg.Empty != "include" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "include")
	}
	// Repo docs category overrides global docs.
	assertSlice(t, "docs.Extensions", cfg.Categories["docs"].Extensions, []string{".rst"})
	// Repo adds tests category.
	assertSlice(t, "tests.Patterns", cfg.Categories["tests"].Patterns, []string{"test/**"})
}

func TestCLIOverridesEverything(t *testing.T) {
	tmp := t.TempDir()

	globalFile := filepath.Join(tmp, "global.yml")
	writeYAML(t, globalFile, `
sort: path
empty: include
exclude:
  - "node_modules/**"
`)

	repoDir := filepath.Join(tmp, "repo")
	os.MkdirAll(repoDir, 0o755)
	writeYAML(t, filepath.Join(repoDir, ".differ.yml"), `
sort: churn
exclude:
  - "vendor/**"
`)

	cli := Config{
		Sort:    "path",
		Empty:   "exclude",
		Exclude: []string{"dist/**"},
		Include: []string{"src/**"},
	}

	cfg, err := load(globalFile, repoDir, cli)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Sort != "path" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "path")
	}
	if cfg.Empty != "exclude" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "exclude")
	}
	assertSlice(t, "Exclude", cfg.Exclude, []string{"dist/**"})
	assertSlice(t, "Include", cfg.Include, []string{"src/**"})
}

func TestMalformedRepoYAML(t *testing.T) {
	tmp := t.TempDir()
	writeYAML(t, filepath.Join(tmp, ".differ.yml"), `
include: [
  this is not valid yaml
  !!!
`)

	_, err := load("", tmp, Config{})
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestMalformedGlobalYAML(t *testing.T) {
	tmp := t.TempDir()
	globalFile := filepath.Join(tmp, "global.yml")
	writeYAML(t, globalFile, `
include: [
  this is not valid yaml
  !!!
`)

	_, err := load(globalFile, "", Config{})
	if err == nil {
		t.Fatal("expected error for malformed global YAML, got nil")
	}
}

func TestMissingConfigFilesAreSkipped(t *testing.T) {
	// Global path points to nonexistent file, repo dir has no .differ.yml.
	cfg, err := load("/nonexistent/global.yml", "/nonexistent/repo", Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should just get defaults.
	if cfg.Empty != "exclude" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "exclude")
	}
	if cfg.Sort != "churn" {
		t.Errorf("Sort = %q, want %q", cfg.Sort, "churn")
	}
}

func TestMergeCategoriesPreservesBase(t *testing.T) {
	base := Config{
		Categories: map[string]CategoryConfig{
			"docs":  {Extensions: []string{".md"}},
			"tests": {Patterns: []string{"test/**"}},
		},
	}
	override := Config{
		Categories: map[string]CategoryConfig{
			"docs": {Extensions: []string{".rst"}},
		},
	}
	result := merge(base, override)

	// docs should be overridden.
	assertSlice(t, "docs.Extensions", result.Categories["docs"].Extensions, []string{".rst"})
	// tests should be preserved from base.
	assertSlice(t, "tests.Patterns", result.Categories["tests"].Patterns, []string{"test/**"})
}

func TestEmptyRepoRoot(t *testing.T) {
	// Empty repoRoot should skip repo config loading.
	cfg, err := load("", "", Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Empty != "exclude" {
		t.Errorf("Empty = %q, want %q", cfg.Empty, "exclude")
	}
}

// --- helpers ---

func writeYAML(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func assertSlice(t *testing.T, name string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s = %v (len %d), want %v (len %d)", name, got, len(got), want, len(want))
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", name, i, got[i], want[i])
		}
	}
}
