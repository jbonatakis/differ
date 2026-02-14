package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Category display names and their corresponding internal keys.
var categoryOrder = []struct {
	key     string
	display string
}{
	{"docs", "Documentation"},
	{"tests", "Tests"},
	{"source", "Source"},
	{"generated", "Generated"},
	{"other", "Uncategorized"},
}

const (
	addColor   = "\033[32m"
	delColor   = "\033[31m"
	resetColor = "\033[0m"
)

// FileStat holds per-file statistics with classification info.
type FileStat struct {
	Path     string
	Added    int
	Deleted  int
	Churn    int
	Category string
	Language string
}

// CategoryTotal holds aggregate stats for a category.
type CategoryTotal struct {
	Added     int
	Deleted   int
	Churn     int
	FileCount int
}

// Meta holds metadata about the diff operation.
type Meta struct {
	Base      string   `json:"base"`
	Head      string   `json:"head"`
	Empty     string   `json:"empty"`
	Pathspecs []string `json:"pathspecs"`
	Timestamp string   `json:"timestamp"`
}

// Summary holds the complete output data.
type Summary struct {
	Totals         CategoryTotal
	CategoryTotals map[string]CategoryTotal
	FileStats      []FileStat
	Meta           Meta
}

// OutputOpts controls text rendering behavior.
type OutputOpts struct {
	List     bool
	ListOnly bool
	Sort     string // "churn" (default) or "path"
	NoColor  bool
}

// RenderText writes the human-readable text output to w.
func RenderText(w io.Writer, summary Summary, opts OutputOpts) {
	if !opts.ListOnly {
		renderSummary(w, summary, opts)
	}

	if opts.List || opts.ListOnly {
		if !opts.ListOnly {
			fmt.Fprintln(w)
		}
		renderFileList(w, summary, opts)
	}
}

func renderSummary(w io.Writer, summary Summary, opts OutputOpts) {
	labelWidth, addWidth, delWidth, churnWidth := summaryWidths(summary)

	for _, cat := range categoryOrder {
		ct, ok := summary.CategoryTotals[cat.key]
		if !ok || ct.Churn == 0 {
			continue
		}
		gap := strings.Repeat(" ", labelWidth-len(cat.display)+1)
		fmt.Fprintf(w, "%s:%s%s (%*d) [%d %s]\n",
			cat.display, gap, formatAddDel(ct.Added, ct.Deleted, addWidth, delWidth, opts.NoColor), churnWidth, ct.Churn, ct.FileCount, fileWord(ct.FileCount))
	}

	t := summary.Totals
	gap := strings.Repeat(" ", labelWidth-len("Total")+1)
	fmt.Fprintf(w, "Total:%s%s (%*d) [%d %s]\n",
		gap, formatAddDel(t.Added, t.Deleted, addWidth, delWidth, opts.NoColor), churnWidth, t.Churn, t.FileCount, fileWord(t.FileCount))
}

func renderFileList(w io.Writer, summary Summary, opts OutputOpts) {
	sorted := make([]FileStat, len(summary.FileStats))
	copy(sorted, summary.FileStats)
	sortFiles(sorted, opts.Sort)
	addWidth, delWidth := fileWidths(sorted)

	// Group by category in display order.
	grouped := make(map[string][]FileStat)
	for _, f := range sorted {
		grouped[f.Category] = append(grouped[f.Category], f)
	}

	first := true
	for _, cat := range categoryOrder {
		files, ok := grouped[cat.key]
		if !ok || len(files) == 0 {
			continue
		}
		if !first {
			fmt.Fprintln(w)
		}
		first = false

		fmt.Fprintf(w, "[%s]\n", cat.display)
		for _, f := range files {
			fmt.Fprintf(w, "%s %s\n", formatAddDel(f.Added, f.Deleted, addWidth, delWidth, opts.NoColor), f.Path)
		}
	}
}

func formatAddDel(added, deleted, addWidth, delWidth int, noColor bool) string {
	if noColor {
		return fmt.Sprintf("+%*d -%*d", addWidth, added, delWidth, deleted)
	}
	return fmt.Sprintf("%s+%*d%s %s-%*d%s", addColor, addWidth, added, resetColor, delColor, delWidth, deleted, resetColor)
}

