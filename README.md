# differ

A CLI tool for diffing directories and files with support for specs.

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap jbonatakis/tap && brew install differ
```

### From source

```bash
go install github.com/jackbonatakis/differ/cmd/differ@latest
```

### Binary releases

Download pre-built binaries from the [Releases](https://github.com/jackbonatakis/differ/releases) page. Binaries are available for:

- macOS (arm64, x86_64)
- Linux (arm64, x86_64)

## Releasing

Releases are automated by `.github/workflows/release.yml`:

- On a GitHub Release creation, binaries are built and uploaded as release assets.
- After assets upload, the workflow updates `jbonatakis/homebrew-tap` with the new `differ` formula version and SHA256 values.

Required repository secret:

- `HOMEBREW_TAP_TOKEN`: a PAT with push access to `jbonatakis/homebrew-tap`.
