package parser

import (
	"bufio"
	"io"
	"strings"
)

// FileStat holds per-file diff statistics.
type FileStat struct {
	Path    string
	Added   int
	Deleted int
	Churn   int
}

// Parse reads unified diff output from r and returns per-file add/delete counts.
// emptyMode controls whether whitespace-only changed lines are counted:
// "exclude" (default) skips them, "include" counts them.
func Parse(r io.Reader, emptyMode string) ([]FileStat, error) {
	scanner := bufio.NewScanner(r)

	var stats []FileStat
	var current *FileStat
	inBinary := false

	flush := func() {
		if current != nil {
			current.Churn = current.Added + current.Deleted
			stats = append(stats, *current)
			current = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// New file header.
		if strings.HasPrefix(line, "diff --git ") {
			flush()
			inBinary = false
			path := parseDiffHeader(line)
			current = &FileStat{Path: path}
			continue
		}

		if current == nil {
			continue
		}

		// Detect rename.
		if strings.HasPrefix(line, "rename to ") {
			current.Path = strings.TrimPrefix(line, "rename to ")
			continue
		}

		// Detect binary files — skip the entire file.
		if strings.HasPrefix(line, "Binary files ") {
			inBinary = true
			continue
		}

		if inBinary {
			continue
		}

		// Skip diff metadata lines.
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}

		// Count additions.
		if strings.HasPrefix(line, "+") {
			content := line[1:]
			if emptyMode != "include" && strings.TrimSpace(content) == "" {
				continue
			}
			current.Added++
			continue
		}

		// Count deletions.
		if strings.HasPrefix(line, "-") {
			content := line[1:]
			if emptyMode != "include" && strings.TrimSpace(content) == "" {
				continue
			}
			current.Deleted++
			continue
		}
	}

	flush()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

// parseDiffHeader extracts the file path from a "diff --git a/... b/..." line.
// It returns the b-side path (the destination).
func parseDiffHeader(line string) string {
	// Format: "diff --git a/path b/path"
	// Find " b/" — the last occurrence handles paths with spaces.
	parts := strings.SplitN(line, " b/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	// Fallback: strip prefix and take second half.
	trimmed := strings.TrimPrefix(line, "diff --git ")
	halves := strings.SplitN(trimmed, " ", 2)
	if len(halves) == 2 {
		return strings.TrimPrefix(halves[1], "b/")
	}
	return trimmed
}
