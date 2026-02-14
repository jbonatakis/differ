package filter

import (
	"github.com/bmatcuk/doublestar/v4"
	"github.com/jackbonatakis/differ/internal/parser"
)

// FilterConfig controls which files to keep or discard.
type FilterConfig struct {
	Include    []string // glob patterns; if non-empty, only matching files are kept
	Exclude    []string // glob patterns; matching files are removed
	Categories []string // category names; if non-empty, only matching categories are kept
}

// CategoryFunc returns the category string for a given file path.
type CategoryFunc func(path string) string

// Filter applies include/exclude glob patterns and category restrictions to stats.
// categoryFn is called to determine each file's category when Categories is non-empty.
func Filter(stats []parser.FileStat, cfg FilterConfig, categoryFn CategoryFunc) []parser.FileStat {
	var result []parser.FileStat
	for _, fs := range stats {
		if !matchInclude(fs.Path, cfg.Include) {
			continue
		}
		if matchExclude(fs.Path, cfg.Exclude) {
			continue
		}
		if !matchCategory(fs.Path, cfg.Categories, categoryFn) {
			continue
		}
		result = append(result, fs)
	}
	return result
}

// matchInclude returns true if the path matches at least one include pattern,
// or if there are no include patterns.
func matchInclude(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if matched, _ := doublestar.Match(p, path); matched {
			return true
		}
	}
	return false
}

// matchExclude returns true if the path matches any exclude pattern.
func matchExclude(path string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := doublestar.Match(p, path); matched {
			return true
		}
	}
	return false
}

// matchCategory returns true if the file's category is in the allowed list,
// or if no category filter is set.
func matchCategory(path string, categories []string, categoryFn CategoryFunc) bool {
	if len(categories) == 0 {
		return true
	}
	if categoryFn == nil {
		return true
	}
	cat := categoryFn(path)
	for _, c := range categories {
		if c == cat {
			return true
		}
	}
	return false
}
