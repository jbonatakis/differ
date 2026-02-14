package gitdiff

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitInDir runs a git command in the given directory.
func gitInDir(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestIntegration_RunDiff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a temporary git repo.
	tmpDir := t.TempDir()

	gitInDir(t, tmpDir, "init")
	gitInDir(t, tmpDir, "config", "user.email", "test@test.com")
	gitInDir(t, tmpDir, "config", "user.name", "Test")

	// Create initial commit on main branch.
	initialFile := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(initialFile, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitInDir(t, tmpDir, "add", "hello.txt")
	gitInDir(t, tmpDir, "commit", "-m", "initial")

	// Rename default branch to main for predictability.
	gitInDir(t, tmpDir, "branch", "-M", "main")

	// Create a feature branch with changes.
	gitInDir(t, tmpDir, "checkout", "-b", "feature")

	// Modify existing file.
	if err := os.WriteFile(initialFile, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Add a new file.
	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	gitInDir(t, tmpDir, "add", ".")
	gitInDir(t, tmpDir, "commit", "-m", "feature changes")

	// Use a real runner that operates in the temp repo.
	runner := &dirRunner{dir: tmpDir}

	// Test ResolveRefs auto-detect (should find main).
	refRange, err := ResolveRefs(runner, "", "", "")
	if err != nil {
		t.Fatalf("ResolveRefs: %v", err)
	}
	if refRange != "main...HEAD" {
		t.Errorf("ResolveRefs got %q, want %q", refRange, "main...HEAD")
	}

	// Test RunDiff.
	result, err := RunDiff(runner, refRange, nil)
	if err != nil {
		t.Fatalf("RunDiff: %v", err)
	}

	output, err := io.ReadAll(result.Stdout)
	if err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	if err := result.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	diffStr := string(output)

	// Verify diff contains expected content.
	if !strings.Contains(diffStr, "hello.txt") {
		t.Errorf("diff output missing hello.txt:\n%s", diffStr)
	}
	if !strings.Contains(diffStr, "new.txt") {
		t.Errorf("diff output missing new.txt:\n%s", diffStr)
	}
	if !strings.Contains(diffStr, "+world") {
		t.Errorf("diff output missing '+world' line:\n%s", diffStr)
	}
	if !strings.Contains(diffStr, "+new file") {
		t.Errorf("diff output missing '+new file' line:\n%s", diffStr)
	}

	// Test RunDiff with pathspec filter.
	result2, err := RunDiff(runner, refRange, []string{"hello.txt"})
	if err != nil {
		t.Fatalf("RunDiff with pathspec: %v", err)
	}

	output2, err := io.ReadAll(result2.Stdout)
	if err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	if err := result2.Wait(); err != nil {
		t.Fatalf("Wait: %v", err)
	}

	diffStr2 := string(output2)
	if !strings.Contains(diffStr2, "hello.txt") {
		t.Errorf("pathspec-filtered diff missing hello.txt:\n%s", diffStr2)
	}
	if strings.Contains(diffStr2, "new.txt") {
		t.Errorf("pathspec-filtered diff should not contain new.txt:\n%s", diffStr2)
	}
}

// dirRunner runs git commands in a specific directory.
type dirRunner struct {
	dir string
}

func (d *dirRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = d.dir
	return cmd.Output()
}

func (d *dirRunner) Start(name string, args ...string) (io.ReadCloser, *exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = d.dir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	return stdout, cmd, nil
}
