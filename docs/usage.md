# differ Usage Guide

This guide covers practical usage for `differ`, including ref selection, filtering, output modes, and common workflows.

## Command Shape

```bash
differ [rev-range] [flags] [-- pathspec...]
```

Examples:

```bash
differ
differ main...HEAD
differ --base main --head feature/my-branch
differ --empty include -l
differ --format json --exclude 'vendor/**'
differ -- docs/ internal/
```

## Ref Selection

`differ` resolves what to compare in this order:

1. `--base` and `--head` if both are provided (`base...head`)
2. Positional `rev-range` if provided
3. Auto mode fallback chain:
   - `origin/HEAD...HEAD`
   - `main...HEAD`
   - `master...HEAD`

### Local Uncommitted Changes In Auto Mode

If you run `differ` with no refs and your repo has staged/unstaged changes, it switches to diff from merge-base to current worktree so local edits are included.

Examples:

```bash
# Include local staged + unstaged changes (default auto behavior in dirty repos)
differ

# Explicitly compare HEAD to worktree
differ HEAD
```

## What Gets Counted

- Added lines: diff hunk lines starting with `+`
- Deleted lines: diff hunk lines starting with `-`
- Metadata lines (`diff --git`, `+++`, `---`, etc.) are ignored
- Binary files are skipped

By default, empty/whitespace-only changed lines are excluded.

```bash
# default
differ --empty exclude

# include whitespace-only changed lines
differ --empty include
```

## Output Modes

### Text Summary (default)

```bash
differ
```

### Summary + Per-file List

```bash
differ -l
```

### Per-file List Only

```bash
differ -L
```

### JSON

```bash
differ --format json
```

JSON includes:

- `meta`: base/head refs, empty-line mode, pathspecs, timestamp
- `total`: added/deleted/churn/files
- `by_category`: totals and file list per category
- `by_file`: per-file stats with category/language

## Sorting

Sorting applies to file list output (`-l` or `-L`):

```bash
# default
differ -l --sort churn

differ -l --sort path
```

## Filtering

### Git Pathspec Filter

Use `--` to pass pathspecs to `git diff`:

```bash
differ -- internal/ docs/
```

### Include/Exclude Globs

```bash
differ --include '**/*.go' --exclude '**/*_test.go'
differ --exclude 'vendor/**' --exclude 'dist/**'
```

### Category Filter

Allowed categories:

- `docs`
- `tests`
- `source`
- `generated`
- `other`

```bash
differ --category source
differ --category source --category tests
```

## Categories

Files are assigned to one category by priority:

1. `generated`
2. `docs`
3. `tests`
4. `source`
5. `other`

Examples of built-in heuristics:

- Generated: `vendor/`, `node_modules/`, `dist/`, `build/`, common lockfiles
- Docs: `.md`, `.mdx`, `.rst`, `.adoc`, `.txt`, `docs/`
- Tests: `*_test.go`, `*.test.*`, `*.spec.*`, `tests/`, `specs/`

## Config Files

Supported config locations:

- Global: `~/.config/differ/config.yml`
- Repo-local: `.differ.yml`

Precedence:

1. CLI flags
2. Repo config
3. Global config
4. Built-in defaults

Example `.differ.yml`:

```yaml
empty: exclude
sort: churn
include:
  - "**/*.go"
exclude:
  - "vendor/**"
categories:
  docs:
    patterns:
      - "handbook/**"
```

## Exit Codes

- `0`: success
- `1`: runtime/usage error
- `2`: invalid config

## Common Workflows

```bash
# Current branch vs main
differ main...HEAD

# Feature branch with file list
differ --base main --head HEAD -l

# Only docs changes
differ --category docs

# JSON for scripting
differ --format json > churn.json

# Include local worktree changes from HEAD
differ HEAD
```

## Troubleshooting

### "0 lines changed" when you edited files

If you compared commit refs (for example `main...HEAD`), uncommitted edits are not included. Use:

```bash
differ
```

or:

```bash
differ HEAD
```

### "cannot resolve base ref: tried origin/HEAD, main, master"

You are likely not in a git repository, or none of those refs exists. Run with explicit refs:

```bash
differ --base <ref> --head <ref>
```
