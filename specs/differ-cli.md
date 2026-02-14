# differ CLI Spec (v1)

## 1. Summary

`differ` is a Git-aware command line tool that reports line-of-code churn between two refs, grouped into practical categories (`docs`, `tests`, `source`, `generated`, `other`).

Primary goals:
- Work across arbitrary repositories without language-specific setup.
- Provide useful defaults with low cognitive overhead.
- Support human-readable and machine-readable output.
- Be easy to install via Homebrew as a single binary.

Implementation language: `Go`.

## 2. Why Go (Decision)

We will build v1 in Go because it is the fastest path to:
- Single static binaries for macOS/Linux.
- Straightforward Homebrew distribution via GitHub Releases and tap formula updates.
- Fast contributor onboarding (minimal toolchain friction).
- Strong CLI ecosystem (`cobra`), plus mature libraries for language/file heuristics.

Rust remains a valid future option, but Go optimizes delivery speed and maintainability for this problem.

## 3. Scope

### In scope (v1)
- Parse churn from `git diff --no-color -U0 -M`.
- Count added/deleted lines with optional empty-line inclusion.
- Auto-categorize files across languages.
- Output summary by category.
- Optional file list mode.
- Optional JSON output.
- Per-repo/global config overrides.

### Out of scope (v1)
- Re-implementing Git diff internals.
- Deep AST-aware semantic diffs.
- Full binary file analysis.

## 4. CLI Contract

Command:

```bash
differ [rev-range] [flags] [-- pathspec...]
```

Examples:

```bash
differ
differ main...HEAD
differ --base main --head feature/my-branch
differ --empty include -l
differ --format json --exclude 'vendor/**' --exclude '*.lock'
differ -- docs/ internal/
```

### Ref selection
Resolution order:
1. `--base` + `--head` if both set.
2. Positional `rev-range` if set.
3. Auto default:
   - `origin/HEAD...HEAD` if resolvable,
   - else `main...HEAD`,
   - else `master...HEAD`.

### Primary flags (v1)
- `--base <rev>`: base ref.
- `--head <rev>`: head ref.
- `--empty <include|exclude>`: count empty/whitespace-only changed lines.
- `-l, --list`: show summary plus per-file list.
- `-L, --list-only`: show per-file list only.
- `--format <text|json>`: output format.
- `--include <glob>`: include path glob (repeatable).
- `--exclude <glob>`: exclude path glob (repeatable).
- `--category <name>`: restrict to category (`docs|tests|source|generated|other`, repeatable).
- `--sort <churn|path>`: file list ordering.
- `--no-color`: disable colorized text output.

### Exit codes
- `0`: success.
- `1`: runtime or usage error.
- `2`: invalid config.

## 5. Counting Semantics

Input source:

```bash
git diff --no-color -U0 -M <resolved-range> -- <optional pathspecs>
```

Rules:
- Count only lines beginning with `+` or `-` in hunks.
- Ignore diff metadata lines (`+++`, `---`, `diff --git`, etc).
- If `--empty exclude` (default), ignore lines matching `^[[:space:]]*$` after removing diff prefix.
- Renames are supported via `-M`; counting remains based on line adds/deletes reported by Git.
- Binary patches are skipped in v1.

## 6. Categorization Model

Each changed file maps to one category using first-match priority:
1. `generated`
2. `docs`
3. `tests`
4. `source`
5. `other`

### 6.1 generated
Heuristics include:
- Common generated directories/files: `vendor/`, `node_modules/`, `dist/`, `build/`.
- Lockfiles: `package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`, `go.sum`, `Cargo.lock`, etc.
- Generated markers in file path/name (configurable).

### 6.2 docs
- Extensions: `.md`, `.markdown`, `.mdx`, `.rst`, `.adoc`, `.txt`.
- Paths including docs-oriented folders (`docs/`, `documentation/`).