func summaryWidths(summary Summary) (labelWidth, addWidth, delWidth, churnWidth int) {
	labelWidth = len("Total")
	addWidth = digitWidth(summary.Totals.Added)
	delWidth = digitWidth(summary.Totals.Deleted)
	churnWidth = digitWidth(summary.Totals.Churn)

	for _, cat := range categoryOrder {
		if len(cat.display) > labelWidth {
			labelWidth = len(cat.display)
		}
		ct, ok := summary.CategoryTotals[cat.key]
		if !ok || ct.Churn == 0 {
			continue
		}
		if w := digitWidth(ct.Added); w > addWidth {
			addWidth = w
		}
		if w := digitWidth(ct.Deleted); w > delWidth {
			delWidth = w
		}
		if w := digitWidth(ct.Churn); w > churnWidth {
			churnWidth = w
		}
	}

	return labelWidth, addWidth, delWidth, churnWidth
}

func fileWidths(files []FileStat) (addWidth, delWidth int) {
	addWidth, delWidth = 1, 1
	for _, f := range files {
		if w := digitWidth(f.Added); w > addWidth {
			addWidth = w
		}
		if w := digitWidth(f.Deleted); w > delWidth {
			delWidth = w
		}
	}
	return addWidth, delWidth
}

func digitWidth(n int) int {
	if n == 0 {
		return 1
	}
	width := 0
	for n > 0 {
		width++
		n /= 10
	}
	return width
}

func fileWord(count int) string {
	if count == 1 {
		return "file"
	}
	return "files"
}

func sortFiles(files []FileStat, sortMode string) {
	switch strings.ToLower(sortMode) {
	case "path":
		sort.Slice(files, func(i, j int) bool {
			return files[i].Path < files[j].Path
		})
	default: // "churn" or empty
		sort.Slice(files, func(i, j int) bool {
			if files[i].Churn != files[j].Churn {
				return files[i].Churn > files[j].Churn
			}
			return files[i].Path < files[j].Path
		})
	}
}

// jsonOutput is the top-level JSON structure.
type jsonOutput struct {
	Meta       jsonMeta                 `json:"meta"`
	Total      jsonTotal                `json:"total"`
	ByCategory map[string]jsonCatDetail `json:"by_category"`
	ByFile     []jsonFile               `json:"by_file"`
}

type jsonMeta struct {
	Base      string   `json:"base"`
	Head      string   `json:"head"`
	Empty     string   `json:"empty"`
	Pathspecs []string `json:"pathspecs"`
	Timestamp string   `json:"timestamp"`
}

type jsonTotal struct {
	Added   int `json:"added"`
	Deleted int `json:"deleted"`
	Churn   int `json:"churn"`
	Files   int `json:"files"`
}

type jsonCatDetail struct {
	Added     int      `json:"added"`
	Deleted   int      `json:"deleted"`
	Churn     int      `json:"churn"`
	Files     []string `json:"files"`
	FileCount int      `json:"file_count"`
}

type jsonFile struct {
	Path     string `json:"path"`
	Added    int    `json:"added"`
	Deleted  int    `json:"deleted"`
	Churn    int    `json:"churn"`
	Category string `json:"category"`
	Language string `json:"language"`
}

// RenderJSON writes JSON output to w.
func RenderJSON(w io.Writer, summary Summary) error {
	byCategory := make(map[string]jsonCatDetail)

	// Build file lists per category.
	catFiles := make(map[string][]string)
	for _, f := range summary.FileStats {
		catFiles[f.Category] = append(catFiles[f.Category], f.Path)
	}

	for cat, ct := range summary.CategoryTotals {
		byCategory[cat] = jsonCatDetail{
			Added:     ct.Added,
			Deleted:   ct.Deleted,
			Churn:     ct.Churn,
			Files:     catFiles[cat],
			FileCount: ct.FileCount,
		}
	}

	byFile := make([]jsonFile, 0, len(summary.FileStats))
	for _, f := range summary.FileStats {
		byFile = append(byFile, jsonFile{
			Path:     f.Path,
			Added:    f.Added,
			Deleted:  f.Deleted,
			Churn:    f.Churn,
			Category: f.Category,
			Language: f.Language,
		})
	}

	pathspecs := summary.Meta.Pathspecs
	if pathspecs == nil {
		pathspecs = []string{}
	}

	out := jsonOutput{
		Meta: jsonMeta{
			Base:      summary.Meta.Base,
			Head:      summary.Meta.Head,
			Empty:     summary.Meta.Empty,
			Pathspecs: pathspecs,
			Timestamp: summary.Meta.Timestamp,
		},
		Total: jsonTotal{
			Added:   summary.Totals.Added,
			Deleted: summary.Totals.Deleted,
			Churn:   summary.Totals.Churn,
			Files:   summary.Totals.FileCount,
		},
		ByCategory: byCategory,
		ByFile:     byFile,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
