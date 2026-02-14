package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// testSummary returns a representative Summary for tests.
func testSummary() Summary {
	return Summary{
		Totals: CategoryTotal{Added: 186, Deleted: 104, Churn: 290, FileCount: 28},
		CategoryTotals: map[string]CategoryTotal{
			"docs":      {Added: 12, Deleted: 3, Churn: 15, FileCount: 4},
			"tests":     {Added: 45, Deleted: 8, Churn: 53, FileCount: 6},
			"source":    {Added: 120, Deleted: 90, Churn: 210, FileCount: 14},
			"generated": {Added: 2, Deleted: 2, Churn: 4, FileCount: 1},
			"other":     {Added: 7, Deleted: 1, Churn: 8, FileCount: 3},
		},
		FileStats: []FileStat{
			{Path: "internal/foo/bar.go", Added: 50, Deleted: 30, Churn: 80, Category: "source", Language: "Go"},
			{Path: "internal/baz/baz.go", Added: 70, Deleted: 60, Churn: 130, Category: "source", Language: "Go"},
			{Path: "pkg/a/a_test.go", Added: 20, Deleted: 2, Churn: 22, Category: "tests", Language: "Go"},
			{Path: "docs/README.md", Added: 12, Deleted: 3, Churn: 15, Category: "docs", Language: ""},
			{Path: "go.sum", Added: 2, Deleted: 2, Churn: 4, Category: "generated", Language: ""},
			{Path: "Makefile", Added: 7, Deleted: 1, Churn: 8, Category: "other", Language: ""},
			{Path: "internal/foo/specs/bar_spec.rb", Added: 25, Deleted: 6, Churn: 31, Category: "tests", Language: "Ruby"},
		},
		Meta: Meta{
			Base:      "main",
			Head:      "HEAD",
			Empty:     "exclude",
			Pathspecs: []string{"docs/", "internal/"},
			Timestamp: "2024-01-15T10:30:00Z",
		},
	}
}

func TestRenderTextSummaryOnly(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{NoColor: true})

	got := buf.String()
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")

	expected := []string{
		"Documentation: +12 -3 (15) [4 files]",
		"Tests: +45 -8 (53) [6 files]",
		"Source: +120 -90 (210) [14 files]",
		"Generated: +2 -2 (4) [1 files]",
		"Uncategorized: +7 -1 (8) [3 files]",
		"Total: +186 -104 (290) [28 files]",
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d:\n%s", len(expected), len(lines), got)
	}

	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("line %d: expected %q, got %q", i, exp, lines[i])
		}
	}
}

func TestRenderTextCategoryOrder(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{NoColor: true})

	got := buf.String()
	docIdx := strings.Index(got, "Documentation:")
	testIdx := strings.Index(got, "Tests:")
	srcIdx := strings.Index(got, "Source:")
	genIdx := strings.Index(got, "Generated:")
	uncatIdx := strings.Index(got, "Uncategorized:")
	totalIdx := strings.Index(got, "Total:")

	if docIdx >= testIdx || testIdx >= srcIdx || srcIdx >= genIdx || genIdx >= uncatIdx || uncatIdx >= totalIdx {
		t.Errorf("categories not in expected order:\ndoc=%d test=%d src=%d gen=%d uncat=%d total=%d",
			docIdx, testIdx, srcIdx, genIdx, uncatIdx, totalIdx)
	}
}

func TestRenderTextOmitsZeroChurn(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{
		Totals: CategoryTotal{Added: 10, Deleted: 5, Churn: 15, FileCount: 2},
		CategoryTotals: map[string]CategoryTotal{
			"source": {Added: 10, Deleted: 5, Churn: 15, FileCount: 2},
			"docs":   {Added: 0, Deleted: 0, Churn: 0, FileCount: 0},
		},
		FileStats: []FileStat{},
	}

	RenderText(&buf, s, OutputOpts{NoColor: true})
	got := buf.String()

	if strings.Contains(got, "Documentation:") {
		t.Error("zero-churn category should be omitted")
	}
	if !strings.Contains(got, "Source:") {
		t.Error("non-zero-churn category should be present")
	}
}

func TestRenderTextWithColor(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{
		Totals: CategoryTotal{Added: 10, Deleted: 5, Churn: 15, FileCount: 2},
		CategoryTotals: map[string]CategoryTotal{
			"source": {Added: 10, Deleted: 5, Churn: 15, FileCount: 2},
		},
		FileStats: []FileStat{},
	}

	RenderText(&buf, s, OutputOpts{NoColor: false})
	got := buf.String()

	if !strings.Contains(got, "\033[32m") {
		t.Error("expected ANSI color code for source category")
	}
	if !strings.Contains(got, resetColor) {
		t.Error("expected ANSI reset code")
	}
}

