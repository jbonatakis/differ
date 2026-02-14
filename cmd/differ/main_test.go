package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temp git repo with two commits and returns:
// - dir: the repo directory
// - baseRef: the SHA of the first commit
// - headRef: the SHA of the second commit
func setupTestRepo(t *testing.T) (dir, baseRef, headRef string) {
	t.Helper()

	dir = t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	git("init")
	git("checkout", "-b", "main")

	// First commit: create initial files.
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	writeFile(t, filepath.Join(dir, "README.md"), "# Test\n")
	writeFile(t, filepath.Join(dir, "main_test.go"), "package main\n")
	git("add", "-A")
	git("commit", "-m", "initial")
	baseRef = git("rev-parse", "HEAD")

	// Second commit: modify files and add new ones.
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n")
	writeFile(t, filepath.Join(dir, "README.md"), "# Test\n\nA description.\n")
	writeFile(t, filepath.Join(dir, "main_test.go"), "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n")
	writeFile(t, filepath.Join(dir, "go.sum"), "some/dep v1.0.0 h1:abc=\n")
	git("add", "-A")
	git("commit", "-m", "add features")
	headRef = git("rev-parse", "HEAD")

	return dir, baseRef, headRef
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildBinary builds the differ binary and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "differ")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/differ")
	cmd.Dir = moduleRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	return bin
}

// moduleRoot returns the root of the Go module (where go.mod lives).
func moduleRoot(t *testing.T) string {
	t.Helper()
	// We are in cmd/differ, so go up two levels.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(filepath.Dir(wd))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("could not find go.mod at %s: %v", root, err)
	}
	return root
}

// runDiffer runs the differ binary in the given directory with the given args.
func runDiffer(t *testing.T, bin, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("failed to run differ: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func TestE2E_DefaultTextOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	stdout, _, exitCode := runDiffer(t, bin, dir, baseRef+"...HEAD", "--no-color")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	// Should contain category summaries and a total line.
	if !strings.Contains(stdout, "Source:") {
		t.Errorf("expected Source category in output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Total:") {
		t.Errorf("expected Total line in output, got:\n%s", stdout)
	}
}

func TestE2E_ExplicitRefs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, headRef := setupTestRepo(t)

	stdout, _, exitCode := runDiffer(t, bin, dir, "--base", baseRef, "--head", headRef, "--no-color")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Total:") {
		t.Errorf("expected Total line in output, got:\n%s", stdout)
	}
}

func TestE2E_ListMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	stdout, _, exitCode := runDiffer(t, bin, dir, baseRef+"...HEAD", "-l", "--no-color")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	// List mode should show file paths.
	if !strings.Contains(stdout, "main.go") {
		t.Errorf("expected main.go in list output, got:\n%s", stdout)
	}
	// Should also have the summary.
	if !strings.Contains(stdout, "Total:") {
		t.Errorf("expected Total line in list output, got:\n%s", stdout)
	}
}

func TestE2E_ListOnlyMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	stdout, _, exitCode := runDiffer(t, bin, dir, baseRef+"...HEAD", "-L", "--no-color")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	// List-only mode should show file paths but NOT the Total summary line.
	if !strings.Contains(stdout, "main.go") {
		t.Errorf("expected main.go in list-only output, got:\n%s", stdout)
	}
	if strings.Contains(stdout, "Total:") {
		t.Errorf("list-only mode should not show Total line, got:\n%s", stdout)
	}
}

func TestE2E_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, headRef := setupTestRepo(t)

	stdout, _, exitCode := runDiffer(t, bin, dir, "--base", baseRef, "--head", headRef, "--format", "json")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, stdout)
	}

	// Verify top-level keys.
	for _, key := range []string{"meta", "total", "by_category", "by_file"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing top-level key %q in JSON output", key)
		}
	}

	// Verify meta contains base and head.
	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not an object")
	}
	if meta["base"] != baseRef {
		t.Errorf("meta.base = %v, want %s", meta["base"], baseRef)
	}
	if meta["head"] != headRef {
		t.Errorf("meta.head = %v, want %s", meta["head"], headRef)
	}
	if meta["empty"] != "exclude" {
		t.Errorf("meta.empty = %v, want 'exclude'", meta["empty"])
	}

	// Verify total has positive churn.
	total := result["total"].(map[string]interface{})
	if total["churn"].(float64) <= 0 {
		t.Errorf("expected positive total churn, got %v", total["churn"])
	}

	// Verify by_file has entries.
	byFile := result["by_file"].([]interface{})
	if len(byFile) == 0 {
		t.Error("expected non-empty by_file array")
	}
}

