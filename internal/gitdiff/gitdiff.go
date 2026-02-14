package gitdiff

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args ...string) ([]byte, error)
	Start(name string, args ...string) (io.ReadCloser, *exec.Cmd, error)
}

// defaultRunner executes real git commands.
type defaultRunner struct{}

func (d defaultRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (d defaultRunner) Start(name string, args ...string) (io.ReadCloser, *exec.Cmd, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("creating stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("starting command: %w", err)
	}
	return stdout, cmd, nil
}

// DefaultRunner is the default CommandRunner that executes real commands.
var DefaultRunner CommandRunner = defaultRunner{}

// ResolveRefs determines the git ref range to diff.
//
// Priority order:
//  1. --base and --head flags → "base...head"
//  2. Positional rev-range → returned directly
//  3. Auto-detect: origin/HEAD...HEAD → main...HEAD → master...HEAD
//
// Returns an error if no ref can be resolved.
func ResolveRefs(runner CommandRunner, base, head, positionalRange string) (string, error) {
	if base != "" && head != "" {
		return base + "..." + head, nil
	}

	if positionalRange != "" {
		return positionalRange, nil
	}

	// Auto-detect fallback chain.
	fallbacks := []struct {
		ref      string
		refRange string
	}{
		{"origin/HEAD", "origin/HEAD...HEAD"},
		{"main", "main...HEAD"},
		{"master", "master...HEAD"},
	}

	for _, fb := range fallbacks {
		if _, err := runner.Run("git", "rev-parse", "--verify", fb.ref); err == nil {
			return fb.refRange, nil
		}
	}

	return "", fmt.Errorf("cannot resolve base ref: tried origin/HEAD, main, master — are you in a git repository?")
}

// WorktreeDirty reports whether the current repository has staged or unstaged changes.
func WorktreeDirty(runner CommandRunner) (bool, error) {
	out, err := runner.Run("git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("checking working tree state: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

// MergeBase returns the merge base commit between two refs.
func MergeBase(runner CommandRunner, base, head string) (string, error) {
	out, err := runner.Run("git", "merge-base", base, head)
	if err != nil {
		return "", fmt.Errorf("resolving merge base for %q and %q: %w", base, head, err)
	}
	mergeBase := strings.TrimSpace(string(out))
	if mergeBase == "" {
		return "", fmt.Errorf("resolving merge base for %q and %q: empty output", base, head)
	}
	return mergeBase, nil
}

// DiffResult holds the output of a git diff command.
type DiffResult struct {
	Stdout io.ReadCloser
	Cmd    *exec.Cmd
}

// Wait waits for the diff command to finish and returns any error.
// The stderr output is included in the error message if the command fails.
func (d *DiffResult) Wait() error {
	err := d.Cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return fmt.Errorf("git diff failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return fmt.Errorf("git diff failed: %w", err)
	}
	return nil
}

// RunDiff executes `git diff --no-color -U0 -M <refRange> -- <pathspecs...>` and
// returns a DiffResult whose Stdout provides streaming access to the diff output.
func RunDiff(runner CommandRunner, refRange string, pathspecs []string) (*DiffResult, error) {
	args := []string{"diff", "--no-color", "-U0", "-M", refRange}
	if len(pathspecs) > 0 {
		args = append(args, "--")
		args = append(args, pathspecs...)
	}

	stdout, cmd, err := runner.Start("git", args...)
	if err != nil {
		return nil, err
	}

	return &DiffResult{Stdout: stdout, Cmd: cmd}, nil
}
