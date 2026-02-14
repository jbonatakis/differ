package parser

import (
	"strings"
	"testing"
)

func TestSimpleSingleFile(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
+
 func main() {
-	println("hello")
+	fmt.Println("hello")
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	f := stats[0]
	if f.Path != "main.go" {
		t.Errorf("path = %q, want %q", f.Path, "main.go")
	}
	// "import "fmt"" and "fmt.Println("hello")" are added (empty line excluded)
	if f.Added != 2 {
		t.Errorf("added = %d, want 2", f.Added)
	}
	// 'println("hello")' is deleted
	if f.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", f.Deleted)
	}
	if f.Churn != 3 {
		t.Errorf("churn = %d, want 3", f.Churn)
	}
}

func TestSimpleSingleFileIncludeEmpty(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
+
 func main() {
-	println("hello")
+	fmt.Println("hello")
`
	stats, err := Parse(strings.NewReader(diff), "include")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	f := stats[0]
	// With include, the empty "+" line is also counted
	if f.Added != 3 {
		t.Errorf("added = %d, want 3", f.Added)
	}
	if f.Deleted != 1 {
		t.Errorf("deleted = %d, want 1", f.Deleted)
	}
	if f.Churn != 4 {
		t.Errorf("churn = %d, want 4", f.Churn)
	}
}

func TestMultiFile(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,2 +1,3 @@
 package a
+func A() {}
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1,2 +1,2 @@
 package b
-func Old() {}
+func New() {}
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 files, got %d", len(stats))
	}
	if stats[0].Path != "a.go" || stats[0].Added != 1 || stats[0].Deleted != 0 {
		t.Errorf("a.go: got %+v", stats[0])
	}
	if stats[1].Path != "b.go" || stats[1].Added != 1 || stats[1].Deleted != 1 {
		t.Errorf("b.go: got %+v", stats[1])
	}
}

func TestRenamedFile(t *testing.T) {
	diff := `diff --git a/old.go b/new.go
similarity index 95%
rename from old.go
rename to new.go
--- a/old.go
+++ b/new.go
@@ -1,3 +1,3 @@
 package pkg
-func Old() {}
+func New() {}
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	if stats[0].Path != "new.go" {
		t.Errorf("path = %q, want %q", stats[0].Path, "new.go")
	}
	if stats[0].Added != 1 || stats[0].Deleted != 1 {
		t.Errorf("stats = %+v, want Added=1, Deleted=1", stats[0])
	}
}

func TestBinaryFileSkipped(t *testing.T) {
	diff := `diff --git a/image.png b/image.png
Binary files /dev/null and b/image.png differ
diff --git a/code.go b/code.go
--- a/code.go
+++ b/code.go
@@ -1 +1,2 @@
 package code
+func F() {}
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 files, got %d", len(stats))
	}
	// Binary file should have zero counts.
	if stats[0].Path != "image.png" {
		t.Errorf("path = %q, want %q", stats[0].Path, "image.png")
	}
	if stats[0].Added != 0 || stats[0].Deleted != 0 {
		t.Errorf("binary file should have zero counts, got %+v", stats[0])
	}
	// Code file should be counted normally.
	if stats[1].Path != "code.go" || stats[1].Added != 1 {
		t.Errorf("code.go: got %+v", stats[1])
	}
}

func TestEmptyLineExclusion(t *testing.T) {
	diff := `diff --git a/f.go b/f.go
--- a/f.go
+++ b/f.go
@@ -1,3 +1,5 @@
 package f
+
+
+func F() {}
-
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	// Only "func F() {}" should be counted as added; "+" and "+\t" are whitespace-only.
	// The deleted empty line "-" should also be excluded.
	if stats[0].Added != 1 {
		t.Errorf("added = %d, want 1", stats[0].Added)
	}
	if stats[0].Deleted != 0 {
		t.Errorf("deleted = %d, want 0", stats[0].Deleted)
	}
}

func TestEmptyLineInclusion(t *testing.T) {
	diff := `diff --git a/f.go b/f.go
--- a/f.go
+++ b/f.go
@@ -1,3 +1,5 @@
 package f
+
+
+func F() {}
-
`
	stats, err := Parse(strings.NewReader(diff), "include")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	if stats[0].Added != 3 {
		t.Errorf("added = %d, want 3", stats[0].Added)
	}
	if stats[0].Deleted != 1 {
		t.Errorf("deleted = %d, want 1", stats[0].Deleted)
	}
}

func TestEmptyDiff(t *testing.T) {
	stats, err := Parse(strings.NewReader(""), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 files, got %d", len(stats))
	}
}

func TestOnlyMetadataLines(t *testing.T) {
	diff := `diff --git a/f.go b/f.go
index 1234567..abcdefg 100644
--- a/f.go
+++ b/f.go
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	if stats[0].Added != 0 || stats[0].Deleted != 0 {
		t.Errorf("expected zero counts, got %+v", stats[0])
	}
}

func TestChurnCalculation(t *testing.T) {
	diff := `diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1,4 +1,4 @@
 package x
-func A() {}
-func B() {}
+func C() {}
+func D() {}
+func E() {}
`
	stats, err := Parse(strings.NewReader(diff), "exclude")
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 file, got %d", len(stats))
	}
	if stats[0].Added != 3 || stats[0].Deleted != 2 || stats[0].Churn != 5 {
		t.Errorf("got %+v, want Added=3 Deleted=2 Churn=5", stats[0])
	}
}