func TestE2E_CategoryFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	// Filter to only docs category.
	stdout, _, exitCode := runDiffer(t, bin, dir, baseRef+"...HEAD", "--category", "docs", "--format", "json")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout)
	}

	// All files should be docs.
	byFile := result["by_file"].([]interface{})
	for _, f := range byFile {
		file := f.(map[string]interface{})
		if file["category"] != "docs" {
			t.Errorf("expected category 'docs', got %q for file %v", file["category"], file["path"])
		}
	}
}

func TestE2E_IncludeExclude(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	// Include only .go files, exclude test files.
	stdout, _, exitCode := runDiffer(t, bin, dir, baseRef+"...HEAD",
		"--include", "*.go", "--exclude", "*_test.go", "--format", "json")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout)
	}

	byFile := result["by_file"].([]interface{})
	for _, f := range byFile {
		file := f.(map[string]interface{})
		path := file["path"].(string)
		if !strings.HasSuffix(path, ".go") {
			t.Errorf("expected only .go files, got %q", path)
		}
		if strings.HasSuffix(path, "_test.go") {
			t.Errorf("expected test files excluded, got %q", path)
		}
	}

	// Should have at least main.go.
	if len(byFile) == 0 {
		t.Error("expected at least one file after include/exclude filtering")
	}
}

func TestE2E_DefaultRefResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir, baseRef, _ := setupTestRepo(t)

	// Create a feature branch off of baseRef so main...HEAD produces a diff.
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	// Create a feature branch from the second commit (HEAD).
	git("checkout", "-b", "feature")
	writeFile(t, filepath.Join(dir, "feature.go"), "package main\n\nfunc feature() {}\n")
	git("add", "-A")
	git("commit", "-m", "feature work")

	// Now we're on 'feature' branch; main exists, so default resolution should use main...HEAD.
	stdout, _, exitCode := runDiffer(t, bin, dir, "--no-color")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; base=%s", exitCode, baseRef)
	}

	if !strings.Contains(stdout, "Total:") {
		t.Errorf("expected Total line from default ref resolution, got:\n%s", stdout)
	}
}

func TestE2E_DefaultIncludesDirtyWorktree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	bin := buildBinary(t)
	dir := t.TempDir()

	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	git("init")
	git("checkout", "-b", "main")
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	git("add", "-A")
	git("commit", "-m", "initial")

	// Leave local changes uncommitted.
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n")

	stdout, _, exitCode := runDiffer(t, bin, dir, "--format", "json")
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout)
	}

	total := result["total"].(map[string]interface{})
	if total["churn"].(float64) <= 0 {
		t.Fatalf("expected positive churn for dirty worktree, got %v", total["churn"])
	}

	meta := result["meta"].(map[string]interface{})
	if meta["head"] != "WORKTREE" {
		t.Fatalf("expected meta.head WORKTREE for dirty auto mode, got %v", meta["head"])
	}

	foundMain := false
	for _, f := range result["by_file"].([]interface{}) {
		file := f.(map[string]interface{})
		if file["path"] == "main.go" {
			foundMain = true
			break
		}
	}
	if !foundMain {
		t.Fatalf("expected main.go in by_file, got %v", result["by_file"])
	}
}

func TestE2E_ParseRefRange(t *testing.T) {
	tests := []struct {
		input    string
		wantBase string
		wantHead string
	}{
		{"main...HEAD", "main", "HEAD"},
		{"abc123..def456", "abc123", "def456"},
		{"single-ref", "single-ref", ""},
	}

	for _, tt := range tests {
		base, head := parseRefRange(tt.input)
		if base != tt.wantBase || head != tt.wantHead {
			t.Errorf("parseRefRange(%q) = (%q, %q), want (%q, %q)",
				tt.input, base, head, tt.wantBase, tt.wantHead)
		}
	}
}