### 6.3 tests
Language-agnostic and language-aware conventions, e.g.:
- Paths: `test/`, `tests/`, `spec/`, `specs/`, `__tests__/`.
- Filenames: `*.test.*`, `*.spec.*`.
- Language-specific defaults:
  - Go: `*_test.go`
  - Python: `test_*.py`, `*_test.py`
  - Java/Kotlin: `*Test.java`, `*Tests.java`, `*Test.kt`
  - JS/TS: `*.test.js|ts|jsx|tsx`, `*.spec.js|ts|jsx|tsx`
  - Ruby: `*_spec.rb`, `test_*.rb`

### 6.4 source
- Recognized source-code extensions (broad mapping; e.g. Go/Rust/Python/JS/TS/Java/C/C++/C#/PHP/Ruby/Swift/Kotlin/etc).
- If extension unknown and not matched elsewhere, falls through to `other`.

## 7. Language Detection Strategy

v1 uses extension and filename mapping with optional integration of a language detection library (e.g. `go-enry`) for richer classification.

Important: category assignment does not require perfect language identity. It requires reliable bucket mapping. The strategy prioritizes deterministic file-pattern rules over probabilistic detection.

## 8. Output Formats

### Text summary

```text
Documentation: +12 -3 (15) [4 files]
Tests: +45 -8 (53) [6 files]
Source: +120 -90 (210) [14 files]
Generated: +2 -2 (4) [1 files]
Uncategorized: +7 -1 (8) [3 files]
Total: +186 -104 (290) [28 files]
```

### Text with list mode (`-l`)
- Show summary first.
- Then grouped file lines:

```text
[Tests]
+20 -2 pkg/a/a_test.go
+25 -6 internal/foo/specs/bar_spec.rb
```

`-L` prints only grouped file lines.

### JSON (`--format json`)
Top-level schema:
- `meta`: refs, flags, timestamp, pathspecs.
- `total`: added, deleted, churn, files.
- `by_category`: per-category totals/files.
- `by_file`: per-file added/deleted/churn/category/language(optional).

## 9. Configuration

Config files:
- Repo-local: `.differ.yml` (repo root).
- Global: `~/.config/differ/config.yml`.

Precedence:
1. CLI flags
2. Repo config
3. Global config
4. Built-in defaults

Supported config fields (v1):
- `include`: list of glob patterns.
- `exclude`: list of glob patterns.
- `categories.<name>.patterns`: custom path/filename globs.
- `categories.<name>.extensions`: extension lists.
- `empty`: `include|exclude` default.
- `sort`: `churn|path` default.

## 10. Packaging and Distribution

Release pipeline:
- Build and publish binaries with GitHub Actions on release creation.
- Publish binaries for macOS (arm64/x86_64) and Linux (arm64/x86_64).
- Update Homebrew tap formula version and SHA256 values in a tap repo.

Install UX target:

```bash
brew tap <org>/tap
brew install differ
```

## 11. Suggested Internal Architecture

Packages/modules:
- `cmd/differ`: CLI entrypoint and flag wiring.
- `internal/gitdiff`: invoke git and stream diff lines.
- `internal/parser`: parse hunks into per-file add/del counts.
- `internal/classify`: category and language heuristics.
- `internal/filter`: include/exclude/category filtering.
- `internal/output`: text/json rendering.
- `internal/config`: config loading + merge precedence.

Data model (core):
- `FileStat{Path, Added, Deleted, Churn, Category, Language}`
- `Summary{Totals, CategoryTotals, FileStats, Meta}`

## 12. Future Options (post-v1)

- `--by language|category` breakdown mode.
- `--staged` / `--unstaged` working-tree comparisons.
- `--top <N>` top changed files.
- `--ignore-whitespace` forwarding to git flags.
- `--fail-on-churn <N>` CI guardrail.
- `--binary` coarse binary change counting.

## 13. Open Questions

- Should lockfiles always be `generated`, or configurable per repo?
- Should docs-like JSON (e.g. API schemas) default to `docs` or `other`?
- Should file list output include unchanged renamed files when rename-only?
