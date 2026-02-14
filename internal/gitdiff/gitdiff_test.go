package gitdiff

import (
	"fmt"
	"io"
	"os/exec"
	"testing"
)

// mockRunner is a mock CommandRunner for unit testing ref resolution.
type mockRunner struct {
	// validRefs is the set of refs that rev-parse --verify will succeed for.
	validRefs map[string]bool
	// statusOutput is returned by `git status --porcelain`.
	statusOutput string
	// mergeBaseOutput is returned by `git merge-base <base> <head>` when set.
	mergeBaseOutput string
	mergeBaseSet    bool
}

func (m *mockRunner) Run(name string, args ...string) ([]byte, error) {
	// Expect: git rev-parse --verify <ref>
	if len(args) == 3 && args[0] == "rev-parse" && args[1] == "--verify" {
		ref := args[2]
		if m.validRefs[ref] {
			return []byte("abc123\n"), nil
		}
		return nil, fmt.Errorf("fatal: Needed a single revision")
	}
	if len(args) == 2 && args[0] == "status" && args[1] == "--porcelain" {
		return []byte(m.statusOutput), nil
	}
	if len(args) == 3 && args[0] == "merge-base" {
		if m.mergeBaseSet {
			return []byte(m.mergeBaseOutput), nil
		}
		return nil, fmt.Errorf("fatal: no merge base")
	}
	return nil, fmt.Errorf("unexpected command: %s %v", name, args)
}

func (m *mockRunner) Start(name string, args ...string) (io.ReadCloser, *exec.Cmd, error) {
	return nil, nil, fmt.Errorf("Start not implemented in mock")
}

func TestResolveRefs_BaseAndHead(t *testing.T) {
	runner := &mockRunner{}
	got, err := ResolveRefs(runner, "v1.0", "feature", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v1.0...feature" {
		t.Errorf("got %q, want %q", got, "v1.0...feature")
	}
}

func TestResolveRefs_PositionalRange(t *testing.T) {
	runner := &mockRunner{}
	got, err := ResolveRefs(runner, "", "", "abc123..def456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123..def456" {
		t.Errorf("got %q, want %q", got, "abc123..def456")
	}
}

func TestResolveRefs_BaseHeadTakesPriorityOverPositional(t *testing.T) {
	runner := &mockRunner{}
	got, err := ResolveRefs(runner, "v1.0", "feature", "some..range")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "v1.0...feature" {
		t.Errorf("got %q, want %q", got, "v1.0...feature")
	}
}

func TestResolveRefs_FallbackOriginHead(t *testing.T) {
	runner := &mockRunner{validRefs: map[string]bool{"origin/HEAD": true, "main": true}}
	got, err := ResolveRefs(runner, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "origin/HEAD...HEAD" {
		t.Errorf("got %q, want %q", got, "origin/HEAD...HEAD")
	}
}

func TestResolveRefs_FallbackMain(t *testing.T) {
	runner := &mockRunner{validRefs: map[string]bool{"main": true}}
	got, err := ResolveRefs(runner, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main...HEAD" {
		t.Errorf("got %q, want %q", got, "main...HEAD")
	}
}

func TestResolveRefs_FallbackMaster(t *testing.T) {
	runner := &mockRunner{validRefs: map[string]bool{"master": true}}
	got, err := ResolveRefs(runner, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "master...HEAD" {
		t.Errorf("got %q, want %q", got, "master...HEAD")
	}
}

func TestResolveRefs_NothingResolves(t *testing.T) {
	runner := &mockRunner{validRefs: map[string]bool{}}
	_, err := ResolveRefs(runner, "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveRefs_FallbackOrder(t *testing.T) {
	// When main and master both exist, main should win.
	runner := &mockRunner{validRefs: map[string]bool{"main": true, "master": true}}
	got, err := ResolveRefs(runner, "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main...HEAD" {
		t.Errorf("got %q, want %q — main should take priority over master", got, "main...HEAD")
	}
}

func TestWorktreeDirty_Dirty(t *testing.T) {
	runner := &mockRunner{statusOutput: " M main.go\n"}
	dirty, err := WorktreeDirty(runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dirty {
		t.Fatal("expected dirty=true, got false")
	}
}

func TestWorktreeDirty_Clean(t *testing.T) {
	runner := &mockRunner{statusOutput: ""}
	dirty, err := WorktreeDirty(runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Fatal("expected dirty=false, got true")
	}
}

func TestMergeBase_Success(t *testing.T) {
	runner := &mockRunner{
		mergeBaseSet:    true,
		mergeBaseOutput: "abc123\n",
	}
	got, err := MergeBase(runner, "main", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestMergeBase_EmptyOutput(t *testing.T) {
	runner := &mockRunner{
		mergeBaseSet:    true,
		mergeBaseOutput: "",
	}
	_, err := MergeBase(runner, "main", "HEAD")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
