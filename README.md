# differ

A CLI tool for breaking down diff summaries.

<pre><code>&gt; differ

Documentation: <span style="color:#26a269">+772</span> <span style="color:#c01c28">-3</span> (775) [9 files]
Tests: <span style="color:#26a269">+1991</span> <span style="color:#c01c28">-0</span> (1991) [21 files]
Source: <span style="color:#26a269">+6512</span> <span style="color:#c01c28">-351</span> (6863) [50 files]
Generated: <span style="color:#26a269">+38</span> <span style="color:#c01c28">-0</span> (38) [1 files]
Uncategorized: <span style="color:#26a269">+12</span> <span style="color:#c01c28">-0</span> (12) [1 files]
Total: <span style="color:#26a269">+9325</span> <span style="color:#c01c28">-354</span> (9679) [82 files]
</code></pre>

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

## Documentation

### Usage

- See the [Usage Guide](docs/usage.md) for detailed usage instructions

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
