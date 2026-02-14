# differ

```bash
> differ

Documentation: +772 -3 (775) [9 files]
Tests: +1991 -0 (1991) [21 files]
Source: +6512 -351 (6863) [50 files]
Generated: +38 -0 (38) [1 files]
Uncategorized: +12 -0 (12) [1 files]
Total: +9325 -354 (9679) [82 files]
```

A CLI tool for breaking down diff summaries.

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap jbonatakis/tap && brew install differ
```

### From source

```bash
go install github.com/jbonatakis/differ/cmd/differ@latest
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/jbonatakis/differ/releases) page. Binaries are available for:

- macOS (arm64, x86_64)
- Linux (arm64, x86_64)

## Usage

Basic examples:

```bash
# Auto-detect a base ref and compare against your current branch/worktree
differ

# Compare an explicit revision range
differ main...HEAD

# Compare explicit refs
differ --base main --head feature/my-branch

# Show summary + per-file list
differ -l

# Show per-file list only
differ -L

# JSON output for scripting/CI
differ --format json

# Restrict to specific paths and exclude globs
differ --exclude 'vendor/**' -- docs/ internal/
```

Common flags:

- `--base <rev>` / `--head <rev>`: select refs explicitly.
- `--empty <include|exclude>`: include or skip empty/whitespace-only changed lines.
- `--format <text|json>`: choose output format.
- `-l, --list`: show summary plus per-file list.
- `-L, --list-only`: show only per-file list.
- `--include <glob>` / `--exclude <glob>`: filter paths (repeatable).
- `--category <docs|tests|source|generated|other>`: restrict categories (repeatable).
- `--sort <churn|path>`: sort file list output.
- `--no-color`: disable ANSI colors in text mode.

Run `differ --help` for the full CLI reference.

## Releasing

Releases are automated by `.github/workflows/release.yml`:

- On a GitHub Release creation, binaries are built and uploaded as release assets.
- After assets upload, the workflow updates `jbonatakis/homebrew-tap` with the new `differ` formula version and SHA256 values.