func TestRenderTextNoColor(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{NoColor: true})
	got := buf.String()

	if strings.Contains(got, "\033[") {
		t.Error("no-color output should not contain ANSI escape codes")
	}
}

func TestRenderTextListMode(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{List: true, NoColor: true})

	got := buf.String()

	// Should have summary.
	if !strings.Contains(got, "Total:") {
		t.Error("list mode should include summary")
	}

	// Should have file list headers.
	if !strings.Contains(got, "[Tests]") {
		t.Error("list mode should include category group headers")
	}
	if !strings.Contains(got, "[Source]") {
		t.Error("list mode should include Source group header")
	}

	// Should have file lines.
	if !strings.Contains(got, "pkg/a/a_test.go") {
		t.Error("list mode should include file paths")
	}
}

func TestRenderTextListOnlyMode(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{ListOnly: true, NoColor: true})

	got := buf.String()

	// Should NOT have summary.
	if strings.Contains(got, "Total:") {
		t.Error("list-only mode should not include summary")
	}

	// Should have file list.
	if !strings.Contains(got, "[Source]") {
		t.Error("list-only mode should include category group headers")
	}
	if !strings.Contains(got, "internal/foo/bar.go") {
		t.Error("list-only mode should include file paths")
	}
}

func TestRenderTextSortByChurn(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{ListOnly: true, Sort: "churn", NoColor: true})

	got := buf.String()

	// In source category, baz.go (churn=130) should appear before bar.go (churn=80).
	bazIdx := strings.Index(got, "internal/baz/baz.go")
	barIdx := strings.Index(got, "internal/foo/bar.go")
	if bazIdx > barIdx {
		t.Error("churn sort: higher churn file should appear first")
	}
}

func TestRenderTextSortByPath(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{ListOnly: true, Sort: "path", NoColor: true})

	got := buf.String()

	// In source category, baz.go (internal/baz/) should appear before bar.go (internal/foo/) alphabetically.
	bazIdx := strings.Index(got, "internal/baz/baz.go")
	barIdx := strings.Index(got, "internal/foo/bar.go")
	if bazIdx > barIdx {
		t.Error("path sort: alphabetically earlier file should appear first")
	}
}

func TestRenderTextFileListGroupOrder(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{ListOnly: true, NoColor: true})

	got := buf.String()

	// Category groups should follow display order.
	docIdx := strings.Index(got, "[Documentation]")
	testIdx := strings.Index(got, "[Tests]")
	srcIdx := strings.Index(got, "[Source]")
	genIdx := strings.Index(got, "[Generated]")
	uncatIdx := strings.Index(got, "[Uncategorized]")

	if docIdx >= testIdx || testIdx >= srcIdx || srcIdx >= genIdx || genIdx >= uncatIdx {
		t.Errorf("file list groups not in expected order: doc=%d test=%d src=%d gen=%d uncat=%d",
			docIdx, testIdx, srcIdx, genIdx, uncatIdx)
	}
}

func TestRenderTextEmptySummary(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{
		Totals:         CategoryTotal{},
		CategoryTotals: map[string]CategoryTotal{},
		FileStats:      []FileStat{},
	}
	RenderText(&buf, s, OutputOpts{NoColor: true})
	got := buf.String()

	// Should only have the Total line.
	if !strings.Contains(got, "Total: +0 -0 (0) [0 files]") {
		t.Errorf("empty summary should show zero total, got: %s", got)
	}
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("empty summary should have 1 line, got %d", len(lines))
	}
}

// --- JSON tests ---

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	err := RenderJSON(&buf, s)
	if err != nil {
		t.Fatalf("RenderJSON error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify top-level keys.
	for _, key := range []string{"meta", "total", "by_category", "by_file"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing top-level key: %s", key)
		}
	}
}

func TestRenderJSONMeta(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderJSON(&buf, s)

	var result struct {
		Meta jsonMeta `json:"meta"`
	}
	json.Unmarshal(buf.Bytes(), &result)

	if result.Meta.Base != "main" {
		t.Errorf("meta.base: expected %q, got %q", "main", result.Meta.Base)
	}
	if result.Meta.Head != "HEAD" {
		t.Errorf("meta.head: expected %q, got %q", "HEAD", result.Meta.Head)
	}
	if result.Meta.Empty != "exclude" {
		t.Errorf("meta.empty: expected %q, got %q", "exclude", result.Meta.Empty)
	}
	if len(result.Meta.Pathspecs) != 2 {
		t.Errorf("meta.pathspecs: expected 2, got %d", len(result.Meta.Pathspecs))
	}
	if result.Meta.Timestamp != "2024-01-15T10:30:00Z" {
		t.Errorf("meta.timestamp: expected %q, got %q", "2024-01-15T10:30:00Z", result.Meta.Timestamp)
	}
}

func TestRenderJSONTotal(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderJSON(&buf, s)

	var result struct {
		Total jsonTotal `json:"total"`
	}
	json.Unmarshal(buf.Bytes(), &result)

	if result.Total.Added != 186 {
		t.Errorf("total.added: expected 186, got %d", result.Total.Added)
	}
	if result.Total.Deleted != 104 {
		t.Errorf("total.deleted: expected 104, got %d", result.Total.Deleted)
	}
	if result.Total.Churn != 290 {
		t.Errorf("total.churn: expected 290, got %d", result.Total.Churn)
	}
	if result.Total.Files != 28 {
		t.Errorf("total.files: expected 28, got %d", result.Total.Files)
	}
}

func TestRenderJSONByCategory(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderJSON(&buf, s)

	var result struct {
		ByCategory map[string]jsonCatDetail `json:"by_category"`
	}
	json.Unmarshal(buf.Bytes(), &result)

	src, ok := result.ByCategory["source"]
	if !ok {
		t.Fatal("by_category missing 'source'")
	}
	if src.Added != 120 {
		t.Errorf("source.added: expected 120, got %d", src.Added)
	}
	if src.FileCount != 14 {
		t.Errorf("source.file_count: expected 14, got %d", src.FileCount)
	}

	tests, ok := result.ByCategory["tests"]
	if !ok {
		t.Fatal("by_category missing 'tests'")
	}
	if tests.Churn != 53 {
		t.Errorf("tests.churn: expected 53, got %d", tests.Churn)
	}
}

func TestRenderJSONByFile(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderJSON(&buf, s)

	var result struct {
		ByFile []jsonFile `json:"by_file"`
	}
	json.Unmarshal(buf.Bytes(), &result)

	if len(result.ByFile) != len(s.FileStats) {
		t.Fatalf("by_file: expected %d entries, got %d", len(s.FileStats), len(result.ByFile))
	}

	// Verify first file.
	f := result.ByFile[0]
	if f.Path != "internal/foo/bar.go" {
		t.Errorf("by_file[0].path: expected %q, got %q", "internal/foo/bar.go", f.Path)
	}
	if f.Category != "source" {
		t.Errorf("by_file[0].category: expected %q, got %q", "source", f.Category)
	}
	if f.Language != "Go" {
		t.Errorf("by_file[0].language: expected %q, got %q", "Go", f.Language)
	}
}

func TestRenderJSONNilPathspecs(t *testing.T) {
	var buf bytes.Buffer
	s := Summary{
		Totals:         CategoryTotal{},
		CategoryTotals: map[string]CategoryTotal{},
		FileStats:      []FileStat{},
		Meta:           Meta{Base: "main", Head: "HEAD"},
	}
	RenderJSON(&buf, s)

	// Pathspecs should be [] not null.
	if strings.Contains(buf.String(), `"pathspecs": null`) {
		t.Error("pathspecs should be empty array, not null")
	}
	if !strings.Contains(buf.String(), `"pathspecs": []`) {
		t.Error("pathspecs should be empty array []")
	}
}

func TestRenderJSONByCategoryFiles(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderJSON(&buf, s)

	var result struct {
		ByCategory map[string]jsonCatDetail `json:"by_category"`
	}
	json.Unmarshal(buf.Bytes(), &result)

	tests := result.ByCategory["tests"]
	if len(tests.Files) != 2 {
		t.Fatalf("tests.files: expected 2 paths, got %d", len(tests.Files))
	}
	// Check that test file paths are included.
	found := false
	for _, f := range tests.Files {
		if f == "pkg/a/a_test.go" {
			found = true
		}
	}
	if !found {
		t.Error("tests.files should contain pkg/a/a_test.go")
	}
}

func TestRenderTextListModeColorInHeaders(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()
	RenderText(&buf, s, OutputOpts{ListOnly: true, NoColor: false})
	got := buf.String()

	// File list category headers should be colored.
	if !strings.Contains(got, "\033[32m") {
		t.Error("expected color in file list category headers")
	}
}
